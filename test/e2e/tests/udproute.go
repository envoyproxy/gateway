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
	"reflect"
	"testing"
	"time"

	"github.com/miekg/dns"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/config"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, UDPRouteTest)
}

var UDPRouteTest = suite.ConformanceTest{
	ShortName:   "UDPRoute",
	Description: "Make sure UDPRoute is working",
	Manifests:   []string{"testdata/udproute.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Simple UDP request matching UDPRoute should reach coredns backend", func(t *testing.T) {
			namespace := "gateway-conformance-infra"
			domain := "foo.bar.com."
			routeNN := types.NamespacedName{Name: "udp-coredns", Namespace: namespace}
			gwNN := types.NamespacedName{Name: "udp-gateway", Namespace: namespace}
			gwAddr := GatewayAndUDPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, NewGatewayRef(gwNN), routeNN)

			msg := new(dns.Msg)
			msg.SetQuestion(domain, dns.TypeA)

			if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, time.Minute, true,
				func(_ context.Context) (done bool, err error) {
					tlog.Logf(t, "performing DNS query %s on %s", domain, gwAddr)
					r, err := dns.Exchange(msg, gwAddr)
					if err != nil {
						tlog.Logf(t, "failed to perform a UDP query: %v", err)
						return false, nil
					}
					tlog.Logf(t, "got DNS response: %s", r.String())
					return true, nil
				}); err != nil {
				t.Errorf("failed to perform DNS query: %v", err)
			}
		})
	},
}

// GatewayRef is a tiny type for specifying a UDP Route ParentRef without
// relying on a specific api version.
type GatewayRef struct {
	types.NamespacedName
	listenerNames []*gwapiv1.SectionName
}

// NewGatewayRef creates a GatewayRef resource.  ListenerNames are optional.
func NewGatewayRef(nn types.NamespacedName, listenerNames ...string) GatewayRef {
	var listeners []*gwapiv1.SectionName

	if len(listenerNames) == 0 {
		listenerNames = append(listenerNames, "")
	}

	for _, listener := range listenerNames {
		sectionName := gwapiv1.SectionName(listener)
		listeners = append(listeners, &sectionName)
	}
	return GatewayRef{
		NamespacedName: nn,
		listenerNames:  listeners,
	}
}

// GatewayAndUDPRoutesMustBeAccepted waits until the specified Gateway has an IP
// address assigned to it and the UDPRoute has a ParentRef referring to the
// Gateway. The test will fail if these conditions are not met before the
// timeouts.
func GatewayAndUDPRoutesMustBeAccepted(t *testing.T, c client.Client, timeoutConfig config.TimeoutConfig, controllerName string, gw GatewayRef, routeNNs ...types.NamespacedName) string {
	t.Helper()

	gwAddr, err := kubernetes.WaitForGatewayAddress(t, c, timeoutConfig, kubernetes.GatewayRef{
		NamespacedName: gw.NamespacedName,
	})
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
				Conditions: []metav1.Condition{
					{
						Type:   string(gwapiv1.RouteConditionAccepted),
						Status: metav1.ConditionTrue,
						Reason: string(gwapiv1.RouteReasonAccepted),
					},
				},
			})
		}
		UDPRouteMustHaveParents(t, c, timeoutConfig, routeNN, parents, namespaceRequired)
	}

	return gwAddr
}

func UDPRouteMustHaveParents(t *testing.T, client client.Client, timeoutConfig config.TimeoutConfig, routeName types.NamespacedName, parents []gwapiv1.RouteParentStatus, namespaceRequired bool) {
	t.Helper()

	var actual []gwapiv1.RouteParentStatus
	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, timeoutConfig.RouteMustHaveParents, true, func(ctx context.Context) (bool, error) {
		route := &gwapiv1a2.UDPRoute{}
		err := client.Get(ctx, routeName, route)
		if err != nil {
			return false, fmt.Errorf("error fetching UDPRoute: %w", err)
		}

		actual = route.Status.Parents
		return parentsForRouteMatch(t, routeName, parents, actual, namespaceRequired), nil
	})
	require.NoErrorf(t, waitErr, "error waiting for UDPRoute to have parents matching expectations")
}

