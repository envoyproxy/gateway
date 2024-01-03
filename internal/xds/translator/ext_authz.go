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

// patchHCM builds and appends the external authorization Filter to the HTTP
// Connection Manager if applicable.
func (*extAuth) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
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

	extAuthzFilter, err := buildHCM(irListener)
	if err != nil {
		return err
	}

	mgr.HttpFilters = append(mgr.HttpFilters, extAuthzFilter)

	return nil
}

// patchRouteConfig patches the provided route configuration with
// the ext authz filter if applicable.
// Note: this method disables the ext authz filters on all routes not explicitly requiring it.
func (*extAuth) patchRouteConfig(routeCfg *routev3.RouteConfiguration, irListener *ir.HTTPListener) error {
	if routeCfg == nil {
		return errors.New("route configuration is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}
	if !listenerContainsExtAuthz(irListener) {
		return nil
	}

	for _, route := range irListener.Routes {
		if !routeContainsExtAuthz(route) {
			perRouteFilterName := extAuthzFilterName(route)
			filterCfg := routeCfg.TypedPerFilterConfig

			routeCfgProto := &extauthzv3.ExtAuthzPerRoute{
				Override: &extauthzv3.ExtAuthzPerRoute_Disabled{Disabled: true},
			}

			routeCfgAny, err := anypb.New(routeCfgProto)
			if err != nil {
				return err
			}

			if filterCfg == nil {
				routeCfg.TypedPerFilterConfig = make(map[string]*anypb.Any)
			}

			routeCfg.TypedPerFilterConfig[perRouteFilterName] = routeCfgAny
		}
	}

	return nil
}

// buildHCM returns an external authorization filter from the provided IR listener.
func buildHCM(irListener *ir.HTTPListener) (*hcmv3.HttpFilter, error) {
	var grpcURI string

	for _, route := range irListener.Routes {
		if !routeContainsExtAuthz(route) {
			continue
		}

		if grpcURI != "" && grpcURI != route.ExtAuthz.GRPCURI {
			return nil, errors.New("multiple grpcURI for this listener, unsupported configuration")
		}

		if grpcURI == "" {
			grpcURI = route.ExtAuthz.GRPCURI
		}

		grpc, err := url2Cluster(grpcURI)
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

		if err := authProto.ValidateAll(); err != nil {
			return nil, err
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
// authorization policies attached to its routes.
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

		if !grpc.tlsDisabled {
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

func extAuthzFilterName(route *ir.HTTPRoute) string {
	return fmt.Sprintf("%s_%s", extAuthzFilter, route.Name)
}

func (*extAuth) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	return nil
}
