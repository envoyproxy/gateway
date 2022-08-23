// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/status/gatewayclassconditions_test.go

package status

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilclock "k8s.io/utils/clock"
	fakeclock "k8s.io/utils/clock/testing"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
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
				Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
				Status: metav1.ConditionTrue,
				Reason: string(gwapiv1b1.GatewayClassReasonAccepted),
			},
		},
		{
			name:     "not accepted gatewayclass",
			accepted: false,
			expect: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
				Status: metav1.ConditionFalse,
				Reason: string(ReasonOlderGatewayClassExists),
			},
		},
	}

	for _, tc := range testCases {
		gc := &gwapiv1b1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Generation: 7,
			},
		}

		got := computeGatewayClassAcceptedCondition(gc, tc.accepted)

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
				Type:   string(gwapiv1b1.GatewayConditionScheduled),
				Status: metav1.ConditionTrue,
			},
		},
		{
			name:  "not scheduled gateway",
			sched: false,
			expect: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayConditionScheduled),
				Status: metav1.ConditionFalse,
			},
		},
	}

	for _, tc := range testCases {
		gw := &gwapiv1b1.Gateway{
			ObjectMeta: metav1.ObjectMeta{
				Namespace: "test",
				Name:      "test",
			},
		}

		got := computeGatewayScheduledCondition(gw, tc.sched)

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
				Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Unix(0, 0),
			},
			b: metav1.Condition{
				Type:               string(gwapiv1b1.GatewayClassConditionStatusAccepted),
				Status:             metav1.ConditionTrue,
				LastTransitionTime: metav1.Unix(1, 0),
			},
		},
		{
			name:     "check condition reason differs",
			expected: true,
			a: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayConditionReady),
				Status: metav1.ConditionFalse,
				Reason: "foo",
			},
			b: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayConditionReady),
				Status: metav1.ConditionFalse,
				Reason: "bar",
			},
		},
		{
			name:     "condition status differs",
			expected: true,
			a: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
				Status: metav1.ConditionTrue,
			},
			b: metav1.Condition{
				Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
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
				newCondition("available", "true", "Reason", "Message", later, gen),
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
				newCondition("available", "false", "New Reason", "Message", start, gen),
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
				newCondition("available", "false", "Reason", "New Message", start, gen),
			},
		},
	}

	// Simulate the passage of time between original condition creation
	// and update processing
	fakeClock.SetTime(later)

	for _, tc := range testCases {
		got := mergeConditions(tc.current, tc.updates...)
		if conditionChanged(tc.expected[0], got[0]) {
			assert.Equal(t, tc.expected, got, tc.name)
		}
	}
}
