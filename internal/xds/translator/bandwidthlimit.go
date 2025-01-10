// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"github.com/envoyproxy/gateway/internal/ir"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	bandwidthlimitfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/bandwidth_limit/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"strings"
)

func patchRouteWithBandWidthLimit(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if irRoute.BandWidthLimit == nil {
		return nil
	}

	routeCfgAny, err := anypb.New(builbandWidthLimit(irRoute))
	if err != nil {
		return err
	}
	route.TypedPerFilterConfig["envoy.filters.http.bandwidth_limit"] = routeCfgAny
	return nil
}
func builbandWidthLimit(irRoute *ir.HTTPRoute) *bandwidthlimitfilterv3.BandwidthLimit {
	banwidthLimitFilterProto := &bandwidthlimitfilterv3.BandwidthLimit{
		StatPrefix:   "bandwidth_limiter_custom_route",
		LimitKbps:    irRoute.BandWidthLimit.Limit,
		FillInterval: irRoute.BandWidthLimit.Interval,
	}
	flag := corev3.RuntimeFeatureFlag{}
	flag.DefaultValue = irRoute.BandWidthLimit.Enable
	banwidthLimitFilterProto.RuntimeEnabled = &flag

	if strings.ToLower(irRoute.BandWidthLimit.Mode) == "request" {
		banwidthLimitFilterProto.EnableMode = bandwidthlimitfilterv3.BandwidthLimit_REQUEST
	} else if strings.ToLower(irRoute.BandWidthLimit.Mode) == "response" {
		banwidthLimitFilterProto.EnableMode = bandwidthlimitfilterv3.BandwidthLimit_RESPONSE
	} else if strings.ToLower(irRoute.BandWidthLimit.Mode) == "all" {
		banwidthLimitFilterProto.EnableMode = bandwidthlimitfilterv3.BandwidthLimit_REQUEST_AND_RESPONSE
	} else {
		banwidthLimitFilterProto.EnableMode = bandwidthlimitfilterv3.BandwidthLimit_DISABLED
	}
	return banwidthLimitFilterProto
}
