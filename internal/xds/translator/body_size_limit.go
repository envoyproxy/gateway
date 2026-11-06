// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"strconv"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	bodysizelimitv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/body_size_limit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&bodySizeLimit{})
}

var _ httpFilter = &bodySizeLimit{}

// bodySizeLimit translates ir.RequestBodyLimit into the Envoy Body Size Limit filter, which rejects a
// request whose body exceeds max_request_bytes with HTTP 413 without buffering the body.
type bodySizeLimit struct{}

// patchHCM adds a Body Size Limit filter to the HCM filter chain for every distinct limit found on the
// listener's routes. The Envoy filter has no per-route configuration of its own, so one filter per limit
// is added, named after the limit and disabled by default; patchRoute then enables the matching filter on
// each route. Routes that share a limit share a filter.
func (*bodySizeLimit) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsRequestBodyLimit(route) {
			continue
		}

		filterName := bodySizeLimitFilterName(route.Traffic.RequestBodyLimit)
		if hcmContainsFilter(mgr, filterName) {
			continue
		}

		filter, err := buildHCMBodySizeLimitFilter(route.Traffic.RequestBodyLimit)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// bodySizeLimitFilterName returns the name of the HCM filter that enforces the provided limit.
// The limit in bytes is the configName, so routes with the same limit share one filter.
func bodySizeLimitFilterName(limit *ir.RequestBodyLimit) string {
	return perRouteFilterName(egv1a1.EnvoyFilterBodySizeLimit, strconv.FormatUint(limit.MaxSize, 10))
}

// buildHCMBodySizeLimitFilter returns a disabled Body Size Limit HTTP filter for the provided limit.
func buildHCMBodySizeLimitFilter(limit *ir.RequestBodyLimit) (*hcmv3.HttpFilter, error) {
	bodySizeLimitProto := &bodysizelimitv3.BodySizeLimit{
		MaxRequestBytes: wrapperspb.UInt64(limit.MaxSize),
	}

	bodySizeLimitAny, err := proto.ToAnyWithValidation(bodySizeLimitProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: bodySizeLimitFilterName(limit),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: bodySizeLimitAny,
		},
		Disabled: true,
	}, nil
}

func (*bodySizeLimit) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

// patchRoute enables the Body Size Limit filter that matches the route's limit on the provided route.
func (*bodySizeLimit) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if !routeContainsRequestBodyLimit(irRoute) {
		return nil
	}

	// The filter has no per-route configuration, so an empty FilterConfig only enables the disabled
	// HCM filter for this route.
	return enableFilterOnRoute(route, bodySizeLimitFilterName(irRoute.Traffic.RequestBodyLimit), &routev3.FilterConfig{
		Config: &anypb.Any{},
	})
}

func routeContainsRequestBodyLimit(route *ir.HTTPRoute) bool {
	return route != nil && route.Traffic != nil && route.Traffic.RequestBodyLimit != nil
}
