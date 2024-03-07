// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"errors"
	"net/url"
	"testing"
	"time"

	"fortio.org/fortio/periodic"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyShutdownTest)
}

var EnvoyShutdownTest = suite.ConformanceTest{
	ShortName:   "EnvoyShutdown",
	Description: "Deleting envoy pod should not lead to failures",
	Manifests:   []string{"testdata/envoy-shutdown.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("All requests must succeed", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			name := "same-namespace"
			routeNN := types.NamespacedName{Name: "http-envoy-shutdown", Namespace: ns}
			gwNN := types.NamespacedName{Name: name, Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			reqURL := url.URL{Scheme: "http", Host: http.CalculateHost(t, gwAddr, "http"), Path: "/envoy-shutdown"}

			dp, err := getDeploymentForGateway(ns, name, suite.Client)
			if err != nil {
				t.Errorf("Failed to get proxy deployment")
			}

			// can be used to abort the test after deployment restart is complete or failed
			aborter := periodic.NewAborter()
			// will contain indication on success or failure of load test
			loadSuccess := make(chan bool)

			t.Log("Starting load generation")
			// Run load async and continue to restart deployment
			go runLoadAndWait(t, suite.TimeoutConfig, loadSuccess, aborter, reqURL.String())

			t.Log("Deleting proxy pod")
			// the deleting pod will gracefully close connections, and the load generator is expected
			// to reconnect to the other replicas
			err = deleteDeploymentPodAndWaitForTermination(t, suite.TimeoutConfig, suite.Client, dp)

			t.Log("Stopping load generation and collecting results")
			aborter.Abort(false) // abort the load either way

			if err != nil {
				t.Errorf("Failed to delete proxy pods")
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

	listOpts := []client.ListOption{
		client.InNamespace("envoy-gateway-system"),
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

// deletes a single pod belonging to a deployment and wait for it to be removed from the cluster
func deleteDeploymentPodAndWaitForTermination(t *testing.T, timeoutConfig config.TimeoutConfig, c client.Client, dp *appsv1.Deployment) error {
	t.Helper()

	ctx := context.Background()

	podList := &corev1.PodList{}
	listOpts := []client.ListOption{
		client.InNamespace(dp.Namespace),
		client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(dp.Spec.Selector.MatchLabels)},
	}

	err := c.List(ctx, podList, listOpts...)
	if err != nil {
		return err
	}

	if len(podList.Items) < 2 {
		return errors.New("insufficient amount of pods for graceful termination test")
	}

	// delete the first pods
	pod := podList.Items[0]
	if err = c.Delete(ctx, &pod); err != nil {
		return err
	}

	// wait for pod to be removed
	return wait.PollUntilContextTimeout(ctx, 1*time.Second, timeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		if err := c.Get(ctx, types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, &corev1.Pod{}); err != nil {
			if kerrors.IsNotFound(err) {
				return true, nil
			}
			return false, err
		}
		return false, nil
	})
}
