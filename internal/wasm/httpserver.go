// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package wasm

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/envoyproxy/gateway/internal/logging"
)

const (
	serverHost = "envoy-gateway"
	serverPort = 18002
)

var _ Cache = &HTTPServer{}

// HTTPServer wraps a local file cache and serves the Wasm modules over HTTP.
type HTTPServer struct {
	sync.Mutex
	// map from the mapping path to the wasm file path in the local cache.
	// The mapping path is a generated unguessable path to prevent unauthorized users
	// from accessing the Wasm module using EnvoyPatchPolicy. Unless the user is
	// an admin who can dump the configuration of the Envoy proxy, the mapping path
	// is not exposed to the user.
	mappingPath2Cache map[string]wasmModuleEntry
	// local file cache
	cache Cache
	// HTTP server to serve the Wasm modules to the Envoy Proxies.
	server *http.Server
	// Salt is used as a hash salt to generate an unguessable path for the Wasm module.
	salt []byte
	// TLSConfig is the TLS configuration for the HTTP server.
	tlsConfig *tls.Config
	// logger
	logger logging.Logger
}

type wasmModuleEntry struct {
	name        string
	originalURL string
	localFile   string
}

// NewHTTPServerWithFileCache creates a HTTP server with a local file cache for Wasm modules.
// The local file cache is used to store the Wasm modules downloaded from the original URL.
// The HTTP server serves the cached Wasm modules over HTTP to the Envoy Proxies.
func NewHTTPServerWithFileCache(salt []byte, tlsConfig *tls.Config, options CacheOptions, logger logging.Logger) *HTTPServer {
	logger = logger.WithName("wasm-cache")

	return &HTTPServer{
		mappingPath2Cache: make(map[string]wasmModuleEntry),
		cache:             newLocalFileCache(options, logger),
		salt:              salt,
		tlsConfig:         tlsConfig,
		logger:            logger,
	}
}

func (s *HTTPServer) Start(ctx context.Context) {
	s.logger.Info(fmt.Sprintf("Listening on :%d", serverPort))

	handler := http.NewServeMux()
	handler.Handle("/", s)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", serverPort),
		Handler:           handler,
		TLSConfig:         s.tlsConfig,
		ReadHeaderTimeout: 15 * time.Second,
	}

	var err error
	go func() {
		if s.enableTLS() {
			err = s.server.ListenAndServeTLS("", "")
		} else {
			err = s.server.ListenAndServe()
		}
		if err != nil {
			s.logger.Error(err, "Failed to start Wasm HTTP server")
			return
		}
	}()
	s.cache.Start(ctx)
}

func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.logger.Sugar().Debugw("Received wasm request", "path", r.URL.Path)

	path := strings.TrimPrefix(r.URL.Path, "/")
	if entry, ok := s.mappingPath2Cache[path]; ok {
		http.ServeFile(w, r, entry.localFile)
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}

// Get returns the HTTP URL of the Wasm module serving by the EG HTTP Wasm server
// and the checksum of the Wasm module.
// EG downloads the Wasm module from its original URL, caches it locally in the
// file system, and serves it through an HTTP server.
func (s *HTTPServer) Get(originalURL string, opts GetOptions) (servingURL string, checksum string, err error) {
	var (
		mappingPath string
		localFile   string
	)

	// Get the local file path of the cached Wasm module.
	// Even it's already cached, the file cache may still download the Wasm module
	// again if it is expired or it needs to be updated.
	if localFile, checksum, err = s.cache.Get(originalURL, opts); err != nil {
		s.logger.Error(err, "Failed to get Wasm module", "URL", originalURL)
		return "", "", err
	}

	s.Lock()
	defer s.Unlock()

	// Generate a new path with the hash of the original url and a salt to
	// make the URL unpredictable.
	// The unguessable path is used to prevent unauthorized users from accessing
	// an unauthorized private Wasm module.
	mappingPath = generateUnguessablePath(originalURL, s.salt)
	s.mappingPath2Cache[mappingPath] = wasmModuleEntry{
		name:        opts.ResourceName,
		originalURL: originalURL,
		localFile:   localFile,
	}

	entry := s.mappingPath2Cache[mappingPath]
	entry.localFile = localFile
	s.mappingPath2Cache[mappingPath] = entry

	scheme := "http"
	if s.enableTLS() {
		scheme = "https"
	}
	servingURL = fmt.Sprintf("%s://%s:%d/%s", scheme, serverHost, serverPort, mappingPath)
	return servingURL, checksum, nil
}

// Generate an unguessable downloading path for a Wasm module.
func generateUnguessablePath(originalURL string, salt []byte) string {
	saltedData := []byte(originalURL)
	saltedData = append(saltedData, salt...)
	hash := sha256.Sum256(saltedData)
	return fmt.Sprintf("%s.wasm", base64.URLEncoding.EncodeToString(hash[:]))
}

func (s *HTTPServer) close() {
	if s != nil {
		_ = s.server.Close()
	}
}

func (s *HTTPServer) enableTLS() bool {
	return s.tlsConfig != nil
}
