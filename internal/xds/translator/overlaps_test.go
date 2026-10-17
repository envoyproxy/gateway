// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDetectMisdirectedRequests(t *testing.T) {
	tests := []struct {
		name     string
		listener *ir.HTTPListener
		want     bool
	}{
		{name: "nil listener", want: false},
		{
			name: "no overlapping TLS",
			listener: &ir.HTTPListener{
				OverlappingTLSHandling: ir.OverlappingTLSHandlingMisdirectedRequest,
			},
			want: false,
		},
		{
			name: "unset handling",
			listener: &ir.HTTPListener{
				TLSOverlaps: true,
			},
			want: false,
		},
		{
			name: "downgrade handling",
			listener: &ir.HTTPListener{
				TLSOverlaps:            true,
				OverlappingTLSHandling: ir.OverlappingTLSHandlingDowngradeToHTTP1,
			},
			want: false,
		},
		{
			name: "misdirected request handling",
			listener: &ir.HTTPListener{
				TLSOverlaps:            true,
				OverlappingTLSHandling: ir.OverlappingTLSHandlingMisdirectedRequest,
			},
			want: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, detectMisdirectedRequests(tc.listener))
		})
	}
}

func TestDomainsMatchHostname(t *testing.T) {
	tests := []struct {
		name     string
		domains  []string
		hostname string
		want     bool
	}{
		{name: "exact", domains: []string{"foo.example.com"}, hostname: "foo.example.com", want: true},
		{name: "exact mismatch", domains: []string{"foo.example.com"}, hostname: "bar.example.com", want: false},
		{name: "any", domains: []string{"*"}, hostname: "foo.example.com", want: true},
		{name: "wildcard single label", domains: []string{"*.example.com"}, hostname: "foo.example.com", want: true},
		{name: "wildcard multiple labels", domains: []string{"*.example.com"}, hostname: "foo.bar.example.com", want: true},
		{name: "wildcard matches narrower wildcard", domains: []string{"*.example.com"}, hostname: "*.foo.example.com", want: true},
		{name: "wildcard does not match apex", domains: []string{"*.example.com"}, hostname: "example.com", want: false},
		{name: "wildcard does not match broader wildcard", domains: []string{"*.foo.example.com"}, hostname: "*.example.com", want: false},
		{name: "exact does not match wildcard", domains: []string{"foo.example.com"}, hostname: "*.example.com", want: false},
		{name: "any of the domains", domains: []string{"bar.com", "*.example.com"}, hostname: "foo.example.com", want: true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, domainsMatchHostname(tc.domains, tc.hostname))
		})
	}
}

func TestBuildMisdirectedRequestRouteAuthority(t *testing.T) {
	tests := []struct {
		hostname   string
		matches    []string
		mismatches []string
	}{
		{
			hostname:   "foo.example.com",
			matches:    []string{"foo.example.com", "foo.example.com:443"},
			mismatches: []string{"bar.example.com", "xfoo.example.com", "foo.example.com.evil", "fooXexample.com"},
		},
		{
			hostname:   "*.example.com",
			matches:    []string{"foo.example.com", "foo.bar.example.com", "foo.example.com:8443"},
			mismatches: []string{"example.com", ".example.com", "foo.example.org", "fooexample.com"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.hostname, func(t *testing.T) {
			route := buildMisdirectedRequestRoute(tc.hostname)
			require.EqualValues(t, 421, route.GetDirectResponse().GetStatus())
			require.Len(t, route.GetMatch().GetHeaders(), 1)
			re := regexp.MustCompile(route.GetMatch().GetHeaders()[0].GetStringMatch().GetSafeRegex().GetRegex())
			for _, authority := range tc.matches {
				require.True(t, re.MatchString(authority), authority)
			}
			for _, authority := range tc.mismatches {
				require.False(t, re.MatchString(authority), authority)
			}
		})
	}

	t.Run("any hostname has no authority match", func(t *testing.T) {
		require.Empty(t, buildMisdirectedRequestRoute("*").GetMatch().GetHeaders())
	})
}
