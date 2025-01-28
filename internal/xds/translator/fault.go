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

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&fault{})
}

type fault struct{}

var _ httpFilter = &fault{}

// patchHCM builds and appends the fault Filters to the HTTP Connection Manager
// if applicable, and it does not already exist.
// Note: this method creates a fault filter for each route that contains an Fault config.
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
		if existingFilter.Name == egv1a1.EnvoyFilterFault.String() {
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
	faultAny, err := proto.ToAnyWithValidation(faultProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterFault.String(),
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
	if irRoute != nil &&
		irRoute.Traffic != nil &&
		irRoute.Traffic.FaultInjection != nil {
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
	if irRoute.Traffic == nil || irRoute.Traffic.FaultInjection == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[wellknown.Fault]; ok {
		// This should not happen since this is the only place where the fault
		// filter is added in a route.
		return fmt.Errorf("route already contains fault config: %+v", route)
	}

	routeCfgProto := &xdshttpfaultv3.HTTPFault{}

	delay := irRoute.Traffic.FaultInjection.Delay
	if delay != nil {
		routeCfgProto.Delay = &xdsfault.FaultDelay{}
		if delay.Percentage != nil {
			routeCfgProto.Delay.Percentage = translatePercentToFractionalPercent(delay.Percentage)
		}
		if delay.FixedDelay != nil {
			routeCfgProto.Delay.FaultDelaySecifier = &xdsfault.FaultDelay_FixedDelay{
				FixedDelay: durationpb.New(delay.FixedDelay.Duration),
			}
		}
	}

	abort := irRoute.Traffic.FaultInjection.Abort
	if abort != nil {
		routeCfgProto.Abort = &xdshttpfaultv3.FaultAbort{}
		if abort.Percentage != nil {
			routeCfgProto.Abort.Percentage = translatePercentToFractionalPercent(abort.Percentage)
		}
		if abort.HTTPStatus != nil {
			routeCfgProto.Abort.ErrorType = &xdshttpfaultv3.FaultAbort_HttpStatus{
				HttpStatus: uint32(*abort.HTTPStatus),
			}
		}
		if abort.GrpcStatus != nil {
			routeCfgProto.Abort.ErrorType = &xdshttpfaultv3.FaultAbort_GrpcStatus{
				GrpcStatus: uint32(*abort.GrpcStatus),
			}
		}
	}

	if routeCfgProto.Delay == nil && routeCfgProto.Abort == nil {
		return nil
	}

	routeCfgAny, err := proto.ToAnyWithValidation(routeCfgProto)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[wellknown.Fault] = routeCfgAny

	return nil
}

// translatePercentToFractionalPercent translates a v1alpha3 Percent instance
// to an envoy.type.FractionalPercent instance.
func translatePercentToFractionalPercent(p *float32) *xdstype.FractionalPercent {
	return &xdstype.FractionalPercent{
		Numerator:   uint32(*p * 10000),
		Denominator: xdstype.FractionalPercent_MILLION,
	}
}
