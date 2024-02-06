// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net/url"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	extAuthFilter = "envoy.filters.http.ext_authz"
)

func init() {
	registerHTTPFilter(&extAuth{})
}

type extAuth struct {
}

var _ httpFilter = &extAuth{}

// patchHCM builds and appends the ext_authz Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an ext_authz filter for each route that contains an ExtAuthz config.
// TODO: zhaohuabing avoid duplicated HTTP filters
func (*extAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsExtAuth(route) {
			continue
		}

		filter, err := buildHCMExtAuthFilter(route)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// buildHCMExtAuthFilter returns an ext_authz HTTP filter from the provided IR HTTPRoute.
func buildHCMExtAuthFilter(route *ir.HTTPRoute) (*hcmv3.HttpFilter, error) {
	extAuthProto := extAuthConfig(route.ExtAuth)
	if err := extAuthProto.ValidateAll(); err != nil {
		return nil, err
	}

	extAuthAny, err := anypb.New(extAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: extAuthFilterName(route),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extAuthAny,
		},
	}, nil
}

func extAuthFilterName(route *ir.HTTPRoute) string {
	return perRouteFilterName(extAuthFilter, route.Name)
}

func extAuthConfig(extAuth *ir.ExtAuth) *extauthv3.ExtAuthz {
	config := &extauthv3.ExtAuthz{}

	var headersToExtAuth []*matcherv3.StringMatcher
	for _, header := range extAuth.HeadersToExtAuth {
		headersToExtAuth = append(headersToExtAuth, &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: header,
			},
		})
	}

	if len(headersToExtAuth) > 0 {
		config.AllowedHeaders = &matcherv3.ListStringMatcher{
			Patterns: headersToExtAuth,
		}
	}

	if extAuth.HTTP != nil {
		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: httpService(extAuth.HTTP),
		}
	} else if extAuth.GRPC != nil {
		config.Services = &extauthv3.ExtAuthz_GrpcService{
			GrpcService: &corev3.GrpcService{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: grpcService(extAuth.GRPC),
				},
				Timeout: &duration.Duration{
					Seconds: defaultExtServiceRequestTimeout,
				},
			},
		}
	}

	return config
}

func httpService(http *ir.HTTPExtAuthService) *extauthv3.HttpService {
	var (
		uri              string
		headersToBackend []*matcherv3.StringMatcher
	)

	u := url.URL{
		// scheme should be decided by the TLS setting, but we don't have that info now.
		// It's safe to set it to http because the ext auth filter doesn't use the
		// uri to make the request. It only uses the cluster.
		Scheme: "http",
		Host:   http.Authority,
		Path:   http.Path,
	}
	uri = u.String()

	for _, header := range http.HeadersToBackend {
		headersToBackend = append(headersToBackend, &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: header,
			},
		})
	}

	return &extauthv3.HttpService{
		ServerUri: &corev3.HttpUri{
			Uri: uri,
			HttpUpstreamType: &corev3.HttpUri_Cluster{
				Cluster: http.Destination.Name,
			},
			Timeout: &duration.Duration{
				Seconds: defaultExtServiceRequestTimeout,
			},
		},
		AuthorizationResponse: &extauthv3.AuthorizationResponse{
			AllowedUpstreamHeaders: &matcherv3.ListStringMatcher{
				Patterns: headersToBackend,
			},
		},
	}
}

func grpcService(grpc *ir.GRPCExtAuthService) *corev3.GrpcService_EnvoyGrpc {
	return &corev3.GrpcService_EnvoyGrpc{
		ClusterName: grpc.Destination.Name,
		Authority:   grpc.Authority,
	}
}

// routeContainsExtAuth returns true if ExtAuth exists for the provided route.
func routeContainsExtAuth(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.ExtAuth != nil {
		return true
	}

	return false
}

// patchResources patches the cluster resources for the external auth services.
func (*extAuth) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtAuth(route) {
			continue
		}
		if route.ExtAuth.HTTP != nil {
			if err := createExtServiceXDSCluster(
				&route.ExtAuth.HTTP.Destination, tCtx); err != nil && !errors.Is(
				err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		} else {
			if err := createExtServiceXDSCluster(
				&route.ExtAuth.GRPC.Destination, tCtx); err != nil && !errors.Is(
				err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

func createExtServiceXDSCluster(rd *ir.RouteDestination, tCtx *types.ResourceVersionTable) error {
	// Get the address type from the first setting.
	// This is safe because no mixed address types in the settings.
	addrTypeState := rd.Settings[0].AddressType

	var endpointType EndpointType
	if addrTypeState != nil && *addrTypeState == ir.FQDN {
		endpointType = EndpointTypeDNS
	} else {
		endpointType = EndpointTypeStatic
	}
	if err := addXdsCluster(tCtx, &xdsClusterArgs{
		name:         rd.Name,
		settings:     rd.Settings,
		tSocket:      nil,
		endpointType: endpointType,
	}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
		return err
	}
	return nil
}

// patchRouteCfg patches the provided route configuration with the extAuth filter
// if applicable.
// Note: this method disables all the extAuth filters by default. The filter will
// be enabled per-route in the typePerFilterConfig of the route.
func (*extAuth) patchRouteConfig(routeCfg *routev3.RouteConfiguration, irListener *ir.HTTPListener) error {
	if routeCfg == nil {
		return errors.New("route configuration is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error
	for _, route := range irListener.Routes {
		if !routeContainsExtAuth(route) {
			continue
		}

		filterName := extAuthFilterName(route)
		filterCfg := routeCfg.TypedPerFilterConfig

		if _, ok := filterCfg[filterName]; ok {
			// This should not happen since this is the only place where the extAuth
			// filter is added in a route.
			errs = errors.Join(errs, fmt.Errorf(
				"route config already contains extAuth config: %+v", route))
			continue
		}

		// Disable all the filters by default. The filter will be enabled
		// per-route in the typePerFilterConfig of the route.
		routeCfgAny, err := anypb.New(&routev3.FilterConfig{Disabled: true})
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if filterCfg == nil {
			routeCfg.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		routeCfg.TypedPerFilterConfig[filterName] = routeCfgAny
	}
	return errs
}

// patchRoute patches the provided route with the extAuth config if applicable.
// Note: this method enables the corresponding extAuth filter for the provided route.
func (*extAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.ExtAuth == nil {
		return nil
	}
	if err := enableFilterOnRoute(extAuthFilter, route, irRoute); err != nil {
		return err
	}
	return nil
}
