// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	extprocv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ext_proc/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
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

// patchHCM adds disabled envoy.filters.http.filter_chain placeholder filters to the HTTP
// Connection Manager: one for per-listener (per-connection) ExtProc and one for per-route
// ExtProc.
//
// Both placeholders are added together as soon as either scope has an ExtProc policy anywhere on
// this listener, even if the other scope currently has none. This keeps the HCM's filter set
// stable across that kind of policy churn too: e.g. adding a per-listener ExtProc policy later to
// a listener that already has per-route ExtProc only changes route/virtual host
// TypedPerFilterConfig (an RDS update), never the listener's filter list (which would require
// an LDS update and a connection drain).
//
// Envoy's ExtProcPerRoute API can only override one processor for one filter instance, while EG's
// EnvoyExtensionPolicy API allows an ordered list of ExtProc filters per listener/route. The
// filter_chain filter wraps an ordered, named sub-chain of ExtProc filters that is supplied
// separately (per virtual host for listener-scoped ExtProc, per route for route-scoped ExtProc).
func (*extProc) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	hasListenerExtProc := listenerContainsExtProc(irListener)
	hasRouteExtProc := slices.ContainsFunc(irListener.Routes, routeContainsExtProc)
	if !hasListenerExtProc && !hasRouteExtProc {
		return nil
	}

	for _, filterName := range []string{eepListenerFCFilterName(), eepFCFilterName()} {
		if hcmContainsFilter(mgr, filterName) {
			continue
		}
		filter, err := buildHCMFilterChainFilter(filterName)
		if err != nil {
			return err
		}
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

// extProcSubFilterName returns the stable top-level filter name for the per-route ExtProc slot
// index. The index is the execution slot within the ordered EnvoyExtensionPolicy ExtProc list, so
// route 0th processors always bind to the same listener-level filter.
func extProcSubFilterName(idx int) string {
	return perRouteFilterName(egv1a1.EnvoyFilterExtProc, strconv.Itoa(idx))
}

// extProcListenerSubFilterName returns the stable HCM-level filter name for a listener-level
// ExtProc slot. Using the envoy.filters.http.ext_proc prefix (instead of the raw policy name)
// ensures sortHTTPFilters assigns it the correct order relative to route-level slots.
func extProcListenerSubFilterName(idx int) string {
	return fmt.Sprintf("%s/listener/%d", egv1a1.EnvoyFilterExtProc, idx)
}

func extProcConfig(extProc *ir.ExtProc) (*extprocv3.ExternalProcessor, error) {
	config := &extprocv3.ExternalProcessor{
		GrpcService: &corev3.GrpcService{
			TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
				EnvoyGrpc: grpcExtProcService(extProc),
			},
			Timeout: durationpb.New(defaultExtServiceRequestTimeout),
		},
	}

	config.ProcessingMode = buildProcessingMode(extProc)

	if extProc.ShadowMode != nil {
		config.ObservabilityMode = *extProc.ShadowMode
	}

	if extProc.FailOpen != nil {
		config.FailureModeAllow = *extProc.FailOpen
	}

	if extProc.MessageTimeout != nil {
		config.MessageTimeout = durationpb.New(extProc.MessageTimeout.Duration)
	}

	if extProc.RequestAttributes != nil {
		config.RequestAttributes = slices.Clone(extProc.RequestAttributes)
	}

	if extProc.ResponseAttributes != nil {
		config.ResponseAttributes = slices.Clone(extProc.ResponseAttributes)
	}

	if extProc.Traffic != nil && extProc.Traffic.Retry != nil {
		rp, err := buildNonRouteRetryPolicy(extProc.Traffic.Retry)
		if err != nil {
			return nil, fmt.Errorf("failed to build retry policy for extproc: %w", err)
		}
		config.GrpcService.RetryPolicy = rp
	}

	if extProc.ForwardingMetadataNamespaces != nil || extProc.ReceivingMetadataNamespaces != nil {
		config.MetadataOptions = &extprocv3.MetadataOptions{}

		if extProc.ForwardingMetadataNamespaces != nil {
			config.MetadataOptions.ForwardingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: slices.Clone(extProc.ForwardingMetadataNamespaces),
			}
		}

		if extProc.ReceivingMetadataNamespaces != nil {
			config.MetadataOptions.ReceivingNamespaces = &extprocv3.MetadataOptions_MetadataNamespaces{
				Untyped: slices.Clone(extProc.ReceivingMetadataNamespaces),
			}
		}
	}
	config.AllowModeOverride = extProc.AllowModeOverride

	if extProc.StatusOnError != nil {
		config.StatusOnError = &typev3.HttpStatus{
			Code: typev3.StatusCode(*extProc.StatusOnError),
		}
	}
	return config, nil
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

