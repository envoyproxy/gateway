// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	basicauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/basic_auth/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&basicAuth{})
}

type basicAuth struct{}

var _ httpFilter = &basicAuth{}

// patchHCM updates the HTTPConnectionManager with a Basic Auth HTTP filter for routes requiring authentication.
// It scans through all routes in the provided HTTPListener, and if a route has a BasicAuth configuration,
// it checks for the presence of a corresponding filter in the manager. If the filter is not already present,
// it generates and appends a new Basic Auth filter.
// The function returns an error if either the HTTPConnectionManager or HTTPListener is nil, or if an error occurs
// during the filter creation process.
func (*basicAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error

	for _, route := range irListener.Routes {
		if route.Security == nil || route.Security.BasicAuth == nil {
			continue
		}

		// Only generates one Basic Auth Envoy filter for each unique name.
		// For example, if there are two routes under the same gateway with the
		// same BasicAuth config, only one BasicAuth filter will be generated.
		if hcmContainsFilter(mgr, basicAuthFilterName(route.Security.BasicAuth)) {
			continue
		}

		filter, err := buildHCMBasicAuthFilter(route.Security.BasicAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// patchHCM builds and appends the basic_auth Filter to the HTTP Connection Manager
// if applicable, and it does not already exist.

// buildHCMBasicAuthFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMBasicAuthFilter(basicAuth *ir.BasicAuth) (*hcmv3.HttpFilter, error) {
	var (
		basicAuthProto *basicauthv3.BasicAuth
		basicAuthAny   *anypb.Any
		err            error
	)

	basicAuthProto = &basicauthv3.BasicAuth{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: basicAuth.Users,
			},
		},
	}
	// Set the ForwardUsernameHeader field if it is specified.
	if basicAuth.ForwardUsernameHeader != nil && *basicAuth.ForwardUsernameHeader != "" {
		basicAuthProto.ForwardUsernameHeader = *basicAuth.ForwardUsernameHeader
	}

	if basicAuthAny, err = proto.ToAnyWithValidation(basicAuthProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: basicAuthFilterName(basicAuth),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: basicAuthAny,
		},
		Disabled: true,
	}, nil
}

func basicAuthFilterName(basicAuth *ir.BasicAuth) string {
	return perRouteFilterName(egv1a1.EnvoyFilterBasicAuth, basicAuth.Name)
}

func (*basicAuth) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the basicAuth config if applicable.
// Note: this method overwrites the HCM level filter config with the per route filter config.
func (*basicAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.BasicAuth == nil {
		return nil
	}

	var (
		perFilterCfg map[string]*anypb.Any
		basicAuthAny *anypb.Any
		err          error
	)
	filterName := basicAuthFilterName(irRoute.Security.BasicAuth)
	perFilterCfg = route.GetTypedPerFilterConfig()
	if _, ok := perFilterCfg[filterName]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf("route already contains filter config: %s, %+v",
			egv1a1.EnvoyFilterBasicAuth.String(), route)
	}

	// Overwrite the HCM level filter config with the per route filter config.
	basicAuthProto := basicAuthPerRouteConfig(irRoute.Security.BasicAuth)
	if basicAuthAny, err = proto.ToAnyWithValidation(basicAuthProto); err != nil {
		return err
	}

	if perFilterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[filterName] = basicAuthAny

	return nil
}

func basicAuthPerRouteConfig(basicAuth *ir.BasicAuth) *basicauthv3.BasicAuthPerRoute {
	return &basicauthv3.BasicAuthPerRoute{
		Users: &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineBytes{
				InlineBytes: basicAuth.Users,
			},
		},
	}
}
