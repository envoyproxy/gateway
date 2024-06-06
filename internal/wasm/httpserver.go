// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package wasm

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/docker/docker/pkg/fileutils"

	"github.com/envoyproxy/gateway/internal/logging"
	tlsUtils "github.com/envoyproxy/gateway/internal/utils/tls"
)

const (
	serverHost           = "envoy-gateway"
	serverPort           = 18002
	serveTLSCertFilename = "/certs/tls.crt"
	serveTLSKeyFilename  = "/certs/tls.key"
	serveTLSCaFilename   = "/certs/ca.crt"
)

var _ Cache = &HTTPServer{}

// HTTPServer wraps a local file cache and serves the Wasm modules over HTTP.
type HTTPServer struct {
	sync.Mutex
	// map from the mapping path to the wasm file path in the local cache.
	// The mapping path is a generated random path to prevent unauthorized users
	// from accessing the Wasm module using EnvoyPatchPolicy. Unless the user is
	// an admin who can dump the configuration of the Envoy proxy, the mapping path
	// is not exposed to the user.
	mappingPath2Cache map[string]wasmModuleEntry
	// map from the original wasm module URL to the EG HTTP server serving URL.
	// We need to keep the generated mapping path the same for the same Wasm module.
	originalURL2MappingPath map[string]string
	// directory path used to store Wasm module.
	dir string
	// logger
	logger logging.Logger
	// local file cache
	cache Cache
}

type wasmModuleEntry struct {
	name        string
	originalURL string
	localFile   string
}

// NewHTTPServerWithFileCache creates a HTTP server with a local file cache for Wasm modules.
// The local file cache is used to store the Wasm modules downloaded from the original URL.
// The HTTP server serves the cached Wasm modules over HTTP to the Envoy Proxies.
func NewHTTPServerWithFileCache(cacheDir string, logger logging.Logger) *HTTPServer {
	logger = logger.WithName("wasm-cache")
	return &HTTPServer{
		mappingPath2Cache:       make(map[string]wasmModuleEntry),
		originalURL2MappingPath: make(map[string]string),
		dir:                     cacheDir,
		cache:                   newLocalFileCache(cacheDir, defaultOptions(), logger), //TODO: zhaohuabing we may expose the options in the future
		logger:                  logger,
	}
}

func (s *HTTPServer) Start(ctx context.Context) {
	s.logger.Info(fmt.Sprintf("Listening on :%d", serverPort))

	var (
		tlsConfig *tls.Config
		err       error
	)

	// Create the file directory if it does not exist.
	if err = fileutils.CreateIfNotExists(s.dir, true); err != nil {
		s.logger.Error(err, "Failed to create Wasm cache directory")
		return
	}

	handler := http.NewServeMux()
	handler.Handle("/", s)

	if tlsConfig, err = tlsUtils.TLSConfig(
		serveTLSCertFilename,
		serveTLSKeyFilename,
		serveTLSCaFilename); err != nil {
		s.logger.Error(err, "Failed to create TLS config")
		return
	}

	server := http.Server{
		Addr:      fmt.Sprintf(":%d", serverPort),
		Handler:   handler,
		TLSConfig: tlsConfig,
	}

	go func() {
		if err = server.ListenAndServe(); err != nil {
			s.logger.Error(err, "Failed to start Wasm HTTP server")
		}
	}()
	s.cache.Start(ctx)

	select {
	case <-ctx.Done():
		if err = server.Close(); err != nil {
			s.logger.Error(err, "Error closing Wasm HTTP server")
		}
	}
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if strings.HasPrefix(path, "/") {
		path = path[1:]
	}
	if entry, ok := s.mappingPath2Cache[path]; ok {
		http.ServeFile(w, r, entry.localFile)
	}
	w.WriteHeader(http.StatusNotFound)
}

// Get returns the HTTP URL of the Wasm module serving by the EG HTTP server.
// EG downloads the Wasm module from its original URL, caches it locally in the file system,
// and serves it through an HTTP server.
func (s *HTTPServer) Get(originalUrl string, opts GetOptions) (servingURL string, checksum string, err error) {
	var (
		mappingPath string
		localFile   string
		mapped      bool
	)

	// Check if the Wasm module is already mapped to a serving URL.
	mappingPath, mapped = s.originalURL2MappingPath[originalUrl]

	// Get the local file path of the cached Wasm module.
	// Even it's already cached, the file cache may still download the Wasm module
	// again if it is expired or it needs to be updated.
	if localFile, checksum, err = s.cache.Get(originalUrl, opts); err != nil {
		s.logger.Error(err, "Failed to get Wasm module", "URL", originalUrl)
		return "", "", err
	}

	s.Lock()
	defer s.Unlock()

	if !mapped {
		// A random path is used to make the URL unpredictable and prevent unauthorized
		// users to access the Wasm module using EnvoyPatchPolicy.
		if mappingPath, err = generateRandomPath(); err != nil {
			return "", "", err
		}

		s.mappingPath2Cache[mappingPath] = wasmModuleEntry{
			name:        opts.ResourceName,
			originalURL: originalUrl,
			localFile:   localFile,
		}
		s.originalURL2MappingPath[originalUrl] = mappingPath
	} else {
		// Update the local file path of the cached Wasm module in case it is changed.
		entry := s.mappingPath2Cache[mappingPath]
		entry.localFile = localFile
		s.mappingPath2Cache[mappingPath] = entry
	}

	// TODO: zhaohuabing: https
	servingURL = fmt.Sprintf("http://%s:%d/%s", serverHost,
		serverPort, mappingPath)
	return servingURL, checksum, nil
}

// Generate a random downloading path for a Wasm module.
func generateRandomPath() (string, error) {
	random := make([]byte, 16) // 16 bytes for the salt
	_, err := io.ReadFull(rand.Reader, random)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("wasm/%s.wasm", hex.EncodeToString(random)), nil
}
