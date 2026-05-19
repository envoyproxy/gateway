// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		wantScheme      string
		wantHostAndPort string
		wantErr         bool
		errContains     string
	}{
		{
			name:            "valid unix domain socket",
			url:             "unix:///var/run/app.sock",
			wantScheme:      "unix",
			wantHostAndPort: "/var/run/app.sock",
			wantErr:         false,
		},
		{
			name:            "valid unix domain socket relative path",
			url:             "unix://./app.sock",
			wantScheme:      "unix",
			wantHostAndPort: "/app.sock",
			wantErr:         false,
		},
		{
			name:            "invalid url",
			url:             "://invalid",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "invalid URL",
		},
		{
			name:            "http url not supported",
			url:             "http://example.com:8080",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "unsupported URL scheme",
		},
		{
			name:            "https url not supported",
			url:             "https://example.com:9443",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "unsupported URL scheme",
		},
		{
			name:            "unix url without path",
			url:             "unix://",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "must contain a path",
		},
		{
			name:            "unsupported scheme ftp",
			url:             "ftp://example.com",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "unsupported URL scheme",
		},
		{
			name:            "unsupported scheme ws",
			url:             "ws://example.com:8080",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "unsupported URL scheme",
		},
		{
			name:            "empty url",
			url:             "",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "unsupported URL scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScheme, gotHostAndPort, err := ParseURL(tt.url)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					require.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.wantScheme, gotScheme)
			require.Equal(t, tt.wantHostAndPort, gotHostAndPort)
		})
	}
}
