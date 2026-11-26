// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestTypedErrorCollectorTypesOrder(t *testing.T) {
	added := []gwapiv1.RouteConditionType{
		"Zeta",
		RouteConditionBackendsAvailable,
		gwapiv1.RouteConditionPartiallyInvalid,
		gwapiv1.RouteConditionResolvedRefs,
		"Alpha",
		gwapiv1.RouteConditionAccepted,
	}
	want := []gwapiv1.RouteConditionType{
		gwapiv1.RouteConditionAccepted,
		gwapiv1.RouteConditionResolvedRefs,
		gwapiv1.RouteConditionPartiallyInvalid,
		"Alpha",
		RouteConditionBackendsAvailable,
		"Zeta",
	}

	// Map iteration order is randomized, so check the order is stable across many collectors.
	for range 100 {
		c := &TypedErrorCollector{}
		for _, ct := range added {
			c.Add(NewRouteStatusError(errors.New(string(ct)), "Reason").WithType(ct))
		}
		require.Equal(t, want, c.Types())

		errs := c.GetAllErrors()
		got := make([]gwapiv1.RouteConditionType, 0, len(errs))
		for _, err := range errs {
			got = append(got, err.Type())
		}
		require.Equal(t, want, got)
	}
}

func TestTypedErrorCollectorTypesEmpty(t *testing.T) {
	require.Nil(t, (&TypedErrorCollector{}).Types())
}

func TestTypedErrorCollectorAddFrom(t *testing.T) {
	headless := NewRouteStatusError(errors.New("service default/headless has no endpoints"), RouteReasonEndpointsNotFound).WithType(RouteConditionBackendsAvailable)
	noEndpoints := NewRouteStatusError(errors.New("service default/noeps has no endpoints"), RouteReasonEndpointsNotFound).WithType(RouteConditionBackendsAvailable)
	invalid := NewRouteStatusError(errors.New("invalid filter"), gwapiv1.RouteReasonUnsupportedValue).WithType(gwapiv1.RouteConditionAccepted)

	first := &TypedErrorCollector{}
	first.Add(headless)
	first.Add(invalid)
	first.Add(invalid)
	second := &TypedErrorCollector{}
	second.Add(headless)
	second.Add(noEndpoints)
	second.Add(invalid)

	merged := &TypedErrorCollector{}
	merged.AddFrom(first)
	merged.AddFrom(second)
	merged.AddFrom(nil)

	// Errors already held are skipped, while duplicates within one translation stay as they were.
	require.Equal(t, "Service default/headless has no endpoints.\nService default/noeps has no endpoints.",
		merged.GetError(RouteConditionBackendsAvailable).Error())
	require.Equal(t, "Invalid filter.\nInvalid filter.", merged.GetError(gwapiv1.RouteConditionAccepted).Error())
}
