// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	v32 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	rbacv3 "github.com/envoyproxy/go-control-plane/envoy/config/rbac/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	frbacv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/rbac/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	aclFilter = "envoy.filters.http.acl"
)

func init() {
	registerHTTPFilter(&acl{})
}

type acl struct {
}

var _ httpFilter = &acl{}

// patchHCM builds and appends the acl Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an acl filter for each route that contains an acl config.
func (*acl) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsACL(route) {
			continue
		}

		allowFilter, err := buildHCMACLFilter(route, rbacv3.RBAC_ALLOW)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if allowFilter != nil {
			mgr.HttpFilters = append(mgr.HttpFilters, allowFilter)
		}

		denyFilter, err := buildHCMACLFilter(route, rbacv3.RBAC_DENY)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if denyFilter != nil {
			mgr.HttpFilters = append(mgr.HttpFilters, denyFilter)
		}
	}

	return errs
}

// routeContainsACL returns true if acl exists for the provided route.
func routeContainsACL(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.ACL != nil {
		return true
	}

	return false
}

// buildHCMACLFilter returns an acl HTTP filter from the provided IR HTTPRoute.
func buildHCMACLFilter(route *ir.HTTPRoute, rbacType rbacv3.RBAC_Action) (*hcmv3.HttpFilter, error) {
	aclProto, empty := aclConfig(route.ACL, rbacType)
	// if we do not have any rules for the action, we do not need to add the filter
	if empty {
		return nil, nil
	}
	if err := aclProto.ValidateAll(); err != nil {
		return nil, err
	}

	aclAny, err := anypb.New(aclProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: aclFilterName(route, rbacType),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: aclAny,
		},
	}, nil
}

func aclFilterName(route *ir.HTTPRoute, rbacType rbacv3.RBAC_Action) string {
	return perRouteFilterName(aclFilter, fmt.Sprintf("%s_%s", route.Name, rbacType))
}

func aclConfig(acl *ir.ACL, rbacType rbacv3.RBAC_Action) (*frbacv3.RBAC, bool) {
	config := &frbacv3.RBAC{
		Rules: &rbacv3.RBAC{
			Action: rbacType,
		},
	}

	var principals []*rbacv3.Principal
	var policyName string
	if rbacType == rbacv3.RBAC_ALLOW {
		policyName = "allow-acl"
		for _, allow := range acl.Allow {
			principals = append(principals, &rbacv3.Principal{
				Identifier: &rbacv3.Principal_SourceIp{
					SourceIp: &v32.CidrRange{
						AddressPrefix: allow.Prefix,
						PrefixLen: &wrapperspb.UInt32Value{
							Value: allow.Length,
						},
					},
				},
			})
		}
	} else if rbacType == rbacv3.RBAC_DENY {
		policyName = "deny-acl"
		for _, deny := range acl.Deny {
			principals = append(principals, &rbacv3.Principal{
				Identifier: &rbacv3.Principal_SourceIp{
					SourceIp: &v32.CidrRange{
						AddressPrefix: deny.Prefix,
						PrefixLen: &wrapperspb.UInt32Value{
							Value: deny.Length,
						},
					},
				},
			})
		}
	}

	config.Rules.Policies = make(map[string]*rbacv3.Policy, 1)
	config.Rules.Policies[policyName] = &rbacv3.Policy{
		Permissions: []*rbacv3.Permission{
			{
				Rule: &rbacv3.Permission_Any{
					Any: true,
				},
			},
		},
		Principals: principals,
	}

	return config, len(config.Rules.Policies[policyName].Principals) == 0
}

// patchResources patches the cluster resources for the acl rules.
func (*acl) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the acl config if applicable.
// Note: this method enables the corresponding acl filter for the provided route.
func (*acl) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.ACL == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	for _, rule := range []rbacv3.RBAC_Action{rbacv3.RBAC_ALLOW, rbacv3.RBAC_DENY} {
		filterName := aclFilterName(irRoute, rule)
		conf, empty := aclConfig(irRoute.ACL, rule)
		// if we do not have any rules for the action, we do not need to add the filter
		if empty {
			continue
		}
		routeCfgProto := &frbacv3.RBACPerRoute{
			Rbac: conf,
		}

		routeCfgAny, err := anypb.New(routeCfgProto)
		if err != nil {
			return err
		}

		if filterCfg == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		route.TypedPerFilterConfig[filterName] = routeCfgAny
	}

	return nil
}

// patchRouteCfg patches the provided route configuration with the acl filter
// if applicable.
func (*acl) patchRouteConfig(routeCfg *routev3.RouteConfiguration, irListener *ir.HTTPListener) error {
	return nil
}
