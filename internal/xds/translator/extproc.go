// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/golang/protobuf/ptypes/duration"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	extProcFilter = "envoy.filters.http.ext_proc"
)

func init() {
	registerHTTPFilter(&extProc{})
}

type extProc struct {
}

var _ httpFilter = &extProc{}

// patchHCM builds and appends the ext_proc Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an ext_proc filter for each route that contains an ExtAuthz config.
// The filter is disabled by default. It is enabled on the route level.
func (*extProc) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsExtProc(route) {
			continue
		}

		for _, ep := range route.ExtProcs {
			if hcmContainsFilter(mgr, extProcFilterName(ep)) {
				continue
			}

			filter, err := buildHCMExtProcFilter(ep)
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}

			mgr.HttpFilters = append(mgr.HttpFilters, filter)
		}
	}

	return errs
}

// buildHCMExtProcFilter returns an ext_proc HTTP filter from the provided IR HTTPRoute.
func buildHCMExtProcFilter(extProc ir.ExtProc) (*hcmv3.HttpFilter, error) {
	extAuthProto := extProcConfig(extProc)
	if err := extAuthProto.ValidateAll(); err != nil {
		return nil, err
	}

	extAuthAny, err := anypb.New(extAuthProto)
	if err != nil {
		return nil, err
	}

	// All extproc filters for all Routes are aggregated on HCM and disabled by default
	// Per-route config is used to enable the relevant filters on appropriate routes
	return &hcmv3.HttpFilter{
		Name:     extProcFilterName(extProc),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: extAuthAny,
		},
	}, nil
}

func extProcFilterName(extProc ir.ExtProc) string {
	return perRouteFilterName(extProcFilter, extProc.Name)
}

func extProcConfig(extProc ir.ExtProc) *extprocv3.ExternalProcessor {
	config := &extprocv3.ExternalProcessor{
		GrpcService: &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcExtProcService(extProc),
			},
			Timeout: &duration.Duration{
				Seconds: defaultExtServiceRequestTimeout,
			},
		},
	}

	return config
}

func grpcExtProcService(extProc ir.ExtProc) *corev3.GrpcService_EnvoyGrpc {
	return &corev3.GrpcService_EnvoyGrpc{
		ClusterName: extProc.Destination.Name,
		Authority:   extProc.Authority,
	}
}

// routeContainsExtProc returns true if ExtProcs exists for the provided route.
func routeContainsExtProc(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil {
		return false
	}

	return len(irRoute.ExtProcs) > 0
}

// patchResources patches the cluster resources for the external services.
func (*extProc) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtProc(route) {
			continue
		}

		for i := range route.ExtProcs {
			ep := route.ExtProcs[i]
			if err := createExtServiceXDSCluster(
				&ep.Destination, tCtx); err != nil && !errors.Is(
				err, ErrXdsClusterExists) {
				errs = errors.Join(errs, err)
			}
		}

	}

	return errs
}

// patchRoute patches the provided route with the extProc config if applicable.
// Note: this method enables the corresponding extProc filter for the provided route.
func (*extProc) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}

	for _, ep := range irRoute.ExtProcs {
		filterName := extProcFilterName(ep)
		if err := enableFilterOnRoute(route, filterName); err != nil {
			return err
		}
	}
	return nil
}
