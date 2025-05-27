// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyProxyDaemonSetTest)
}

var EnvoyProxyDaemonSetTest = suite.ConformanceTest{
	ShortName:   "EnvoyProxyDaemonSet",
	Description: "Test running Envoy as a DaemonSet",
	Manifests:   []string{"testdata/envoyproxy-daemonset.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("RunAndDelete", func(t *testing.T) {
			t.Cleanup(func() {
				if t.Failed() {
					CollectAndDump(t, suite.RestConfig)
				}
			})

			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "daemonset-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "eg-daemonset", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			gwPodNamespace := GetGatewayResourceNamespace()
			gwPodSelector := map[string]string{
				"app.kubernetes.io/managed-by":                   "envoy-gateway",
				"app.kubernetes.io/name":                         "envoy",
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}
			// Make sure there's no deployment for the gateway
			err := wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
				deploys := &appsv1.DeploymentList{}
				err := suite.Client.List(ctx, deploys, &client.ListOptions{
					Namespace:     gwPodNamespace,
					LabelSelector: labels.SelectorFromSet(gwPodSelector),
				})
				if err != nil {
					return false, err
				}

				return len(deploys.Items) == 0, err
			})
			if err != nil {
				t.Fatalf("Failed to check no deployments for the Gateway: %v", err)
			}

			WaitForPods(t, suite.Client, gwPodNamespace, gwPodSelector, corev1.PodRunning, PodReady)

			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/daemonset",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			})

			// Delete the Gateway and wait for the DaemonSet to be deleted
			gtw := &gwapiv1.Gateway{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: gwNN.Namespace,
					Name:      gwNN.Name,
				},
			}
			err = suite.Client.Delete(context.TODO(), gtw)
			if err != nil {
				t.Fatalf("Failed to delete Gateway: %v", err)
			}
			err = wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.DeleteTimeout, true, func(ctx context.Context) (bool, error) {
				dsList := &appsv1.DaemonSetList{}
				err := suite.Client.List(ctx, dsList, &client.ListOptions{
					Namespace:     gwPodNamespace,
					LabelSelector: labels.SelectorFromSet(gwPodSelector),
				})
				if err != nil {
					return false, err
				}

				return len(dsList.Items) == 0, nil
			})
			if err != nil {
				t.Fatalf("Failed to delete DaemonSet Gateway: %v", err)
			}
		})
	},
}
