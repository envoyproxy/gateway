// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ExtProcAccessLogTest)
}

// ExtProcAccessLogTest verifies %EG_EXT_FILTER_STATE(name:attribute)% operator resolution:
// (1) straightforward expansion, (2) first-match on shared name, (3) filter-chain isolation.
var ExtProcAccessLogTest = suite.ConformanceTest{
	ShortName:   "ExtProcAccessLog",
	Description: "Verify %EG_EXT_FILTER_STATE% operator expansion in access logs with ext-proc",
	Manifests:   []string{"testdata/ext-proc-service.yaml", "testdata/ext-proc-accesslog.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		gatewayNS := GetGatewayResourceNamespace()

		// File sink logs land in the envoy container stdout which fluent-bit collects.
		fileLogLabels := map[string]string{
			"job":       fmt.Sprintf("%s/envoy", gatewayNS),
			"namespace": gatewayNS,
			"container": "envoy",
		}

		// Wait for the shared grpc-ext-proc backend to be ready.
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, &PodReady)

		// ── Scenario 1: straightforward expansion ────────────────────────────────
		t.Run("Straightforward expansion", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "ext-proc-accesslog-route1", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep1", Namespace: ns},
				suite.ControllerName, ancestorRef)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "accesslog.example.com",
					Path: "/route1",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			// Match prefix + path to confirm the operator resolved and this specific route logged.
			runLogTest(t, suite, gwAddr, &expectedResponse, fileLogLabels, "ep_lat=.*path=/route1", 1)
		})

		// ── Scenario 2: first-match / override ───────────────────────────────────
		t.Run("First-match on shared name", func(t *testing.T) {
			route1NN := types.NamespacedName{Name: "ext-proc-accesslog-route1", Namespace: ns}
			route2NN := types.NamespacedName{Name: "ext-proc-accesslog-route2", Namespace: ns}
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-gtw", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false,
				route1NN, route2NN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep2", Namespace: ns},
				suite.ControllerName, ancestorRef)

			// Both routes execute their own ext-proc (per-route override); access log
			// operator resolution uses first-match. Match per-path to verify both log.
			for _, path := range []string{"/route1", "/route2"} {
				resp := httputils.ExpectedResponse{
					Request: httputils.Request{
						Host: "accesslog.example.com",
						Path: path,
					},
					Response: httputils.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				}
				// Match the path in the log line to confirm this specific route logged.
				runLogTest(t, suite, gwAddr, &resp, fileLogLabels, "ep_lat=.*path="+path, 1)
			}
		})

		// ── Scenario 3: filter-chain level isolation ─────────────────────────────
		t.Run("Filter-chain isolation", func(t *testing.T) {
			routeANN := types.NamespacedName{Name: "ext-proc-accesslog-route-a", Namespace: ns}
			routeBNN := types.NamespacedName{Name: "ext-proc-accesslog-route-b", Namespace: ns}
			// Routes bind to separate listeners (8084/8085) via sectionName; each listener
			// has its own HCM and resolves "isolated-proc" independently.
			gwNN := types.NamespacedName{Name: "ext-proc-accesslog-iso-gtw", Namespace: ns}

			gwAddrA := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeANN)
			gwAddrB := kubernetes.GatewayAndRoutesMustBeAccepted(
				t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeBNN)

			ancestorRefA := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8084"),
			}
			ancestorRefB := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: gatewayapi.SectionNamePtr("http-8085"),
			}
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep-a", Namespace: ns},
				suite.ControllerName, ancestorRefA)
			EnvoyExtensionPolicyMustBeAccepted(t, suite.Client,
				types.NamespacedName{Name: "ext-proc-accesslog-eep-b", Namespace: ns},
				suite.ControllerName, ancestorRefB)

			// Both listeners share the same format; each resolves "isolated-proc" against
			// its own filter chain, producing independent log lines.
			respA := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "accesslog-iso.example.com",
					Path: "/isolated-a",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			runLogTest(t, suite, gwAddrA, &respA, fileLogLabels, "ep_lat_iso=.*path=/isolated-a", 1)

			respB := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "accesslog-iso.example.com",
					Path: "/isolated-b",
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}
			runLogTest(t, suite, gwAddrB, &respB, fileLogLabels, "ep_lat_iso=.*path=/isolated-b", 1)
		})
	},
}
