// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	networkrbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildTCPRBACFilter(statPrefix string, authorization *ir.Authorization) (*listenerv3.Filter, error) {
	if authorization == nil {
		return nil, nil
	}

	rbacCfg, err := buildTCPRBACConfig(statPrefix, authorization)
	if err != nil {
		return nil, err
	}
	if rbacCfg == nil {
		return nil, nil
	}

	return toNetworkFilter(wellknown.RoleBasedAccessControl, rbacCfg)
}

func buildTCPRBACConfig(statPrefix string, authorization *ir.Authorization) (*networkrbacv3.RBAC, error) {
	matchers := make([]*matcherv3.Matcher_MatcherList_FieldMatcher, 0, len(authorization.Rules))
	for _, rule := range authorization.Rules {
		predicate, err := buildTCPPrincipalPredicate(&rule.Principal)
		if err != nil {
			return nil, err
		}

		ruleAction, err := buildRBACAction(rule.Name, ruleAllows(rule))
		if err != nil {
			return nil, err
		}

		matchers = append(matchers, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: predicate,
			OnMatch: &matcherv3.Matcher_OnMatch{
				OnMatch: &matcherv3.Matcher_OnMatch_Action{
					Action: &cncfv3.TypedExtensionConfig{
						Name:        rule.Name,
						TypedConfig: ruleAction,
					},
				},
			},
		})
	}

	defaultAction, err := buildRBACAction(rbacDefaultActionName, defaultAllows(authorization))
	if err != nil {
		return nil, err
	}

	matcher := &matcherv3.Matcher{
		OnNoMatch: &matcherv3.Matcher_OnMatch{
			OnMatch: &matcherv3.Matcher_OnMatch_Action{
				Action: &cncfv3.TypedExtensionConfig{
					Name:        rbacDefaultActionName,
					TypedConfig: defaultAction,
				},
			},
		},
	}
	if len(matchers) > 0 {
		matcher.MatcherType = &matcherv3.Matcher_MatcherList_{
			MatcherList: &matcherv3.Matcher_MatcherList{
				Matchers: matchers,
			},
		}
	} else {
		matcher.MatcherType = nil
	}

	return &networkrbacv3.RBAC{
		StatPrefix: statPrefix,
		Matcher:    matcher,
	}, nil
}

func buildTCPPrincipalPredicate(principal *ir.Principal) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	// only build predicate for CIDR
	if len(principal.ClientCIDRs) == 0 {
		return nil, nil
	}
	return buildIPPredicate(principal.ClientCIDRs)
}
