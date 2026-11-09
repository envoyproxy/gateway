// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"slices"

	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// authenticationGateName is an RBAC filter that rejects a request on which
// neither JWT nor BasicAuth recorded a verified identity. It shares the RBAC
// prefix, so orderAuthenticationGate positions it after sorting.
var authenticationGateName = perRouteFilterName(egv1a1.EnvoyFilterRBAC, "authentication-any")

func init() {
	registerHTTPFilter(&authenticationGate{})
}

type authenticationGate struct{}

var _ httpFilter = &authenticationGate{}

func (*authenticationGate) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if hcmContainsFilter(mgr, authenticationGateName) ||
		!slices.ContainsFunc(irListener.Routes, func(route *ir.HTTPRoute) bool { return authenticationAny(route.Security) }) {
		return nil
	}
	gateAny, err := anypb.New(&rbacv3.RBAC{})
	if err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, &hcmv3.HttpFilter{
		Name:       authenticationGateName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: gateAny},
		Disabled:   true,
	})
	return nil
}

func authenticationAny(security *ir.SecurityFeatures) bool {
	return security != nil && security.AuthenticationAny
}

func (*authenticationGate) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	security := irRoute.Security
	if !authenticationAny(security) {
		return nil
	}
	// gatewayapi rejects Any without both methods and closes the route, but the
	// route is still patched; with no principal the gate denies every request.
	policies := map[string]*rbacconfigv3.Policy{}
	if principals := authenticationPrincipals(security); len(principals) > 0 {
		policies["authenticated"] = &rbacconfigv3.Policy{
			Permissions: []*rbacconfigv3.Permission{{Rule: &rbacconfigv3.Permission_Any{Any: true}}},
			Principals:  principals,
		}
	}

	return enableFilterOnRoute(route, authenticationGateName, &rbacv3.RBACPerRoute{
		Rbac: &rbacv3.RBAC{
			Rules: &rbacconfigv3.RBAC{Action: rbacconfigv3.RBAC_ALLOW, Policies: policies},
		},
	})
}

func authenticationPrincipals(security *ir.SecurityFeatures) []*rbacconfigv3.Principal {
	var principals []*rbacconfigv3.Principal
	if security.BasicAuth != nil {
		principals = append(principals, metadataPresentPrincipal(basicAuthFilterName(security.BasicAuth), "username"))
	}
	if security.JWT != nil {
		for _, provider := range security.JWT.Providers {
			// jwt_authn writes each verified payload under its provider name (payload_in_metadata).
			principals = append(principals, metadataPresentPrincipal(egv1a1.EnvoyFilterJWTAuthn.String(), provider.Name))
		}
	}
	return principals
}

func metadataPresentPrincipal(namespace, key string) *rbacconfigv3.Principal {
	return &rbacconfigv3.Principal{
		Identifier: &rbacconfigv3.Principal_Metadata{
			Metadata: &matcherv3.MetadataMatcher{
				Filter: namespace,
				Path:   []*matcherv3.MetadataMatcher_PathSegment{{Segment: &matcherv3.MetadataMatcher_PathSegment_Key{Key: key}}},
				Value:  &matcherv3.ValueMatcher{MatchPattern: &matcherv3.ValueMatcher_PresentMatch{PresentMatch: true}},
			},
		},
	}
}

func (*authenticationGate) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// orderAuthenticationGate moves the gate directly after the last JWT or BasicAuth
// filter, whatever order EnvoyProxy.spec.filterOrder produced, so the gate always
// sees the metadata of both methods.
func orderAuthenticationGate(filters []*hcmv3.HttpFilter) []*hcmv3.HttpFilter {
	var gate *hcmv3.HttpFilter
	ordered := make([]*hcmv3.HttpFilter, 0, len(filters))
	lastAuthentication := -1
	for _, filter := range filters {
		if filter.Name == authenticationGateName {
			gate = filter
			continue
		}
		ordered = append(ordered, filter)
		if isFilterType(filter, egv1a1.EnvoyFilterJWTAuthn) || isFilterType(filter, egv1a1.EnvoyFilterBasicAuth) {
			lastAuthentication = len(ordered) - 1
		}
	}
	if gate == nil {
		return filters
	}
	return append(ordered[:lastAuthentication+1], append([]*hcmv3.HttpFilter{gate}, ordered[lastAuthentication+1:]...)...)
}
