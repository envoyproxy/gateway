package translator

import (
	"errors"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
	"github.com/envoyproxy/gateway/internal/xds/types"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extauthzv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/tetratelabs/multierror"
	"google.golang.org/protobuf/types/known/anypb"
)

const (
	extAuthzFilter = "envoy.filters.http.ext_authz"
)

// patchHCMWithExtAuthzFilter builds and appends the external authorization Filter to the HTTP
// Connection Manager if applicable.
func patchHCMWithExtAuthzFilter(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsExtAuthz(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == extAuthzFilter {
			return nil
		}
	}

	extAuthzFilter, err := buildHCMExtAuthzFilter(irListener)
	if err != nil {
		return err
	}

	// Ensure the external authorization filter is the first one in the filter chain.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{extAuthzFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMExtAuthzFilter returns an external authorization filter from the provided IR listener.
func buildHCMExtAuthzFilter(irListener *ir.HTTPListener) (*hcmv3.HttpFilter, error) {
	for _, route := range irListener.Routes {
		grpc, err := url2Cluster(route.ExtAuthz.GRPCURI)
		if err != nil {
			return nil, err
		}

		authProto := &extauthzv3.ExtAuthz{
			Services: &extauthzv3.ExtAuthz_GrpcService{
				GrpcService: &corev3.GrpcService{
					TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{ClusterName: grpc.name},
					},
				},
			},
		}

		authAny, err := anypb.New(authProto)
		if err != nil {
			return nil, err
		}

		return &hcmv3.HttpFilter{
			Name: extAuthzFilter,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: authAny,
			},
		}, nil
	}

	return nil, nil
}

// patchRouteWithExtAuthz patches the provided route with an external authorization PerRouteConfig, if the
// route doesn't contain it.
func patchRouteWithExtAuthz(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[extAuthzFilter]; !ok {
		if !routeContainsExtAuthz(irRoute) {
			return nil
		}

		routeCfgProto := &extauthzv3.ExtAuthzPerRoute{}

		routeCfgAny, err := anypb.New(routeCfgProto)
		if err != nil {
			return err
		}

		if filterCfg == nil {
			route.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		route.TypedPerFilterConfig[extAuthzFilter] = routeCfgAny
	}

	return nil
}

// routeContainsExtAuthz returns true if external authorizations exists for the
// provided route.
func routeContainsExtAuthz(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	if irRoute != nil &&
		irRoute.ExtAuthz != nil &&
		irRoute.ExtAuthz.GRPCURI != "" {
		return true
	}

	return false
}

// listenerContainsExtAuthz returns true if the provided listener has external
// authroization policies attached to its routes.
func listenerContainsExtAuthz(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.ExtAuthz != nil {
			return true
		}
	}

	return false
}

// createExtAuthzClusters creates external authorizations clusters from the provided routes, if needed.
func createExtAuthzClusters(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtAuthz(route) {
			continue
		}

		var (
			grpc    *urlCluster
			ds      *ir.DestinationSetting
			tSocket *corev3.TransportSocket
			err     error
		)

		url := route.ExtAuthz.GRPCURI
		grpc, err = url2Cluster(url)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		ds = &ir.DestinationSetting{
			Weight:    ptr.To(uint32(1)),
			Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(grpc.hostname, grpc.port)},
			Protocol:  ir.GRPC,
		}

		tSocket, err = buildXdsUpstreamTLSSocket()
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		if err = addXdsCluster(tCtx, &xdsClusterArgs{
			name:         grpc.name,
			settings:     []*ir.DestinationSetting{ds},
			tSocket:      tSocket,
			endpointType: grpc.endpointType,
		}); err != nil && !errors.Is(err, ErrXdsClusterExists) {
			errs = multierror.Append(errs, err)
		}
	}

	return errs
}
