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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
		gatewayNS := GetGatewayResourceNamespace()

		t.Run("Deployment", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "deploy-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "eg-deployment", Namespace: ns}
			okResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/deploy",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Make sure there's deployment for the gateway
			err := checkDeployment(t, suite, gwNN, gatewayNS, expectedGatewayName(gwNN), 0, 0)
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)

			// Update the Gateway to use a custom name
			updateGateway(t, suite, gwNN, &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "deploy-custom-name",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			})

			err = checkDeployment(t, suite, gwNN, gatewayNS, "custom-", 1, 1)
			if err != nil {
				t.Fatalf("Failed to delete Gateway: %v", err)
			}
			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)

			// Rollback the Gateway to without custom name
			updateGateway(t, suite, gwNN, &gwapiv1.GatewayInfrastructure{})

			// Make sure there's deployment for the gateway
			err = checkDeployment(t, suite, gwNN, gatewayNS, expectedGatewayName(gwNN), 0, 0)
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}
			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)
		})

		t.Run("DaemonSet", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "ds-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "eg-daemonset", Namespace: ns}
			okResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/daemonset",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Make sure there's DaemonSet for the gateway
			err := checkDaemonSet(t, suite, gwNN, gatewayNS, expectedGatewayName(gwNN))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}

			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)

			// Update the Gateway to use a custom name
			updateGateway(t, suite, gwNN, &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "ds-custom-name",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			})

			err = checkDaemonSet(t, suite, gwNN, gatewayNS, "ds-custom-name")
			if err != nil {
				t.Fatalf("Failed to delete Gateway: %v", err)
			}
			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)

			// Rollback the Gateway to without custom name
			updateGateway(t, suite, gwNN, &gwapiv1.GatewayInfrastructure{
				ParametersRef: &gwapiv1.LocalParametersReference{
					Name:  "eg-daemonset",
					Kind:  "EnvoyProxy",
					Group: "gateway.envoyproxy.io",
				},
			})

			// Make sure there's DaemonSet for the gateway
			err = checkDaemonSet(t, suite, gwNN, gatewayNS, expectedGatewayName(gwNN))
			if err != nil {
				t.Fatalf("Failed to check EnvoyProxy deployment: %v", err)
			}

			// Send a request to a valid path and expect a successful response
			ExpectEventuallyConsistentResponse(t, suite, gwNN, routeNN, &okResp)
		})
	},
}

func expectedGatewayName(gwNN types.NamespacedName) string {
	if IsGatewayNamespaceMode() {
		return gwNN.Name
	}

	return fmt.Sprintf("envoy-%s-%s", gwNN.Namespace, gwNN.Name)
}

func updateGateway(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, paramRef *gwapiv1.GatewayInfrastructure) {
	err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.CreateTimeout, true,
		func(ctx context.Context) (bool, error) {
			gw := &gwapiv1.Gateway{}
			err := suite.Client.Get(context.Background(), gwNN, gw)
			if err != nil {
				tlog.Logf(t, "Failed to get Gateway %s: %v", gwNN, err)
				return false, err
			}
			gw.Spec.Infrastructure = paramRef
			err = suite.Client.Update(context.Background(), gw)
			if err != nil {
				tlog.Logf(t, "Failed to update Gateway %s: %v", gwNN, err)
				return false, nil
			}
			return true, nil
		})
	if err != nil {
		t.Fatalf("Failed to patch Gateway %s: %v", gwNN, err)
	}
}

// ExpectEventuallyConsistentResponse sends a request to the gateway and waits for an eventually consistent response.
// This's different from because of the name may change, so we query the gateway address every time.
func ExpectEventuallyConsistentResponse(t *testing.T, suite *suite.ConformanceTestSuite,
	gwNN, routeNN types.NamespacedName, expected *http.ExpectedResponse,
) {
	t.Helper()

	if expected == nil {
		t.Fatalf("expected response cannot be nil")
	}

	err := wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		req := http.MakeRequest(t, expected, gwAddr, "HTTP", "http")

		cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "Request failed: %v", err.Error())
			return false, nil
		}

		if err := http.CompareRoundTrip(t, &req, cReq, cRes, *expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v  %v", req, err)
			return false, nil
		}

		return true, nil
	})
	if err != nil {
		t.Fatalf("Failed to get expected response: %v", err)
		return
	}
	tlog.Logf(t, "Request passed")
}

