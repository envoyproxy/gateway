// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthCheckerHandlerAddsClientResponseHeader(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/auth/myapp", nil)
	req.Header.Set("Authorization", "Bearer token1")
	res := httptest.NewRecorder()

	authCheckerHandler(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("status code = %d, want %d", res.Code, http.StatusOK)
	}
	const want = "ext-auth-session=user1; Path=/; HttpOnly"
	if got := res.Header().Get("Set-Cookie"); got != want {
		t.Fatalf("Set-Cookie = %q, want %q", got, want)
	}
}
