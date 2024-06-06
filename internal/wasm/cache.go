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

	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/utils/sets"
)

const (
	// oci URL prefix
	ociURLPrefix = "oci://"

	// sha256 scheme prefix
	sha256SchemePrefix = "sha256:"
)

// Cache models a Wasm module cache.
type Cache interface {
	// Get returns the path to the Wasm module file if it's the local file cache.
	// If it's the EG HTTP server, it returns the serving URL to the cached Wasm module.
	Get(downloadURL string, opts GetOptions) (url string, checksum string, err error)
	Start(ctx context.Context)
}

// localFileCache for downloaded Wasm modules. Currently it stores the Wasm module as local file.
type localFileCache struct {
	// Map from Wasm module checksum to cache entry.
	modules map[moduleKey]*cacheEntry
	// Map from tagged URL to checksum
	checksums map[string]*checksumEntry
	// http fetcher fetches Wasm module with HTTP get.
	httpFetcher *HTTPFetcher

	// directory path used to store Wasm module.
	dir string

	// mux is needed because stale Wasm module files will be purged periodically.
	mux sync.Mutex

	// option sets for configuring the cache.
	cacheOptions
	// stopChan currently is only used by test
	stopChan chan struct{}

	// logger
	logger logging.Logger
}

func (c *localFileCache) Start(ctx context.Context) {
	go c.purge(ctx)
}

var _ Cache = &localFileCache{}

type checksumEntry struct {
	checksum string
	// Keeps the resource version per each resource for dealing with multiple resources which pointing the same image.
	resourceVersionByResource map[string]string
}

type moduleKey struct {
	// Identifier for the module. It should be neutral for the checksum.
	// e.g.) oci://docker.io/test@sha256:0123456789 is not allowed.
	//       oci://docker.io/test:latest (tagged form) is allowed.
	name     string
	checksum string
}

type cacheKey struct {
	moduleKey
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
	referencingURLs sets.String
	// isPrivate is true if the module is from a private registry.
	isPrivate bool
	// checksum is the sha256 checksum of the module.
	checksum string
}

type cacheOptions struct {
	Options
	allowAllInsecureRegistries bool
}

func (o cacheOptions) sanitize() cacheOptions {
	ret := cacheOptions{
		Options: defaultOptions(),
	}
	if o.InsecureRegistries != nil {
		ret.InsecureRegistries = o.InsecureRegistries
	}
	ret.allowAllInsecureRegistries = ret.InsecureRegistries.Contains("*")

	if o.PurgeInterval != 0 {
		ret.PurgeInterval = o.PurgeInterval
	}
	if o.ModuleExpiry != 0 {
		ret.ModuleExpiry = o.ModuleExpiry
	}
	if o.HTTPRequestTimeout != 0 {
		ret.HTTPRequestTimeout = o.HTTPRequestTimeout
	}
	if o.HTTPRequestMaxRetries != 0 {
		ret.HTTPRequestMaxRetries = o.HTTPRequestMaxRetries
	}

	return ret
}

func (o cacheOptions) allowInsecure(host string) bool {
	return o.allowAllInsecureRegistries || o.InsecureRegistries.Contains(host)
}

