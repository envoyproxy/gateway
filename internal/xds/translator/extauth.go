// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

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
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/utils/str"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var extAuthRouteMetadataContextNamespaces = []string{envoyGatewayXdsMetadataNamespace}

// extAuthClusterPrefix namespaces the deduplicated ext auth cluster names
// produced by extServiceClusterName.
const extAuthClusterPrefix = "extauth"

// extServiceNameHashLength is the number of hex characters kept from the content
// hash in a deduplicated ext service cluster name (64-bit, matching the JWT
// provider dedup width).
const extServiceNameHashLength = 16

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

		// Only generate one ext_authz filter per unique filter name. Routes that
		// share the same SecurityPolicy (and therefore the same ExtAuth name)
		// reuse a single ext_authz filter on this HCM.
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
	extAuthProto, err := extAuthConfig(extAuth)
	if err != nil {
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
	return perRouteFilterName(egv1a1.EnvoyFilterExtAuthz, extAuth.Name)
}

// extServiceClusterName derives a deduplicated upstream cluster name for an
// external-service backend (ext auth, ext proc, oidc, ...) from a hash of the
// backend identity in the IR.
//
// The identity is the destination Settings plus the traffic features, with two
// deliberate exclusions per setting:
//   - the policy-derived Name (it embeds the SecurityPolicy and surfaces as
//     locality.Region), and
//   - the resolved Endpoints.
//
// Endpoints are runtime membership, not backend identity: for a Service/
// ServiceImport (EDS) backend they change on every scale or rollout. Folding them
// into the name would rename the cluster on routine endpoint churn, forcing Envoy
// to recreate the cluster (CDS/LDS churn, resetting cluster stats and connection
// pools) instead of applying a plain EDS update.
//
// The hash covers the cluster-shaping settings (TLS, protocol, ...) and traffic features.
// The RouteDestination Name/StatName/Metadata are policy-derived and excluded entirely.
func extServiceClusterName(prefix, authority string, rd *ir.RouteDestination, traffic *ir.TrafficFeatures) (string, error) {
	// rewriteSettingNames returns copies (safe to further mutate) with the
	// policy-derived Name cleared; drop the endpoints too so the name is stable
	// across endpoint churn.
	settings := rewriteSettingNames(rd.Settings, "")
	for _, ds := range settings {
		ds.Endpoints = nil
	}
	identity := struct {
		Settings []*ir.DestinationSetting `json:"settings,omitempty"`
		Traffic  *ir.TrafficFeatures      `json:"traffic,omitempty"`
	}{
		Settings: settings,
		Traffic:  traffic,
	}
	b, err := json.Marshal(identity)
	if err != nil {
		return "", err
	}
	hash := utils.Digest256(string(b))[:extServiceNameHashLength]
	return fmt.Sprintf("%s/%s/%s", prefix, str.SanitizeLabelName(authority), hash), nil
}

// rewriteSettingNames returns copies of settings with the policy-derived setting
// Name rewritten to "<destName>/backend/<i>" (or cleared when destName is empty).
// The Name embeds the SecurityPolicy and surfaces as locality.Region, so rebasing
// it off the content-hashed cluster name lets policies sharing a backend produce
// byte-identical clusters.
func rewriteSettingNames(settings []*ir.DestinationSetting, destName string) []*ir.DestinationSetting {
	out := make([]*ir.DestinationSetting, len(settings))
	for i, ds := range settings {
		cp := *ds
		if destName == "" {
			cp.Name = ""
		} else {
			cp.Name = fmt.Sprintf("%s/backend/%d", destName, i)
		}
		out[i] = &cp
	}
	return out
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
		clusterName, err := extServiceClusterName(extAuthClusterPrefix, extAuth.HTTP.Authority, &extAuth.HTTP.Destination, extAuth.Traffic)
		if err != nil {
			return nil, err
		}
		hs := httpService(extAuth.HTTP, clusterName, timeout)
		hs.RetryPolicy = rp

		config.Services = &extauthv3.ExtAuthz_HttpService{
			HttpService: hs,
		}
	} else if extAuth.GRPC != nil {
		clusterName, err := extServiceClusterName(extAuthClusterPrefix, extAuth.GRPC.Authority, &extAuth.GRPC.Destination, extAuth.Traffic)
		if err != nil {
			return nil, err
		}
		service := &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcService(extAuth.GRPC, clusterName),
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

func httpService(http *ir.HTTPExtAuthService, clusterName string, timeout *durationpb.Duration) *extauthv3.HttpService {
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
			Cluster: clusterName,
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

func grpcService(grpc *ir.GRPCExtAuthService, clusterName string) *corev3.GrpcService_EnvoyGrpc {
	return &corev3.GrpcService_EnvoyGrpc{
		ClusterName: clusterName,
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
			if err := createDedupedExtServiceCluster(
				extAuthClusterPrefix, http.Authority, &http.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
			if err := processClientCertificates(tCtx, http.Destination.Settings); err != nil {
				errs = errors.Join(errs, err)
			}
		} else {
			grpc := route.Security.ExtAuth.GRPC
			if err := createDedupedExtServiceCluster(
				extAuthClusterPrefix, grpc.Authority, &grpc.Destination, route.Security.ExtAuth.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
			if err := processClientCertificates(tCtx, grpc.Destination.Settings); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// createDedupedExtServiceCluster creates the upstream cluster for an
// external-service backend using the content-hashed cluster name
// (extServiceClusterName), so that identical backends across policies collapse
// onto a single cluster via addXdsCluster's name-based dedup. The IR destination
// is not mutated.
func createDedupedExtServiceCluster(prefix, authority string, rd *ir.RouteDestination, traffic *ir.TrafficFeatures, tCtx *types.ResourceVersionTable) error {
	name, err := extServiceClusterName(prefix, authority, rd, traffic)
	if err != nil {
		return err
	}
	// Build the real cluster with the hashed name and content-derived setting
	// names. The route-destination Metadata (SecurityPolicy attribution) is kept
	// as-is: when the same backend is referenced by multiple policies they dedup
	// to a single cluster (addXdsCluster keeps the first-encountered), so the
	// cluster carries attribution from only one of the referencing policies. That
	// is acceptable — the metadata is informational and not used for routing.
	named := *rd
	named.Name = name
	named.Settings = rewriteSettingNames(rd.Settings, name)
	return createExtServiceXDSCluster(&named, traffic, tCtx)
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
	contextExtensions := convertContextExtensions(irRoute.Security.ExtAuth.ContextExtensions)
	if err := enableFilterOnRoute(route, filterName, &extauthv3.ExtAuthzPerRoute{
		Override: &extauthv3.ExtAuthzPerRoute_CheckSettings{
			CheckSettings: &extauthv3.CheckSettings{
				ContextExtensions: contextExtensions,
			},
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
