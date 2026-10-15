// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net/url"
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var extAuthRouteMetadataContextNamespaces = []string{envoyGatewayXdsMetadataNamespace}

func init() {
	registerHTTPFilter(&extAuth{})
}

type extAuth struct{}

var _ httpFilter = &extAuth{}

// patchHCM builds and appends the ext_authz Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: routes share one disabled ext_authz filter per distinct set of
// filter-level settings. Each route enables its filter and supplies its own
// authorization service through ExtAuthzPerRoute.CheckSettings, so creating or
// deleting a SecurityPolicy changes routes only and does not update the
// listener, which would drain its connections.
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

		filter, err := buildHCMExtAuthFilter(route.Security.ExtAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		if hcmContainsFilter(mgr, filter.Name) {
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMExtAuthFilter returns the shared ext_authz HTTP filter for the provided IR ExtAuth.
func buildHCMExtAuthFilter(extAuth *ir.ExtAuth) (*hcmv3.HttpFilter, error) {
	filterName, extAuthProto, _, err := splitExtAuthConfig(extAuth)
	if err != nil {
		return nil, err
	}
	extAuthAny, err := anypb.New(extAuthProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     filterName,
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extAuthAny,
		},
	}, nil
}

// extAuthPlaceholderCluster backs the service of the shared ext_authz filters.
// Every route that enables one of those filters overrides the service, so a
// request reaches this cluster only if Envoy cannot build the route's client,
// and then fails like any unreachable authorization service.
const extAuthPlaceholderCluster = "ext-authz/placeholder"

// splitExtAuthConfig splits the ext_authz config for extAuth into the
// listener-level filter and the route-level CheckSettings that carry the
// authorization service and context extensions. The filter is named by a hash
// of its content, so policies that differ only in their service share it.
func splitExtAuthConfig(extAuth *ir.ExtAuth) (string, *extauthv3.ExtAuthz, *extauthv3.CheckSettings, error) {
	config, err := extAuthConfig(extAuth)
	if err != nil {
		return "", nil, nil, err
	}

	checkSettings := &extauthv3.CheckSettings{
		ContextExtensions: convertContextExtensions(extAuth.ContextExtensions),
	}
	// The filter keeps the service type, because Envoy derives filter-level
	// behavior such as allowed request headers and body encoding from it.
	switch services := config.Services.(type) {
	case *extauthv3.ExtAuthz_HttpService:
		checkSettings.ServiceOverride = &extauthv3.CheckSettings_HttpService{HttpService: services.HttpService}
		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: &extauthv3.HttpService{
				ServerUri: &corev3.HttpUri{
					Uri: "http://127.0.0.1:1",
					HttpUpstreamType: &corev3.HttpUri_Cluster{
						Cluster: extAuthPlaceholderCluster,
					},
					Timeout: durationpb.New(defaultExtServiceRequestTimeout),
				},
			},
		}
	case *extauthv3.ExtAuthz_GrpcService:
		checkSettings.ServiceOverride = &extauthv3.CheckSettings_GrpcService{GrpcService: services.GrpcService}
		config.Services = &extauthv3.ExtAuthz_GrpcService{
			GrpcService: &corev3.GrpcService{
				TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
					EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
						ClusterName: extAuthPlaceholderCluster,
					},
				},
			},
		}
	}

	hash, err := protoHash(config)
	if err != nil {
		return "", nil, nil, err
	}
	return perRouteFilterName(egv1a1.EnvoyFilterExtAuthz, hash), config, checkSettings, nil
}

