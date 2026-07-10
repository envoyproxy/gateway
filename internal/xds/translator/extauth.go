// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"net/url"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var extAuthRouteMetadataContextNamespaces = []string{envoyGatewayXdsMetadataNamespace}

// extAuthClusterPrefix namespaces the deduplicated ext auth cluster names
// produced by extServiceClusterName.
const extAuthClusterPrefix = "extauth"

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
// built cluster with the policy-derived identifiers normalized out. The name is
// "<prefix>/<sanitized-authority>/<hash>": the prefix namespaces the consumer,
// the authority keeps it human-readable, and the content hash makes two policies
// referencing the same backend+settings collapse onto one cluster via
// addXdsCluster's name-based dedup.
func extServiceClusterName(prefix, authority string, rd *ir.RouteDestination, traffic *ir.TrafficFeatures) (string, error) {
	args := extServiceClusterArgs(rd, traffic)
	// Normalize every policy-derived identifier so the hash reflects only the
	// backend/config content, not the owning policy. Clearing the cluster name
	// also normalizes the derived TransportSocketMatch names ("<name>/tls/<i>"),
	// and clearing the per-setting names normalizes the locality Region
	// (buildWeightedLocalities sets Region = setting name).
	args.name = ""
	args.metadata = nil
	args.settings = normalizeExtServiceSettings(args.settings, "")

	result, err := buildXdsCluster(args)
	if err != nil {
		return "", err
	}
	// Fold the endpoints into the cluster proto so a single hash covers both the
	// cluster and its load assignment. protoHash is the shared helper (also used
	// by JWT provider dedup) and already truncates to a short, stable prefix.
	lb := ptr.Deref(args.loadBalancer, ir.LoadBalancer{})
	result.cluster.LoadAssignment = buildXdsClusterLoadAssignment("", args.settings, args.healthCheck, lb.PreferLocal, lb.WeightedZones)

	hash, err := protoHash(result.cluster)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s", prefix, sanitizeStatName(authority), hash), nil
}

// normalizeExtServiceSettings returns copies of settings with the policy-derived
// setting Name replaced by an identifier derived from destName (or cleared when
// destName is empty). This makes a deduplicated ext service cluster depend only
// on the backend content, so two SecurityPolicies pointing at the same backend
// produce byte-identical clusters — satisfying addXdsCluster's "same name ⇒
// same args" contract.
func normalizeExtServiceSettings(settings []*ir.DestinationSetting, destName string) []*ir.DestinationSetting {
	out := make([]*ir.DestinationSetting, len(settings))
	for i, ds := range settings {
		cp := *ds
		// Only the setting Name is policy-derived (it embeds the SecurityPolicy
		// namespace/name and surfaces as locality.Region), so it must be
		// normalized for dedup. Backend-derived fields such as Metadata are
		// identical across policies referencing the same backend, so they are
		// left intact and remain part of the hashed content.
		if destName == "" {
			cp.Name = ""
		} else {
			cp.Name = fmt.Sprintf("%s/backend/%d", destName, i)
		}
		out[i] = &cp
	}
	return out
}

// sanitizeStatName replaces the characters Envoy uses as stat-path separators
// (".", ":") so an authority (host:port) embedded in a resource name does not
// fragment the resulting Prometheus stat labels.
func sanitizeStatName(name string) string {
	return strings.NewReplacer(".", "_", ":", "_").Replace(name)
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
	// names, clearing the policy-derived route-destination metadata, so two
	// policies pointing at the same backend produce byte-identical clusters that
	// dedup cleanly. The shared IR destination is not mutated.
	named := *rd
	named.Name = name
	named.Metadata = nil
	named.Settings = normalizeExtServiceSettings(rd.Settings, name)
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
