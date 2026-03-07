// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyProxyInvalidUpdateTest)
}

var EnvoyProxyInvalidUpdateTest = suite.ConformanceTest{
	ShortName:   "EnvoyProxyInvalidUpdate",
	Description: "Update EnvoyProxy with invalid value, infra should skip provision",
	Manifests:   []string{"testdata/envoyproxy-invalid.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "eg-invalid", Namespace: ns}
		gwNN := types.NamespacedName{Name: "eg-invalid", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		okResp := http.ExpectedResponse{
			Request: http.Request{
				Path: "/invalid",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		// Send a request to a valid path and expect a successful response
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, okResp)

		// get the pod name for specific Gateway
		expectedNs := GetGatewayResourceNamespace()
		podName, err := getPodNameForGateway(t, suite, expectedNs, gwNN)
		require.NoError(t, err)

		// Update EnvoyProx with invalid bootstrap
		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/envoyproxy-invalid-update.yaml", true)

		// Make sure that the Gateway contains InvalidParameters
		kubernetes.GatewayMustHaveCondition(t, suite.Client, suite.TimeoutConfig, gwNN, metav1.Condition{
			Type:   string(gwapiv1.GatewayConditionAccepted),
			Status: metav1.ConditionFalse,
			Reason: string(gwapiv1.GatewayReasonInvalidParameters),
		})

		podShouldBeSame, err := getPodNameForGateway(t, suite, expectedNs, gwNN)
		require.NoError(t, err)
		require.Equal(t, podName, podShouldBeSame)
	},
}

func getPodNameForGateway(t *testing.T, suite *suite.ConformanceTestSuite, expectedNs string, gtw types.NamespacedName) (string, error) {
	pods := &corev1.PodList{}
	err := suite.Client.List(t.Context(), pods, &client.ListOptions{
		Namespace: expectedNs,
		LabelSelector: labels.SelectorFromSet(map[string]string{
			"app.kubernetes.io/managed-by":                   "envoy-gateway",
			"app.kubernetes.io/name":                         "envoy",
			"gateway.envoyproxy.io/owning-gateway-name":      gtw.Name,
			"gateway.envoyproxy.io/owning-gateway-namespace": gtw.Namespace,
		}),
	})
	if err != nil {
		return "", err
	}

	if len(pods.Items) != 1 {
		return "", fmt.Errorf("get unexpected count of pods: %d", len(pods.Items))
	}

	return pods.Items[0].Name, err
}
