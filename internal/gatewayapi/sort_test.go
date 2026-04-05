// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestPathMatchCount(t *testing.T) {
	cases := []struct {
		name     string
		match    *ir.StringMatch
		expected int
	}{
		{
			name:     "nil match returns 0",
			match:    nil,
			expected: 0,
		},
		{
			name:     "exact match returns length",
			match:    &ir.StringMatch{Exact: ptr.To("/foo/bar")},
			expected: 8,
		},
		{
			name:     "regex match returns length",
			match:    &ir.StringMatch{SafeRegex: ptr.To("/foo/.+")},
			expected: 7,
		},
		{
			name:     "prefix match returns length",
			match:    &ir.StringMatch{Prefix: ptr.To("/api")},
			expected: 4,
		},
		{
			name:     "root prefix returns 0",
			match:    &ir.StringMatch{Prefix: ptr.To("/")},
			expected: 0,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, pathMatchCount(tc.match))
		})
	}
}

func TestNumberOfExactMatches(t *testing.T) {
	cases := []struct {
		name     string
		matches  []*ir.StringMatch
		expected int
	}{
		{
			name:     "nil slice returns 0",
			matches:  nil,
			expected: 0,
		},
		{
			name: "counts only exact matches",
			matches: []*ir.StringMatch{
				{Exact: ptr.To("val1")},
				{Prefix: ptr.To("val2")},
				{Exact: ptr.To("val3")},
			},
			expected: 2,
		},
		{
			name: "nil element skipped",
			matches: []*ir.StringMatch{
				nil,
				{Exact: ptr.To("val1")},
			},
			expected: 1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, numberOfExactMatches(tc.matches))
		})
	}
}

func TestXdsIRRoutesSort(t *testing.T) {
	cases := []struct {
		name          string
		routes        []*ir.HTTPRoute
		expectedOrder []string // route names in expected order after descending sort
	}{
		{
			name: "exact before regex before prefix",
			routes: []*ir.HTTPRoute{
				{Name: "prefix", PathMatch: &ir.StringMatch{Prefix: ptr.To("/foo")}},
				{Name: "exact", PathMatch: &ir.StringMatch{Exact: ptr.To("/foo")}},
				{Name: "regex", PathMatch: &ir.StringMatch{SafeRegex: ptr.To("/foo")}},
			},
			expectedOrder: []string{"exact", "regex", "prefix"},
		},
		{
			name: "longer path wins within same type",
			routes: []*ir.HTTPRoute{
				{Name: "short", PathMatch: &ir.StringMatch{Prefix: ptr.To("/a")}},
				{Name: "long", PathMatch: &ir.StringMatch{Prefix: ptr.To("/api/v1")}},
			},
			expectedOrder: []string{"long", "short"},
		},
		{
			name: "root prefix treated as zero length",
			routes: []*ir.HTTPRoute{
				{Name: "root", PathMatch: &ir.StringMatch{Prefix: ptr.To("/")}},
				{Name: "api", PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")}},
			},
			expectedOrder: []string{"api", "root"},
		},
		{
			name: "more headers wins when path equal",
			routes: []*ir.HTTPRoute{
				{
					Name:      "one-header",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					HeaderMatches: []*ir.StringMatch{
						{Name: "h1", Exact: ptr.To("v1")},
					},
				},
				{
					Name:      "two-headers",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					HeaderMatches: []*ir.StringMatch{
						{Name: "h1", Exact: ptr.To("v1")},
						{Name: "h2", Exact: ptr.To("v2")},
					},
				},
			},
			expectedOrder: []string{"two-headers", "one-header"},
		},
		{
			name: "more exact header matches wins when header count equal",
			routes: []*ir.HTTPRoute{
				{
					Name:      "prefix-headers",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					HeaderMatches: []*ir.StringMatch{
						{Name: "h1", Prefix: ptr.To("v1")},
						{Name: "h2", Prefix: ptr.To("v2")},
					},
				},
				{
					Name:      "exact-headers",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					HeaderMatches: []*ir.StringMatch{
						{Name: "h1", Exact: ptr.To("v1")},
						{Name: "h2", Exact: ptr.To("v2")},
					},
				},
			},
			expectedOrder: []string{"exact-headers", "prefix-headers"},
		},
		{
			name: "more cookie matches wins when headers equal",
			routes: []*ir.HTTPRoute{
				{
					Name:      "no-cookies",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
				},
				{
					Name:      "with-cookies",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					CookieMatches: []*ir.StringMatch{
						{Name: "c1", Exact: ptr.To("v1")},
					},
				},
			},
			expectedOrder: []string{"with-cookies", "no-cookies"},
		},
		{
			name: "more query param matches wins when cookies equal",
			routes: []*ir.HTTPRoute{
				{
					Name:      "no-query",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
				},
				{
					Name:      "with-query",
					PathMatch: &ir.StringMatch{Prefix: ptr.To("/api")},
					QueryParamMatches: []*ir.StringMatch{
						{Name: "q1", Exact: ptr.To("v1")},
					},
				},
			},
			expectedOrder: []string{"with-query", "no-query"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			routes := XdsIRRoutes(tc.routes)
			sort.Stable(sort.Reverse(routes))

			gotOrder := make([]string, len(routes))
			for i, r := range routes {
				gotOrder[i] = r.Name
			}
			require.Equal(t, tc.expectedOrder, gotOrder)
		})
	}
}