func parentsForRouteMatch(t *testing.T, routeName types.NamespacedName, expected, actual []gwapiv1.RouteParentStatus, namespaceRequired bool) bool {
	t.Helper()

	if len(expected) != len(actual) {
		tlog.Logf(t, "Route %s/%s expected %d Parents got %d", routeName.Namespace, routeName.Name, len(expected), len(actual))
		return false
	}

	for i, expectedParent := range expected {
		actualParent := actual[i]
		if actualParent.ControllerName != expectedParent.ControllerName {
			tlog.Logf(t, "Route %s/%s ControllerName doesn't match", routeName.Namespace, routeName.Name)
			return false
		}
		if !reflect.DeepEqual(actualParent.ParentRef.Group, expectedParent.ParentRef.Group) {
			tlog.Logf(t, "Route %s/%s expected ParentReference.Group to be %v, got %v", routeName.Namespace, routeName.Name, expectedParent.ParentRef.Group, actualParent.ParentRef.Group)
			return false
		}
		if !reflect.DeepEqual(actualParent.ParentRef.Kind, expectedParent.ParentRef.Kind) {
			tlog.Logf(t, "Route %s/%s expected ParentReference.Kind to be %v, got %v", routeName.Namespace, routeName.Name, expectedParent.ParentRef.Kind, actualParent.ParentRef.Kind)
			return false
		}
		if actualParent.ParentRef.Name != expectedParent.ParentRef.Name {
			tlog.Logf(t, "Route %s/%s ParentReference.Name doesn't match", routeName.Namespace, routeName.Name)
			return false
		}
		if !reflect.DeepEqual(actualParent.ParentRef.Namespace, expectedParent.ParentRef.Namespace) {
			if namespaceRequired || actualParent.ParentRef.Namespace != nil {
				tlog.Logf(t, "Route %s/%s expected ParentReference.Namespace to be %v, got %v", routeName.Namespace, routeName.Name, expectedParent.ParentRef.Namespace, actualParent.ParentRef.Namespace)
				return false
			}
		}
		if !conditionsMatch(t, expectedParent.Conditions, actualParent.Conditions) {
			return false
		}
	}

	tlog.Logf(t, "Route %s/%s Parents matched expectations", routeName.Namespace, routeName.Name)
	return true
}

func conditionsMatch(t *testing.T, expected, actual []metav1.Condition) bool {
	t.Helper()

	if len(actual) < len(expected) {
		tlog.Logf(t, "Expected more conditions to be present")
		return false
	}
	for _, condition := range expected {
		if !findConditionInList(t, actual, condition.Type, string(condition.Status), condition.Reason) {
			return false
		}
	}

	tlog.Logf(t, "Conditions matched expectations")
	return true
}

// findConditionInList finds a condition in a list of Conditions, checking
// the Name, Value, and Reason. If an empty reason is passed, any Reason will match.
// If an empty status is passed, any Status will match.
func findConditionInList(t *testing.T, conditions []metav1.Condition, condName, expectedStatus, expectedReason string) bool {
	t.Helper()

	for _, cond := range conditions {
		if cond.Type == condName {
			// an empty Status string means "Match any status".
			if expectedStatus == "" || cond.Status == metav1.ConditionStatus(expectedStatus) {
				// an empty Reason string means "Match any reason".
				if expectedReason == "" || cond.Reason == expectedReason {
					return true
				}
				tlog.Logf(t, "%s condition Reason set to %s, expected %s", condName, cond.Reason, expectedReason)
			}

			tlog.Logf(t, "%s condition set to Status %s with Reason %v, expected Status %s", condName, cond.Status, cond.Reason, expectedStatus)
		}
	}

	tlog.Logf(t, "%s was not in conditions list [%v]", condName, conditions)
	return false
}
