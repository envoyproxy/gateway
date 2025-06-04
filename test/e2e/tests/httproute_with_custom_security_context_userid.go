// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyGatewayCustomSecurityContextUseridTest)
}

var EnvoyGatewayCustomSecurityContextUseridTest = suite.ConformanceTest{
	ShortName:   "EnvoyGatewayCustomSecurityContextUserid",
	Description: "Envoy Gateway container with custom security context user id",
	Manifests: []string{
		"testdata/custom-container-security-contex-userid.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("route with custom security context user id", func(t *testing.T) {
			// set envoy-gateway deployment security context user id to 65534 to test custom user has the necessary permissions
			// to run the envoy-gateway container
			setEGSecurityContextUserID(t, suite, 65534)

			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "custom-eg-security-context-userid", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// we don't set the envoy-gateway deployment security context user id back to default because this will
			// cause the envoy proxies failed to be deleted after the Gateway resources in the base are deleted.
			// This is acceptable because this won't affect the later tests in the same suite.
		})
	},
}

func setEGSecurityContextUserID(t *testing.T, suite *suite.ConformanceTestSuite, uid int64) {
	// update envoy-gateway deployment with custom security context user id
	egDeployment := &appsv1.Deployment{}
	err := suite.Client.Get(
		context.Background(),
		types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"},
		egDeployment)
	require.NoError(t, err)
	egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = ptr.To(uid)
	egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup = ptr.To(uid)
	err = suite.Client.Update(context.Background(), egDeployment)
	require.NoError(t, err)
	// test that envoy-gateway pod is running with custom security context user id
	WaitForPods(t, suite.Client, "envoy-gateway-system", map[string]string{"control-plane": "envoy-gateway"}, corev1.PodRunning, PodReady)

	// test that envoy-gateway deployment is updated with custom security context user id
	egDeployment = &appsv1.Deployment{}
	err = suite.Client.Get(
		context.Background(),
		types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"},
		egDeployment)
	require.NoError(t, err)
	require.Equal(t, uid, *egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser, "envoy-gateway deployment is not updated with custom security context user id")
	require.Equal(t, uid, *egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup, "envoy-gateway deployment is not updated with custom security context group id")
}
