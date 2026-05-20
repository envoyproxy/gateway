// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
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
	ConformanceTests = append(ConformanceTests, EnvoyExtensionPolicyMergedTest)
}

var EnvoyExtensionPolicyMergedTest = suite.ConformanceTest{
	ShortName:   "EnvoyExtensionPolicyMerged",
	Description: "Test route-level EnvoyExtensionPolicy merge with parent Gateway policy",
	Manifests:   []string{"testdata/ext-proc-service.yaml", "testdata/envoyextensionpolicy-merged.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("EnvoyExtensionPolicyMerged", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "eep-merged-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
				t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				routeNN,
			)

			ancestorRef := gwapiv1.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: new(gwapiv1.SectionName("http")),
			}

			EnvoyExtensionPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "eep-merged-route", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			EnvoyExtensionPolicyMustBeMerged(t,
				suite.Client,
				types.NamespacedName{Name: "eep-merged-route", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			podReady := corev1.PodCondition{Type: corev1.PodReady, Status: corev1.ConditionTrue}
			WaitForPods(t, suite.Client, ns, map[string]string{"app": "grpc-ext-proc"}, corev1.PodRunning, &podReady)

			expectedResponse := httputils.ExpectedResponse{
				Request: httputils.Request{
					Host: "www.example.com",
					Path: "/merged",
					Headers: map[string]string{
						"x-request-client-header": "original",
					},
				},
				ExpectedRequest: &httputils.ExpectedRequest{
					Request: httputils.Request{
						Path: "/merged",
						Headers: map[string]string{
							"x-request-ext-processed":          "true",
							"x-request-client-header-received": "original",
							"x-request-client-header":          "mutated",
						},
					},
				},
				Response: httputils.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"X-Custom-Lua-Header":      "merged-parent",
						"x-response-ext-processed": "true",
					},
				},
				Namespace: ns,
			}

			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
