// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, TCPRouteTest)
}

var TCPRouteTest = suite.ConformanceTest{
	ShortName:   "TCPRoute",
	Description: "Testing TCP Route",
	Manifests:   []string{"testdata/tcproute.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("tcp-route-1", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tcp-app-1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "my-tcp-gateway", Namespace: ns}
			gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
		t.Run("tcp-route-2", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "tcp-app-2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "my-tcp-gateway", Namespace: ns}
			gwAddr := GatewayAndTCPRoutesMustBeAccepted(t, suite.Client, &suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)
			OkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			// Send a request to an valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)
		})
	},
}

func GatewayAndTCPRoutesMustBeAccepted(t *testing.T, c client.Client, timeoutConfig *config.TimeoutConfig, controllerName string, gw GatewayRef, routeNNs ...types.NamespacedName) string {
	t.Helper()

	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	tcpRoute := &gwapiv1a2.TCPRoute{}
	err := c.Get(context.Background(), routeNNs[0], tcpRoute)
	if err != nil {
		tlog.Logf(t, "error fetching TCPRoute: %v", err)
	}

	gwAddr, err := WaitForGatewayAddress(t, c, timeoutConfig, gw.NamespacedName, string(*tcpRoute.Spec.ParentRefs[0].SectionName))
	require.NoErrorf(t, err, "timed out waiting for Gateway address to be assigned")

	ns := gwapiv1.Namespace(gw.Namespace)
	kind := gwapiv1.Kind("Gateway")

	for _, routeNN := range routeNNs {
		namespaceRequired := true
		if routeNN.Namespace == gw.Namespace {
			namespaceRequired = false
		}

		var parents []gwapiv1.RouteParentStatus
		for _, listener := range gw.listenerNames {
			parents = append(parents, gwapiv1.RouteParentStatus{
				ParentRef: gwapiv1.ParentReference{
					Group:       (*gwapiv1.Group)(&gwapiv1.GroupVersion.Group),
					Kind:        &kind,
					Name:        gwapiv1.ObjectName(gw.Name),
					Namespace:   &ns,
					SectionName: listener,
				},
				ControllerName: gwapiv1.GatewayController(controllerName),
				Conditions: []metav1.Condition{{
					Type:   string(gwapiv1.RouteConditionAccepted),
					Status: metav1.ConditionTrue,
					Reason: string(gwapiv1.RouteReasonAccepted),
				}},
			})
		}
		TCPRouteMustHaveParents(t, c, timeoutConfig, routeNN, parents, namespaceRequired)
	}

	return gwAddr
}

// WaitForGatewayAddress waits until at least one IP Address has been set in the
// status of the specified Gateway.
func WaitForGatewayAddress(t *testing.T, client client.Client, timeoutConfig *config.TimeoutConfig, gwName types.NamespacedName, sectionName string) (string, error) {
	t.Helper()

	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	var ipAddr, port string
	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeoutConfig.GatewayMustHaveAddress, true, func(ctx context.Context) (bool, error) {
		gw := &gwapiv1.Gateway{}
		err := client.Get(ctx, gwName, gw)
		if err != nil {
			tlog.Logf(t, "error fetching Gateway: %v", err)
			return false, fmt.Errorf("error fetching Gateway: %w", err)
		}

		if err := kubernetes.ConditionsHaveLatestObservedGeneration(gw, gw.Status.Conditions); err != nil {
			t.Log("Gateway", err)
			return false, nil
		}
		if sectionName != "" {
			for i, val := range gw.Spec.Listeners {
				if val.Name == gwapiv1.SectionName(sectionName) {
					port = strconv.FormatInt(int64(gw.Spec.Listeners[i].Port), 10)
				}
			}
		} else {
			port = strconv.FormatInt(int64(gw.Spec.Listeners[0].Port), 10)
		}

		// TODO: Support more than IPAddress
		for _, address := range gw.Status.Addresses {
			if address.Type != nil && *address.Type == gwapiv1.IPAddressType {
				ipAddr = address.Value
				return true, nil
			}
		}

		return false, nil
	})
	require.NoErrorf(t, waitErr, "error waiting for Gateway to have at least one IP address in status")
	return net.JoinHostPort(ipAddr, port), waitErr
}

func TCPRouteMustHaveParents(t *testing.T, client client.Client, timeoutConfig *config.TimeoutConfig, routeName types.NamespacedName, parents []gwapiv1.RouteParentStatus, namespaceRequired bool) {
	t.Helper()

	if timeoutConfig == nil {
		t.Fatalf("timeoutConfig cannot be nil")
	}

	var actual []gwapiv1.RouteParentStatus
	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeoutConfig.RouteMustHaveParents, true, func(ctx context.Context) (bool, error) {
		route := &gwapiv1a2.TCPRoute{}
		err := client.Get(ctx, routeName, route)
		if err != nil {
			return false, fmt.Errorf("error fetching HTTPRoute: %w", err)
		}

		actual = route.Status.Parents
		return parentsForRouteMatch(t, routeName, parents, actual, namespaceRequired), nil
	})
	require.NoErrorf(t, waitErr, "error waiting for TCPRoute to have parents matching expectations")
}
