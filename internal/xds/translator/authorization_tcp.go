// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	networkrbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/rbac/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
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

	return toNetworkFilter(wellknown.RBAC, rbacCfg)
}

func buildTCPRBACConfig(statPrefix string, authorization *ir.Authorization) (*networkrbacv3.RBAC, error) {
	allowAction, denyAction, err := buildTCPRBACActions()
	if err != nil {
		return nil, err
	}

	matchers := make([]*matcherv3.Matcher_MatcherList_FieldMatcher, 0, len(authorization.Rules))
	for idx, rule := range authorization.Rules {
		predicate, err := buildTCPPrincipalPredicate(rule.Principal)
		if err != nil {
			return nil, err
		}

		ruleAction := allowAction
		if rule.Action == egv1a1.AuthorizationActionDeny {
			ruleAction = denyAction
		}

		actionName := rule.Name
		if actionName == "" {
			actionName = fmt.Sprintf("tcp-authz-rule-%d", idx)
		}

		matchers = append(matchers, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: predicate,
			OnMatch: &matcherv3.Matcher_OnMatch{
				OnMatch: &matcherv3.Matcher_OnMatch_Action{
					Action: &cncfv3.TypedExtensionConfig{
						Name:        actionName,
						TypedConfig: ruleAction,
					},
				},
			},
		})
	}

	defaultAction := denyAction
	if authorization.DefaultAction == egv1a1.AuthorizationActionAllow {
		defaultAction = allowAction
	}

	matcher := &matcherv3.Matcher{
		OnNoMatch: &matcherv3.Matcher_OnMatch{
			OnMatch: &matcherv3.Matcher_OnMatch_Action{
				Action: &cncfv3.TypedExtensionConfig{
					Name:        "default",
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
		StatPrefix: buildTCPRBACStatPrefix(statPrefix),
		Matcher:    matcher,
	}, nil
}

func buildTCPRBACActions() (*anypb.Any, *anypb.Any, error) {
	allowAction, err := proto.ToAnyWithValidation(&rbacconfigv3.Action{
		Name:   "ALLOW",
		Action: rbacconfigv3.RBAC_ALLOW,
	})
	if err != nil {
		return nil, nil, err
	}

	denyAction, err := proto.ToAnyWithValidation(&rbacconfigv3.Action{
		Name:   "DENY",
		Action: rbacconfigv3.RBAC_DENY,
	})
	if err != nil {
		return nil, nil, err
	}

	return allowAction, denyAction, nil
}

func buildTCPPrincipalPredicate(principal ir.Principal) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	// only build predicate for CIDR
	if len(principal.ClientCIDRs) == 0 {
		return nil, nil
	}
	return buildIPPredicate(principal.ClientCIDRs)
}

func buildTCPRBACStatPrefix(statPrefix string) string {
	if statPrefix == "" {
		return "tcp_authz"
	}
	return fmt.Sprintf("%s_authz", statPrefix)
}
