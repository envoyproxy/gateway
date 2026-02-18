// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	adaptiveconcurrencyv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/adaptive_concurrency/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&adaptiveConcurrency{})
}

type adaptiveConcurrency struct{}

var _ httpFilter = &adaptiveConcurrency{}

// patchHCM builds and appends the adaptive concurrency filter to the HTTP Connection Manager
// if applicable. Since the adaptive concurrency filter does not have native per-route
// configuration support, a separate filter instance is created for each route that
// requires it. Each filter is disabled by default and enabled on the route level.
func (*adaptiveConcurrency) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	var errs error

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	for _, route := range irListener.Routes {
		if !routeContainsAdaptiveConcurrency(route) {
			continue
		}

		filterName := adaptiveConcurrencyFilterName(route.Traffic.AdaptiveConcurrency)

		// Only add one filter per unique name.
		if hcmContainsFilter(mgr, filterName) {
			continue
		}

		filter, err := buildHCMAdaptiveConcurrencyFilter(route.Traffic.AdaptiveConcurrency)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

func buildHCMAdaptiveConcurrencyFilter(ac *ir.AdaptiveConcurrency) (*hcmv3.HttpFilter, error) {
	acProto := buildAdaptiveConcurrencyProto(ac)

	acAny, err := proto.ToAnyWithValidation(acProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name:     adaptiveConcurrencyFilterName(ac),
		Disabled: true,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: acAny,
		},
	}, nil
}

func buildAdaptiveConcurrencyProto(ac *ir.AdaptiveConcurrency) *adaptiveconcurrencyv3.AdaptiveConcurrency {
	gradientConfig := &adaptiveconcurrencyv3.GradientControllerConfig{}

	// Sample aggregate percentile
	if ac.SampleAggregatePercentile != nil {
		gradientConfig.SampleAggregatePercentile = &typev3.Percent{
			Value: *ac.SampleAggregatePercentile,
		}
	}

	// Concurrency limit params
	concurrencyLimitParams := &adaptiveconcurrencyv3.GradientControllerConfig_ConcurrencyLimitCalculationParams{}
	if ac.MaxConcurrencyLimit != nil {
		concurrencyLimitParams.MaxConcurrencyLimit = wrapperspb.UInt32(*ac.MaxConcurrencyLimit)
	}
	if ac.ConcurrencyUpdateInterval != nil {
		concurrencyLimitParams.ConcurrencyUpdateInterval = durationpb.New(ac.ConcurrencyUpdateInterval.Duration)
	}
	gradientConfig.ConcurrencyLimitParams = concurrencyLimitParams

	// MinRTT calculation params
	minRttParams := &adaptiveconcurrencyv3.GradientControllerConfig_MinimumRTTCalculationParams{}
	if ac.MinRTTCalcInterval != nil {
		minRttParams.Interval = durationpb.New(ac.MinRTTCalcInterval.Duration)
	}
	if ac.FixedMinRTT != nil {
		minRttParams.FixedValue = durationpb.New(ac.FixedMinRTT.Duration)
	}
	if ac.RequestCount != nil {
		minRttParams.RequestCount = wrapperspb.UInt32(*ac.RequestCount)
	}
	if ac.Jitter != nil {
		minRttParams.Jitter = &typev3.Percent{
			Value: *ac.Jitter,
		}
	}
	if ac.MinConcurrency != nil {
		minRttParams.MinConcurrency = wrapperspb.UInt32(*ac.MinConcurrency)
	}
	if ac.Buffer != nil {
		minRttParams.Buffer = &typev3.Percent{
			Value: *ac.Buffer,
		}
	}
	gradientConfig.MinRttCalcParams = minRttParams

	acProto := &adaptiveconcurrencyv3.AdaptiveConcurrency{
		ConcurrencyControllerConfig: &adaptiveconcurrencyv3.AdaptiveConcurrency_GradientControllerConfig{
			GradientControllerConfig: gradientConfig,
		},
	}

	if ac.ConcurrencyLimitExceededStatus != nil {
		acProto.ConcurrencyLimitExceededStatus = &typev3.HttpStatus{
			Code: typev3.StatusCode(*ac.ConcurrencyLimitExceededStatus),
		}
	}

	return acProto
}

func adaptiveConcurrencyFilterName(ac *ir.AdaptiveConcurrency) string {
	return perRouteFilterName(egv1a1.EnvoyFilterAdaptiveConcurrency, ac.Name)
}

// patchRoute enables the adaptive concurrency filter on routes that need it.
func (*adaptiveConcurrency) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if !routeContainsAdaptiveConcurrency(irRoute) {
		return nil
	}

	filterName := adaptiveConcurrencyFilterName(irRoute.Traffic.AdaptiveConcurrency)
	if err := enableFilterOnRoute(route, filterName, &routev3.FilterConfig{
		Config: &anypb.Any{},
	}); err != nil {
		return err
	}
	return nil
}

func (*adaptiveConcurrency) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	return nil
}

func routeContainsAdaptiveConcurrency(route *ir.HTTPRoute) bool {
	return route != nil && route.Traffic != nil && route.Traffic.AdaptiveConcurrency != nil
}
