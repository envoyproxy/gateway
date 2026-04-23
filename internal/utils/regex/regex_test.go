// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package regex

import (
	"regexp"
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		regex   string
		wantErr bool
	}{
		{
			name:    "Valid regex",
			regex:   "^[a-z0-9-]+$",
			wantErr: false,
		},
		{
			name:    "Invalid regex",
			regex:   "^[a-z0-9-++$",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Validate(tt.regex); (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestPathSeparatedPrefixRegex(t *testing.T) {
	tests := []struct {
		name     string
		prefix   string
		testPath string
		want     bool
	}{
		// Tests for prefix "/api/v1" - what should match
		{
			name:     "exact match",
			prefix:   "/api/v1",
			testPath: "/api/v1",
			want:     true,
		},
		{
			name:     "with trailing slash",
			prefix:   "/api/v1",
			testPath: "/api/v1/",
			want:     true,
		},
		{
			name:     "with sub-path",
			prefix:   "/api/v1",
			testPath: "/api/v1/users",
			want:     true,
		},
		{
			name:     "with deep sub-path",
			prefix:   "/api/v1",
			testPath: "/api/v1/users/123/profile",
			want:     true,
		},
		{
			name:     "with query params",
			prefix:   "/api/v1",
			testPath: "/api/v1?version=latest",
			want:     true,
		},
		{
			name:     "with complex query",
			prefix:   "/api/v1",
			testPath: "/api/v1?param1=value1&param2=value2",
			want:     true,
		},
		{
			name:     "with fragment",
			prefix:   "/api/v1",
			testPath: "/api/v1#section",
			want:     true,
		},
		{
			name:     "with semicolon parameter",
			prefix:   "/api/v1",
			testPath: "/api/v1;sessionid=123",
			want:     true,
		},
		{
			name:     "with semicolon and sub-path",
			prefix:   "/api/v1",
			testPath: "/api/v1;sessionid=123/profile",
			want:     true,
		},

		// Tests for prefix "/api/v1" - what should NOT match
		{
			name:     "alphanumeric continuation",
			prefix:   "/api/v1",
			testPath: "/api/v1abc",
			want:     false,
		},
		{
			name:     "underscore continuation",
			prefix:   "/api/v1",
			testPath: "/api/v1_test",
			want:     false,
		},
		{
			name:     "dash continuation",
			prefix:   "/api/v1",
			testPath: "/api/v1-beta",
			want:     false,
		},
		{
			name:     "dot continuation",
			prefix:   "/api/v1",
			testPath: "/api/v1.1",
			want:     false,
		},
		{
			name:     "different path completely",
			prefix:   "/api/v1",
			testPath: "/api/v2",
			want:     false,
		},
		{
			name:     "prefix longer than path",
			prefix:   "/api/v1",
			testPath: "/api",
			want:     false,
		},
		{
			name:     "similar but different",
			prefix:   "/api/v1",
			testPath: "/api/v10",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pattern := PathSeparatedPrefixRegex(tt.prefix)

			regex, err := regexp.Compile(pattern)
			if err != nil {
				t.Fatalf("Failed to compile regex pattern %q: %v", pattern, err)
			}

			got := regex.MatchString(tt.testPath)
			if got != tt.want {
				t.Errorf("PathSeparatedPrefixRegex(%q).MatchString(%q) = %v, want %v (pattern: %q)",
					tt.prefix, tt.testPath, got, tt.want, pattern)
			}
		})
	}
}