func extAuthConfig(extAuth *ir.ExtAuth) (*extauthv3.ExtAuthz, error) {
	config := &extauthv3.ExtAuthz{
		TransportApiVersion: corev3.ApiVersion_V3,
	}

	if extAuth.FailOpen != nil {
		config.FailureModeAllow = *extAuth.FailOpen
	}

	if extAuth.RecomputeRoute != nil {
		config.ClearRouteCache = *extAuth.RecomputeRoute
	}

	if extAuth.IncludeRouteMetadata != nil && *extAuth.IncludeRouteMetadata {
		config.RouteMetadataContextNamespaces = extAuthRouteMetadataContextNamespaces
	}

	headersToExtAuth := make([]*matcherv3.StringMatcher, 0, len(extAuth.HeadersToExtAuth))
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

	var rp *corev3.RetryPolicy
	// Set the retry policy if it exists.
	if extAuth.Traffic != nil && extAuth.Traffic.Retry != nil {
		var err error
		rp, err = buildNonRouteRetryPolicy(extAuth.Traffic.Retry)
		if err != nil {
			return nil, fmt.Errorf("build retry policy for http service: %w", err)
		}
	}

	if extAuth.HTTP != nil {
		hs := httpService(extAuth.HTTP, timeout)
		hs.RetryPolicy = rp

		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: hs,
		}
	} else if extAuth.GRPC != nil {
		service := &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcService(extAuth.GRPC),
			},
			Timeout: timeout,
		}
		service.RetryPolicy = rp

		config.Services = &extauthv3.ExtAuthz_GrpcService{
			GrpcService: service,
		}
	}

	if extAuth.StatusOnError != nil {
		config.StatusOnError = &typev3.HttpStatus{
			Code: typev3.StatusCode(*extAuth.StatusOnError),
		}
	}
	return config, nil
}

func httpService(http *ir.HTTPExtAuthService, timeout *durationpb.Duration) *extauthv3.HttpService {
	var (
		uri     string
		service *extauthv3.HttpService
	)

	service = &extauthv3.HttpService{
		PathPrefix:   http.Path,
		PathOverride: http.PathOverride,
	}

	u := url.URL{
		// scheme should be decided by the TLS setting, but we don't have that info now.
		// It's safe to set it to http because the ext auth filter doesn't use the
		// uri to make the request. It only uses the cluster.
		Scheme: "http",
		Host:   http.Authority,
	}
	if http.PathOverride != "" {
		u.Path = http.PathOverride
	} else {
		u.Path = http.Path
	}
	uri = u.String()

	service.ServerUri = &corev3.HttpUri{
		Uri: uri,
		HttpUpstreamType: &corev3.HttpUri_Cluster{
			Cluster: http.Destination.Name,
		},
		Timeout: timeout,
	}

	headersToBackend := make([]*matcherv3.StringMatcher, 0, len(http.HeadersToBackend))
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
	if slices.ContainsFunc(routes, routeContainsExtAuth) {
		errs = errors.Join(errs, addXdsCluster(tCtx, &xdsClusterArgs{
			name: extAuthPlaceholderCluster,
			settings: []*ir.DestinationSetting{{
				Name:      destinationSettingName(extAuthPlaceholderCluster),
				Weight:    new(uint32(1)),
				Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, "127.0.0.1", 1, false, nil)},
			}},
			endpointType: EndpointTypeStatic,
		}))
	}
	for _, route := range routes {
		if !routeContainsExtAuth(route) {
			continue
		}
		if http := route.Security.ExtAuth.HTTP; http != nil {
			if err := createExtServiceXDSCluster(
				&http.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
			if err := processClientCertificates(tCtx, http.Destination.Settings); err != nil {
				errs = errors.Join(errs, err)
			}
		} else {
			grpc := route.Security.ExtAuth.GRPC
			if err := createExtServiceXDSCluster(
				&grpc.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
			if err := processClientCertificates(tCtx, grpc.Destination.Settings); err != nil {
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
	filterName, _, checkSettings, err := splitExtAuthConfig(irRoute.Security.ExtAuth)
	if err != nil {
		return err
	}
	if err := enableFilterOnRoute(route, filterName, &extauthv3.ExtAuthzPerRoute{
		Override: &extauthv3.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: checkSettings,
		},
	}); err != nil {
		return err
	}
	return nil
}

// convertContextExtensions converts the provided context extensions
// [ir.PrivateBytes] values to regular string values.
func convertContextExtensions(irCtxExts []*ir.ContextExtention) map[string]string {
	if irCtxExts == nil {
		return nil
	}

	ctxExts := make(map[string]string, len(irCtxExts))
	for _, ext := range irCtxExts {
		if ext != nil {
			ctxExts[ext.Name] = string(ext.Value)
		}
	}

	return ctxExts
}
