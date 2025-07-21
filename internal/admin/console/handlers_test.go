// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package console

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestNewHandler(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	providerResources := &message.ProviderResources{}

	handler := NewHandler(cfg, providerResources)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.cfg)
	assert.Equal(t, providerResources, handler.providerResources)
}

func TestHandleIndex(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	handler.handleIndex(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Envoy Gateway Admin Console")
}

func TestHandleIndexMethodNotAllowed(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, nil)

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	w := httptest.NewRecorder()

	handler.handleIndex(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}

func TestHandlePprof(t *testing.T) {
	tests := []struct {
		name        string
		enablePprof bool
		expectBody  string
	}{
		{
			name:        "pprof enabled",
			enablePprof: true,
			expectBody:  "Performance Profiling",
		},
		{
			name:        "pprof disabled",
			enablePprof: false,
			expectBody:  "Performance Profiling",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Server{
				EnvoyGateway: &egv1a1.EnvoyGateway{
					EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
						Admin: &egv1a1.EnvoyGatewayAdmin{
							EnablePprof: tt.enablePprof,
						},
					},
				},
				Logger: logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
			}

			handler := NewHandler(cfg, (*message.ProviderResources)(nil))

			req := httptest.NewRequest(http.MethodGet, "/pprof", nil)
			w := httptest.NewRecorder()

			handler.handlePprof(w, req)

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
			assert.Contains(t, w.Body.String(), tt.expectBody)
		})
	}
}

func TestHandleServerInfo(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/server_info", nil)
	w := httptest.NewRecorder()

	handler.handleServerInfo(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Server Information")
}

func TestHandleConfigDump(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/config_dump", nil)
	w := httptest.NewRecorder()

	handler.handleConfigDump(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Configuration Dump")
}

func TestHandleStats(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))

	req := httptest.NewRequest(http.MethodGet, "/stats", nil)
	w := httptest.NewRecorder()

	handler.handleStats(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Content-Type"), "text/html")
	assert.Contains(t, w.Body.String(), "Statistics")
}

func TestRegisterRoutes(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))
	mux := http.NewServeMux()

	// This should not panic
	require.NotPanics(t, func() {
		handler.RegisterRoutes(mux)
	})

	// Test that routes are registered by making requests
	testCases := []struct {
		path           string
		expectedStatus int
	}{
		{"/", http.StatusOK},
		{"/pprof", http.StatusOK},
		{"/server_info", http.StatusOK},
		{"/config_dump", http.StatusOK},
		{"/stats", http.StatusOK},
		{"/api/info", http.StatusOK},
		{"/api/server_info", http.StatusOK},
		{"/api/config_dump", http.StatusOK},
		{"/api/config_dump_all", http.StatusOK},
		{"/api/metrics", http.StatusOK}, // Now uses Prometheus registry directly
	}

	for _, tc := range testCases {
		t.Run(tc.path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, req)

			assert.Equal(t, tc.expectedStatus, w.Code)
		})
	}
}
