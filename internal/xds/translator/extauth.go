// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
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
// The filter is disabled by default. It is enabled on the route level.
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

		// Only generates one OAuth2 Envoy filter for each unique name.
		// For example, if there are two routes under the same gateway with the
		// same OIDC config, only one OAuth2 filter will be generated.
		if hcmContainsFilter(mgr, extAuthFilterName(route.ExtAuth)) {
			continue
		}

		filter, err := buildHCMExtAuthFilter(route.ExtAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMExtAuthFilter returns an ext_authz HTTP filter from the provided IR HTTPRoute.
func buildHCMExtAuthFilter(extAuth *ir.ExtAuth) (*hcmv3.HttpFilter, error) {
	extAuthProto := extAuthConfig(extAuth)
	if err := extAuthProto.ValidateAll(); err != nil {
		return nil, err
	}

	extAuthAny, err := anypb.New(extAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     extAuthFilterName(extAuth),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extAuthAny,
		},
	}, nil
}

func extAuthFilterName(extAuth *ir.ExtAuth) string {
	return perRouteFilterName(extAuthFilter, extAuth.Name)
}

func extAuthConfig(extAuth *ir.ExtAuth) *extauthv3.ExtAuthz {
	config := &extauthv3.ExtAuthz{
		TransportApiVersion: corev3.ApiVersion_V3,
	}

	if extAuth.FailOpen != nil {
		config.FailureModeAllow = *extAuth.FailOpen
	}

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
		service          *extauthv3.HttpService
	)

	service = &extauthv3.HttpService{
		PathPrefix: http.Path,
	}

	u := url.URL{
		// scheme should be decided by the TLS setting, but we don't have that info now.
		// It's safe to set it to http because the ext auth filter doesn't use the
		// uri to make the request. It only uses the cluster.
		Scheme: "http",
		Host:   http.Authority,
		Path:   http.Path,
	}
	uri = u.String()

	service.ServerUri = &corev3.HttpUri{
		Uri: uri,
		HttpUpstreamType: &corev3.HttpUri_Cluster{
			Cluster: http.Destination.Name,
		},
		Timeout: &duration.Duration{
			Seconds: defaultExtServiceRequestTimeout,
		},
	}

	for _, header := range http.HeadersToBackend {
		headersToBackend = append(headersToBackend, &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: header,
			},
		})
	}

	if len(headersToBackend) > 0 {
		service.AuthorizationResponse = &extauthv3.AuthorizationResponse{
			AllowedUpstreamHeaders: &matcherv3.ListStringMatcher{
				Patterns: headersToBackend,
			},
		}
	}

	return service
}

func grpcService(grpc *ir.GRPCExtAuthService) *corev3.GrpcService_EnvoyGrpc {
	return &corev3.GrpcService_EnvoyGrpc{
		ClusterName: grpc.Destination.Name,
		Authority:   grpc.Authority,
	}
}

// routeContainsExtAuth returns true if ExtAuth exists for the provided route.
func routeContainsExtAuth(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil && irRoute.ExtAuth != nil {
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
	var (
		endpointType EndpointType
		tSocket      *corev3.TransportSocket
		err          error
	)

	// Get the address type from the first setting.
	// This is safe because no mixed address types in the settings.
	addrTypeState := rd.Settings[0].AddressType
	if addrTypeState != nil && *addrTypeState == ir.FQDN {
		endpointType = EndpointTypeDNS
	} else {
		endpointType = EndpointTypeStatic
	}

	if rd.Settings[0].TLS != nil {
		tSocket, err = processTLSSocket(rd.Settings[0].TLS, tCtx)
		if err != nil {
			return err
		}
	}

	if err = addXdsCluster(tCtx, &xdsClusterArgs{
		name:         rd.Name,
		settings:     rd.Settings,
		tSocket:      tSocket,
		endpointType: endpointType,
	}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
		return err
	}
	return nil
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
	filterName := extAuthFilterName(irRoute.ExtAuth)
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
