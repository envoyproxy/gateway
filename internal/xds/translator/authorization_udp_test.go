// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func cidrRule(action egv1a1.AuthorizationAction, cidr string, maskLen uint32) *ir.AuthorizationRule {
	return &ir.AuthorizationRule{
		Action: action,
		Principal: ir.Principal{
			ClientCIDRs: []*ir.CIDRMatch{{CIDR: cidr, MaskLen: maskLen}},
		},
	}
}

// udp_proxy can only route or not route, so a denial is expressed as the absence of a
// match. These cases pin down how the ordered Allow/Deny rules collapse into
// allow-only predicates, and in particular that on_no_match is set only when the
// default action routes unconditionally.
func TestBuildUDPProxyMatcher(t *testing.T) {
	routeAction := &cncfv3.TypedExtensionConfig{Name: "route"}

	tests := []struct {
		name          string
		authorization *ir.Authorization
		wantMatchers  int
		wantOnNoMatch bool
	}{
		{
			name:          "no authorization routes everything",
			authorization: nil,
			wantMatchers:  0,
			wantOnNoMatch: true,
		},
		{
			name: "allowlist emits one matcher per allow rule and drops the rest",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules: []*ir.AuthorizationRule{
					cidrRule(egv1a1.AuthorizationActionAllow, "192.168.100.0/24", 24),
					cidrRule(egv1a1.AuthorizationActionAllow, "10.1.0.0/16", 16),
				},
			},
			wantMatchers:  2,
			wantOnNoMatch: false,
		},
		{
			name: "denylist collapses to a single negated matcher",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionAllow,
				Rules: []*ir.AuthorizationRule{
					cidrRule(egv1a1.AuthorizationActionDeny, "10.0.0.0/24", 24),
					cidrRule(egv1a1.AuthorizationActionDeny, "172.16.0.0/12", 12),
				},
			},
			wantMatchers:  1,
			wantOnNoMatch: false,
		},
		{
			name: "a preceding deny rule narrows the allow rule and the default",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionAllow,
				Rules: []*ir.AuthorizationRule{
					cidrRule(egv1a1.AuthorizationActionDeny, "10.0.0.0/8", 8),
					cidrRule(egv1a1.AuthorizationActionAllow, "10.1.0.0/16", 16),
				},
			},
			// The narrowed allow rule, plus the catch-all for the permissive default.
			wantMatchers:  2,
			wantOnNoMatch: false,
		},
		{
			name: "an allow rule before a deny rule keeps its own entry",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionAllow,
				Rules: []*ir.AuthorizationRule{
					cidrRule(egv1a1.AuthorizationActionAllow, "10.1.0.0/16", 16),
					cidrRule(egv1a1.AuthorizationActionDeny, "10.0.0.0/8", 8),
					cidrRule(egv1a1.AuthorizationActionAllow, "10.0.0.0/7", 7),
				},
			},
			// The first allow rule needs no exclusion, the second is narrowed by
			// the deny between them, and the permissive default adds a catch-all.
			wantMatchers:  3,
			wantOnNoMatch: false,
		},
		{
			name: "an allow default with nothing denied is the same as no authorization",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionAllow,
				Rules: []*ir.AuthorizationRule{
					cidrRule(egv1a1.AuthorizationActionAllow, "192.168.100.0/24", 24),
				},
			},
			wantMatchers:  0,
			wantOnNoMatch: true,
		},
		{
			name: "a deny default with no allow rules drops everything",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
			},
			wantMatchers:  0,
			wantOnNoMatch: false,
		},
		{
			name: "rules without client CIDRs cannot match and are skipped",
			authorization: &ir.Authorization{
				DefaultAction: egv1a1.AuthorizationActionDeny,
				Rules: []*ir.AuthorizationRule{
					{Action: egv1a1.AuthorizationActionAllow},
					cidrRule(egv1a1.AuthorizationActionAllow, "192.168.100.0/24", 24),
				},
			},
			wantMatchers:  1,
			wantOnNoMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m, err := buildUDPProxyMatcher(routeAction, tt.authorization)
			require.NoError(t, err)

			require.Len(t, m.GetMatcherList().GetMatchers(), tt.wantMatchers)
			require.Equal(t, tt.wantOnNoMatch, m.GetOnNoMatch() != nil)

			// An empty matcher list fails proto validation, so it must be left unset
			// rather than emitted empty.
			if tt.wantMatchers == 0 {
				require.Nil(t, m.GetMatcherList())
			}
			for _, fm := range m.GetMatcherList().GetMatchers() {
				require.NotNil(t, fm.GetPredicate())
				require.NotNil(t, fm.GetOnMatch())
			}
		})
	}
}

// A predicate list must hold at least two entries, so a lone predicate has to be
// returned bare rather than wrapped.
func TestPredicateCombinatorsSkipSingletonLists(t *testing.T) {
	single := &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{},
	}

	require.Same(t, single, andPredicates([]*matcherv3.Matcher_MatcherList_Predicate{single}))
	require.Same(t, single, orPredicates([]*matcherv3.Matcher_MatcherList_Predicate{single}))

	require.Len(t, andPredicates([]*matcherv3.Matcher_MatcherList_Predicate{single, single}).GetAndMatcher().GetPredicate(), 2)
	require.Len(t, orPredicates([]*matcherv3.Matcher_MatcherList_Predicate{single, single}).GetOrMatcher().GetPredicate(), 2)
	require.Same(t, single, notPredicate(single).GetNotMatcher())
}
