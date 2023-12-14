// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func TestGatewaysOfClass(t *testing.T) {
	gc := &gwapiv1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
	}
	testCases := []struct {
		name   string
		gws    []gwapiv1.Gateway
		expect int
	}{
		{
			name: "no matching gateways",
			gws: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName("no-match"),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName("no-match2"),
					},
				},
			},
			expect: 0,
		},
		{
			name: "one of two matching gateways",
			gws: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName("no-match"),
					},
				},
			},
			expect: 1,
		},
		{
			name: "two of two matching gateways",
			gws: []gwapiv1.Gateway{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName(gc.Name),
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test2",
						Namespace: "test",
					},
					Spec: gwapiv1.GatewaySpec{
						GatewayClassName: gwapiv1.ObjectName(gc.Name),
					},
				},
			},
			expect: 2,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			gwList := &gwapiv1.GatewayList{Items: tc.gws}
			actual := gatewaysOfClass(gc, gwList)
			require.Equal(t, tc.expect, len(actual))
		})
	}
}

func TestIsGatewayClassAccepted(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1.GatewayClass
		expect bool
	}{
		{
			name: "gatewayclass accepted condition",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
				},
				Status: gwapiv1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expect: true,
		},
		{
			name: "gatewayclass not accepted condition",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
				},
				Status: gwapiv1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expect: false,
		},
		{
			name: "no gatewayclass accepted condition type",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gwapiv1.GatewayController(egv1a1.GatewayControllerName),
				},
				Status: gwapiv1.GatewayClassStatus{
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
	createGatewayClass := func(name string, creationTime time.Time) *gwapiv1.GatewayClass {
		return &gwapiv1.GatewayClass{
			ObjectMeta: metav1.ObjectMeta{
				Name:              name,
				CreationTimestamp: metav1.NewTime(creationTime),
			},
			Spec: gwapiv1.GatewayClassSpec{
				ControllerName: egv1a1.GatewayControllerName,
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

func TestRefsEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name   string
		gc     *gwapiv1.GatewayClass
		expect bool
	}{
		{
			name:   "nil gatewayclass",
			gc:     nil,
			expect: false,
		},
		{
			name: "valid envoyproxy parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expect: true,
		},
		{
			name: "unspecified parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
				},
			},
			expect: false,
		},
		{
			name: "unsupported group parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group("Unsupported"),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expect: false,
		},
		{
			name: "unsupported group parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind("Unsupported"),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expect: false,
		},
		{
			name: "unsupported group parameters ref",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind("Unsupported"),
						Name:      "test",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expect: false,
		},
		{
			name: "empty parameters ref name",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			expect: false,
		},
		{
			name: "unspecified parameters ref namespace",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: "test",
					ParametersRef: &gwapiv1.ParametersReference{
						Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:  "test",
					},
				},
			},
			expect: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			result := refsEnvoyProxy(tc.gc)
			require.Equal(t, tc.expect, result)
		})
	}
}

func TestClassRefsEnvoyProxy(t *testing.T) {
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)

	testCases := []struct {
		name     string
		gc       *gwapiv1.GatewayClass
		ep       *egv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil gatewayclass and envoyproxy",
			gc:       nil,
			ep:       nil,
			expected: false,
		},
		{
			name: "gatewayclass references envoyproxy",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test-ep",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-ep",
				},
			},
			expected: true,
		},
		{
			name: "gatewayclass does not reference envoyproxy",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "not-test-ep",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-ep",
				},
			},
			expected: false,
		},
		{
			name: "gatewayclass references invalid kind",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:      gwapiv1.Kind("UnsupportedKind"),
						Name:      "test-ep",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-ep",
				},
			},
			expected: false,
		},
		{
			name: "gatewayclass references invalid group",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group:     gwapiv1.Group("UnsupportedGroup"),
						Kind:      gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:      "test-ep",
						Namespace: gatewayapi.NamespacePtr(config.DefaultNamespace),
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-ep",
				},
			},
			expected: false,
		},
		{
			name: "gatewayclass references envoyproxy without namespace",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
					ParametersRef: &gwapiv1.ParametersReference{
						Group: gwapiv1.Group(egv1a1.GroupVersion.Group),
						Kind:  gwapiv1.Kind(egv1a1.KindEnvoyProxy),
						Name:  "test-ep",
					},
				},
			},
			ep: &egv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.DefaultNamespace,
					Name:      "test-ep",
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Process the test case objects.
			res := classRefsEnvoyProxy(tc.gc, tc.ep)
			require.Equal(t, tc.expected, res)
		})
	}
}

func TestClassAccepted(t *testing.T) {
	gcCtrlName := gwapiv1.GatewayController(egv1a1.GatewayControllerName)

	testCases := []struct {
		name     string
		gc       *gwapiv1.GatewayClass
		expected bool
	}{
		{
			name:     "nil gatewayclass",
			gc:       nil,
			expected: false,
		},
		{
			name: "gatewayclass accepted",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
				},
				Status: gwapiv1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionTrue,
						},
					},
				},
			},
			expected: true,
		},
		{
			name: "gatewayclass not accepted",
			gc: &gwapiv1.GatewayClass{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-gc",
				},
				Spec: gwapiv1.GatewayClassSpec{
					ControllerName: gcCtrlName,
				},
				Status: gwapiv1.GatewayClassStatus{
					Conditions: []metav1.Condition{
						{
							Type:   string(gwapiv1.GatewayClassConditionStatusAccepted),
							Status: metav1.ConditionFalse,
						},
					},
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]

		// Run the test cases.
		t.Run(tc.name, func(t *testing.T) {
			// Process the test case objects.
			res := classAccepted(tc.gc)
			require.Equal(t, tc.expected, res)
		})
	}
}
