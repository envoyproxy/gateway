// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// Helper functions to get LastTransitionTime for different resource types
func getConditionTimesForGatewayClass(obj interface{}) []metav1.Time {
	gc := obj.(*gwapiv1.GatewayClass)
	times := make([]metav1.Time, len(gc.Status.Conditions))
	for i, cond := range gc.Status.Conditions {
		times[i] = cond.LastTransitionTime
	}
	return times
}

func getConditionTimesForGateway(obj interface{}) []metav1.Time {
	gw := obj.(*gwapiv1.Gateway)
	times := make([]metav1.Time, len(gw.Status.Conditions))
	for i, cond := range gw.Status.Conditions {
		times[i] = cond.LastTransitionTime
	}
	return times
}

func getConditionTimesForEnvoyPatchPolicy(obj interface{}) []metav1.Time {
	epp := obj.(*egv1a1.EnvoyPatchPolicy)
	var times []metav1.Time
	for _, ancestor := range epp.Status.Ancestors {
		for _, cond := range ancestor.Conditions {
			times = append(times, cond.LastTransitionTime)
		}
	}
	return times
}

func getConditionTimesForHTTPRoute(obj interface{}) []metav1.Time {
	hr := obj.(*gwapiv1.HTTPRoute)
	var times []metav1.Time
	for _, parent := range hr.Status.Parents {
		for _, cond := range parent.Conditions {
			times = append(times, cond.LastTransitionTime)
		}
	}
	return times
}

func getConditionTimesForTLSRoute(obj interface{}) []metav1.Time {
	tr := obj.(*gwapiv1a3.TLSRoute)
	var times []metav1.Time
	for _, parent := range tr.Status.Parents {
		for _, cond := range parent.Conditions {
			times = append(times, cond.LastTransitionTime)
		}
	}
	return times
}

func getConditionTimesForBackendTLSPolicy(obj interface{}) []metav1.Time {
	btp := obj.(*gwapiv1.BackendTLSPolicy)
	var times []metav1.Time
	for _, ancestor := range btp.Status.Ancestors {
		for _, cond := range ancestor.Conditions {
			times = append(times, cond.LastTransitionTime)
		}
	}
	return times
}

