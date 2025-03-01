// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package wasm

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/go-containerregistry/pkg/name"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/envoyproxy/gateway/internal/logging"
)

const (
	// oci URL prefix
	ociURLPrefix = "oci://"
	// sha256 scheme prefix
	sha256SchemePrefix = "sha256:"
)

// Cache models a Wasm module cache.
type Cache interface {
	Get(downloadURL string, opts GetOptions) (url string, checksum string, err error)
	Start(ctx context.Context)
}

// localFileCache for downloaded Wasm modules. It stores the Wasm module as local files.
type localFileCache struct {
	// Map from Wasm module key to cache entry.
	modules map[moduleKey]*cacheEntry
	// Map from downloading URL to checksum
	checksums map[string]*checksumEntry
	// http fetcher fetches Wasm module with HTTP get.
	httpFetcher *HTTPFetcher

	// mux is needed because stale Wasm module files will be purged periodically.
	mux sync.Mutex

	// option sets for configuring the cache.
	CacheOptions

	// permissionCheckCache is the cache for permission check for private OCI images.
	// The permission check is run periodically by a background goroutine and the result is cached.
	permissionCheckCache *permissionCache

	// logger
	logger logging.Logger
}

func (c *localFileCache) Start(ctx context.Context) {
	c.permissionCheckCache.Start(ctx)
	go c.purge(ctx)
}

var _ Cache = &localFileCache{}

type checksumEntry struct {
	checksum string
	// Keeps the resource version per each resource for dealing with multiple resources which pointing the same image.
	resourceVersionByResource map[string]string
}

// moduleKey is a unique identifier for a Wasm module consisting of the name and checksum.
type moduleKey struct {
	// Identifier for the wasm module.
	// If the wasm module is an HTTP URL, it is the original download URL.
	// e.g.  http://example.com/test.wasm
	// If the wasm module is an OCI image, it should be the image name without tag or digest.
	// e.g.  oci://docker.io/test
	name string
	// sha256 checksum of the wasm file or the image.
	// Note that the checksum is different from the checksum of the wasm file if
	// the module is extracted from an OCI image.
	checksum string
}

type cacheKey struct {
	moduleKey
	// URL to download the wasm module.
	// e.g. http://example.com/test.wasm or oci://docker.io/test:v1.0.0
	downloadURL string
	// Resource name of the wasm module. This should be a fully-qualified name.
	// e.g. "envoyextensionpolicy/envoy-gateway/policy-for-gateway/wasm/0"
	resourceName string
	// Resource version of EnvoyExtensionPolicy resource. Even though PullPolicy is Always,
	// if there is no change of resource state, a cached entry is used instead of pulling newly.
	resourceVersion string
}

// cacheEntry contains information about a Wasm module cache entry.
type cacheEntry struct {
	// File path to the downloaded wasm modules.
	modulePath string
	// Last time that this local Wasm module is referenced.
	last time.Time
	// set of URLs referencing this entry
	referencingURLs sets.Set[string]
	// isPrivate is true if the module is from a private registry.
	isPrivate bool
	// checksum is the sha256 checksum of the module.
	// It is different from the checksum of the image if the module is from an OCI image.
	checksum string
	// size is the size of the module.
	size int
}

// newLocalFileCache create a new Wasm module cache which downloads and stores Wasm module files locally.
func newLocalFileCache(options CacheOptions, logger logging.Logger) *localFileCache {
	options = options.sanitize()
	cache := &localFileCache{
		httpFetcher: NewHTTPFetcher(options.HTTPRequestTimeout, options.HTTPRequestMaxRetries, logger),
		modules:     make(map[moduleKey]*cacheEntry),
		checksums:   make(map[string]*checksumEntry),
		permissionCheckCache: newPermissionCache(
			permissionCacheOptions{},
			logger),
		CacheOptions: options,
		logger:       logger,
	}

	return cache
}

func moduleNameFromURL(fullURLStr string) string {
	if strings.HasPrefix(fullURLStr, ociURLPrefix) {
		if tag, err := name.ParseReference(fullURLStr[len(ociURLPrefix):]); err == nil {
			// remove tag or sha
			return ociURLPrefix + tag.Context().Name()
		}
	}
	return fullURLStr
}

