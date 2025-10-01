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
	"sigs.k8s.io/controller-runtime/pkg/client"
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
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// reset envoy-gateway deployment security context user id to the default value 65532
			setEGSecurityContextUserID(t, suite, 65532)
			// We have to manually delete the envoy proxy deployment to ensure that the test suite can clean up properly.
			// This is because the rollout restart of the envoy-gateway deployment may cause Envoy Gateway fail to delete
			// the envoy proxy deployments after the Gateway resources are deleted in ControllerNamspace mod, which can
			// lead to failure of the upgrade test.
			if suite.Cleanup {
				proxies := appsv1.DeploymentList{}
				err := suite.Client.List(
					context.Background(),
					&proxies,
					client.InNamespace("envoy-gateway-system"),
					client.MatchingLabels{"app.kubernetes.io/component": "proxy", "app.kubernetes.io/managed-by": "envoy-gateway"})
				require.NoError(t, err, "failed to list envoy proxy deployments")
				for _, proxy := range proxies.Items {
					err = suite.Client.Delete(context.Background(), &proxy)
					require.NoError(t, err, "failed to delete envoy proxy deployment %s", proxy.Name)
				}
			}
		})
	},
}

func setEGSecurityContextUserID(t *testing.T, suite *suite.ConformanceTestSuite, uid int64) {
	// update envoy-gateway deployment with custom security context user id
	retries := 5
	var err error
	for i := 0; i < retries; i++ { // retry a few times to avoid update conflicts
		egDeployment := &appsv1.Deployment{}
		err = suite.Client.Get(
			context.Background(),
			types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"},
			egDeployment,
		)
		require.NoError(t, err)

		egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser = ptr.To(uid)
		egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup = ptr.To(uid)

		if err = suite.Client.Update(context.Background(), egDeployment); err == nil {
			break
		}
	}
	require.NoError(t, err)

	// test that envoy-gateway pod is running with custom security context user id
	WaitForPods(t, suite.Client, "envoy-gateway-system", map[string]string{"control-plane": "envoy-gateway"}, corev1.PodRunning, &PodReady)

	// test that envoy-gateway deployment is updated with custom security context user id
	egDeployment := &appsv1.Deployment{}
	err = suite.Client.Get(
		context.Background(),
		types.NamespacedName{Name: "envoy-gateway", Namespace: "envoy-gateway-system"},
		egDeployment)
	require.NoError(t, err)
	require.Equal(t, uid, *egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsUser, "envoy-gateway deployment is not updated with custom security context user id")
	require.Equal(t, uid, *egDeployment.Spec.Template.Spec.Containers[0].SecurityContext.RunAsGroup, "envoy-gateway deployment is not updated with custom security context group id")
}
