// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net/url"
	"sort"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var extAuthRouteMetadataContextNamespaces = []string{envoyGatewayXdsMetadataNamespace}

func init() {
	registerHTTPFilter(&extAuth{})
}

type extAuth struct{}

var _ httpFilter = &extAuth{}

// patchHCM builds and appends disabled ext_authz filters to the HTTP Connection Manager.
// One filter is added for each unique set of filter-level settings that Envoy cannot
// override per route. The auth service, body buffering, and context extensions are
// supplied per route through ExtAuthzPerRoute.CheckSettings.
func (*extAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error
	buckets := map[string]*extauthv3.ExtAuthz{}
	for _, route := range irListener.Routes {
		if !routeContainsExtAuth(route) {
			continue
		}

		filterName, err := extAuthFilterName(route.Security.ExtAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		config, err := extAuthFilterConfig(route.Security.ExtAuth)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		existing, ok := buckets[filterName]
		if !ok {
			buckets[filterName] = config
			continue
		}
		replace, err := extAuthConfigLess(config, existing)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		if replace {
			buckets[filterName] = config
		}
	}

	filterNames := make([]string, 0, len(buckets))
	for filterName := range buckets {
		filterNames = append(filterNames, filterName)
	}
	sort.Strings(filterNames)

	for _, filterName := range filterNames {
		if hcmContainsFilter(mgr, filterName) {
			continue
		}

		filter, err := buildHCMExtAuthFilter(filterName, buckets[filterName])
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMExtAuthFilter returns a disabled ext_authz HTTP filter for a filter-level config bucket.
func buildHCMExtAuthFilter(filterName string, extAuthProto *extauthv3.ExtAuthz) (*hcmv3.HttpFilter, error) {
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

func extAuthFilterName(extAuth *ir.ExtAuth) (string, error) {
	config := extAuthFilterKeyConfig(extAuth)
	data, err := proto.MarshalOptions{Deterministic: true}.Marshal(config)
	if err != nil {
		return "", err
	}

	return perRouteFilterName(egv1a1.EnvoyFilterExtAuthz, utils.Digest256(string(data))[:16]), nil
}

func extAuthFilterKeyConfig(extAuth *ir.ExtAuth) *extauthv3.ExtAuthz {
	config := extAuthFilterLevelConfig(extAuth)

	// Envoy derives some filter-level behavior, including allowed header handling,
	// from the configured auth service type. Keep only the service type in the
	// bucket key so HTTP and gRPC policies do not share an HCM filter; backend
	// details are intentionally supplied by the per-route service override.
	if extAuth.HTTP != nil {
		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: &extauthv3.HttpService{},
		}
	} else if extAuth.GRPC != nil {
		config.Services = &extauthv3.ExtAuthz_GrpcService{
			GrpcService: &corev3.GrpcService{},
		}
	}
	return config
}

func extAuthFilterConfig(extAuth *ir.ExtAuth) (*extauthv3.ExtAuthz, error) {
	config := extAuthFilterLevelConfig(extAuth)

	timeout := durationpb.New(defaultExtServiceRequestTimeout)
	if extAuth.Timeout != nil {
		timeout = durationpb.New(extAuth.Timeout.Duration)
	}

	var rp *corev3.RetryPolicy
	if extAuth.Traffic != nil && extAuth.Traffic.Retry != nil {
		var err error
		rp, err = buildNonRouteRetryPolicy(extAuth.Traffic.Retry)
		if err != nil {
			return nil, fmt.Errorf("build retry policy for http service: %w", err)
		}
	}

	if extAuth.HTTP != nil {
		// Keep a placeholder filter-level service so Envoy initializes
		// service-type-dependent filter state such as allowed header matchers.
		// Routes provide the effective backend through CheckSettings.
		hs := httpService(extAuth.HTTP, timeout)
		hs.RetryPolicy = rp

		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: hs,
		}
	} else if extAuth.GRPC != nil {
		// Keep a placeholder filter-level service so Envoy initializes
		// service-type-dependent filter state such as allowed header matchers.
		// Routes provide the effective backend through CheckSettings.
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

	return config, nil
}

func extAuthFilterLevelConfig(extAuth *ir.ExtAuth) *extauthv3.ExtAuthz {
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

	if len(headersToExtAuth) > 0 {
		config.AllowedHeaders = &matcherv3.ListStringMatcher{
			Patterns: headersToExtAuth,
		}
	}

	if extAuth.StatusOnError != nil {
		config.StatusOnError = &typev3.HttpStatus{
			Code: typev3.StatusCode(*extAuth.StatusOnError),
		}
	}
	return config
}

func extAuthConfigLess(a, b *extauthv3.ExtAuthz) (bool, error) {
	aData, err := proto.MarshalOptions{Deterministic: true}.Marshal(a)
	if err != nil {
		return false, err
	}
	bData, err := proto.MarshalOptions{Deterministic: true}.Marshal(b)
	if err != nil {
		return false, err
	}
	return string(aData) < string(bData), nil
}

func extAuthCheckSettings(extAuth *ir.ExtAuth) (*extauthv3.CheckSettings, error) {
	checkSettings := &extauthv3.CheckSettings{
		ContextExtensions: convertContextExtensions(extAuth.ContextExtensions),
	}

	if extAuth.BodyToExtAuth != nil {
		checkSettings.WithRequestBody = &extauthv3.BufferSettings{
			MaxRequestBytes: extAuth.BodyToExtAuth.MaxRequestBytes,
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

		checkSettings.ServiceOverride = &extauthv3.CheckSettings_HttpService{
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

		checkSettings.ServiceOverride = &extauthv3.CheckSettings_GrpcService{
			GrpcService: service,
		}
	}

	return checkSettings, nil
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
	filterName, err := extAuthFilterName(irRoute.Security.ExtAuth)
	if err != nil {
		return err
	}
	checkSettings, err := extAuthCheckSettings(irRoute.Security.ExtAuth)
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
