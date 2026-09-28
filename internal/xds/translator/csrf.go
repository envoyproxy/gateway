// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	csrfv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/csrf/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
	"github.com/envoyproxy/gateway/internal/xds/utils/fractionalpercent"
)

const csrfFilterName = "envoy.filters.http.csrf"

func init() {
	registerHTTPFilter(&csrf{})
}

type csrf struct{}

var _ httpFilter = &csrf{}

// patchHCM builds and appends the CSRF Filter to the HTTP Connection Manager if applicable.
func (*csrf) patchHCM(
	mgr *hcmv3.HttpConnectionManager,
	irListener *ir.HTTPListener,
) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsCSRF(irListener) {
		return nil
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == csrfFilterName {
			return nil
		}
	}

	csrfFilter, err := buildHCMCSRFFilter()
	if err != nil {
		return err
	}

	// Insert before the router filter.
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{csrfFilter}, mgr.HttpFilters...)

	return nil
}

// buildHCMCSRFFilter returns a CSRF filter for the HCM with 0% enforcement.
// The actual policy is enabled per-route via typed_per_filter_config.
func buildHCMCSRFFilter() (*hcmv3.HttpFilter, error) {
	csrfProto := &csrfv3.CsrfPolicy{
		FilterEnabled: &corev3.RuntimeFractionalPercent{
			DefaultValue: fractionalpercent.FromIn32(0),
		},
	}

	csrfAny, err := anypb.New(csrfProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: csrfFilterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: csrfAny,
		},
	}, nil
}

// listenerContainsCSRF returns true if the provided listener has CSRF
// policies attached to its routes.
func listenerContainsCSRF(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Security != nil && route.Security.CSRF != nil {
			return true
		}
	}

	return false
}

// patchRoute patches the provided route with the CSRF config if applicable.
func (*csrf) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.Security == nil || irRoute.Security.CSRF == nil {
		return nil
	}

	filterCfg := route.GetTypedPerFilterConfig()
	if _, ok := filterCfg[csrfFilterName]; ok {
		return fmt.Errorf("route already contains csrf config: %+v", route)
	}

	routeCfgProto := buildXdsCSRFPolicy(irRoute.Security.CSRF)

	routeCfgAny, err := anypb.New(routeCfgProto)
	if err != nil {
		return err
	}

	if filterCfg == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	route.TypedPerFilterConfig[csrfFilterName] = routeCfgAny

	return nil
}

func (*csrf) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

// buildXdsCSRFPolicy builds the full Envoy CSRF policy from the IR.
func buildXdsCSRFPolicy(csrf *ir.CSRF) *csrfv3.CsrfPolicy {
	// Envoy rolls the dice separately for FilterEnabled and ShadowEnabled on every request.
	// FilterEnabled decides whether the request is enforced. ShadowEnabled only applies to
	// requests that FilterEnabled did not pick, because Envoy checks it as "shadow && !enabled".
	//
	// So FilterEnabled alone carries the fraction, set to the complement of the shadow
	// fraction, while ShadowEnabled stays at 100% to pick up everything FilterEnabled left.
	// A shadow fraction of 40% gives FilterEnabled 60%: 60% of requests are enforced and the
	// other 40% are checked and recorded in the csrf stats but allowed through, so every
	// request lands in one group or the other.
	// With no shadow fraction, all requests are enforced.
	enforcedFraction := fractionalpercent.FromIn32(100)
	var shadowEnabled *corev3.RuntimeFractionalPercent
	if csrf.ShadowFraction != nil {
		enforcedFraction = fractionalpercent.FromFraction(complementFraction(csrf.ShadowFraction))
		shadowEnabled = &corev3.RuntimeFractionalPercent{
			DefaultValue: fractionalpercent.FromIn32(100),
		}
	}

	policy := &csrfv3.CsrfPolicy{
		FilterEnabled: &corev3.RuntimeFractionalPercent{
			DefaultValue: enforcedFraction,
		},
		ShadowEnabled: shadowEnabled,
	}

	for _, origin := range csrf.AdditionalOrigins {
		policy.AdditionalOrigins = append(policy.AdditionalOrigins, buildXdsStringMatcher(origin))
	}

	return policy
}

// complementFraction returns the remainder of the given fraction, i.e. 1 - fraction.
// The numerator is never greater than the denominator, so the result is never negative.
func complementFraction(fraction *gwapiv1.Fraction) *gwapiv1.Fraction {
	denominator := ptr.Deref(fraction.Denominator, fractionalpercent.Hundred)
	return &gwapiv1.Fraction{
		Numerator:   denominator - fraction.Numerator,
		Denominator: new(denominator),
	}
}