func TestSetLastTransitionTimeInStatusConditions(t *testing.T) {
	timestamp := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	now := metav1.NewTime(timestamp)

	tests := []struct {
		name              string
		obj               interface{}
		getConditionTimes func(interface{}) []metav1.Time
		expectTimes       bool
		expectedError     error
	}{
		{
			name: "GatewayClass with conditions",
			obj: &gwapiv1.GatewayClass{
				Status: gwapiv1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
						{Type: "Scheduled", Status: metav1.ConditionFalse},
					},
				},
			},
			getConditionTimes: getConditionTimesForGatewayClass,
			expectTimes:       true,
			expectedError:     nil,
		},
		{
			name: "Gateway with conditions",
			obj: &gwapiv1.Gateway{
				Status: gwapiv1.GatewayStatus{
					Conditions: []metav1.Condition{
						{Type: "Ready", Status: metav1.ConditionTrue},
					},
				},
			},
			getConditionTimes: getConditionTimesForGateway,
			expectTimes:       true,
			expectedError:     nil,
		},
		{
			name: "EnvoyPatchPolicy with ancestor conditions",
			obj: &egv1a1.EnvoyPatchPolicy{
				Status: gwapiv1.PolicyStatus{
					Ancestors: []gwapiv1.PolicyAncestorStatus{
						{
							Conditions: []metav1.Condition{
								{Type: "Ready", Status: metav1.ConditionTrue},
							},
						},
					},
				},
			},
			getConditionTimes: getConditionTimesForEnvoyPatchPolicy,
			expectTimes:       true,
			expectedError:     nil,
		},
		{
			name: "HTTPRoute with parent conditions",
			obj: &gwapiv1.HTTPRoute{
				Status: gwapiv1.HTTPRouteStatus{
					RouteStatus: gwapiv1.RouteStatus{
						Parents: []gwapiv1.RouteParentStatus{
							{
								Conditions: []metav1.Condition{
									{Type: "Accepted", Status: metav1.ConditionTrue},
									{Type: "ResolvedRefs", Status: metav1.ConditionFalse},
								},
							},
							{
								Conditions: []metav1.Condition{
									{Type: "Accepted", Status: metav1.ConditionTrue},
								},
							},
						},
					},
				},
			},
			getConditionTimes: getConditionTimesForHTTPRoute,
			expectTimes:       true,
		},
		{
			name: "TLSRoute with parent conditions",
			obj: &gwapiv1a3.TLSRoute{
				Status: gwapiv1a2.TLSRouteStatus{
					RouteStatus: gwapiv1.RouteStatus{
						Parents: []gwapiv1.RouteParentStatus{
							{
								Conditions: []metav1.Condition{
									{Type: "Accepted", Status: metav1.ConditionTrue},
									{Type: "ResolvedRefs", Status: metav1.ConditionFalse},
								},
							},
						},
					},
				},
			},
			getConditionTimes: getConditionTimesForTLSRoute,
			expectTimes:       true,
		},
		{
			name: "BackendTLSPolicy with ancestor conditions",
			obj: &gwapiv1.BackendTLSPolicy{
				Status: gwapiv1.PolicyStatus{
					Ancestors: []gwapiv1.PolicyAncestorStatus{
						{
							Conditions: []metav1.Condition{
								{Type: "Ready", Status: metav1.ConditionTrue},
								{Type: "ResolvedRefs", Status: metav1.ConditionFalse},
							},
						},
					},
				},
			},
			getConditionTimes: getConditionTimesForBackendTLSPolicy,
			expectTimes:       true,
		},
		{
			name: "Empty conditions",
			obj: &gwapiv1.Gateway{
				Status: gwapiv1.GatewayStatus{
					Conditions: []metav1.Condition{},
				},
			},
			getConditionTimes: getConditionTimesForGateway,
			expectTimes:       false,
			expectedError:     nil,
		},
		{
			name: "Nil conditions",
			obj: &gwapiv1.Gateway{
				Status: gwapiv1.GatewayStatus{},
			},
			getConditionTimes: getConditionTimesForGateway,
			expectTimes:       false,
			expectedError:     nil,
		},
		{
			name: "HTTPRoute with empty parents",
			obj: &gwapiv1.HTTPRoute{
				Status: gwapiv1.HTTPRouteStatus{
					RouteStatus: gwapiv1.RouteStatus{
						Parents: []gwapiv1.RouteParentStatus{},
					},
				},
			},
			getConditionTimes: getConditionTimesForHTTPRoute,
			expectTimes:       false,
			expectedError:     nil,
		},
		{
			name: "HTTPRoute with nil parents",
			obj: &gwapiv1.HTTPRoute{
				Status: gwapiv1.HTTPRouteStatus{},
			},
			getConditionTimes: getConditionTimesForHTTPRoute,
			expectTimes:       false,
			expectedError:     nil,
		},
		{
			name: "Unknown type",
			obj: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"status": map[string]interface{}{
						"conditions": []interface{}{
							map[string]interface{}{
								"type":   "Ready",
								"status": string(metav1.ConditionTrue),
							},
						},
					},
				},
			},
			getConditionTimes: func(obj interface{}) []metav1.Time {
				u := obj.(*unstructured.Unstructured)
				var times []metav1.Time
				if status, ok := u.Object["status"].(map[string]interface{}); ok {
					if conditions, ok := status["conditions"].([]interface{}); ok {
						for _, c := range conditions {
							if condition, ok := c.(map[string]interface{}); ok {
								if t, ok := condition["lastTransitionTime"].(metav1.Time); ok {
									times = append(times, t)
								}
							}
						}
					}
				}
				return times
			},
			expectTimes:   true,
			expectedError: nil,
		},
		{
			name: "Unknown type",
			obj: struct {
				Status struct {
					Conditions []metav1.Condition
				}
			}{},
			getConditionTimes: func(interface{}) []metav1.Time { return nil },
			expectTimes:       false,
			expectedError:     fmt.Errorf("cannot set last transition time for unknown object type: struct { Status struct { Conditions []v1.Condition } }"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newObjWithUpdatedLastTransitionTime, err := setLastTransitionTimeInStatusConditions(tt.obj, timestamp)
			assert.Equal(t, tt.expectedError, err)
			times := tt.getConditionTimes(newObjWithUpdatedLastTransitionTime)
			if tt.expectTimes {
				assert.NotEmpty(t, times, "expected condition times to be set")
				for _, time := range times {
					assert.Equal(t, now, time)
				}
			} else {
				assert.Empty(t, times, "expected no condition times")
			}
		})
	}
}
