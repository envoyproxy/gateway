// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyProxyCustomNameTest)
}

var EnvoyProxyCustomNameTest = suite.ConformanceTest{
	ShortName:   "EnvoyProxyCustomName",
	Description: "Test running Envoy with custom name",
	Manifests:   []string{"testdata/envoyproxy-custom-name.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Deployment", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "deploy-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "deploy-custom-name", Namespace: ns}
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/deploy",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Make sure there's deployment for the gateway
			err := checkEnvoyProxyDeployment(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			// Update the Gateway to use a custom name
			gw := &gwapiv1.Gateway{}
			err = suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				t.Fatalf("Failed to get Gateway: %v", err)
			}
			gw.Spec.Infrastructure = &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "deploy-custom-name",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			}
			err = suite.Client.Update(context.Background(), gw)
			if err != nil {
				t.Fatalf("Failed to update Gateway: %v", err)
			}

			err = checkEnvoyProxyDeployment(t, suite, gwNN, "deploy-custom-name")
			if err != nil {
				t.Fatalf("Failed to delete Gateway: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, "deploy-custom-name")
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			// Rollback the Gateway to without custom name
			gw = &gwapiv1.Gateway{}
			err = suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				t.Fatalf("Failed to get Gateway: %v", err)
			}
			gw.Spec.Infrastructure = &gwapiv1.GatewayInfrastructure{}
			err = suite.Client.Update(context.Background(), gw)
			if err != nil {
				t.Fatalf("Failed to update Gateway: %v", err)
			}

			// Make sure there's deployment for the gateway
			err = checkEnvoyProxyDeployment(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})

		t.Run("DaemonSet", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "ds-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ds-custom-name", Namespace: ns}
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/daemonset",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Make sure there's DaemonSet for the gateway
			err := checkEnvoyProxyDaemonSet(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			// Update the Gateway to use a custom name
			gw := &gwapiv1.Gateway{}
			err = suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				t.Fatalf("Failed to get Gateway: %v", err)
			}
			gw.Spec.Infrastructure = &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "ds-custom-name",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			}
			err = suite.Client.Update(context.Background(), gw)
			if err != nil {
				t.Fatalf("Failed to update Gateway: %v", err)
			}

			err = checkEnvoyProxyDaemonSet(t, suite, gwNN, "ds-custom-name")
			if err != nil {
				t.Fatalf("Failed to delete Gateway: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, "ds-custom-name")
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			// Rollback the Gateway to without custom name
			gw = &gwapiv1.Gateway{}
			err = suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				t.Fatalf("Failed to get Gateway: %v", err)
			}
			gw.Spec.Infrastructure = &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "eg-daemonset",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			}
			err = suite.Client.Update(context.Background(), gw)
			if err != nil {
				t.Fatalf("Failed to update Gateway: %v", err)
			}

			// Make sure there's DaemonSet for the gateway
			err = checkEnvoyProxyDaemonSet(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			err = checkEnvoyProxyService(t, suite, gwNN, fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy service: %v", err)
			}
			gwAddr = kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
	},
}

func checkEnvoyProxyDeployment(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, exceptName string) error {
	// Make sure there's deployment for the gateway
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		deploys := &appsv1.DeploymentList{}
		err := suite.Client.List(ctx, deploys, &client.ListOptions{
			Namespace: "envoy-gateway-system",
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/managed-by":                   "envoy-gateway",
				"app.kubernetes.io/name":                         "envoy",
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}),
		})
		if err != nil {
			return false, err
		}
		if len(deploys.Items) != 1 {
			tlog.Logf(t, "Expected 1 Deployment for the Gateway, got %d", len(deploys.Items))
			return false, nil
		}

		dp := deploys.Items[0]
		if !strings.HasPrefix(dp.Name, exceptName) {
			tlog.Logf(t, "Expected Deployment name has prefix '%s', got %s", exceptName, dp.Name)
			return false, nil
		}

		if dp.Status.ReadyReplicas <= 0 {
			tlog.Logf(t, "Expected Deployment %s ready", dp.Name)
		}

		tlog.Logf(t, "Check Envoy proxy deployment pass, name: %s", dp.Name)
		return true, nil
	})
}

func checkEnvoyProxyService(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, exceptName string) error {
	// Make sure there's deployment for the gateway
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		svcList := &corev1.ServiceList{}
		err := suite.Client.List(ctx, svcList, &client.ListOptions{
			Namespace: "envoy-gateway-system",
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/managed-by":                   "envoy-gateway",
				"app.kubernetes.io/name":                         "envoy",
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}),
		})
		if err != nil {
			return false, err
		}
		if len(svcList.Items) != 1 {
			tlog.Logf(t, "Expected 1 Service for the Gateway, got %d", len(svcList.Items))
			return false, nil
		}

		svc := svcList.Items[0]
		if !strings.HasPrefix(svc.Name, exceptName) {
			tlog.Logf(t, "Expected Service name has prefix '%s', got %s", exceptName, svc.Name)
			return false, nil
		}

		tlog.Logf(t, "Check envoy proxy service pass, name: %s", svc.Name)
		return true, nil
	})
}

func checkEnvoyProxyDaemonSet(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, exceptName string) error {
	// Make sure there's deployment for the gateway
	return wait.PollUntilContextTimeout(context.TODO(), time.Second, suite.TimeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		dsList := &appsv1.DaemonSetList{}
		err := suite.Client.List(ctx, dsList, &client.ListOptions{
			Namespace: "envoy-gateway-system",
			LabelSelector: labels.SelectorFromSet(map[string]string{
				"app.kubernetes.io/managed-by":                   "envoy-gateway",
				"app.kubernetes.io/name":                         "envoy",
				"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
				"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
			}),
		})
		if err != nil {
			return false, err
		}
		if len(dsList.Items) != 1 {
			tlog.Logf(t, "Expected 1 DaemonSet for the Gateway, got %d", len(dsList.Items))
			return false, nil
		}

		ds := dsList.Items[0]
		if !strings.HasPrefix(ds.Name, exceptName) {
			tlog.Logf(t, "Expected DaemonSet name has prefix '%s', got %s", exceptName, ds.Name)
			return false, nil
		}

		if ds.Status.NumberReady <= 0 {
			tlog.Logf(t, "Expected DaemonSet %s ready", ds.Name)
		}

		tlog.Logf(t, "Check envoy proxy DaemonSet pass, name: %s", ds.Name)
		return true, nil
	})
}
