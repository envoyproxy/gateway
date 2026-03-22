// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestTruncateRouteParents(t *testing.T) {
	t.Run("no-op when below cap", func(t *testing.T) {
		routeStatus := &gwapiv1.RouteStatus{
			Parents: []gwapiv1.RouteParentStatus{
				makeRouteParent("default", "gw-a", "http", []metav1.Condition{acceptedCondition(7)}),
			},
		}

		TruncateRouteParents(routeStatus, 7)

		require.Len(t, routeStatus.Parents, 1)
		require.Empty(t, aggregatedConditions(routeStatus.Parents[0].Conditions))
	})

	t.Run("no-op when at cap", func(t *testing.T) {
		routeStatus := &gwapiv1.RouteStatus{Parents: make([]gwapiv1.RouteParentStatus, 0, 32)}
		for i := 0; i < 32; i++ {
			routeStatus.Parents = append(routeStatus.Parents, makeRouteParent("default", fmt.Sprintf("gw-%02d", i), "http", []metav1.Condition{acceptedCondition(9), resolvedRefsCondition(9)}))
		}

		TruncateRouteParents(routeStatus, 9)

		require.Len(t, routeStatus.Parents, 32)
		for _, parent := range routeStatus.Parents {
			require.Empty(t, aggregatedConditions(parent.Conditions))
		}
	})

	t.Run("truncates to 32, prioritizes failures, and annotates last retained parent", func(t *testing.T) {
		routeStatus := &gwapiv1.RouteStatus{Parents: make([]gwapiv1.RouteParentStatus, 0, 35)}
		for i := 0; i < 31; i++ {
			routeStatus.Parents = append(routeStatus.Parents, makeRouteParent("default", fmt.Sprintf("ok-%02d", i), "http", []metav1.Condition{acceptedCondition(11), resolvedRefsCondition(11)}))
		}
		routeStatus.Parents = append(routeStatus.Parents,
			makeRouteParent("default", "unresolved-z", "http", []metav1.Condition{acceptedCondition(11), unresolvedRefsCondition(11)}),
			makeRouteParent("default", "rejected-z", "http", []metav1.Condition{rejectedCondition(11), resolvedRefsCondition(11)}),
			makeRouteParent("default", "zz-ignored-ok-1", "http", []metav1.Condition{acceptedCondition(11), resolvedRefsCondition(11)}),
			makeRouteParent("default", "zz-ignored-ok-2", "http", []metav1.Condition{acceptedCondition(11), resolvedRefsCondition(11)}),
		)

		TruncateRouteParents(routeStatus, 11)

		require.Len(t, routeStatus.Parents, 32)
		require.Equal(t, "rejected-z", string(routeStatus.Parents[0].ParentRef.Name))
		require.Equal(t, "unresolved-z", string(routeStatus.Parents[1].ParentRef.Name))
		require.Equal(t, "ok-00", string(routeStatus.Parents[2].ParentRef.Name))
		require.Equal(t, "ok-29", string(routeStatus.Parents[31].ParentRef.Name))

		lastAggregated := aggregatedConditions(routeStatus.Parents[31].Conditions)
		require.Len(t, lastAggregated, 1)
		require.Equal(t, string(routeConditionAggregated), lastAggregated[0].Type)
		require.Equal(t, metav1.ConditionTrue, lastAggregated[0].Status)
		require.Equal(t, string(routeReasonAggregated), lastAggregated[0].Reason)
		require.Equal(t, int64(11), lastAggregated[0].ObservedGeneration)
		require.Equal(t, "Parents have been truncated because the number of route parents exceeds 32.", lastAggregated[0].Message)

		for i := 0; i < 31; i++ {
			require.Empty(t, aggregatedConditions(routeStatus.Parents[i].Conditions))
		}
	})

	t.Run("sorts ties deterministically", func(t *testing.T) {
		routeStatus := &gwapiv1.RouteStatus{
			Parents: []gwapiv1.RouteParentStatus{
				makeRouteParent("b", "gw-b", "b", []metav1.Condition{acceptedCondition(13)}),
				makeRouteParent("a", "gw-b", "a", []metav1.Condition{acceptedCondition(13)}),
				makeRouteParent("a", "gw-a", "b", []metav1.Condition{acceptedCondition(13)}),
				makeRouteParent("a", "gw-a", "a", []metav1.Condition{acceptedCondition(13)}),
			},
		}
		for i := 0; i < 29; i++ {
			routeStatus.Parents = append(routeStatus.Parents, makeRouteParent("z", fmt.Sprintf("gw-z-%02d", i), "http", []metav1.Condition{acceptedCondition(13)}))
		}

		TruncateRouteParents(routeStatus, 13)

		require.Equal(t, "a", string(*routeStatus.Parents[0].ParentRef.Namespace))
		require.Equal(t, "gw-a", string(routeStatus.Parents[0].ParentRef.Name))
		require.Equal(t, "a", string(*routeStatus.Parents[0].ParentRef.SectionName))
		require.Equal(t, "a", string(*routeStatus.Parents[1].ParentRef.Namespace))
		require.Equal(t, "gw-a", string(routeStatus.Parents[1].ParentRef.Name))
		require.Equal(t, "b", string(*routeStatus.Parents[1].ParentRef.SectionName))
		require.Equal(t, "a", string(*routeStatus.Parents[2].ParentRef.Namespace))
		require.Equal(t, "gw-b", string(routeStatus.Parents[2].ParentRef.Name))
	})

	t.Run("sorts ties by port before truncating", func(t *testing.T) {
		routeStatus := &gwapiv1.RouteStatus{Parents: make([]gwapiv1.RouteParentStatus, 0, 34)}
		for i := 0; i < 31; i++ {
			routeStatus.Parents = append(routeStatus.Parents, makeRouteParent("z", fmt.Sprintf("gw-z-%02d", i), "http", []metav1.Condition{acceptedCondition(17)}))
		}
		routeStatus.Parents = append(routeStatus.Parents,
			makeRouteParentWithPort("default", "gw-port", "http", 8443, []metav1.Condition{acceptedCondition(17)}),
			makeRouteParentWithPort("default", "gw-port", "http", 443, []metav1.Condition{acceptedCondition(17)}),
			makeRouteParent("zz", "gw-tail", "http", []metav1.Condition{acceptedCondition(17)}),
		)

		TruncateRouteParents(routeStatus, 17)

		require.Len(t, routeStatus.Parents, 32)
		require.Equal(t, "default", string(*routeStatus.Parents[0].ParentRef.Namespace))
		require.Equal(t, "gw-port", string(routeStatus.Parents[0].ParentRef.Name))
		require.Equal(t, gwapiv1.PortNumber(443), *routeStatus.Parents[0].ParentRef.Port)
		require.Equal(t, gwapiv1.PortNumber(8443), *routeStatus.Parents[1].ParentRef.Port)
		require.Equal(t, "gw-z-29", string(routeStatus.Parents[31].ParentRef.Name))
	})
}

