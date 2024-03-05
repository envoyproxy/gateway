// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"io"
	"net/url"
	"testing"
	"time"

	"fortio.org/fortio/fhttp"
	"fortio.org/fortio/periodic"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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
	ConformanceTests = append(ConformanceTests, BackendUpgradeTest)
}

var BackendUpgradeTest = suite.ConformanceTest{
	ShortName:   "BackendUpgrade",
	Description: "Rolling backend pods should not lead to failures",
	Manifests:   []string{"testdata/backend-upgrade.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("All requests must succeed", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-backend-upgrade", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			dNN := types.NamespacedName{Name: "infra-backend-v1", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			reqURL := url.URL{Scheme: "http", Host: http.CalculateHost(t, gwAddr, "http"), Path: "/backend-upgrade"}

			// can be used to abort the test after deployment restart is complete or failed
			aborter := periodic.NewAborter()
			// will contain indication on success or failure of load test
			loadSuccess := make(chan bool)

			t.Log("Starting load generation")
			// Run load async and continue to restart deployment
			go runLoadAndWait(t, suite.TimeoutConfig, loadSuccess, aborter, reqURL.String())

			t.Log("Restarting deployment")
			err := restartDeploymentAndWaitForNewPods(t, suite.TimeoutConfig, suite.Client, dNN)

			t.Log("Stopping load generation and collecting results")
			aborter.Abort(false) // abort the load either way

			if err != nil {
				t.Errorf("Failed to restart deployment")
			}

			// Wait for the goroutine to finish
			result := <-loadSuccess
			if !result {
				t.Errorf("Load test failed")
			}
		})
	},
}

// runs a load test with options described in opts
// the done channel is used to notify caller of execution result
// the execution may end due to an external abort or timeout
func runLoadAndWait(t *testing.T, timeoutConfig config.TimeoutConfig, done chan bool, aborter *periodic.Aborter, reqURL string) {
	opts := fhttp.HTTPRunnerOptions{
		RunnerOptions: periodic.RunnerOptions{
			QPS: 5000,
			// allow some overhead time for setting up workers and tearing down after restart
			Duration:   timeoutConfig.CreateTimeout + timeoutConfig.CreateTimeout/2,
			NumThreads: 50,
			Stop:       aborter,
			Out:        io.Discard,
		},
		HTTPOptions: fhttp.HTTPOptions{
			URL: reqURL,
		},
	}
	res, err := fhttp.RunHTTPTest(&opts)
	if err != nil {
		done <- false
		t.Logf("failed to create load: %v", err)
	}

	// collect stats
	okReq := res.RetCodes[200]
	totalReq := res.DurationHistogram.Count
	failedReq := totalReq - okReq
	errorReq := res.ErrorsDurationHistogram.Count
	timedOut := res.ActualDuration == opts.Duration
	t.Logf("Backend upgrade completed after %s with %d requests, %d success, %d failures and %d errors", res.ActualDuration, totalReq, okReq, failedReq, errorReq)

	if okReq == totalReq && errorReq == 0 && !timedOut {
		done <- true
	}
	done <- false
}

func restartDeploymentAndWaitForNewPods(t *testing.T, timeoutConfig config.TimeoutConfig, c client.Client, dNN types.NamespacedName) error {
	t.Helper()
	const kubeRestartAnnotation = "kubectl.kubernetes.io/restartedAt"

	ctx := context.Background()
	dp := &appsv1.Deployment{}

	err := c.Get(ctx, dNN, dp)
	if err != nil {
		return err
	}

	if dp.Spec.Template.ObjectMeta.Annotations == nil {
		dp.Spec.Template.ObjectMeta.Annotations = make(map[string]string)
	}
	restartTime := time.Now().Format(time.RFC3339)
	dp.Spec.Template.ObjectMeta.Annotations[kubeRestartAnnotation] = restartTime

	if err = c.Update(ctx, dp); err != nil {
		return err
	}

	return wait.PollUntilContextTimeout(ctx, 1*time.Second, timeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		// wait for replicaset with the same annotation to reach ready status
		podList := &corev1.PodList{}
		listOpts := []client.ListOption{
			client.InNamespace(dNN.Namespace),
			client.MatchingLabelsSelector{Selector: labels.SelectorFromSet(labels.Set{"app": dNN.Name})},
		}

		err = c.List(ctx, podList, listOpts...)
		if err != nil {
			return false, err
		}

		rolled := int32(0)
		for _, rs := range podList.Items {
			if rs.Annotations[kubeRestartAnnotation] == restartTime {
				rolled++
			}
		}

		if rolled == *dp.Spec.Replicas {
			return true, nil
		}

		return false, nil
	})
}
