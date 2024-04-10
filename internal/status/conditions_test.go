// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions_test.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package status

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilclock "k8s.io/utils/clock"
	fakeclock "k8s.io/utils/clock/testing"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var clock utilclock.Clock = utilclock.RealClock{}

func TestComputeGatewayClassAcceptedCondition(t *testing.T) {
	testCases := []struct {
		name     string
		accepted bool
		expect   metav1.Condition
	}{
		{
			name:     "accepted gatewayclass",
			accepted: true,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionTrue,
				Reason:  string(gwapiv1.GatewayClassReasonAccepted),
				Message: MsgValidGatewayClass,
			},
		},
		{
			name:     "not accepted gatewayclass",
			accepted: false,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionFalse,
				Reason:  string(ReasonOlderGatewayClassExists),
				Message: MsgOlderGatewayClassExists,
			},
		},
		{
			name:     "invalid parameters gatewayclass",
			accepted: false,
			expect: metav1.Condition{
				Type:    string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:  metav1.ConditionFalse,
				Reason:  string(gwapiv1.GatewayClassReasonInvalidParameters),
				Message: MsgGatewayClassInvalidParams,
			},
		},
	}

	for _, tc := range testCases {
		gc := &gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 7,
			},
		}

		got := computeGatewayClassAcceptedCondition(gc, tc.accepted, tc.expect.Reason, tc.expect.Message)

		assert.Equal(t, tc.expect.Type, got.Type)
		assert.Equal(t, tc.expect.Status, got.Status)
		assert.Equal(t, tc.expect.Reason, got.Reason)
		assert.Equal(t, gc.Generation, got.ObservedGeneration)
	}
}

func TestComputeGatewayScheduledCondition(t *testing.T) {
	testCases := []struct {
		name   string
		sched  bool
		expect metav1.Condition
	}{
		{
			name:  "scheduled gateway",
			sched: true,
			expect: metav1.Condition{
				Type:   string(gwapiv1.GatewayReasonAccepted),
				Status: metav1.ConditionTrue,
			},
		},
		{
			name:  "not scheduled gateway",
			sched: false,
			expect: metav1.Condition{
				Type:   string(gwapiv1.GatewayReasonAccepted),
				Status: metav1.ConditionFalse,
			},
		},
	}

	for _, tc := range testCases {
		gw := &gwapiv1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "test",
			},
		}

		got := computeGatewayAcceptedCondition(gw, tc.sched)

		assert.Equal(t, tc.expect.Type, got.Type)
		assert.Equal(t, tc.expect.Status, got.Status)
	}
}

func TestConditionChanged(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
		a, b     metav1.Condition
	}{
		{
			name:     "nil and non-nil current are equal",
			expected: false,
			a:        metav1.Condition{},
		},
		{
			name:     "empty slices should be equal",
			expected: false,
			a:        metav1.Condition{},
			b:        metav1.Condition{},
		},
		{
			name:     "condition LastTransitionTime should be ignored",
			expected: false,
			a: metav1.Condition{
				Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Unix(0, 0),
			},
			b: metav1.Condition{
				Type:               string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Unix(1, 0),
			},
		},
		{
			name:     "check condition reason differs",
			expected: true,
			a: metav1.Condition{
				Type:   string(gwapiv1.GatewayConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: "foo",
			},
			b: metav1.Condition{
				Type:   string(gwapiv1.GatewayConditionProgrammed),
				Status: metav1.ConditionFalse,
				Reason: "bar",
			},
		},
		{
			name:     "condition status differs",
			expected: true,
			a: metav1.Condition{
				Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status: metav1.ConditionTrue,
			},
			b: metav1.Condition{
				Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
				Status: metav1.ConditionFalse,
			},
		},
	}

	for _, tc := range testCases {
		if got := conditionChanged(tc.a, tc.b); got != tc.expected {
			assert.Equal(t, tc.expected, got, tc.name)
		}
	}
}