func getModulePath(baseDir string, mkey moduleKey) (string, error) {
	// Use sha256 checksum as the name of the module.
	sha := sha256.Sum256([]byte(mkey.name))
	hashedName := hex.EncodeToString(sha[:])
	moduleDir := filepath.Join(baseDir, hashedName)
	if err := os.Mkdir(moduleDir, 0o755); err != nil && !os.IsExist(err) {
		return "", err
	}
	return filepath.Join(moduleDir, fmt.Sprintf("%s.wasm", mkey.checksum)), nil
}

// Get returns path the local Wasm module file and its checksum.
func (c *localFileCache) Get(downloadURL string, opts GetOptions) (localFile string, checksum string, err error) {
	// If the checksum is not provided, try to extract it from the OCI image URL.
	originalChecksum := opts.Checksum
	if len(opts.Checksum) == 0 && strings.HasPrefix(downloadURL, ociURLPrefix) {
		if d, err := name.NewDigest(downloadURL[len(ociURLPrefix):]); err == nil {
			// If there is no checksum and the digest is suffixed in URL, use the digest.
			dstr := d.DigestStr()
			if strings.HasPrefix(dstr, sha256SchemePrefix) {
				originalChecksum = dstr[len(sha256SchemePrefix):]
			}
		}
	}

	// Construct Wasm cache key with downloading URL and provided checksum of the module.
	key := cacheKey{
		downloadURL: downloadURL,
		moduleKey: moduleKey{
			name:     moduleNameFromURL(downloadURL),
			checksum: originalChecksum,
		},
		resourceName:    opts.ResourceName,
		resourceVersion: opts.ResourceVersion,
	}

	entry, err := c.getOrFetch(key, opts)
	if err != nil {
		return "", "", err
	}

	return entry.modulePath, entry.checksum, err
}

