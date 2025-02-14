// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	healthcheckv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/health_check/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&healthCheck{})
}

type healthCheck struct{}

var _ httpFilter = &healthCheck{}

// patchHCM builds and appends the health_check Filter to the HTTP Connection Manager if applicable.
func (*healthCheck) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	// Return early if filter already exists.
	if hcmContainsFilter(mgr, wellknown.HealthCheck) {
		return nil
	}

	var (
		filter *hcmv3.HttpFilter
		err    error
	)

	if irListener.HealthCheck == nil {
		return nil
	}

	if filter, err = buildHealthCheckFilter(irListener.HealthCheck); err != nil {
		return err
	}

	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return nil
}

// buildHealthCheckFilter builds a health_check filter from provided ir Listener.
func buildHealthCheckFilter(healthCheck *ir.HealthCheckSettings) (*hcmv3.HttpFilter, error) {
	var (
		healthCheckProto *healthcheckv3.HealthCheck
		healthCheckAny   *anypb.Any
		err              error
	)

	healthCheckProto = &healthcheckv3.HealthCheck{
		PassThroughMode: &wrapperspb.BoolValue{Value: false},
		Headers: []*routev3.HeaderMatcher{{
			Name: ":path",
			HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
				StringMatch: &matcherv3.StringMatcher{
					MatchPattern: &matcherv3.StringMatcher_Exact{
						Exact: healthCheck.Path,
					},
				},
			},
		}},
	}

	if healthCheckAny, err = anypb.New(healthCheckProto); err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: wellknown.HealthCheck,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: healthCheckAny,
		},
	}, nil
}

func (*healthCheck) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func (*healthCheck) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	return nil
}
