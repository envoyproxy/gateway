// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// buildUDPProxyMatcher builds the matcher that udp_proxy uses to pick a route,
// folding client-IP authorization into it.
//
// A UDP listener has no network filter chain, so the network RBAC filter used for
// TCPRoutes cannot be reused here. Instead the authorization decision rides along
// with route selection: a datagram whose source IP matches is routed to the
// cluster, and a datagram that matches nothing is dropped by udp_proxy, which
// counts it as downstream_sess_no_route.
//
// Since a denial can only be expressed as the absence of a match, the ordered
// Allow/Deny rules are compiled into allow-only predicates. A Deny rule never
// becomes an entry of its own; it subtracts from the Allow rules that follow it
// and from a permissive default action.
func buildUDPProxyMatcher(routeAction *cncfv3.TypedExtensionConfig, authorization *ir.Authorization) (*matcherv3.Matcher, error) {
	onRoute := &matcherv3.Matcher_OnMatch{
		OnMatch: &matcherv3.Matcher_OnMatch_Action{Action: routeAction},
	}

	if authorization == nil {
		return &matcherv3.Matcher{OnNoMatch: onRoute}, nil
	}

	var (
		matchers []*matcherv3.Matcher_MatcherList_FieldMatcher
		// Predicates of the Deny rules seen so far, in rule order.
		denied []*matcherv3.Matcher_MatcherList_Predicate
	)

	for _, rule := range authorization.Rules {
		// Only client CIDRs are enforceable on the UDP path. The Gateway API layer
		// rejects every other principal for L4 targets, so anything else here is a
		// rule that cannot match rather than one that matches everything.
		if len(rule.Principal.ClientCIDRs) == 0 {
			continue
		}

		predicate, err := buildIPPredicate(rule.Principal.ClientCIDRs)
		if err != nil {
			return nil, err
		}

		if rule.Action == egv1a1.AuthorizationActionDeny {
			denied = append(denied, predicate)
			continue
		}

		// Rules are first-match-wins, so an Allow rule only covers the sources that
		// no preceding Deny rule already claimed. Preceding Allow rules need no such
		// exclusion: had one matched, the datagram would already have been routed.
		conjuncts := []*matcherv3.Matcher_MatcherList_Predicate{predicate}
		for _, d := range denied {
			conjuncts = append(conjuncts, notPredicate(d))
		}

		matchers = append(matchers, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: andPredicates(conjuncts),
			OnMatch:   onRoute,
		})
	}

	if authorization.DefaultAction == egv1a1.AuthorizationActionAllow {
		// Nothing is denied, so every datagram is routed either way and the Allow
		// rules are redundant. Emit the same matcher as an unauthorized listener.
		if len(denied) == 0 {
			return &matcherv3.Matcher{OnNoMatch: onRoute}, nil
		}

		// Route whatever no Deny rule claimed. This goes last so the Allow rules
		// above keep their precedence.
		matchers = append(matchers, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: notPredicate(orPredicates(denied)),
			OnMatch:   onRoute,
		})
	}

	matcher := &matcherv3.Matcher{}
	// An empty matcher list fails proto validation, so the matcher type is left
	// unset when nothing can be allowed.
	if len(matchers) > 0 {
		matcher.MatcherType = &matcherv3.Matcher_MatcherList_{
			MatcherList: &matcherv3.Matcher_MatcherList{Matchers: matchers},
		}
	}
	// on_no_match is deliberately left unset: that absence is what makes udp_proxy
	// drop a datagram no Allow rule accounted for.
	return matcher, nil
}

// notPredicate negates a predicate.
func notPredicate(predicate *matcherv3.Matcher_MatcherList_Predicate) *matcherv3.Matcher_MatcherList_Predicate {
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_NotMatcher{
			NotMatcher: predicate,
		},
	}
}

// andPredicates conjoins predicates. A predicate list must hold at least two
// entries, so a lone predicate is returned as-is.
func andPredicates(predicates []*matcherv3.Matcher_MatcherList_Predicate) *matcherv3.Matcher_MatcherList_Predicate {
	if len(predicates) == 1 {
		return predicates[0]
	}
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_AndMatcher{
			AndMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: predicates,
			},
		},
	}
}

// orPredicates disjoins predicates. A predicate list must hold at least two
// entries, so a lone predicate is returned as-is.
func orPredicates(predicates []*matcherv3.Matcher_MatcherList_Predicate) *matcherv3.Matcher_MatcherList_Predicate {
	if len(predicates) == 1 {
		return predicates[0]
	}
	return &matcherv3.Matcher_MatcherList_Predicate{
		MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
			OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
				Predicate: predicates,
			},
		},
	}
}