func makeRouteParent(namespace, name, section string, conditions []metav1.Condition) gwapiv1.RouteParentStatus {
	return makeRouteParentWithPort(namespace, name, section, 0, conditions)
}

func makeRouteParentWithPort(namespace, name, section string, port gwapiv1.PortNumber, conditions []metav1.Condition) gwapiv1.RouteParentStatus {
	group := gwapiv1.Group(gwapiv1.GroupVersion.Group)
	kind := gwapiv1.Kind("Gateway")
	ns := gwapiv1.Namespace(namespace)
	sectionName := gwapiv1.SectionName(section)
	var portPtr *gwapiv1.PortNumber
	if port != 0 {
		portCopy := port
		portPtr = &portCopy
	}

	return gwapiv1.RouteParentStatus{
		ParentRef: gwapiv1.ParentReference{
			Group:       &group,
			Kind:        &kind,
			Name:        gwapiv1.ObjectName(name),
			Namespace:   &ns,
			SectionName: &sectionName,
			Port:        portPtr,
		},
		ControllerName: gwapiv1.GatewayController("example.com/gateway"),
		Conditions:     conditions,
	}
}

func acceptedCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               string(gwapiv1.RouteConditionAccepted),
		Status:             metav1.ConditionTrue,
		Reason:             string(gwapiv1.RouteReasonAccepted),
		ObservedGeneration: generation,
	}
}

func rejectedCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               string(gwapiv1.RouteConditionAccepted),
		Status:             metav1.ConditionFalse,
		Reason:             string(gwapiv1.RouteReasonNotAllowedByListeners),
		ObservedGeneration: generation,
	}
}

func resolvedRefsCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               string(gwapiv1.RouteConditionResolvedRefs),
		Status:             metav1.ConditionTrue,
		Reason:             string(gwapiv1.RouteReasonResolvedRefs),
		ObservedGeneration: generation,
	}
}

func unresolvedRefsCondition(generation int64) metav1.Condition {
	return metav1.Condition{
		Type:               string(gwapiv1.RouteConditionResolvedRefs),
		Status:             metav1.ConditionFalse,
		Reason:             string(gwapiv1.RouteReasonInvalidKind),
		ObservedGeneration: generation,
	}
}

func aggregatedConditions(conditions []metav1.Condition) []metav1.Condition {
	var aggregated []metav1.Condition
	for _, condition := range conditions {
		if condition.Type == string(routeConditionAggregated) {
			aggregated = append(aggregated, condition)
		}
	}
	return aggregated
}
