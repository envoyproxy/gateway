// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"math"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	xdsbufferhttpv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/buffer/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&requestBuffer{})
}

var _ httpFilter = &requestBuffer{}

type requestBuffer struct{}

// patchHCM will add a Buffer filter to the HCM filter chain if any of the routes contain a request buffer.
// The Buffer filter is required in the HCM chain for any BufferPerRoute filter to work, so we will add the
// filter using the first request buffer settings found but set the filter to be disabled. patchRoute will
// then enable and set the correct settings for that particular route
func (r *requestBuffer) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, existingFilter := range mgr.HttpFilters {
		if existingFilter.Name == egv1a1.EnvoyFilterBuffer.String() {
			return nil
		}
	}

	for _, route := range irListener.Routes {
		if !routeContainsRequestBuffer(route) {
			continue
		}

		requestBufferFilter, err := buildHCMRequestBufferFilter(route.Traffic.RequestBuffer)
		if err != nil {
			return err
		}

		mgr.HttpFilters = append(mgr.HttpFilters, requestBufferFilter)

		return nil
	}

	return nil
}

func buildHCMRequestBufferFilter(spec *ir.RequestBuffer) (*hcmv3.HttpFilter, error) {
	maxBytes, ok := spec.Limit.AsInt64()
	if !ok {
		return nil, fmt.Errorf("invalid Limit value %s", spec.Limit.String())
	}

	if maxBytes < 0 || maxBytes > math.MaxUint32 {
		return nil, fmt.Errorf("limit value %s is out of range, must be between 0 and %d",
			spec.Limit.String(), math.MaxUint32)
	}

	bufferProto := &xdsbufferhttpv3.Buffer{
		MaxRequestBytes: wrapperspb.UInt32(uint32(maxBytes)),
	}

	bufferAny, err := proto.ToAnyWithValidation(bufferProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterBuffer.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: bufferAny,
		},
		Disabled: true,
	}, nil
}

func (r *requestBuffer) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
	return nil
}

// patchRoute will add a BufferPerRoute filter for a particular route
func (r *requestBuffer) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if !routeContainsRequestBuffer(irRoute) {
		return nil
	}

	filter, err := buildRequestBufferPerRouteProto(irRoute.Traffic.RequestBuffer)
	if err != nil {
		return err
	}

	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[egv1a1.EnvoyFilterBuffer.String()] = filter

	return nil
}

func buildRequestBufferPerRouteProto(spec *ir.RequestBuffer) (*anypb.Any, error) {
	maxBytes, ok := spec.Limit.AsInt64()
	if !ok {
		return nil, fmt.Errorf("invalid Limit value %s", spec.Limit.String())
	}

	if maxBytes < 0 || maxBytes > math.MaxUint32 {
		return nil, fmt.Errorf("limit value %s is out of range, must be between 0 and %d",
			spec.Limit.String(), math.MaxUint32)
	}

	bufferProto := &xdsbufferhttpv3.BufferPerRoute{
		Override: &xdsbufferhttpv3.BufferPerRoute_Buffer{
			Buffer: &xdsbufferhttpv3.Buffer{
				MaxRequestBytes: wrapperspb.UInt32(uint32(maxBytes)),
			},
		},
	}

	return proto.ToAnyWithValidation(bufferProto)
}

func routeContainsRequestBuffer(route *ir.HTTPRoute) bool {
	return route.Traffic != nil && route.Traffic.RequestBuffer != nil
}
