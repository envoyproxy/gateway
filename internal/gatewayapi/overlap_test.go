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

func TestRouteMatchesOverlap(t *testing.T) {
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
			name: "prefix paths are not detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
			},
			overlap: false,
		},
		{
			name: "nil path matches are not detected as overlap",
			a: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			b: &ir.HTTPRoute{
				Hostname: "example.com",
			},
			overlap: false,
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
			name: "regex path matches are not detected as overlap",
			a: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: ptr.To("/foo.*")},
			},
			b: &ir.HTTPRoute{
				Hostname:  "example.com",
				PathMatch: &ir.StringMatch{SafeRegex: ptr.To("/foo.*")},
			},
			overlap: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.overlap, routeMatchesOverlap(tt.a, tt.b))
		})
	}
}

func TestStringMatchEqual(t *testing.T) {
	tests := []struct {
		name  string
		a     *ir.StringMatch
		b     *ir.StringMatch
		equal bool
	}{
		{
			name:  "both nil",
			a:     nil,
			b:     nil,
			equal: true,
		},
		{
			name:  "one nil",
			a:     &ir.StringMatch{Exact: ptr.To("foo")},
			b:     nil,
			equal: false,
		},
		{
			name:  "identical exact",
			a:     &ir.StringMatch{Name: "h", Exact: ptr.To("v")},
			b:     &ir.StringMatch{Name: "h", Exact: ptr.To("v")},
			equal: true,
		},
		{
			name:  "different name",
			a:     &ir.StringMatch{Name: "a", Exact: ptr.To("v")},
			b:     &ir.StringMatch{Name: "b", Exact: ptr.To("v")},
			equal: false,
		},
		{
			name:  "different value",
			a:     &ir.StringMatch{Name: "h", Exact: ptr.To("v1")},
			b:     &ir.StringMatch{Name: "h", Exact: ptr.To("v2")},
			equal: false,
		},
		{
			name:  "different distinct",
			a:     &ir.StringMatch{Name: "h", Distinct: true},
			b:     &ir.StringMatch{Name: "h", Distinct: false},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, stringMatchEqual(tt.a, tt.b))
		})
	}
}

func TestStringMatchSliceEqual(t *testing.T) {
	tests := []struct {
		name  string
		a     []*ir.StringMatch
		b     []*ir.StringMatch
		equal bool
	}{
		{
			name:  "both empty",
			a:     nil,
			b:     nil,
			equal: true,
		},
		{
			name:  "different lengths",
			a:     []*ir.StringMatch{{Name: "a", Exact: ptr.To("1")}},
			b:     nil,
			equal: false,
		},
		{
			name: "identical",
			a: []*ir.StringMatch{
				{Name: "a", Exact: ptr.To("1")},
				{Name: "b", Exact: ptr.To("2")},
			},
			b: []*ir.StringMatch{
				{Name: "a", Exact: ptr.To("1")},
				{Name: "b", Exact: ptr.To("2")},
			},
			equal: true,
		},
		{
			name: "same elements different order",
			a: []*ir.StringMatch{
				{Name: "a", Exact: ptr.To("1")},
				{Name: "b", Exact: ptr.To("2")},
			},
			b: []*ir.StringMatch{
				{Name: "b", Exact: ptr.To("2")},
				{Name: "a", Exact: ptr.To("1")},
			},
			equal: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.equal, stringMatchSliceEqual(tt.a, tt.b))
		})
	}
}
