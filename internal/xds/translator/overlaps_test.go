// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/stretchr/testify/assert"

	"github.com/envoyproxy/gateway/internal/xds/filters"
)

func TestDomainMatchHostname(t *testing.T) {
	tests := []struct {
		name             string
		vhDomain         string
		overlapsHostname string
		want             bool
	}{
		{
			name:             "wildcard matches any hostname",
			vhDomain:         "*",
			overlapsHostname: "example.com",
			want:             true,
		},
		{
			name:             "wildcard matches empty hostname",
			vhDomain:         "*",
			overlapsHostname: "",
			want:             true,
		},
		{
			name:             "exact match",
			vhDomain:         "example.com",
			overlapsHostname: "example.com",
			want:             true,
		},
		{
			name:             "no match - different domains",
			vhDomain:         "example.com",
			overlapsHostname: "other.com",
			want:             false,
		},
		{
			name:             "wildcard subdomain matches single level",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "www.wildcard.com",
			want:             true,
		},
		{
			name:             "wildcard subdomain matches another single level",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "api.wildcard.com",
			want:             true,
		},
		{
			name:             "wildcard subdomain does NOT match two levels (Gateway API spec)",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "www.sub.wildcard.com",
			want:             false,
		},
		{
			name:             "wildcard subdomain does not match base domain",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "wildcard.com",
			want:             false,
		},
		{
			name:             "wildcard subdomain does not match different domain",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "example.com",
			want:             false,
		},
		{
			name:             "wildcard subdomain does not match partial suffix",
			vhDomain:         "*.wildcard.com",
			overlapsHostname: "notwildcard.com",
			want:             false,
		},
		{
			name:             "empty vhDomain does not match",
			vhDomain:         "",
			overlapsHostname: "example.com",
			want:             false,
		},
		{
			name:             "empty vhDomain and empty overlapsHostname match",
			vhDomain:         "",
			overlapsHostname: "",
			want:             true,
		},
		{
			name:             "vhDomain with single character wildcard",
			vhDomain:         "*",
			overlapsHostname: "a.b.c.d.e.com",
			want:             true,
		},
		{
			name:             "wildcard with short hostname",
			vhDomain:         "*.com",
			overlapsHostname: "a.com",
			want:             true,
		},
		{
			name:             "wildcard with hostname equals suffix length",
			vhDomain:         "*.example.com",
			overlapsHostname: ".example.com",
			want:             false,
		},
		{
			name:             "case sensitive exact match - different case",
			vhDomain:         "Example.com",
			overlapsHostname: "example.com",
			want:             false,
		},
		{
			name:             "wildcard subdomain with uppercase",
			vhDomain:         "*.Example.com",
			overlapsHostname: "www.Example.com",
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainMatchHostname(tt.vhDomain, tt.overlapsHostname)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDomainsMatched(t *testing.T) {
	tests := []struct {
		name             string
		vhDomains        []string
		overlapsHostname string
		want             bool
	}{
		{
			name:             "single domain exact match",
			vhDomains:        []string{"example.com"},
			overlapsHostname: "example.com",
			want:             true,
		},
		{
			name:             "single domain no match",
			vhDomains:        []string{"example.com"},
			overlapsHostname: "other.com",
			want:             false,
		},
		{
			name:             "multiple domains with match",
			vhDomains:        []string{"example.com", "test.com", "demo.com"},
			overlapsHostname: "test.com",
			want:             true,
		},
		{
			name:             "multiple domains no match",
			vhDomains:        []string{"example.com", "test.com", "demo.com"},
			overlapsHostname: "other.com",
			want:             false,
		},
		{
			name:             "wildcard in domains",
			vhDomains:        []string{"example.com", "*"},
			overlapsHostname: "anything.com",
			want:             true,
		},
		{
			name:             "wildcard subdomain match",
			vhDomains:        []string{"example.com", "*.wildcard.com"},
			overlapsHostname: "www.wildcard.com",
			want:             true,
		},
		{
			name:             "wildcard subdomain does not match multi-level",
			vhDomains:        []string{"example.com", "*.wildcard.com"},
			overlapsHostname: "www.sub.wildcard.com",
			want:             false,
		},
		{
			name:             "empty domains list",
			vhDomains:        []string{},
			overlapsHostname: "example.com",
			want:             false,
		},
		{
			name:             "nil domains list",
			vhDomains:        nil,
			overlapsHostname: "example.com",
			want:             false,
		},
		{
			name:             "first domain matches",
			vhDomains:        []string{"match.com", "other.com"},
			overlapsHostname: "match.com",
			want:             true,
		},
		{
			name:             "last domain matches",
			vhDomains:        []string{"other.com", "match.com"},
			overlapsHostname: "match.com",
			want:             true,
		},
		{
			name:             "middle domain matches",
			vhDomains:        []string{"first.com", "match.com", "last.com"},
			overlapsHostname: "match.com",
			want:             true,
		},
		{
			name:             "multiple wildcards with match",
			vhDomains:        []string{"*.example.com", "*.test.com"},
			overlapsHostname: "api.test.com",
			want:             true,
		},
		{
			name:             "complex domains list with wildcard match",
			vhDomains:        []string{"example.com", "*.wildcard.com", "test.com", "*"},
			overlapsHostname: "completely.different.com",
			want:             true,
		},
		{
			name:             "empty string in domains list",
			vhDomains:        []string{"", "example.com"},
			overlapsHostname: "",
			want:             true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := domainsMatched(tt.vhDomains, tt.overlapsHostname)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetReturn421RouteWithHost(t *testing.T) {
	tests := []struct {
		name                 string
		hostname             string
		expectHeaderMatcher  bool
		expectedRegexPattern string // expected regex pattern or empty if no matcher
	}{
		{
			name:                "wildcard asterisk - no header matcher",
			hostname:            "*",
			expectHeaderMatcher: false,
		},
		{
			name:                 "wildcard subdomain - regex with single label",
			hostname:             "*.example.com",
			expectHeaderMatcher:  true,
			expectedRegexPattern: `^[^.]+\.example\.com(:\d+)?$`,
		},
		{
			name:                 "wildcard subdomain complex - regex with single label",
			hostname:             "*.foo.bar.example.com",
			expectHeaderMatcher:  true,
			expectedRegexPattern: `^[^.]+\.foo\.bar\.example\.com(:\d+)?$`,
		},
		{
			name:                 "exact hostname - regex with optional port",
			hostname:             "foo.example.com",
			expectHeaderMatcher:  true,
			expectedRegexPattern: `^foo\.example\.com(:\d+)?$`,
		},
		{
			name:                 "exact hostname simple - regex with optional port",
			hostname:             "example.com",
			expectHeaderMatcher:  true,
			expectedRegexPattern: `^example\.com(:\d+)?$`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			route := getReturn421RouteWithHost(tt.hostname)

			// Verify basic route structure
			assert.NotNil(t, route)
			assert.NotNil(t, route.Match)
			assert.NotNil(t, route.Action)

			// Verify 421 status code
			directResp, ok := route.Action.(*routev3.Route_DirectResponse)
			assert.True(t, ok, "Expected DirectResponse action")
			assert.Equal(t, uint32(421), directResp.DirectResponse.Status)

			// Verify path prefix
			pathPrefix, ok := route.Match.PathSpecifier.(*routev3.RouteMatch_Prefix)
			assert.True(t, ok, "Expected Prefix path specifier")
			assert.Equal(t, "/", pathPrefix.Prefix)

			// Verify filter state for HTTP/2
			assert.Len(t, route.Match.FilterState, 1)
			assert.Equal(t, filters.DownstreamProtocolKey, route.Match.FilterState[0].Key)

			// Verify header matcher
			if !tt.expectHeaderMatcher {
				assert.Nil(t, route.Match.Headers, "Expected no header matchers for wildcard '*'")
			} else {
				assert.Len(t, route.Match.Headers, 1, "Expected exactly one header matcher")
				headerMatcher := route.Match.Headers[0]
				assert.Equal(t, ":authority", headerMatcher.Name)

				stringMatch, ok := headerMatcher.HeaderMatchSpecifier.(*routev3.HeaderMatcher_StringMatch)
				assert.True(t, ok, "Expected StringMatch header matcher")

				safeRegexMatch, ok := stringMatch.StringMatch.MatchPattern.(*matcherv3.StringMatcher_SafeRegex)
				assert.True(t, ok, "Expected SafeRegex matcher")
				assert.Equal(t, tt.expectedRegexPattern, safeRegexMatch.SafeRegex.Regex)
			}
		})
	}
}
