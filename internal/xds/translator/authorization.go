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
	configv3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	networkinput "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/common_inputs/network/v3"
	ipmatcherv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/matching/input_matchers/ip/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/protocov"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&rbac{})
}

type rbac struct{}

var _ httpFilter = &rbac{}

// patchHCM builds and appends the RBAC Filter to the HTTP Connection Manager if
// applicable.
func (*rbac) patchHCM(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener,
) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsRBAC(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, f := range mgr.HttpFilters {
		if f.Name == string(egv1a1.EnvoyFilterRBAC) {
			return nil
		}
	}

	rbacFilter, err := buildHCMRBACFilter()
	if err != nil {
		return err
	}

	mgr.HttpFilters = append([]*hcmv3.HttpFilter{rbacFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMRBACFilter returns a RBAC filter from the provided IR listener.
func buildHCMRBACFilter() (*hcmv3.HttpFilter, error) {
	rbacProto := &rbacv3.RBAC{}
	rbacAny, err := protocov.ToAnyWithValidation(rbacProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: string(egv1a1.EnvoyFilterRBAC),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rbacAny,
		},
	}, nil
}

// listenerContainsRBAC returns true if the provided listener has RBAC
// policies attached to its routes.
func listenerContainsRBAC(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.Authorization != nil {
			return true
		}
	}

	return false
}

// patchRoute patches the provided route with the RBAC config if applicable.
func (*rbac) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.Authorization == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[string(egv1a1.EnvoyFilterRBAC)]; ok {
		// This should not happen since this is the only place where the RBAC
		// filter is added in a route.
		return fmt.Errorf("route already contains rbac config: %+v", route)
	}

	var (
		authorization = irRoute.Security.Authorization
		allowAction   *anypb.Any
		denyAction    *anypb.Any
		sourceIPInput *anypb.Any
		ipMatcher     *anypb.Any
		matcherList   []*matcherv3.Matcher_MatcherList_FieldMatcher
		err           error
	)

	allow := &rbacconfigv3.Action{
		Name:   "ALLOW",
		Action: rbacconfigv3.RBAC_ALLOW,
	}
	if allowAction, err = protocov.ToAnyWithValidation(allow); err != nil {
		return err
	}

	deny := &rbacconfigv3.Action{
		Name:   "DENY",
		Action: rbacconfigv3.RBAC_DENY,
	}
	if denyAction, err = protocov.ToAnyWithValidation(deny); err != nil {
		return err
	}

	// Build a list of matchers based on the rules.
	// The matchers will be evaluated in order, and the first one that matches
	// will be used to determine the action, the rest of the matchers will be
	// skipped.
	// If no matcher matches, the default action will be used.
	for _, rule := range authorization.Rules {
		// Build the IPMatcher based on the client CIDRs.
		ipRangeMatcher := &ipmatcherv3.Ip{
			StatPrefix: "client_ip",
		}

		for _, cidr := range rule.Principal.ClientCIDRs {
			ipRangeMatcher.CidrRanges = append(ipRangeMatcher.CidrRanges, &configv3.CidrRange{
				AddressPrefix: cidr.IP,
				PrefixLen: &wrapperspb.UInt32Value{
					Value: cidr.MaskLen,
				},
			})
		}

		if ipMatcher, err = protocov.ToAnyWithValidation(ipRangeMatcher); err != nil {
			return err
		}

		if sourceIPInput, err = protocov.ToAnyWithValidation(&networkinput.SourceIPInput{}); err != nil {
			return err
		}

		// Determine the action for the current rule.
		ruleAction := allowAction
		if rule.Action == egv1a1.AuthorizationActionDeny {
			ruleAction = denyAction
		}

		// Add the matcher generated with the current rule to the matcher list.
		matcherList = append(matcherList, &matcherv3.Matcher_MatcherList_FieldMatcher{
			Predicate: &matcherv3.Matcher_MatcherList_Predicate{
				MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
					SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
						Input: &cncfv3.TypedExtensionConfig{
							Name:        "client_ip",
							TypedConfig: sourceIPInput,
						},
						Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_CustomMatch{
							CustomMatch: &cncfv3.TypedExtensionConfig{
								Name:        "ip_matcher",
								TypedConfig: ipMatcher,
							},
						},
					},
				},
			},
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

	// Set the default action.
	defaultAction := denyAction
	if authorization.DefaultAction == egv1a1.AuthorizationActionAllow {
		defaultAction = allowAction
	}

	routeCfgProto := &rbacv3.RBACPerRoute{
		Rbac: &rbacv3.RBAC{
			Matcher: &matcherv3.Matcher{
				MatcherType: &matcherv3.Matcher_MatcherList_{
					MatcherList: &matcherv3.Matcher_MatcherList{
						Matchers: matcherList,
					},
				},
				// If no matcher matches, the default action will be used.
				OnNoMatch: &matcherv3.Matcher_OnMatch{
					OnMatch: &matcherv3.Matcher_OnMatch_Action{
						Action: &cncfv3.TypedExtensionConfig{
							Name:        "default",
							TypedConfig: defaultAction,
						},
					},
				},
			},
		},
	}

	// If there are no matchers, the default action will be used for all requests.
	// Setting the matcher type to nil since Proto validation will fail if the list
	// is empty.
	if len(matcherList) == 0 {
		routeCfgProto.Rbac.Matcher.MatcherType = nil
	}

	routeCfgAny, err := protocov.ToAnyWithValidation(routeCfgProto)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[string(egv1a1.EnvoyFilterRBAC)] = routeCfgAny

	return nil
}

func (c *rbac) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}
