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
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/envoyproxy/gateway/internal/logging"
)

const (
	serverHost                   = "envoy-gateway"
	serverPort                   = 18002
	defaultMaxFailedAttempts     = 10
	defaultAttemptsResetInterval = 5 * time.Minute
	defaultAttemptResetDelay     = 1 * time.Hour
)

var _ Cache = &HTTPServer{}

type SeverOptions struct {
	// Salt is used as a hash salt to generate an unguessable path for the Wasm module.
	Salt []byte
	// TLSConfig is the TLS configuration for the HTTP server.
	TLSConfig                   *tls.Config
	MaxFailedAttempts           int
	FailedAttemptsResetInterval time.Duration
	FailedAttemptResetDelay     time.Duration
}

// setDefault sets the default values for the server options if they are not set.
func (o *SeverOptions) setDefault() {
	if o.MaxFailedAttempts == 0 {
		o.MaxFailedAttempts = defaultMaxFailedAttempts
	}
	if o.FailedAttemptsResetInterval == 0 {
		o.FailedAttemptsResetInterval = defaultAttemptsResetInterval
	}
	if o.FailedAttemptResetDelay == 0 {
		o.FailedAttemptResetDelay = defaultAttemptResetDelay
	}
}

// HTTPServer wraps a local file cache and serves the Wasm modules over HTTP.
type HTTPServer struct {
	SeverOptions
	sync.Mutex
	// map from the mapping path to the wasm file path in the local cache.
	// The mapping path is a generated unguessable path to prevent unauthorized users
	// from accessing the Wasm module using EnvoyPatchPolicy. Unless the user is
	// an admin who can dump the configuration of the Envoy proxy, the mapping path
	// is not exposed to the user.
	mappingPath2Cache map[string]wasmModuleEntry
	// map from the original URL to the number of failed attempts to download the Wasm module.
	// If the number of failed attempts exceeds the maximum number of attempts, we will not
	// try to download the Wasm module again for attemptResetDelay. This is used
	// to prevent the cache from being flooded by failed requests.
	failedAttempts map[string]attemptEntry
	// local file cache
	cache Cache
	// HTTP server to serve the Wasm modules to the Envoy Proxies.
	server *http.Server
	// The namespace where the Envoy Gateway is running.
	controllerNamespace string
	// logger
	logger logging.Logger
}

type attemptEntry struct {
	fails int
	last  time.Time
	delay time.Duration
}

func (a *attemptEntry) expired() bool {
	return time.Since(a.last) > a.delay
}

type wasmModuleEntry struct {
	name        string
	originalURL string
	localFile   string
}

// NewHTTPServerWithFileCache creates a HTTP server with a local file cache for Wasm modules.
// The local file cache is used to store the Wasm modules downloaded from the original URL.
// The HTTP server serves the cached Wasm modules over HTTP to the Envoy Proxies.
func NewHTTPServerWithFileCache(serverOptions SeverOptions, cacheOptions CacheOptions, controllerNamespace string, logger logging.Logger) *HTTPServer {
	logger = logger.WithName("wasm-cache")
	serverOptions.setDefault()
	return &HTTPServer{
		SeverOptions:        serverOptions,
		mappingPath2Cache:   make(map[string]wasmModuleEntry),
		failedAttempts:      make(map[string]attemptEntry),
		cache:               newLocalFileCache(cacheOptions, logger),
		controllerNamespace: controllerNamespace,
		logger:              logger,
	}
}

func (s *HTTPServer) Start(ctx context.Context) {
	s.logger.Info(fmt.Sprintf("Listening on :%d", serverPort))

	handler := http.NewServeMux()
	handler.Handle("/", s)

	s.server = &http.Server{
		Addr:              fmt.Sprintf(":%d", serverPort),
		Handler:           handler,
		TLSConfig:         s.TLSConfig,
		ReadHeaderTimeout: 15 * time.Second,
	}

	var err error
	go func() {
		if s.enableTLS() {
			err = s.server.ListenAndServeTLS("", "")
		} else {
			err = s.server.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error(err, "Failed to start Wasm HTTP server")
			return
		}
	}()

	go func() {
		// waiting for shutdown
		<-ctx.Done()
		_ = s.server.Shutdown(context.Background())
	}()
	s.cache.Start(ctx)
	go s.resetFailedAttempts(ctx)
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
func (s *HTTPServer) Get(originalURL string, opts *GetOptions) (servingURL, checksum string, err error) {
	var (
		mappingPath string
		localFile   string
	)

	s.Lock()
	defer s.Unlock()
	attempt, attempted := s.failedAttempts[originalURL]

	if attempted && attempt.fails > s.MaxFailedAttempts {
		err = fmt.Errorf("failed to get Wasm module %s after %d attempts", originalURL, s.MaxFailedAttempts)
		s.logger.Error(err, "")
		return "", "", err
	}

	// Get the local file path of the cached Wasm module.
	// Even it's already cached, the file cache may still download the Wasm module
	// again if it is expired or it needs to be updated.
	if localFile, checksum, err = s.cache.Get(originalURL, opts); err != nil {
		s.logger.Error(err, "Failed to get Wasm module", "URL", originalURL)
		attempt, attempted = s.failedAttempts[originalURL]
		if !attempted {
			attempt = attemptEntry{fails: 0, last: time.Now(), delay: s.FailedAttemptResetDelay}
		}
		attempt.fails++
		attempt.last = time.Now()
		s.failedAttempts[originalURL] = attempt
		return "", "", err
	}
	delete(s.failedAttempts, originalURL)

	// Generate a new path with the hash of the original url and a salt to
	// make the URL unpredictable.
	// The unguessable path is used to prevent unauthorized users from accessing
	// an unauthorized private Wasm module.
	mappingPath = generateUnguessablePath(originalURL, s.Salt)
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
	serverHostFQDN := fmt.Sprintf("%s.%s.svc.cluster.local", serverHost, s.controllerNamespace)
	servingURL = fmt.Sprintf("%s://%s:%d/%s", scheme, serverHostFQDN, serverPort, mappingPath)
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
	return s.TLSConfig != nil
}

// resetFailedAttempts resets the failed attempts.
// After reset, the cache will try to download the failed Wasm module again the
// next time it is requested.
func (s *HTTPServer) resetFailedAttempts(ctx context.Context) {
	ticker := time.NewTicker(s.FailedAttemptsResetInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.Lock()
			for k, m := range s.failedAttempts {
				if m.expired() {
					s.logger.Info("Reset failed attempts", "URL", k)
					delete(s.failedAttempts, k)
				}
			}
			s.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
