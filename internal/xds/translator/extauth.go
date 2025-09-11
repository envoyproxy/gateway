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
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&extAuth{})
}

type extAuth struct{}

var _ httpFilter = &extAuth{}

// patchHCM builds and appends the ext_authz Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an ext_authz filter for each route that contains an ExtAuthz config.
// The filter is disabled by default. It is enabled on the route level.
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
		if hcmContainsFilter(mgr, extAuthFilterName(route.Security.ExtAuth)) {
			continue
		}

		filter, err := buildHCMExtAuthFilter(route.Security.ExtAuth)
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
	return perRouteFilterName(egv1a1.EnvoyFilterExtAuthz, extAuth.Name)
}

func extAuthConfig(extAuth *ir.ExtAuth) *extauthv3.ExtAuthz {
	config := &extauthv3.ExtAuthz{
		TransportApiVersion: corev3.ApiVersion_V3,
	}

	if extAuth.FailOpen != nil {
		config.FailureModeAllow = *extAuth.FailOpen
	}

	if extAuth.RecomputeRoute != nil {
		config.ClearRouteCache = *extAuth.RecomputeRoute
	}

	var headersToExtAuth []*matcherv3.StringMatcher
	for _, header := range extAuth.HeadersToExtAuth {
		headersToExtAuth = append(headersToExtAuth, &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: header,
			},
			IgnoreCase: true,
		})
	}

	if extAuth.BodyToExtAuth != nil {
		config.WithRequestBody = &extauthv3.BufferSettings{
			MaxRequestBytes: extAuth.BodyToExtAuth.MaxRequestBytes,
		}
	}

	if len(headersToExtAuth) > 0 {
		config.AllowedHeaders = &matcherv3.ListStringMatcher{
			Patterns: headersToExtAuth,
		}
	}

	timeout := durationpb.New(defaultExtServiceRequestTimeout)
	if extAuth.Timeout != nil {
		timeout = durationpb.New(extAuth.Timeout.Duration)
	}

	if extAuth.HTTP != nil {
		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: httpService(extAuth.HTTP, timeout),
		}
	} else if extAuth.GRPC != nil {
		config.Services = &extauthv3.ExtAuthz_GrpcService{
			GrpcService: &corev3.GrpcService{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: grpcService(extAuth.GRPC),
				},
				Timeout: timeout,
			},
		}
	}

	return config
}

func httpService(http *ir.HTTPExtAuthService, timeout *durationpb.Duration) *extauthv3.HttpService {
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
		Timeout: timeout,
	}

	for _, header := range http.HeadersToBackend {
		headersToBackend = append(headersToBackend, &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: header,
			},
			IgnoreCase: true,
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
	if irRoute != nil &&
		irRoute.Security != nil &&
		irRoute.Security.ExtAuth != nil {
		return true
	}
	return false
}

// patchResources patches the cluster resources for the external auth services.
func (*extAuth) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtAuth(route) {
			continue
		}
		if route.Security.ExtAuth.HTTP != nil {
			if err := createExtServiceXDSCluster(
				&route.Security.ExtAuth.HTTP.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		} else {
			if err := createExtServiceXDSCluster(
				&route.Security.ExtAuth.GRPC.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// patchRoute patches the provided route with the extAuth config if applicable.
// Note: this method enables the corresponding extAuth filter for the provided route.
func (*extAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.ExtAuth == nil {
		return nil
	}
	filterName := extAuthFilterName(irRoute.Security.ExtAuth)
	if err := enableFilterOnRoute(route, filterName); err != nil {
		return err
	}
	return nil
}