// listenerContainsExtProc returns true if ExtProcs exist at listener scope.
func listenerContainsExtProc(irListener *ir.HTTPListener) bool {
	return irListener != nil && irListener.EnvoyExtensions != nil && len(irListener.EnvoyExtensions.ExtProcs) > 0
}

// patchResources patches the cluster resources for the external services.
func (*extProc) patchResources(tCtx *types.ResourceVersionTable,
	irListener *ir.HTTPListener, routes []*ir.HTTPRoute,
) error {
	if tCtx == nil || tCtx.XdsResources == nil {
		return errors.New("xds resource table is nil")
	}

	var errs error
	addClusters := func(extProcs []ir.ExtProc) {
		for i := range extProcs {
			ep := extProcs[i]
			if err := createExtServiceXDSCluster(&ep.Destination, ep.Traffic, tCtx); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}

	if listenerContainsExtProc(irListener) {
		addClusters(irListener.EnvoyExtensions.ExtProcs)
	}
	for _, route := range routes {
		if !routeContainsExtProc(route) {
			continue
		}
		addClusters(route.EnvoyExtensions.ExtProcs)
	}

	return errs
}

// patchRoute patches the provided route with the extProc config if applicable.
//
// A nil EnvoyExtensions means no route-scoped policy owns this route: it keeps inheriting the
// listener-scoped ExtProcs delivered at VirtualHost scope by patchVirtualHost.
//
// A non-nil EnvoyExtensions means a more specific (xRoute or route rule) policy owns this route
// and fully replaces — never merges with — the listener-scoped policy. The extension count is
// intentionally not checked: an empty result (e.g. fail-open invalid Wasm) still represents a
// more specific policy that owns this route and must suppress the lower-scope ExtProcs.
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

	// A non-nil EnvoyExtensions means a more specific route policy owns this route and fully
	// replaces the listener-scoped policy. The extension count is intentionally not checked
	// here: an empty result (e.g. fail-open invalid Wasm) still represents a more specific
	// policy that owns this route and must suppress the lower-scope ExtProc.
	if err := disableFilterOnRouteOnce(route, eepListenerFCFilterName()); err != nil {
		return err
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range irRoute.EnvoyExtensions.ExtProcs {
		cfg, err := extProcConfig(&irRoute.EnvoyExtensions.ExtProcs[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        extProcSubFilterName(idx),
			TypedConfig: cfgAny,
		})
	}

	if len(newFilters) == 0 {
		return nil
	}

	merged, err := mergeFilterChainConfigPerRoute(route.GetTypedPerFilterConfig()[eepFCFilterName()], newFilters)
	if err != nil {
		return err
	}
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[eepFCFilterName()] = merged
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

// patchVirtualHost enables the listener-scoped ExtProc filters at VirtualHost scope so a
// listener's policy does not bleed into virtual hosts belonging to a different listener that
// shares the same RouteConfiguration. Delivery via VirtualHost TypedPerFilterConfig goes through
// RDS, so policy changes do not trigger listener drains.
func (*extProc) patchVirtualHost(vh *routev3.VirtualHost, httpListener *ir.HTTPListener) error {
	if !listenerContainsExtProc(httpListener) {
		return nil
	}

	filterName := eepListenerFCFilterName()
	existing := vh.GetTypedPerFilterConfig()[filterName]
	alreadyDelivered, err := filterChainAlreadyHasType(existing, egv1a1.EnvoyFilterExtProc)
	if err != nil {
		return err
	}
	if alreadyDelivered {
		return nil
	}

	var newFilters []*corev3.TypedExtensionConfig
	for idx := range httpListener.EnvoyExtensions.ExtProcs {
		cfg, err := extProcConfig(&httpListener.EnvoyExtensions.ExtProcs[idx])
		if err != nil {
			return err
		}
		cfgAny, err := anypb.New(cfg)
		if err != nil {
			return err
		}
		newFilters = append(newFilters, &corev3.TypedExtensionConfig{
			Name:        extProcListenerSubFilterName(idx),
			TypedConfig: cfgAny,
		})
	}

	merged, err := mergeFilterChainConfigPerRoute(existing, newFilters)
	if err != nil {
		return err
	}
	if vh.TypedPerFilterConfig == nil {
		vh.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	vh.TypedPerFilterConfig[filterName] = merged
	return nil
}
