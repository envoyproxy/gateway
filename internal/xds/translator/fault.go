// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	xdsfault "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/fault/v3"
	xdshttpfaultv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/fault/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	faultFilter = "envoy.filters.http.fault"
)

func init() {
	registerHTTPFilter(&fault{})
}

type fault struct {
}

var _ httpFilter = &fault{}

// patchHCM builds and appends the fault Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates an fault filter for each route that contains an Fault config.
func (*fault) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {

	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsFault(irListener) {
		return nil
	}

	// Return early if the fault filter already exists.
	for _, existingFilter := range mgr.HttpFilters {
		if existingFilter.Name == faultFilter {
			return nil
		}
	}

	faultFilter, err := buildHCMFaultFilter()
	if err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, faultFilter)

	return nil
}

// buildHCMFaultFilter returns a basic_auth HTTP filter from the provided IR HTTPRoute.
func buildHCMFaultFilter() (*hcmv3.HttpFilter, error) {
	faultProto := &xdshttpfaultv3.HTTPFault{}

	if err := faultProto.ValidateAll(); err != nil {
		return nil, err
	}

	faultAny, err := anypb.New(faultProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: faultFilter,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: faultAny,
		},
	}, nil
}

// listenerContainsFault returns true if Fault exists for the provided listener.
func listenerContainsFault(irListener *ir.HTTPListener) bool {
	for _, route := range irListener.Routes {
		if routeContainsFault(route) {
			return true
		}
	}
	return false
}

// routeContainsFault returns true if Fault exists for the provided route.
func routeContainsFault(irRoute *ir.HTTPRoute) bool {
	if irRoute != nil && irRoute.FaultInjection != nil {
		return true
	}
	return false
}

func (*fault) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// patchRoute patches the provided route with the fault config if applicable.
// Note: this method enables the corresponding fault filter for the provided route.
func (*fault) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.FaultInjection == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[wellknown.Fault]; ok {
		// This should not happen since this is the only place where the fault
		// filter is added in a route.
		return fmt.Errorf("route already contains fault config: %+v", route)
	}

	routeCfgProto := &xdshttpfaultv3.HTTPFault{}

	if irRoute.FaultInjection.Delay != nil {
		routeCfgProto.Delay = &xdsfault.FaultDelay{}
		if irRoute.FaultInjection.Delay.Percentage != nil {
			routeCfgProto.Delay.Percentage = translatePercentToFractionalPercent(irRoute.FaultInjection.Delay.Percentage)
		}
		if irRoute.FaultInjection.Delay.FixedDelay != nil {
			routeCfgProto.Delay.FaultDelaySecifier = &xdsfault.FaultDelay_FixedDelay{
				FixedDelay: durationpb.New(irRoute.FaultInjection.Delay.FixedDelay.Duration),
			}
		}
	}

	if irRoute.FaultInjection.Abort != nil {
		routeCfgProto.Abort = &xdshttpfaultv3.FaultAbort{}
		if irRoute.FaultInjection.Abort.Percentage != nil {
			routeCfgProto.Abort.Percentage = translatePercentToFractionalPercent(irRoute.FaultInjection.Abort.Percentage)
		}
		if irRoute.FaultInjection.Abort.HTTPStatus != nil {
			routeCfgProto.Abort.ErrorType = &xdshttpfaultv3.FaultAbort_HttpStatus{
				HttpStatus: uint32(*irRoute.FaultInjection.Abort.HTTPStatus),
			}
		}
		if irRoute.FaultInjection.Abort.GrpcStatus != nil {
			routeCfgProto.Abort.ErrorType = &xdshttpfaultv3.FaultAbort_GrpcStatus{
				GrpcStatus: uint32(*irRoute.FaultInjection.Abort.GrpcStatus),
			}
		}

	}

	if routeCfgProto.Delay == nil && routeCfgProto.Abort == nil {
		return nil
	}

	routeCfgAny, err := anypb.New(routeCfgProto)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[wellknown.Fault] = routeCfgAny

	return nil

}

// translatePercentToFractionalPercent translates an v1alpha3 Percent instance
// to an envoy.type.FractionalPercent instance.
func translatePercentToFractionalPercent(p *float32) *xdstype.FractionalPercent {
	return &xdstype.FractionalPercent{
		Numerator:   uint32(*p * 10000),
		Denominator: xdstype.FractionalPercent_MILLION,
	}
}
