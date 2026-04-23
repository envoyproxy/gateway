// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	cncfv3 "github.com/cncf/xds/go/xds/core/v3"
	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	dfpLoopbackRBACFilterName = "envoy.filters.http.rbac.dfp_loopback_deny"
)

func init() {
	registerHTTPFilter(&dynamicForwardProxy{})
}

type dynamicForwardProxy struct{}

var _ httpFilter = &dynamicForwardProxy{}

// patchHCM appends the loopback protection filter to the HTTP Connection Manager if applicable.
func (*dynamicForwardProxy) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	// Add the RBAC filter once (no-op by default); per-route config will enable it only for dynamic resolver routes.
	if listenerHasDynamicResolverRoute(irListener) {
		loopbackRBAC, err := buildDFPLoopbackRBAC()
		if err != nil {
			return err
		}
		if !hcmContainsFilter(mgr, loopbackRBAC.Name) {
			mgr.HttpFilters = append([]*hcmv3.HttpFilter{loopbackRBAC}, mgr.HttpFilters...)
		}
	}

	return nil
}

func (*dynamicForwardProxy) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}

	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	// Add per-route RBAC config to deny loopback addresses when DFP is used.
	if irRoute.IsDynamicResolverRoute() {
		hostFromLiteral := irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil && irRoute.URLRewrite.Host.Name != nil
		// We don't enforce check for host rewrite from literal since it's static and known at config time.
		// The loopback check is mainly to prevent dynamic hostnames that may resolve to loopback addresses.
		if !hostFromLiteral {
			rbacPerRouteCfg, err := buildDFPLoopbackRBACPerRoute(irRoute)
			if err != nil {
				return err
			}
			rbacAny, err := anypb.New(rbacPerRouteCfg)
			if err != nil {
				return err
			}
			if route.TypedPerFilterConfig == nil {
				route.TypedPerFilterConfig = make(map[string]*anypb.Any)
			}
			route.TypedPerFilterConfig[dfpLoopbackRBACFilterName] = rbacAny
		}
	}

	return nil
}

func (*dynamicForwardProxy) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// listenerHasDynamicResolverRoute checks if a given listener has any route with a dynamic resolver backend.
func listenerHasDynamicResolverRoute(listener *ir.HTTPListener) bool {
	for _, route := range listener.Routes {
		if route.IsDynamicResolverRoute() {
			return true
		}
	}
	return false
}

func buildDFPLoopbackRBAC() (*hcmv3.HttpFilter, error) {
	rbac := &rbacv3.RBAC{
		// No policies: acts as a placeholder; per-route config enables denial.
	}

	rbacAny, err := anypb.New(rbac)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: dfpLoopbackRBACFilterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rbacAny,
		},
	}, nil
}

func buildDFPLoopbackRBACPerRoute(irRoute *ir.HTTPRoute) (*rbacv3.RBACPerRoute, error) {
	loopbackPatterns := []string{
		`^127\.0\.0\.1(?::\d+)?$`,
		`^localhost(?::\d+)?$`,
		`^localhost\.localdomain(?::\d+)?$`,
		`^ip6-localhost(?::\d+)?$`,
		`^ip6-loopback(?::\d+)?$`,
		`^\[::1\](?::\d+)?$`,
		`^::1(?::\d+)?$`,
	}

	denyAction, err := proto.ToAnyWithValidation(&rbacconfigv3.Action{
		Name:   "DENY",
		Action: rbacconfigv3.RBAC_DENY,
	})
	if err != nil {
		return nil, err
	}

	allowAction, err := proto.ToAnyWithValidation(&rbacconfigv3.Action{
		Name: "ALLOW",
	})
	if err != nil {
		return nil, err
	}

	headerToCheck := ":authority"
	hostFromCustomHeader := irRoute.URLRewrite != nil && irRoute.URLRewrite.Host != nil && irRoute.URLRewrite.Host.Header != nil
	if hostFromCustomHeader {
		headerToCheck = *irRoute.URLRewrite.Host.Header
	}

	predicates := make([]*matcherv3.Matcher_MatcherList_Predicate, 0, len(loopbackPatterns))

	headerInput, err := proto.ToAnyWithValidation(&envoymatcherv3.HttpRequestHeaderMatchInput{
		HeaderName: headerToCheck,
	})
	if err != nil {
		return nil, err
	}

	for _, pattern := range loopbackPatterns {
		predicates = append(predicates, &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_{
				SinglePredicate: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate{
					Input: &cncfv3.TypedExtensionConfig{
						Name:        "http_header",
						TypedConfig: headerInput,
					},
					Matcher: &matcherv3.Matcher_MatcherList_Predicate_SinglePredicate_ValueMatch{
						ValueMatch: &matcherv3.StringMatcher{
							MatchPattern: &matcherv3.StringMatcher_SafeRegex{
								SafeRegex: &matcherv3.RegexMatcher{
									EngineType: &matcherv3.RegexMatcher_GoogleRe2{
										GoogleRe2: &matcherv3.RegexMatcher_GoogleRE2{},
									},
									Regex: pattern,
								},
							},
						},
					},
				},
			},
		})
	}

	var predicate *matcherv3.Matcher_MatcherList_Predicate
	if len(predicates) == 1 {
		predicate = predicates[0]
	} else {
		predicate = &matcherv3.Matcher_MatcherList_Predicate{
			MatchType: &matcherv3.Matcher_MatcherList_Predicate_OrMatcher{
				OrMatcher: &matcherv3.Matcher_MatcherList_Predicate_PredicateList{
					Predicate: predicates,
				},
			},
		}
	}

	return &rbacv3.RBACPerRoute{
		Rbac: &rbacv3.RBAC{
			Matcher: &matcherv3.Matcher{
				MatcherType: &matcherv3.Matcher_MatcherList_{
					MatcherList: &matcherv3.Matcher_MatcherList{
						Matchers: []*matcherv3.Matcher_MatcherList_FieldMatcher{
							{
								Predicate: predicate,
								OnMatch: &matcherv3.Matcher_OnMatch{
									OnMatch: &matcherv3.Matcher_OnMatch_Action{
										Action: &cncfv3.TypedExtensionConfig{
											Name:        "deny-loopback-host",
											TypedConfig: denyAction,
										},
									},
								},
							},
						},
					},
				},
				OnNoMatch: &matcherv3.Matcher_OnMatch{
					OnMatch: &matcherv3.Matcher_OnMatch_Action{
						Action: &cncfv3.TypedExtensionConfig{
							Name:        "default",
							TypedConfig: allowAction,
						},
					},
				},
			},
		},
	}, nil
}
