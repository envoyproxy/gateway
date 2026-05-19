// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package net

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			name:            "valid http url with port",
			url:             "http://example.com:8080",
			wantScheme:      "http",
			wantHostAndPort: "example.com:8080",
			wantErr:         false,
		},
		{
			name:            "valid http url without port",
			url:             "http://example.com",
			wantScheme:      "http",
			wantHostAndPort: "example.com:80",
			wantErr:         false,
		},
		{
			name:            "valid https url with port",
			url:             "https://example.com:9443",
			wantScheme:      "https",
			wantHostAndPort: "example.com:9443",
			wantErr:         false,
		},
		{
			name:            "valid https url without port",
			url:             "https://example.com",
			wantScheme:      "https",
			wantHostAndPort: "example.com:443",
			wantErr:         false,
		},
		{
			name:            "valid http url with path",
			url:             "http://example.com:8080/path/to/resource",
			wantScheme:      "http",
			wantHostAndPort: "example.com:8080",
			wantErr:         false,
		},
		{
			name:            "valid http url with query",
			url:             "http://example.com:8080?key=value",
			wantScheme:      "http",
			wantHostAndPort: "example.com:8080",
			wantErr:         false,
		},
		{
			name:            "valid http url with ipv4",
			url:             "http://192.168.1.1:8080",
			wantScheme:      "http",
			wantHostAndPort: "192.168.1.1:8080",
			wantErr:         false,
		},
		{
			name:            "valid http url with ipv6",
			url:             "http://[::1]:8080",
			wantScheme:      "http",
			wantHostAndPort: "[::1]:8080",
			wantErr:         false,
		},
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
			name:            "http url without host",
			url:             "http://",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "must contain a host",
		},
		{
			name:            "https url without host",
			url:             "https://",
			wantScheme:      "",
			wantHostAndPort: "",
			wantErr:         true,
			errContains:     "must contain a host",
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
		{
			name:            "http with localhost",
			url:             "http://localhost:3000",
			wantScheme:      "http",
			wantHostAndPort: "localhost:3000",
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotScheme, gotHostAndPort, err := ParseURL(tt.url)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantScheme, gotScheme)
			assert.Equal(t, tt.wantHostAndPort, gotHostAndPort)
		})
	}
}
