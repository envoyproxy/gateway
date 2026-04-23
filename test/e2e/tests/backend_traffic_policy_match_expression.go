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
	ConformanceTests = append(ConformanceTests, BackendTrafficPolicyMatchExpressionTest)
}

var BackendTrafficPolicyMatchExpressionTest = suite.ConformanceTest{
	ShortName:   "BackendTrafficPolicyMatchExpression",
	Description: "Use match expression to select one HTTPRoute",
	Manifests:   []string{"testdata/backend-traffic-policy-match-expression.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("BackendTrafficPolicyMatchExpression", func(t *testing.T) {
			ns := "gateway-conformance-infra"

			routeNormal := types.NamespacedName{Name: "backend-traffic-match-expression-normal", Namespace: ns}
			routeInjected := types.NamespacedName{Name: "backend-traffic-match-expression-injected", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				routeNormal,
				routeInjected,
			)

			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "backend-traffic-match-expression", Namespace: ns},
				suite.ControllerName,
				gwapiv1.ParentReference{
					Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:      gatewayapi.KindPtr(resource.KindGateway),
					Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:      gwapiv1.ObjectName(gwNN.Name),
				},
			)

			normalResponse := http.ExpectedResponse{
				Namespace: ns,
				Request: http.Request{
					Host: "normal.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t,
				suite.RoundTripper,
				suite.TimeoutConfig,
				gwAddr,
				normalResponse,
			)

			injectedResponse := http.ExpectedResponse{
				Namespace: ns,
				Request: http.Request{
					Host: "injected.example.com",
					Path: "/",
				},
				Response: http.Response{
					StatusCodes: []int{418},
				},
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t,
				suite.RoundTripper,
				suite.TimeoutConfig,
				gwAddr,
				injectedResponse,
			)
		})
	},
}
