// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

// routesOverlap is a test helper that returns true if two routes produce the
// same canonical overlap key (i.e., they match the same set of requests).
func routesOverlap(a, b *ir.HTTPRoute) bool {
	return buildOverlapKey(a) == buildOverlapKey(b)
}

func TestBuildOverlapKey(t *testing.T) {
	tests := []struct {
		name    string
		a       *ir.HTTPRoute
		b       *ir.HTTPRoute
		overlap bool
	}{
		{
			name: "identical exact path and hostname",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			overlap: true,
		},
		{
			name: "different exact paths",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/bar")},
			},
			overlap: false,
		},
		{
			name: "different hostnames",
			a: &ir.HTTPRoute{
				Hostname:  "a.example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "b.example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			overlap: false,
		},
		{
			name: "identical prefix paths are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
			},
			overlap: true,
		},
		{
			name: "nil path matches are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			overlap: true,
		},
		{
			name: "exact vs prefix same value not overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: ptr.To("/foo")},
			},
			overlap: false,
		},
		{
			name: "identical exact path with identical header matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical exact path with different header matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val2")},
				},
			},
			overlap: false,
		},
		{
			name: "identical exact path one has headers other does not",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
			},
			overlap: false,
		},
		{
			name: "header names are compared case-insensitively",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-Custom", Exact: ptr.To("val1")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "x-custom", Exact: ptr.To("val1")},
				},
			},
			overlap: true,
		},
		{
			name: "header matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-A", Exact: ptr.To("1")},
					{Name: "X-B", Exact: ptr.To("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				HeaderMatches: []*ir.StringMatch{
					{Name: "X-B", Exact: ptr.To("2")},
					{Name: "X-A", Exact: ptr.To("1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical exact path with identical query param matches",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				QueryParamMatches: []*ir.StringMatch{
					{Name: "key", Exact: ptr.To("value")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")},
				QueryParamMatches: []*ir.StringMatch{
					{Name: "key", Exact: ptr.To("value")},
				},
			},
			overlap: true,
		},
		{
			name: "query param matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
				QueryParamMatches: []*ir.StringMatch{
					{Name: "a", Exact: ptr.To("1")},
					{Name: "b", Exact: ptr.To("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
				QueryParamMatches: []*ir.StringMatch{
					{Name: "b", Exact: ptr.To("2")},
					{Name: "a", Exact: ptr.To("1")},
				},
			},
			overlap: true,
		},
		{
			name: "cookie matches in different order still overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
				CookieMatches: []*ir.StringMatch{
					{Name: "a", Exact: ptr.To("1")},
					{Name: "b", Exact: ptr.To("2")},
				},
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
				CookieMatches: []*ir.StringMatch{
					{Name: "b", Exact: ptr.To("2")},
					{Name: "a", Exact: ptr.To("1")},
				},
			},
			overlap: true,
		},
		{
			name: "identical regex path matches are detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: ptr.To("/foo.*")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: ptr.To("/foo.*")},
			},
			overlap: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.overlap, routesOverlap(tt.a, tt.b))
		})
	}
}
