// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gateway

import (
	"io"
	"net/http"
	"time"

	"github.com/go-logr/logr"
)

// headersForwardedToAuth are the request headers forwarded to go-authentication's
// /v1/auth/check endpoint so it can authenticate and authorize the caller.
var headersForwardedToAuth = []string{
	"Authorization",
	"x-api-key",
	"x_gmp_tenant",
	"x-date",
	"x_gmp_correlation_id",
	"x_auth_request_type",
}

// authMiddleware validates the caller's bearer token by delegating to
// go-authentication's /v1/auth/check endpoint. A 200 response allows the
// request to proceed; any other status is forwarded back to the caller.
type authMiddleware struct {
	authURL string
	client  *http.Client
	logger  logr.Logger
}

// newAuthMiddleware creates an authMiddleware that calls the given go-auth base URL.
func newAuthMiddleware(authURL string, logger logr.Logger) *authMiddleware {
	return &authMiddleware{
		authURL: authURL,
		client:  &http.Client{Timeout: 10 * time.Second},
		logger:  logger,
	}
}

// wrap returns an http.HandlerFunc that runs the auth check before invoking next.
func (a *authMiddleware) wrap(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" && r.Header.Get("x-api-key") == "" {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authorization header required")
			return
		}

		checkReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, a.authURL+"/v1/auth/check", nil)
		if err != nil {
			a.logger.Error(err, "failed to build auth check request")
			writeError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to build authentication request")
			return
		}

		// Forward auth-relevant headers from the caller.
		for _, h := range headersForwardedToAuth {
			if v := r.Header.Get(h); v != "" {
				checkReq.Header.Set(h, v)
			}
		}
		// Tell go-auth what the caller is trying to do so it can authorize via
		// the api_permission table. The Lua filter normally injects these for
		// data-plane traffic; here we set them directly.
		checkReq.Header.Set("method", r.Method)
		checkReq.Header.Set("path", r.URL.Path)

		resp, err := a.client.Do(checkReq)
		if err != nil {
			a.logger.Error(err, "authentication service unreachable", "authURL", a.authURL)
			writeError(w, http.StatusBadGateway, "AUTH_UNAVAILABLE", "Authentication service unreachable")
			return
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			// Forward go-auth's error response verbatim.
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(resp.StatusCode)
			_, _ = io.Copy(w, resp.Body)
			return
		}

		// Auth passed. Propagate the authz key header if go-auth returned one.
		if key := resp.Header.Get("x_gmp_authz_key"); key != "" {
			r.Header.Set("x_gmp_authz_key", key)
		}
		next(w, r)
	}
}