// newLocalFileCache create a new Wasm module cache which downloads and stores Wasm module files locally.
func newLocalFileCache(dir string, options Options, logger logging.Logger) *localFileCache {
	cacheOptions := cacheOptions{Options: options}
	cache := &localFileCache{
		httpFetcher:  NewHTTPFetcher(options.HTTPRequestTimeout, options.HTTPRequestMaxRetries, logger),
		modules:      make(map[moduleKey]*cacheEntry),
		checksums:    make(map[string]*checksumEntry),
		dir:          dir,
		cacheOptions: cacheOptions.sanitize(),
		stopChan:     make(chan struct{}),
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

func shouldIgnoreResourceVersion(pullPolicy PullPolicy, u *url.URL) bool {
	switch pullPolicy {
	case Always:
		// When Always, pull a wasm module when the resource version is changed.
		return false
	case IfNotPresent:
		// When IfNotPresent, use the cached one regardless of the resource version.
		return true
	default:
		// Default is IfNotPresent except OCI images tagged with `latest`.
		return u.Scheme != "oci" || !strings.HasSuffix(u.Path, ":latest")
	}
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

// Get returns path the local Wasm module file.
func (c *localFileCache) Get(downloadURL string, opts GetOptions) (url string, checksum string, err error) {
	// Construct Wasm cache key with downloading URL and provided checksum of the module.
	key := cacheKey{
		downloadURL: downloadURL,
		moduleKey: moduleKey{
			name:     moduleNameFromURL(downloadURL),
			checksum: opts.Checksum,
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
		return nil, fmt.Errorf("fail to parse Wasm module fetch url: %s, error: %v", key.downloadURL, err)
	}
	insecure = c.allowInsecure(u.Host)

	requestTimout := DefaultPullTimeout
	if opts.RequestTimeout != 0 {
		requestTimout = opts.RequestTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), requestTimout)
	defer cancel()

	// First check if the cache entry is already downloaded and policy does not require to pull always.
	ce, checksum := c.getEntry(key, shouldIgnoreResourceVersion(opts.PullPolicy, u))
	if ce != nil {
		// We still need to check if the pull secret is correct if it is a private OCI image.
		if u.Scheme == "oci" && ce.isPrivate {
			if _, _, err = c.prepareFetch(ctx, u, insecure, opts); err != nil {
				wasmRemoteFetchCount.With(resultTag.Value(manifestFailure)).Increment()
				return nil, fmt.Errorf("could not fetch Wasm OCI image: %v", err)
			}
		}
		return ce, nil
	}

	// Fetch the image now as it is not available in cache.
	key.checksum = checksum
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
			wasmRemoteFetchCount.With(resultTag.Value(downloadFailure)).Increment()
			return nil, err
		}

		// Get sha256 checksum and check if it is the same as provided one.
		sha := sha256.Sum256(b)
		dChecksum = hex.EncodeToString(sha[:])
	case "oci":
		if opts.PullSecret != nil && len(opts.PullSecret) > 0 {
			isPrivate = true
		}
		if imageBinaryFetcher, dChecksum, err = c.prepareFetch(ctx, u, insecure, opts); err != nil {
			wasmRemoteFetchCount.With(resultTag.Value(manifestFailure)).Increment()
			return nil, fmt.Errorf("could not fetch Wasm OCI image: %v", err)
		}
	default:
		return nil, fmt.Errorf("unsupported Wasm module downloading URL scheme: %v", u.Scheme)
	}

	if key.checksum == "" {
		key.checksum = dChecksum
		// check again if the cache is having the checksum.
		if ce, _ := c.getEntry(key, true); ce != nil {
			return ce, nil
		}
	} else if dChecksum != key.checksum {
		wasmRemoteFetchCount.With(resultTag.Value(checksumMismatch)).Increment()
		return nil, fmt.Errorf("module downloaded from %v has checksum %v, which does not match: %v", key.downloadURL, dChecksum, key.checksum)
	}

	if imageBinaryFetcher != nil {
		b, err = imageBinaryFetcher()
		if err != nil {
			wasmRemoteFetchCount.With(resultTag.Value(downloadFailure)).Increment()
			return nil, fmt.Errorf("could not fetch Wasm binary: %v", err)
		}
	}

	if !isValidWasmBinary(b) {
		wasmRemoteFetchCount.With(resultTag.Value(fetchFailure)).Increment()
		return nil, fmt.Errorf("fetched Wasm binary from %s is invalid", key.downloadURL)
	}

	wasmRemoteFetchCount.With(resultTag.Value(fetchSuccess)).Increment()

	key.checksum = dChecksum
	return c.addEntry(key, b, isPrivate)
}

// prepareFetch won't fetch the binary, but it will prepare the binaryFetcher and actualDigest.
func (c *localFileCache) prepareFetch(
	ctx context.Context, url *url.URL, insecure bool, opts GetOptions) (
	binaryFetcher func() ([]byte, error), actualDigest string, err error) {
	imgFetcherOps := ImageFetcherOption{
		Insecure: insecure,
	}
	if opts.PullSecret != nil && len(opts.PullSecret) > 0 {
		imgFetcherOps.PullSecret = opts.PullSecret
	}
	fetcher := NewImageFetcher(ctx, imgFetcherOps, c.logger)
	if binaryFetcher, actualDigest, err = fetcher.PrepareFetch(url.Host + url.Path); err != nil {
		return nil, "", err
	}
	return binaryFetcher, actualDigest, nil
}

func (c *localFileCache) updateChecksum(key cacheKey) bool {
	// If OCI URL having a tag or just http/https URL, we need to update checksum.
	needChecksumUpdate := !strings.HasPrefix(key.downloadURL, ociURLPrefix) || !strings.Contains(key.downloadURL, "@")
	if needChecksumUpdate {
		ce := c.checksums[key.downloadURL]
		if ce == nil {
			ce = new(checksumEntry)
			ce.resourceVersionByResource = make(map[string]string)
			c.checksums[key.downloadURL] = ce
		}
		ce.checksum = key.checksum
		ce.resourceVersionByResource[key.resourceName] = key.resourceVersion
	}
	return needChecksumUpdate
}

// addEntry adds a wasmModule to cache with cacheKey, writes the module to the local file system,
// and returns the created entry.
func (c *localFileCache) addEntry(key cacheKey, wasmModule []byte, isPrivate bool) (*cacheEntry, error) {
	c.mux.Lock()
	defer c.mux.Unlock()
	needChecksumUpdate := c.updateChecksum(key)

	// Check if the module has already been added. If so, avoid writing the file again.
	if ce, ok := c.modules[key.moduleKey]; ok {
		// Update last touched time.
		ce.last = time.Now()
		if needChecksumUpdate {
			ce.referencingURLs.Insert(key.downloadURL)
		}
		return ce, nil
	}

	modulePath, err := getModulePath(c.dir, key.moduleKey)
	if err != nil {
		return nil, err
	}
	// Materialize the Wasm module into a local file. Use checksum as name of the module.
	if err := os.WriteFile(modulePath, wasmModule, 0o644); err != nil {
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
	}
	if needChecksumUpdate {
		ce.referencingURLs.Insert(key.downloadURL)
	}
	c.modules[key.moduleKey] = &ce
	wasmCacheEntries.Record(float64(len(c.modules)))
	return &ce, nil
}

// getEntry finds a cached module, and returns the found cache entry and its checksum.
func (c *localFileCache) getEntry(key cacheKey, ignoreResourceVersion bool) (*cacheEntry, string) {
	cacheHit := false

	c.mux.Lock()
	defer func() {
		c.mux.Unlock()
		wasmCacheLookupCount.With(hitTag.Value(strconv.FormatBool(cacheHit))).Increment()
	}()

	// If the checksum is not provided, try to extract it from the OCI image URL.
	if len(key.checksum) == 0 && strings.HasPrefix(key.downloadURL, ociURLPrefix) {
		if d, err := name.NewDigest(key.downloadURL[len(ociURLPrefix):]); err == nil {
			// If there is no checksum and the digest is suffixed in URL, use the digest.
			dstr := d.DigestStr()
			if strings.HasPrefix(dstr, sha256SchemePrefix) {
				key.checksum = dstr[len(sha256SchemePrefix):]
			}
			// For other digest scheme, give up to use cache.
		}
	}

	// If no checksum, try the checksum cache.
	if len(key.checksum) == 0 {
		// If the image was pulled before, there should be a checksum of the most recently pulled image.
		if ce, found := c.checksums[key.downloadURL]; found {
			// If we don't need to update the wasm module(IfNotPresent), or,
			// If the resource version is not changed(Latest), use the cached checksum.
			if ignoreResourceVersion || key.resourceVersion == ce.resourceVersionByResource[key.resourceName] {
				// update checksum
				key.checksum = ce.checksum
			}
			// update resource version here
			ce.resourceVersionByResource[key.resourceName] = key.resourceVersion
		}
	}

	if ce, ok := c.modules[key.moduleKey]; ok {
		// Update last touched time.
		ce.last = time.Now()
		cacheHit = true
		c.updateChecksum(key)
		return ce, key.checksum
	}
	return nil, key.checksum
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
			// wasmCacheEntries.Record(float64(len(c.modules)))
			c.mux.Unlock()
		case <-c.stopChan:
			// Currently this will only happen in test.
			return
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
