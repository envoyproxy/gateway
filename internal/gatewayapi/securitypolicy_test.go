// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_wildcard2regex(t *testing.T) {
	tests := []struct {
		name     string
		wildcard string
		origin   string
		want     int
	}{
		{
			name:     "test1",
			wildcard: "http://*.example.com",
			origin:   "http://foo.example.com",
			want:     1,
		},
		{
			name:     "test2",
			wildcard: "http://*.example.com",
			origin:   "http://foo.bar.example.com",
			want:     1,
		},
		{
			name:     "test3",
			wildcard: "http://*.example.com",
			origin:   "http://foo.bar.com",
			want:     0,
		},
		{
			name:     "test4",
			wildcard: "http://*.example.com",
			origin:   "https://foo.example.com",
			want:     0,
		},
		{
			name:     "test5",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.example.com:8080",
			want:     1,
		},
		{
			name:     "test6",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.bar.example.com:8080",
			want:     1,
		},
		{
			name:     "test7",
			wildcard: "http://*.example.com:8080",
			origin:   "http://foo.example.com",
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			regexStr := wildcard2regex(tt.wildcard)
			regex, err := regexp.Compile(regexStr)
			require.Nil(t, err)
			finds := regex.FindAllString(tt.origin, -1)
			assert.Equalf(t, tt.want, len(finds), "wildcard2regex(%v)", tt.wildcard)
		})
	}
}
