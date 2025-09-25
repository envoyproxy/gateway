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
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&extProc{})
}

type extProc struct{}

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

		for i := range route.EnvoyExtensions.ExtProcs {
			ep := &route.EnvoyExtensions.ExtProcs[i]
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
func buildHCMExtProcFilter(extProc *ir.ExtProc) (*hcmv3.HttpFilter, error) {
	extAuthProto := extProcConfig(extProc)
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

func extProcFilterName(extProc *ir.ExtProc) string {
	return perRouteFilterName(egv1a1.EnvoyFilterExtProc, extProc.Name)
}

func extProcConfig(extProc *ir.ExtProc) *extprocv3.ExternalProcessor {
	config := &extprocv3.ExternalProcessor{
		GrpcService: &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcExtProcService(extProc),
			},
			Timeout: &durationpb.Duration{
				Seconds: defaultExtServiceRequestTimeout,
			},
		},
	}

	config.ProcessingMode = buildProcessingMode(extProc)

	if extProc.FailOpen != nil {
		config.FailureModeAllow = *extProc.FailOpen
	}

	if extProc.MessageTimeout != nil {
		config.MessageTimeout = durationpb.New(extProc.MessageTimeout.Duration)
	}

	if extProc.RequestAttributes != nil {
		var attrs []string
		attrs = append(attrs, extProc.RequestAttributes...)
		config.RequestAttributes = attrs
	}

	if extProc.ResponseAttributes != nil {
		var attrs []string
		attrs = append(attrs, extProc.ResponseAttributes...)
		config.ResponseAttributes = attrs
	}

	if extProc.ForwardingMetadataNamespaces != nil || extProc.ReceivingMetadataNamespaces != nil {
		config.MetadataOptions = &extprocv3.MetadataOptions{}

		if extProc.ForwardingMetadataNamespaces != nil {
			var ns []string
			ns = append(ns, extProc.ForwardingMetadataNamespaces...)
			config.MetadataOptions.ForwardingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: ns,
			}
		}

		if extProc.ReceivingMetadataNamespaces != nil {
			var ns []string
			ns = append(ns, extProc.ReceivingMetadataNamespaces...)
			config.MetadataOptions.ReceivingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: ns,
			}
		}
	}
	config.AllowModeOverride = extProc.AllowModeOverride
	return config
}

func grpcExtProcService(extProc *ir.ExtProc) *corev3.GrpcService_EnvoyGrpc {
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

	return irRoute.EnvoyExtensions != nil && len(irRoute.EnvoyExtensions.ExtProcs) > 0
}

// patchResources patches the cluster resources for the external services.
func (*extProc) patchResources(tCtx *types.ResourceVersionTable,
	routes []*ir.HTTPRoute,
) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	for _, route := range routes {
		if !routeContainsExtProc(route) {
			continue
		}

		for i := range route.EnvoyExtensions.ExtProcs {
			ep := route.EnvoyExtensions.ExtProcs[i]
			if err := createExtServiceXDSCluster(
				&ep.Destination, ep.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	return errs
}

// patchRoute patches the provided route with the extProc config if applicable.
// Note: this method enables the corresponding extProc filter for the provided route.
func (*extProc) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.EnvoyExtensions == nil {
		return nil
	}

	for i := range irRoute.EnvoyExtensions.ExtProcs {
		ep := &irRoute.EnvoyExtensions.ExtProcs[i]
		filterName := extProcFilterName(ep)
		if err := enableFilterOnRoute(route, filterName); err != nil {
			return err
		}
	}
	return nil
}

func buildProcessingMode(extProc *ir.ExtProc) *extprocv3.ProcessingMode {
	processingMode := &extprocv3.ProcessingMode{
		RequestHeaderMode:   extprocv3.ProcessingMode_SKIP,
		ResponseHeaderMode:  extprocv3.ProcessingMode_SKIP,
		RequestBodyMode:     extprocv3.ProcessingMode_NONE,
		ResponseBodyMode:    extprocv3.ProcessingMode_NONE,
		RequestTrailerMode:  extprocv3.ProcessingMode_SKIP,
		ResponseTrailerMode: extprocv3.ProcessingMode_SKIP,
	}

	if extProc.RequestBodyProcessingMode != nil {
		processingMode.RequestBodyMode = translateExtProcBodyProcessingMode(extProc.RequestBodyProcessingMode)
		//
		if processingMode.RequestBodyMode == extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED {
			processingMode.RequestTrailerMode = extprocv3.ProcessingMode_SEND
		}
	}

	if extProc.RequestHeaderProcessing {
		processingMode.RequestHeaderMode = extprocv3.ProcessingMode_SEND
	}

	if extProc.ResponseBodyProcessingMode != nil {
		processingMode.ResponseBodyMode = translateExtProcBodyProcessingMode(extProc.ResponseBodyProcessingMode)
		if processingMode.ResponseBodyMode == extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED {
			processingMode.ResponseTrailerMode = extprocv3.ProcessingMode_SEND
		}
	}

	if extProc.ResponseHeaderProcessing {
		processingMode.ResponseHeaderMode = extprocv3.ProcessingMode_SEND
	}

	return processingMode
}

func translateExtProcBodyProcessingMode(mode *ir.ExtProcBodyProcessingMode) extprocv3.ProcessingMode_BodySendMode {
	lookup := map[ir.ExtProcBodyProcessingMode]extprocv3.ProcessingMode_BodySendMode{
		ir.ExtProcBodyBuffered:           extprocv3.ProcessingMode_BUFFERED,
		ir.ExtProcBodyBufferedPartial:    extprocv3.ProcessingMode_BUFFERED_PARTIAL,
		ir.ExtProcBodyStreamed:           extprocv3.ProcessingMode_STREAMED,
		ir.ExtProcBodyFullDuplexStreamed: extprocv3.ProcessingMode_FULL_DUPLEX_STREAMED,
	}
	if r, found := lookup[*mode]; found {
		return r
	}
	return extprocv3.ProcessingMode_NONE
}
