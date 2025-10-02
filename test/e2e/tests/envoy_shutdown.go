// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"testing"
	"time"

	"fortio.org/fortio/periodic"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
)

var EnvoyShutdownTest = suite.ConformanceTest{
	ShortName:   "EnvoyShutdown",
	Description: "Deleting envoy pod should not lead to failures",
	Manifests:   []string{"testdata/envoy-shutdown.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("All requests must succeed", func(t *testing.T) {
			ns := "gateway-upgrade-infra"
			name := "ha-gateway"
			routeNN := types.NamespacedName{Name: "http-envoy-shutdown", Namespace: ns}
			gwNN := types.NamespacedName{Name: name, Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
			reqURL := url.URL{Scheme: "http", Host: http.CalculateHost(t, gwAddr, "http"), Path: "/envoy-shutdown"}
			epNN := types.NamespacedName{Name: "upgrade-config", Namespace: "envoy-gateway-system"}
			dp, err := getDeploymentForGateway(ns, name, suite.Client)
			if err != nil {
				t.Errorf("Failed to get proxy deployment")
			}

			WaitForPods(t, suite.Client, dp.Namespace, map[string]string{"gateway.envoyproxy.io/owning-gateway-name": name}, corev1.PodRunning, &PodReady)

			// wait for route to be programmed on envoy
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/envoy-shutdown",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// can be used to abort the test after deployment restart is complete or failed
			aborter := periodic.NewAborter()
			// will contain indication on success or failure of load test
			loadSuccess := make(chan bool)

			t.Log("Starting load generation")
			// Run load async and continue to restart deployment
			go runLoadAndWait(t, &suite.TimeoutConfig, loadSuccess, aborter, reqURL.String())

			t.Log("Rolling out proxy deployment")
			err = restartProxyAndWaitForRollout(t, &suite.TimeoutConfig, suite.Client, epNN, dp)

			t.Log("Stopping load generation and collecting results")
			aborter.Abort(false) // abort the load either way

			if err != nil {
				t.Errorf("Failed to rollout proxy deployment: %v", err)
			}

			// Wait for the goroutine to finish
			result := <-loadSuccess
			if !result {
				t.Errorf("Load test failed")
			}
		})
	},
}

// gets the proxy deployment created for a gateway, assuming merge-gateways is not used
func getDeploymentForGateway(namespace, name string, c client.Client) (*appsv1.Deployment, error) {
	dpLabels := proxy.EnvoyAppLabel()
	owningLabels := gatewayapi.GatewayOwnerLabels(namespace, name)
	for k, v := range owningLabels {
		dpLabels[k] = v
	}
	ctx := context.Background()

	gwNs := "envoy-gateway-system"
	if IsGatewayNamespaceMode() {
		// use the namespace of Gateway resource
		gwNs = namespace
	}
	listOpts := []client.ListOption{
		client.InNamespace(gwNs),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(dpLabels)},
	}

	depList := &appsv1.DeploymentList{}
	err := c.List(ctx, depList, listOpts...)
	if err != nil {
		return nil, err
	}
	if len(depList.Items) != 1 {
		return nil, errors.New("unexpected number of matching deployments found")
	}
	ret := depList.Items[0]
	return &ret, nil
}

// sets the "gateway.envoyproxy.io/restartedAt" annotation in the EnvoyProxy resource's deployment patch spec
// leading to EG triggering a rollout restart of the deployment
func restartProxyAndWaitForRollout(t *testing.T, timeoutConfig *config.TimeoutConfig, c client.Client, epNN types.NamespacedName, dp *appsv1.Deployment) error {
	t.Helper()
	const egRestartAnnotation = "gateway.envoyproxy.io/restartedAt"
	restartTime := time.Now().Format(time.RFC3339)
	ctx := context.Background()

	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}
	ep := egv1a1.EnvoyProxy{}
	if err := c.Get(context.Background(), epNN, &ep); err != nil {
		return err
	}

	jsonData := fmt.Sprintf("{\"metadata\": {\"annotations\": {\"gateway.envoyproxy.io/restartedAt\": \"%s\"}}, \"spec\": {\"template\": {\"metadata\": {\"annotations\": {\"gateway.envoyproxy.io/restartedAt\": \"%s\"}}}}}", restartTime, restartTime)

	ep.Spec.Provider.Kubernetes.EnvoyDeployment.Patch = &egv1a1.KubernetesPatchSpec{
		Value: apiextensionsv1.JSON{
			Raw: []byte(jsonData),
		},
	}

	if err := c.Update(ctx, &ep); err != nil {
		return err
	}

	// increase timeout for IPv6 first cluster
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 2*timeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		// wait for replicaset with the same annotation to reach ready status
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(dp.Namespace),
			client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(dp.Spec.Selector.MatchLabels)},
		}

		err := c.List(ctx, podList, listOpts...)
		if err != nil {
			return false, err
		}

		rolled := int32(0)
		for i := range podList.Items {
			rs := &podList.Items[i]
			if rs.Annotations[egRestartAnnotation] == restartTime {
				rolled++
			}
		}

		// all pods are rolled
		if rolled == int32(len(podList.Items)) && rolled >= *dp.Spec.Replicas {
			return true, nil
		}

		return false, nil
	})
}
