// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"
	"time"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestGatewaysOfClass(t *testing.T) {
	gc := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	testCases := []struct {
		name   string
		gws    []gwapiv1b1.Gateway
		expect int
	}{
		{
			name: "no matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match2"),
					},
				},
			},
			expect: 0,
		},
		{
			name: "one of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName("no-match"),
					},
				},
			},
			expect: 1,
		},
		{
			name: "two of two matching gateways",
			gws: []gwapiv1b1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1b1.GatewaySpec{
						GatewayClassName: gwapiv1b1.ObjectName(gc.Name),
					},
				},
			},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gwList := &gwapiv1b1.GatewayList{Items: tc.gws}
			actual := gatewaysOfClass(gc, gwList)
			require.Equal(t, tc.expect, len(actual))
		})
	}
}

func TestIsGatewayClassAccepted(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1b1.GatewayClass
		expect bool
	}{
		{
			name: "gatewayclass accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "gatewayclass not accepted condition",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1b1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no gatewayclass accepted condition type",
			gc: &gwapiv1b1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1b1.GatewayClassSpec{
					ControllerName: gwapiv1b1.GatewayController(v1alpha1.GatewayControllerName),
				},
				Status: gwapiv1b1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   "SomeOtherType",
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: false,
		},
		{
			name:   "nil gatewayclass",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			actual := isAccepted(tc.gc)
			require.Equal(t, tc.expect, actual)
		})
	}
}

func TestGatewayOldestClass(t *testing.T) {
	createGatewayClass := func(name string, creationTime time.Time) *gwapiv1b1.GatewayClass {
		return &gwapiv1b1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: metav1.NewTime(creationTime),
			},
			Spec: gwapiv1b1.GatewayClassSpec{
				ControllerName: v1alpha1.GatewayControllerName,
			},
		}
	}

	currentTime := metav1.Now()
	addDuration := time.Duration(10)
	testCases := []struct {
		name    string
		classes map[string]time.Time
		remove  map[string]time.Time
		oldest  string
	}{
		{
			name: "normal",
			classes: map[string]time.Time{
				"class-b": currentTime.Time,
				"class-a": currentTime.Add(1 * addDuration),
			},
			remove: nil,
			oldest: "class-b",
		},
		{
			name: "tie breaker",
			classes: map[string]time.Time{
				"class-aa": currentTime.Time,
				"class-ab": currentTime.Time,
			},
			remove: nil,
			oldest: "class-aa",
		},
		{
			name: "remove from matched",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
				"class-b": currentTime.Add(1 * addDuration),
				"class-c": currentTime.Add(2 * addDuration),
			},
			remove: map[string]time.Time{
				"class-b": currentTime.Add(1 * addDuration),
			},
			oldest: "class-a",
		},
		{
			name: "remove oldest",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
				"class-b": currentTime.Add(1 * addDuration),
				"class-c": currentTime.Add(2 * addDuration),
			},
			remove: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			oldest: "class-b",
		},
		{
			name: "remove oldest last",
			classes: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			remove: map[string]time.Time{
				"class-a": currentTime.Time,
			},
			oldest: "",
		},
	}

	for _, tc := range testCases {
		var cc controlledClasses
		for name, timestamp := range tc.classes {
			cc.addMatch(createGatewayClass(name, timestamp))
		}

		for name, timestamp := range tc.remove {
			cc.removeMatch(createGatewayClass(name, timestamp))
		}

		if tc.oldest == "" {
			require.Nil(t, cc.oldestClass)
			return
		}

		require.Equal(t, tc.oldest, cc.oldestClass.Name)
	}
}
