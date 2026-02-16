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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestStaticFiles(t *testing.T) {
	cfg := &config.Server{
		EnvoyGateway: egv1a1.DefaultEnvoyGateway(),
		Logger:       logging.DefaultLogger(nil, egv1a1.LogLevelInfo),
	}

	handler := NewHandler(cfg, (*message.ProviderResources)(nil))
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Test CSS file
	t.Run("CSS file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/css/admin.css", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "text/css")
		assert.Contains(t, w.Body.String(), "Envoy Gateway Admin Console Styles")
	})

	// Test JS file
	t.Run("JS file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/js/admin.js", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Header().Get("Content-Type"), "javascript")
		assert.Contains(t, w.Body.String(), "EnvoyGatewayAdmin")
	})

	// Test non-existent file
	t.Run("Non-existent file", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/static/nonexistent.txt", nil)
		w := httptest.NewRecorder()

		mux.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
