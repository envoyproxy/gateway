// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
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
		t.Run(tc.name, func(t *testing.T) {
			gwList := &gwapiv1.GatewayList{Items: tc.gws}
			actual := gatewaysOfClass(gc, gwList)
			require.Len(t, actual, tc.expect)
		})
	}
}

// gatewaysOfClass returns a list of gateways that reference gc from the provided gwList.
func gatewaysOfClass(gc *gwapiv1.GatewayClass, gwList *gwapiv1.GatewayList) []gwapiv1.Gateway {
	var gateways []gwapiv1.Gateway
	if gwList == nil || gc == nil {
		return gateways
	}
	for i := range gwList.Items {
		gw := gwList.Items[i]
		if string(gw.Spec.GatewayClassName) == gc.Name {
			gateways = append(gateways, gw)
		}
	}
	return gateways
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
		t.Run(tc.name, func(t *testing.T) {
			actual := isGatewayClassAccepted(tc.gc)
			require.Equal(t, tc.expect, actual)
		})
	}
}

// isAccepted returns true if the provided gatewayclass contains the Accepted=true
// status condition.
func isGatewayClassAccepted(gc *gwapiv1.GatewayClass) bool {
	if gc == nil {
		return false
	}
	for _, cond := range gc.Status.Conditions {
		if cond.Type == string(gwapiv1.GatewayClassConditionStatusAccepted) && cond.Status == metav1.ConditionTrue {
			return true
		}
	}
	return false
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
			expect: true,
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

func TestTransformConfigMapData(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected map[string]string
	}{
		{
			name:     "nil configmap",
			input:    nil,
			expected: nil,
		},
		{
			name: "non-configmap object",
			input: &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
			},
			expected: nil,
		},
		{
			name: "configmap with single key - no filtering",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					"key1": "value1",
				},
			},
			expected: map[string]string{
				"key1": "value1",
			},
		},
		{
			name: "configmap with cached keys only",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					gatewayapi.JWKSConfigMapKey:         "jwks-data",
					gatewayapi.LuaConfigMapKey:          "lua-data",
					gatewayapi.ResponseBodyConfigMapKey: "response-body-data",
				},
			},
			expected: map[string]string{
				gatewayapi.JWKSConfigMapKey:         "jwks-data",
				gatewayapi.LuaConfigMapKey:          "lua-data",
				gatewayapi.ResponseBodyConfigMapKey: "response-body-data",
			},
		},
		{
			name: "configmap with cached and non-cached keys",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					gatewayapi.JWKSConfigMapKey:         "jwks-data",
					gatewayapi.LuaConfigMapKey:          "lua-data",
					"unwanted-key-1":                    "unwanted-value-1",
					"unwanted-key-2":                    "unwanted-value-2",
					gatewayapi.ResponseBodyConfigMapKey: "response-body-data",
				},
			},
			expected: map[string]string{
				gatewayapi.JWKSConfigMapKey:         "jwks-data",
				gatewayapi.LuaConfigMapKey:          "lua-data",
				gatewayapi.ResponseBodyConfigMapKey: "response-body-data",
				// Note: First key "jwks" is expected, so no fallback key is added
			},
		},
		{
			name: "configmap with only non-cached keys",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					"unwanted-key-1": "unwanted-value-1",
					"unwanted-key-2": "unwanted-value-2",
					"unwanted-key-3": "unwanted-value-3",
				},
			},
			expected: map[string]string{
				"unwanted-key-1": "unwanted-value-1", // first fallback key
			},
		},
		{
			name: "configmap with CACertKey and CRLKey",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					gatewayapi.CACertKey: "ca-cert-data",
					gatewayapi.CRLKey:    "crl-data",
					"other-key":          "other-value",
				},
			},
			expected: map[string]string{
				gatewayapi.CACertKey: "ca-cert-data",
				gatewayapi.CRLKey:    "crl-data",
				// Note: First key "ca.crl" is expected, so no fallback key is added
			},
		},
		{
			name: "configmap with non-expected key first",
			input: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test",
					Namespace: "default",
				},
				Data: map[string]string{
					"unwanted-key":              "unwanted-value",
					gatewayapi.JWKSConfigMapKey: "jwks-data",
					gatewayapi.LuaConfigMapKey:  "lua-data",
					"another-unwanted-key":      "another-value",
				},
			},
			expected: map[string]string{
				"another-unwanted-key":      "another-value", // first key in sorted order, added as fallback
				gatewayapi.JWKSConfigMapKey: "jwks-data",
				gatewayapi.LuaConfigMapKey:  "lua-data",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			transform := composeTransforms(cache.TransformStripManagedFields(), transformConfigMapData)

			result, err := transform(tc.input)
			require.NoError(t, err)

			if tc.expected == nil {
				require.Equal(t, tc.input, result)
				return
			}

			cm, ok := result.(*corev1.ConfigMap)
			require.True(t, ok, "result should be a ConfigMap")
			if tc.expected != nil {
				require.Equal(t, tc.expected, cm.Data)
			}
		})
	}
}
