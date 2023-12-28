// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_redirectPathMatcher(t *testing.T) {

	tests := []struct {
		name        string
		redirectURI string
		want        string
	}{
		{
			name:        "redirectURI with path",
			redirectURI: "https://example.com/test/oauth2/callback",
			want:        "/test/oauth2/callback",
		},
		{
			name:        "redirectURI with header tokens",
			redirectURI: "%REQ(x-forwarded-proto)%://%REQ(:authority)%/test/oauth2/callback",
			want:        "/test/oauth2/callback",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, redirectPath(tt.redirectURI), "redirectPathMatcher(%v)", tt.redirectURI)
		})
	}
}
