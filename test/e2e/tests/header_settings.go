// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, HeaderSettingsTest)
}

var HeaderSettingsTest = suite.ConformanceTest{
	ShortName:   "HeaderSettings",
	Description: "Modify headers before regular processing",
	Manifests:   []string{"testdata/header-settings.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Early header modifications should apply", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-early-headers", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "early-header-modifier-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/early-header",
					Headers: map[string]string{
						"early-added-header":   "client",
						"early-set-header":     "client",
						"early-removed-header": "client",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/early-header",
						Headers: map[string]string{
							"early-added-header":   "client,early,late", // client, early and late are all added to header
							"early-set-header":     "early,late",        // early set overwrites client value
							"early-removed-header": "late",              // removed by early, so only late value exists
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})

		t.Run("Late header modifications should apply", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-late-headers", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "early-header-modifier-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/late-header",
				},
				BackendSetResponseHeaders: map[string]string{
					"late-added-header":   "backend",
					"late-set-header":     "backend",
					"late-removed-header": "backend",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"late-added-header": "backend,filter,late",
						"late-set-header":   "late",
					},
					AbsentHeaders: []string{
						"late-removed-header",
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)
		})
	},
}
