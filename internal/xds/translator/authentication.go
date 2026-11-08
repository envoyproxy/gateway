// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	matcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	rbacconfigv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const authenticationGateName = "envoy.filters.http.rbac/authentication"

type authenticationGate struct{}

func init() { registerHTTPFilter(&authenticationGate{}) }

func (*authenticationGate) patchHCM(manager *hcmv3.HttpConnectionManager, listener *ir.HTTPListener) error {
	if manager == nil || listener == nil {
		return errors.New("authentication gate requires a listener and HCM")
	}
	for _, route := range listener.Routes {
		if needsAuthenticationGate(route.Security) && !hcmContainsFilter(manager, authenticationGateName) {
			configuration, err := anypb.New(&rbacv3.RBAC{})
			if err != nil {
				return err
			}
			manager.HttpFilters = append(manager.HttpFilters, &hcmv3.HttpFilter{Name: authenticationGateName, ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: configuration}})
		}
	}
	return nil
}

func needsAuthenticationGate(security *ir.SecurityFeatures) bool {
	return security != nil && (security.AuthenticationAny || (security.ExtAuth != nil && len(security.ExtAuth.AuthenticationMetadataNamespaces) > 0))
}

func (*authenticationGate) patchRoute(route *routev3.Route, input *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil || input == nil {
		return errors.New("authentication gate requires a route")
	}
	if !needsAuthenticationGate(input.Security) {
		return nil
	}
	principals := []*rbacconfigv3.Principal{}
	security := input.Security
	if security.BasicAuth != nil {
		principals = append(principals, metadataPrincipal(basicAuthFilterName(security.BasicAuth), []string{"username"}))
	}
	if security.JWT != nil {
		for _, provider := range security.JWT.Providers {
			principals = append(principals, metadataPrincipal(egv1a1.EnvoyFilterJWTAuthn.String(), []string{provider.Name}))
		}
	}
	if len(principals) == 0 {
		return errors.New("authentication gate requires an authentication method")
	}
	configuration, err := anypb.New(&rbacv3.RBACPerRoute{Rbac: &rbacv3.RBAC{Rules: &rbacconfigv3.RBAC{Action: rbacconfigv3.RBAC_ALLOW, Policies: map[string]*rbacconfigv3.Policy{"authenticated": {Permissions: []*rbacconfigv3.Permission{{Rule: &rbacconfigv3.Permission_Any{Any: true}}}, Principals: principals}}}}})
	if err != nil {
		return err
	}
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = map[string]*anypb.Any{}
	}
	route.TypedPerFilterConfig[authenticationGateName] = configuration
	return nil
}

func metadataPrincipal(namespace string, path []string) *rbacconfigv3.Principal {
	segments := make([]*envoymatcherv3.MetadataMatcher_PathSegment, 0, len(path))
	for _, key := range path {
		segments = append(segments, &envoymatcherv3.MetadataMatcher_PathSegment{Segment: &envoymatcherv3.MetadataMatcher_PathSegment_Key{Key: key}})
	}
	return &rbacconfigv3.Principal{Identifier: &rbacconfigv3.Principal_Metadata{Metadata: &envoymatcherv3.MetadataMatcher{Filter: namespace, Path: segments, Value: &envoymatcherv3.ValueMatcher{MatchPattern: &envoymatcherv3.ValueMatcher_PresentMatch{PresentMatch: true}}}}}
}

func (*authenticationGate) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func buildBasicAuthPredicate(namespace string, users []string) (*matcherv3.Matcher_MatcherList_Predicate, error) {
	if namespace == "" || len(users) == 0 {
		return nil, errors.New("BasicAuth principal requires a filter and usernames")
	}
	quoted := make([]string, len(users))
	for i, user := range users {
		quoted[i] = strconv.Quote(user)
	}
	expression := fmt.Sprintf("metadata.filter_metadata[%s]['username'] in [%s]", strconv.Quote(namespace), strings.Join(quoted, ","))
	return buildCELPredicate(expression)
}

func validateAuthenticationFilterOrder(manager *hcmv3.HttpConnectionManager, listener *ir.HTTPListener) error {
	indexes := map[string]int{}
	for index, filter := range manager.HttpFilters {
		indexes[filter.Name] = index
	}
	gate, hasGate := indexes[authenticationGateName]
	for _, route := range listener.Routes {
		security := route.Security
		if security == nil {
			continue
		}
		lastAuthentication := -1
		if security.JWT != nil {
			lastAuthentication = indexes[egv1a1.EnvoyFilterJWTAuthn.String()]
		}
		if security.BasicAuth != nil {
			if index := indexes[basicAuthFilterName(security.BasicAuth)]; index > lastAuthentication {
				lastAuthentication = index
			}
		}
		if needsAuthenticationGate(security) && (!hasGate || gate <= lastAuthentication) {
			return errors.New("authentication gate must run after JWT and BasicAuth")
		}
		if needsAuthenticationGate(security) && security.Authorization != nil && indexes[egv1a1.EnvoyFilterRBAC.String()] <= gate {
			return errors.New("local authorization must run after the authentication gate")
		}
		if security.Authorization != nil {
			for _, rule := range security.Authorization.Rules {
				if rule.Principal.BasicAuth != nil && indexes[egv1a1.EnvoyFilterRBAC.String()] <= lastAuthentication {
					return errors.New("BasicAuth authorization must run after authentication")
				}
			}
		}
		if security.ExtAuth != nil && len(security.ExtAuth.AuthenticationMetadataNamespaces) > 0 {
			external := indexes[extAuthFilterName(security.ExtAuth)]
			if !hasGate || external <= gate {
				return errors.New("extAuth authentication metadata requires filterOrder after JWT, BasicAuth and the authentication gate")
			}
			if security.Authorization != nil && external >= indexes[egv1a1.EnvoyFilterRBAC.String()] {
				return errors.New("extAuth authentication metadata must precede local authorization")
			}
		}
	}
	return nil
}

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
	ordered = append(ordered, nil)
	copy(ordered[lastAuthentication+2:], ordered[lastAuthentication+1:])
	ordered[lastAuthentication+1] = gate
	return ordered
}
