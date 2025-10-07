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
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilclock "k8s.io/utils/clock"
	fakeclock "k8s.io/utils/clock/testing"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

var clock utilclock.Clock = utilclock.RealClock{}

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
		if got := conditionChanged(&tc.a, &tc.b); got != tc.expected {
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

func TestMergeConditionsTruncatesMessages(t *testing.T) {
	longMsg := strings.Repeat("x", conditionMessageMaxLength+5)
	cond := newCondition("available", metav1.ConditionTrue, "Reason", longMsg, time.Now(), 1)
	conditions := MergeConditions(nil, cond)

	if assert.Len(t, conditions, 1) {
		assert.Equal(t, conditionMessageMaxLength, len(conditions[0].Message))
		prefixLen := conditionMessageMaxLength - len(conditionMessageTruncationSuffix)
		expectedPrefix := strings.Repeat("x", prefixLen)
		assert.True(t, strings.HasSuffix(conditions[0].Message, conditionMessageTruncationSuffix))
		assert.Equal(t, expectedPrefix, conditions[0].Message[:prefixLen])
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
			expect: "Something is wrong.",
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.expect, Error2ConditionMsg(tt.err), "Error2ConditionMsg(%v)", tt.err)
		})
	}
}

func TestError2ConditionMsgTruncation(t *testing.T) {
	base := strings.Repeat("a", conditionMessageMaxLength+10)
	got := Error2ConditionMsg(errors.New(base))

	assert.Equal(t, conditionMessageMaxLength, len(got))
	assert.EqualValues(t, 'A', got[0])
	assert.True(t, strings.HasSuffix(got, conditionMessageTruncationSuffix))
}