func checkDaemonSet(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, exceptNs, exceptName string) error {
	if err := checkObject(t, suite, appsv1.SchemeGroupVersion.WithKind("DaemonSet"), gwNN, 1, // service always exists
		exceptNs, exceptName); err != nil {
		return err
	}

	if err := checkObject(t, suite, serviceGVK, gwNN, 1, // service always exists
		exceptNs, exceptName); err != nil {
		return err
	}

	return nil
}

var (
	serviceGVK        = schema.FromAPIVersionAndKind("v1", "Service")
	serviceAccountGVK = schema.FromAPIVersionAndKind("v1", "ServiceAccount")
	hpaGVK            = schema.FromAPIVersionAndKind("autoscaling/v2", "HorizontalPodAutoscaler")
	pdbGVK            = schema.FromAPIVersionAndKind("policy/v1", "PodDisruptionBudget")
)

func checkDeployment(t *testing.T, suite *suite.ConformanceTestSuite, gwNN types.NamespacedName, exceptNs, exceptName string,
	exceptHpaCount, exceptPdbCount int,
) error {
	if err := checkObject(t, suite, appsv1.SchemeGroupVersion.WithKind("Deployment"), gwNN, 1, // service always exists
		exceptNs, exceptName); err != nil {
		return err
	}

	if err := checkObject(t, suite, serviceGVK, gwNN, 1, // service always exists
		exceptNs, exceptName); err != nil {
		return err
	}
	if err := checkObject(t, suite, hpaGVK, gwNN, exceptHpaCount, exceptNs, exceptName); err != nil {
		return err
	}
	if err := checkObject(t, suite, pdbGVK, gwNN, exceptPdbCount, exceptNs, exceptName); err != nil {
		return err
	}
	if err := checkObject(t, suite, serviceAccountGVK, gwNN, 1, exceptNs, exceptName); err != nil {
		return err
	}
	return nil
}

func checkObject(t *testing.T, suite *suite.ConformanceTestSuite, gvk schema.GroupVersionKind, gwNN types.NamespacedName, exceptCount int, exceptNs, exceptName string) error {
	return wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.CreateTimeout, true, func(ctx context.Context) (bool, error) {
		objList := &unstructured.UnstructuredList{}
		objList.SetGroupVersionKind(gvk)

		selector := map[string]string{
			"app.kubernetes.io/managed-by":                   "envoy-gateway",
			"app.kubernetes.io/name":                         "envoy",
			"gateway.envoyproxy.io/owning-gateway-name":      gwNN.Name,
			"gateway.envoyproxy.io/owning-gateway-namespace": gwNN.Namespace,
		}

		opts := &client.ListOptions{
			Namespace:     exceptNs,
			LabelSelector: labels.SelectorFromSet(selector),
		}
		err := suite.Client.List(t.Context(), objList, opts)
		if err != nil {
			return false, err
		}
		if len(objList.Items) != exceptCount {
			tlog.Logf(t, "Expected %d %s for the Gateway (%v), got %d", exceptCount, gvk, opts, len(objList.Items))
			return false, nil
		}

		if exceptCount > 0 {
			obj := objList.Items[0]

			if !strings.HasPrefix(obj.GetName(), exceptName) {
				tlog.Logf(t, "Expected %s name has prefix '%s', got %s", gvk, exceptName, obj.GetName())
				return false, nil
			}
			tlog.Logf(t, "Check envoy proxy %s pass, name: %s", gvk, obj.GetName())
		}

		return true, nil
	})
}
