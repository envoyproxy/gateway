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
	extauthzv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_authz/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"github.com/tetratelabs/multierror"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	extAuthzFilter = "envoy.filters.http.ext_authz"
)

func init() {
	registerHTTPFilter(&extAuth{})
}

type extAuth struct {
}

var _ httpFilter = &extAuth{}

// patchHCM builds and appends external auth Filters to the HTTP Connection
// Manager if applicable, and it does not already exist.
// Note: this method creates an external auth filter for each route that contains
// an external auth config.
func (*extAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsExtAuthz(route) {
			continue
		}

		filter, err := buildHCMExtAuthzFilter(route)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		// skip if the filter already exists
		for _, existingFilter := range mgr.HttpFilters {
			if filter.Name == existingFilter.Name {
				continue
			}
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// buildHCMExtAuthzFilter returns an external auth HTTP filter from the provided IR HTTPRoute.
func buildHCMExtAuthzFilter(route *ir.HTTPRoute) (*hcmv3.HttpFilter, error) {
	extAuthzProto, err := extAuthzConfig(route)
	if err != nil {
		return nil, err
	}

	if err := extAuthzProto.ValidateAll(); err != nil {
		return nil, err
	}

	extAuthzAny, err := anypb.New(extAuthzProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: extAuthzFilterName(route),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extAuthzAny,
		},
	}, nil
}

func extAuthzFilterName(route *ir.HTTPRoute) string {
	return fmt.Sprintf("%s_%s", extAuthzFilter, route.Name)
}

func extAuthzConfig(route *ir.HTTPRoute) (*extauthzv3.ExtAuthz, error) {
	grpcURI := route.ExtAuthz.GRPCURI
	grpc, err := url2Cluster(grpcURI, false)
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
		TransportApiVersion: corev3.ApiVersion_V3,
		FailureModeAllow:    false, // do not let the request pass when authz unavailable
		StatusOnError:       &typev3.HttpStatus{Code: typev3.StatusCode_ServiceUnavailable},
	}
	return authProto, nil
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

// patchResources creates external authorizations clusters from the provided routes, if needed.
func (*extAuth) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
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
		grpc, err = url2Cluster(url, false)
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		ds = &ir.DestinationSetting{
			Weight:    ptr.To(uint32(1)),
			Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(grpc.hostname, grpc.port)},
			Protocol:  ir.GRPC,
		}

		if grpc.tls {
			// grpcURI is using TLS gRPC HT
			tSocket, err = buildExtAuthzTLSocket()
			if err != nil {
				errs = multierror.Append(errs, err)
				continue
			}
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

// buildExtAuthzTLSocket builds the TLS socket for the ext authx.
func buildExtAuthzTLSocket() (*corev3.TransportSocket, error) {
	tlsCtxProto := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{
							Filename: envoyTrustBundle,
						},
					},
				},
			},
		},
	}

	tlsCtxAny, err := anypb.New(tlsCtxProto)
	if err != nil {
		return nil, err
	}

	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

// patchRouteCfg patches the provided route configuration with the ext auth filter
// if applicable.
// Note: this method disables all the ext auth filters by default. The filter will
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
		if !routeContainsExtAuthz(route) {
			continue
		}

		filterName := extAuthzFilterName(route)
		filterCfg := routeCfg.TypedPerFilterConfig

		if _, ok := filterCfg[filterName]; ok {
			// This should not happen since this is the only place where the ext
			// auth filter is added in a route.
			errs = multierror.Append(errs, fmt.Errorf(
				"route config already contains oauth2 config: %+v", route))
			continue
		}

		// Disable all the filters by default. The filter will be enabled
		// per-route in the typePerFilterConfig of the route.
		routeCfgAny, err := anypb.New(&routev3.FilterConfig{Disabled: true})
		if err != nil {
			errs = multierror.Append(errs, err)
			continue
		}

		if filterCfg == nil {
			routeCfg.TypedPerFilterConfig = make(map[string]*anypb.Any)
		}

		routeCfg.TypedPerFilterConfig[filterName] = routeCfgAny
	}
	return errs
}

// patchRoute patches the provided route with the ext auth config if applicable.
// Note: this method enables the corresponding ext auth filter for the provided route.
func (*extAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.ExtAuthz == nil {
		return nil
	}

	if err := enableFilterOnRoute(extAuthzFilter, route, irRoute); err != nil {
		return err
	}
	return nil
}
