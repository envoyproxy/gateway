// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_hasTag(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
		want     bool
	}{
		{
			name:     "image with scheme and tag",
			imageURL: "oci://www.example.com/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image with scheme, host port and tag",
			imageURL: "oci://www.example.com:8080/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image with scheme without tag",
			imageURL: "oci://www.example.com/wasm",
			want:     false,
		},
		{
			name:     "image with scheme, host port without tag",
			imageURL: "oci://www.example.com:8080/wasm",
			want:     false,
		},
		{
			name:     "image without scheme with tag",
			imageURL: "www.example.com/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image without scheme with host port and tag",
			imageURL: "www.example.com:8080/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image without scheme without tag",
			imageURL: "www.example.com/wasm",
			want:     false,
		},
		{
			name:     "image without scheme with host port without tag",
			imageURL: "www.example.com:8080/wasm",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasTag(tt.imageURL), "hasTag(%v)", tt.imageURL)
		})
	}
}

func TestValidateDynamicModuleRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr string
	}{
		{
			name:   "valid https URL",
			rawURL: "https://modules.example.com/libremote_auth.so",
		},
		{
			name:   "valid http URL with port",
			rawURL: "http://modules.example.com:8443/libremote_auth.so",
		},
		{
			name:    "missing hostname",
			rawURL:  "https:///libremote_auth.so",
			wantErr: "hostname",
		},
		{
			name:    "unsupported scheme",
			rawURL:  "ftp://modules.example.com/libremote_auth.so",
			wantErr: "unsupported URL scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDynamicModuleRemoteURL(tt.rawURL)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}
