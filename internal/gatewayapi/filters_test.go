// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestValidateMirrorClusterHeader(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		wantErr string
	}{
		{
			name:  "regular header",
			value: "x-shadow-cluster",
		},
		{
			name:  "authority pseudo-header",
			value: ":authority",
		},
		{
			name:    "empty",
			wantErr: "must not be empty",
		},
		{
			name:    "space",
			value:   "x shadow cluster",
			wantErr: "valid HTTP header name or pseudo-header name",
		},
		{
			name:    "slash",
			value:   "x/shadow-cluster",
			wantErr: "valid HTTP header name or pseudo-header name",
		},
		{
			name:    "too long",
			value:   strings.Repeat("a", mirrorHeaderNameMaxLength+1),
			wantErr: "must be no more than 256 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMirrorClusterHeader(tc.value)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestValidateMirrorHostRewriteLiteral(t *testing.T) {
	testCases := []struct {
		name    string
		value   string
		wantErr string
	}{
		{
			name:  "valid",
			value: "shadow.example.com",
		},
		{
			name:    "empty",
			wantErr: "must not be empty",
		},
		{
			name:    "space",
			value:   "shadow example.com",
			wantErr: "visible ASCII characters without spaces",
		},
		{
			name:    "non-ascii",
			value:   "shadow.example.com\u2603",
			wantErr: "visible ASCII characters without spaces",
		},
		{
			name:    "too long",
			value:   strings.Repeat("a", mirrorHostRewriteLiteralMaxLength+1),
			wantErr: "must be no more than 255 characters",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMirrorHostRewriteLiteral(tc.value)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}

func TestValidateMirrorRequestHeaderMutations(t *testing.T) {
	testCases := []struct {
		name    string
		filter  *egv1a1.HTTPHeaderFilter
		wantErr string
	}{
		{
			name: "valid mutations",
			filter: &egv1a1.HTTPHeaderFilter{
				Add: []gwapiv1.HTTPHeader{
					{Name: "x-shadow-added", Value: "true"},
				},
				Set: []gwapiv1.HTTPHeader{
					{Name: "x-shadow-set", Value: "shadow"},
				},
				AddIfAbsent: []gwapiv1.HTTPHeader{
					{Name: "x-shadow-default", Value: "fallback"},
				},
				Remove: []string{"x-shadow-remove"},
			},
		},
		{
			name: "host mutation",
			filter: &egv1a1.HTTPHeaderFilter{
				Set: []gwapiv1.HTTPHeader{
					{Name: "Host", Value: "shadow.example.com"},
				},
			},
			wantErr: "To modify the Host header use hostRewriteLiteral",
		},
		{
			name: "pseudo-header mutation",
			filter: &egv1a1.HTTPHeaderFilter{
				Remove: []string{":authority"},
			},
			wantErr: "headers with a '/' or ':' character",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateMirrorRequestHeaderMutations(tc.filter)
			if tc.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.ErrorContains(t, err, tc.wantErr)
		})
	}
}