func (c *localFileCache) getOrFetch(key cacheKey, opts GetOptions) (*cacheEntry, error) {
	var (
		u         *url.URL
		insecure  bool
		isPrivate bool
		err       error
	)

	if u, err = url.Parse(key.downloadURL); err != nil {
		return nil, fmt.Errorf("fail to parse Wasm module fetch url: %s, error: %w", key.downloadURL, err)
	}
	insecure = c.allowInsecure(u.Host)

	requestTimout := DefaultPullTimeout
	if opts.RequestTimeout != 0 {
		requestTimout = opts.RequestTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimout)
	defer cancel()

	// First check if the cache entry is already downloaded and policy does not require pulling always.
	ce := c.getEntry(key, opts.PullPolicy, u)
	if ce != nil {
		// We still need to check if the pull secret is correct if it is a private OCI image.
		if u.Scheme == "oci" && ce.isPrivate {
			if _, err := c.permissionCheckCache.IsAllowed(ctx, u, opts.PullSecret, insecure); err != nil {
				return nil, err
			}
		}
		return ce, nil
	}

	// Fetch the image now as it is not available in cache.
	var (
		b                  []byte // Byte array of Wasm binary.
		dChecksum          string // Hex-Encoded sha256 checksum of binary.
		imageBinaryFetcher func() ([]byte, error)
	)

	switch u.Scheme {
	case "http", "https":
		// Download the Wasm module with http fetcher.
		b, err = c.httpFetcher.Fetch(ctx, key.downloadURL, insecure)
		if err != nil {
			wasmRemoteFetchTotal.WithFailure(reasonDownloadError).Increment()
			return nil, err
		}

		// Get sha256 checksum and check if it is the same as the provided one.
		sha := sha256.Sum256(b)
		dChecksum = hex.EncodeToString(sha[:])
	case "oci":
		if len(opts.PullSecret) > 0 {
			isPrivate = true
		}

		imageBinaryFetcher, dChecksum, err = c.prepareFetch(ctx, u, insecure, opts.PullSecret)

		if isPrivate {
			e := &permissionCacheEntry{
				image: u,
				fetcherOption: &ImageFetcherOption{
					Insecure:   insecure,
					PullSecret: opts.PullSecret,
				},
				lastCheck:  time.Now(),
				lastAccess: time.Now(),
				checkError: err,
			}
			c.permissionCheckCache.Put(e)
		}

		if err != nil {
			wasmRemoteFetchTotal.WithFailure(reasonManifestError).Increment()
			return nil, fmt.Errorf("could not fetch Wasm OCI image: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported Wasm module downloading URL scheme: %v", u.Scheme)
	}

	// If the checksum is provided, check if it matches the downloaded binary.
	if key.checksum != "" {
		if dChecksum != key.checksum {
			wasmRemoteFetchTotal.WithFailure(reasonChecksumMismatch).Increment()
			return nil, fmt.Errorf("module downloaded from %v has checksum %v, which does not match: %v", key.downloadURL, dChecksum, key.checksum)
		}
	} else {
		// Update the checksum with the one from the downloaded binary.
		key.checksum = dChecksum
	}

	if imageBinaryFetcher != nil {
		b, err = imageBinaryFetcher()
		if err != nil {
			wasmRemoteFetchTotal.WithFailure(reasonDownloadError).Increment()
			return nil, fmt.Errorf("could not fetch Wasm binary: %w", err)
		}
	}

	if !isValidWasmBinary(b) {
		wasmRemoteFetchTotal.WithFailure(reasonFetchError).Increment()
		return nil, fmt.Errorf("fetched Wasm binary from %s is invalid", key.downloadURL)
	}

	wasmRemoteFetchTotal.WithSuccess().Increment()

	return c.addEntry(key, b, isPrivate)
}

// prepareFetch won't fetch the binary, but it will prepare the binaryFetcher and actualDigest.
func (c *localFileCache) prepareFetch(
	ctx context.Context, url *url.URL, insecure bool, pullSecret []byte) (
	binaryFetcher func() ([]byte, error), actualDigest string, err error,
) {
	imgFetcherOps := ImageFetcherOption{
		Insecure: insecure,
	}
	if len(pullSecret) > 0 {
		imgFetcherOps.PullSecret = pullSecret
	}
	fetcher := NewImageFetcher(ctx, imgFetcherOps, c.logger)
	if binaryFetcher, actualDigest, err = fetcher.PrepareFetch(url.Host + url.Path); err != nil {
		return nil, "", err
	}
	return binaryFetcher, actualDigest, nil
}

func (c *localFileCache) updateChecksum(key cacheKey) {
	ce := c.checksums[key.downloadURL]
	if ce == nil {
		ce = new(checksumEntry)
		ce.resourceVersionByResource = make(map[string]string)
		c.checksums[key.downloadURL] = ce
	}
	ce.checksum = key.checksum
	ce.resourceVersionByResource[key.resourceName] = key.resourceVersion
}

// addEntry adds a wasmModule to cache with cacheKey, writes the module to the local file system,
// and returns the created entry.
func (c *localFileCache) addEntry(key cacheKey, wasmModule []byte, isPrivate bool) (*cacheEntry, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	// Check if the cache size exceeds the limit.
	if c.size()+len(wasmModule) > c.MaxCacheSize {
		return nil, fmt.Errorf("wasm cache size exceeded the limit: %d", c.MaxCacheSize)
	}

	c.updateChecksum(key)

	// Check if the module has already been added. If so, avoid writing the file again.
	if ce, ok := c.modules[key.moduleKey]; ok {
		// Update last touched time.
		ce.last = time.Now()
		ce.referencingURLs.Insert(key.downloadURL)
		return ce, nil
	}

	modulePath, err := getModulePath(c.CacheDir, key.moduleKey)
	if err != nil {
		return nil, err
	}
	// Materialize the Wasm module into a local file. Use checksum as name of the module.
	if err := os.WriteFile(modulePath, wasmModule, 0o600); err != nil {
		return nil, err
	}

	// Calculate the checksum of the wasm module. It is different from the checksum of the image.
	wasmChecksum := strings.ToLower(fmt.Sprintf("%x", sha256.Sum256(wasmModule)))
	ce := cacheEntry{
		modulePath:      modulePath,
		last:            time.Now(),
		referencingURLs: sets.New[string](),
		isPrivate:       isPrivate,
		checksum:        wasmChecksum,
		size:            len(wasmModule),
	}
	ce.referencingURLs.Insert(key.downloadURL)
	c.modules[key.moduleKey] = &ce

	wasmCacheEntries.Record(float64(len(c.modules)))

	return &ce, nil
}

// getEntry finds a cached module, and returns the found cache entry and its checksum.
// If the module is not found in the cache, it returns nil.
// If the module is found in the cache, but the module needs to be re-pulled, it returns nil.
func (c *localFileCache) getEntry(key cacheKey, pullPolicy PullPolicy, u *url.URL) *cacheEntry {
	cacheHit := false

	c.mux.Lock()
	defer func() {
		c.mux.Unlock()
		wasmCacheLookupTotal.With(hitTag.Value(strconv.FormatBool(cacheHit))).Increment()
	}()

	// If no checksum is provided, check if a wasm module with the same downloading URL has been pulled before.
	if len(key.checksum) == 0 {
		// If an image with the same downloading URL was pulled before, there should be a checksum of the most recently pulled image.
		if ce, found := c.checksums[key.downloadURL]; found {
			// If it is an OCI image and the tag is "latest", default pull policy is Always.
			// Otherwise, default pull policy is IfNotPresent.
			if pullPolicy == Unspecified {
				if u.Scheme == "oci" && strings.HasSuffix(u.Path, ":latest") {
					pullPolicy = Always
				} else {
					pullPolicy = IfNotPresent
				}
			}

			// Check if we need to re-pull the wasm module.
			needPull := true
			switch pullPolicy {
			case IfNotPresent:
				needPull = false
			case Always:
				// If the resource version is not changed, use the cached wasm module.
				// Otherwise, pull the new one from its original URL.
				if key.resourceVersion == ce.resourceVersionByResource[key.resourceName] {
					needPull = false
				}
			}

			// If we need to re-pull this wasm module, return nil.
			if needPull {
				return nil
			}

			// If we don't need to pull the module again, return the cached module.
			key.checksum = ce.checksum
			existingModule := c.modules[key.moduleKey]
			// Update last touched time.
			existingModule.last = time.Now()
			cacheHit = true
			// Update the checksum map as the same downloading URL can be referenced
			// by multiple EnvoyExtensionPolicy resources.
			c.updateChecksum(key)
			return existingModule
		}

		// If no previous checksum is found, return nil.
		return nil
	}

	// If the checksum is provided, check if the module with the same checksum has been pulled before.
	if existingModule, ok := c.modules[key.moduleKey]; ok {
		// Update last touched time.
		existingModule.last = time.Now()
		cacheHit = true
		// Update the checksum map as the same downloading URL can be referenced
		// by multiple EnvoyExtensionPolicy resources.
		c.updateChecksum(key)
		return existingModule
	}
	return nil
}

func (c *localFileCache) size() int {
	cacheSize := 0
	for _, entry := range c.modules {
		cacheSize += entry.size
	}
	return cacheSize
}

// Purge periodically clean up the stale Wasm modules local file and the cache map.
func (c *localFileCache) purge(ctx context.Context) {
	ticker := time.NewTicker(c.PurgeInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			c.mux.Lock()
			for k, m := range c.modules {
				if !m.expired(c.ModuleExpiry) {
					continue
				}
				// The module has not be touched for expiry duration, delete it from the map as well as the local dir.
				if err := os.Remove(m.modulePath); err != nil {
					c.logger.Error(err, "failed to purge Wasm module", "path", m.modulePath)
				} else {
					for downloadURL := range m.referencingURLs {
						delete(c.checksums, downloadURL)
					}
					delete(c.modules, k)
					c.logger.Info("successfully removed stale Wasm module", "path", m.modulePath)
				}
			}
			wasmCacheEntries.Record(float64(len(c.modules)))
			c.mux.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

// Expired returns true if the module has not been touched for Wasm module Expiry.
func (ce *cacheEntry) expired(expiry time.Duration) bool {
	now := time.Now()
	return now.Sub(ce.last) > expiry
}

var wasmMagicNumber = []byte{0x00, 0x61, 0x73, 0x6d}

func isValidWasmBinary(in []byte) bool {
	// Wasm file header is 8 bytes (magic number + version).
	return len(in) >= 8 && bytes.Equal(in[:4], wasmMagicNumber)
}
