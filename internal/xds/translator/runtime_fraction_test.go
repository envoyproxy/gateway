// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestApplyRuntimeFractionToRouteMatch(t *testing.T) {
	t.Run("sets runtime fraction when absent", func(t *testing.T) {
		route := &routev3.Route{Match: &routev3.RouteMatch{}}

		applyRuntimeFractionToRouteMatch(route, &gwapiv1.Fraction{
			Numerator:   30,
			Denominator: ptr.To[int32](100),
		})

		rf := route.GetMatch().GetRuntimeFraction()
		require.NotNil(t, rf)
		require.Equal(t, uint32(30), rf.DefaultValue.Numerator)
		require.Equal(t, xdstype.FractionalPercent_HUNDRED, rf.DefaultValue.Denominator)
	})

	t.Run("keeps tighter existing runtime fraction", func(t *testing.T) {
		route := &routev3.Route{
			Match: &routev3.RouteMatch{
				RuntimeFraction: &corev3.RuntimeFractionalPercent{
					DefaultValue: &xdstype.FractionalPercent{
						Numerator:   100000, // 10%
						Denominator: xdstype.FractionalPercent_MILLION,
					},
				},
			},
		}

		applyRuntimeFractionToRouteMatch(route, &gwapiv1.Fraction{
			Numerator:   20,
			Denominator: ptr.To[int32](100), // 20%
		})

		rf := route.GetMatch().GetRuntimeFraction()
		require.NotNil(t, rf)
		require.Equal(t, uint32(100000), rf.DefaultValue.Numerator)
		require.Equal(t, xdstype.FractionalPercent_MILLION, rf.DefaultValue.Denominator)
	})

	t.Run("tightens looser existing runtime fraction", func(t *testing.T) {
		route := &routev3.Route{
			Match: &routev3.RouteMatch{
				RuntimeFraction: &corev3.RuntimeFractionalPercent{
					DefaultValue: &xdstype.FractionalPercent{
						Numerator:   400000, // 40%
						Denominator: xdstype.FractionalPercent_MILLION,
					},
				},
			},
		}

		applyRuntimeFractionToRouteMatch(route, &gwapiv1.Fraction{
			Numerator:   25,
			Denominator: ptr.To[int32](100), // 25%
		})

		rf := route.GetMatch().GetRuntimeFraction()
		require.NotNil(t, rf)
		require.Equal(t, uint32(25), rf.DefaultValue.Numerator)
		require.Equal(t, xdstype.FractionalPercent_HUNDRED, rf.DefaultValue.Denominator)
	})
}

func TestPatchRouteAppliesRuntimeFractionForExtensions(t *testing.T) {
	testCases := []struct {
		name    string
		patcher interface {
			patchRoute(*routev3.Route, *ir.HTTPRoute, *ir.HTTPListener) error
		}
		irRoute *ir.HTTPRoute
	}{
		{
			name:    "extproc",
			patcher: &extProc{},
			irRoute: &ir.HTTPRoute{
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					ExtProcs: []ir.ExtProc{{
						Name: "ep",
						Destination: ir.RouteDestination{
							Name: "ep-cluster",
						},
						Percentage: &gwapiv1.Fraction{
							Numerator:   12,
							Denominator: ptr.To[int32](100),
						},
					}},
				},
			},
		},
		{
			name:    "lua",
			patcher: &lua{},
			irRoute: &ir.HTTPRoute{
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Luas: []ir.Lua{{
						Name: "lua",
						Code: ptr.To("function envoy_on_request() end"),
						Percentage: &gwapiv1.Fraction{
							Numerator:   23,
							Denominator: ptr.To[int32](100),
						},
					}},
				},
			},
		},
		{
			name:    "wasm",
			patcher: &wasm{},
			irRoute: &ir.HTTPRoute{
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{
					Wasms: []ir.Wasm{{
						Name: "wasm",
						Code: &ir.HTTPWasmCode{
							ServingURL: "http://example.com/module.wasm",
							SHA256:     "abc123",
						},
						Percentage: &gwapiv1.Fraction{
							Numerator:   34,
							Denominator: ptr.To[int32](100),
						},
					}},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			route := &routev3.Route{Match: &routev3.RouteMatch{}}
			err := tc.patcher.patchRoute(route, tc.irRoute, nil)
			require.NoError(t, err)
			require.NotNil(t, route.GetMatch().GetRuntimeFraction())
		})
	}
}