func TestMergeConditions(t *testing.T) {
	// Inject a fake clock and don't forget to reset it
	fakeClock := fakeclock.NewFakeClock(time.Time{})
	clock = fakeClock
	defer func() {
		clock = utilclock.RealClock{}
	}()

	start := fakeClock.Now()
	middle := start.Add(1 * time.Minute)
	later := start.Add(2 * time.Minute)

	gen := int64(1)

	testCases := []struct {
		name     string
		current  []metav1.Condition
		updates  []metav1.Condition
		expected []metav1.Condition
	}{
		{
			name: "status updated",
			current: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
			updates: []metav1.Condition{
				newCondition("available", "true", "Reason", "Message", middle, gen),
			},
			expected: []metav1.Condition{
				newCondition("available", "true", "Reason", "Message", middle, gen),
			},
		},
		{
			name: "reason updated",
			current: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
			updates: []metav1.Condition{
				newCondition("available", "false", "New Reason", "Message", middle, gen),
			},
			expected: []metav1.Condition{
				newCondition("available", "false", "New Reason", "Message", middle, gen),
			},
		},
		{
			name: "message updated",
			current: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
			updates: []metav1.Condition{
				newCondition("available", "false", "Reason", "New Message", middle, gen),
			},
			expected: []metav1.Condition{
				newCondition("available", "false", "Reason", "New Message", middle, gen),
			},
		},
		{
			name: "observed generation updated",
			current: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
			updates: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", middle, gen+1),
			},
			expected: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", middle, gen+1),
			},
		},
		{
			name: "status unchanged",
			current: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
			updates: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", middle, gen),
			},
			expected: []metav1.Condition{
				newCondition("available", "false", "Reason", "Message", start, gen),
			},
		},
	}

	// Simulate the passage of time between original condition creation
	// and update processing
	fakeClock.SetTime(later)

	for _, tc := range testCases {
		got := MergeConditions(tc.current, tc.updates...)
		assert.ElementsMatch(t, tc.expected, got, tc.name)
	}
}

func TestGatewayReadyCondition(t *testing.T) {
	testCases := []struct {
		name             string
		serviceAddress   bool
		deploymentStatus appsv1.DeploymentStatus
		expect           metav1.Condition
	}{
		{
			name:             "ready gateway",
			serviceAddress:   true,
			deploymentStatus: appsv1.DeploymentStatus{AvailableReplicas: 1},
			expect: metav1.Condition{
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1.GatewayConditionProgrammed),
			},
		},
		{
			name:             "not ready gateway without address",
			serviceAddress:   false,
			deploymentStatus: appsv1.DeploymentStatus{AvailableReplicas: 1},
			expect: metav1.Condition{
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.GatewayReasonAddressNotAssigned),
			},
		},
		{
			name:             "not ready gateway with address unavailable pods",
			serviceAddress:   true,
			deploymentStatus: appsv1.DeploymentStatus{AvailableReplicas: 0},
			expect: metav1.Condition{
				Status: metav1.ConditionFalse,
				Reason: string(gwapiv1.GatewayReasonNoResources),
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			gtw := &gwapiv1.Gateway{}
			if tc.serviceAddress {
				gtw.Status = gwapiv1.GatewayStatus{
					Addresses: []gwapiv1.GatewayStatusAddress{
						{
							Type:  ptr.To(gwapiv1.IPAddressType),
							Value: "1.1.1.1",
						},
					},
				}
			}

			deployment := &appsv1.Deployment{Status: tc.deploymentStatus}
			got := computeGatewayProgrammedCondition(gtw, deployment)

			assert.Equal(t, string(gwapiv1.GatewayConditionProgrammed), got.Type)
			assert.Equal(t, tc.expect.Status, got.Status)
			assert.Equal(t, tc.expect.Reason, got.Reason)
		})
	}
}

func TestError2ConditionMsg(t *testing.T) {
	testCases := []struct {
		name   string
		err    error
		expect string
	}{
		{
			name:   "nil error",
			err:    nil,
			expect: "",
		},
		{
			name:   "error with message",
			err:    errors.New("something is wrong"),
			expect: "Something is wrong",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.expect, Error2ConditionMsg(tt.err), "Error2ConditionMsg(%v)", tt.err)
		})
	}
}
