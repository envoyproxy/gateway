// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, SecurityPolicyMergedTest)
}

var SecurityPolicyMergedTest = suite.ConformanceTest{
	ShortName:   "SecurityPolicyMerged",
	Description: "Test section level policy attach and merged parent policy for SecurityPolicy",
	Manifests:   []string{"testdata/securitypolicy-merged.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("SecurityPolicyMerged", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "sp-merged-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "sp-merged", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				routeNN,
			)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:        gatewayapi.KindPtr(resource.KindGateway),
				Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:        gwapiv1.ObjectName(gwNN.Name),
				SectionName: ptr.To(gwapiv1.SectionName("listener-1")),
			}

			// Verify gateway-level policy is accepted
			SecurityPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "sp-merged-gateway", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			// Verify route-level policy has Merged condition
			SecurityPolicyMustBeMerged(t,
				suite.Client,
				types.NamespacedName{Name: "sp-merged-route", Namespace: ns},
				suite.ControllerName,
				ancestorRef,
			)

			// Test that merged policies work - Authorization (from gateway) allows, CORS (from route) adds headers
			expectedResponse := httputils.ExpectedResponse{
				Namespace: ns,
				Request: httputils.Request{
					Host: "listener1.merged.example.com",
					Path: "/merged",
					Headers: map[string]string{
						"Origin": "https://www.example.com",
					},
				},
				Response: httputils.Response{
					StatusCode: 200,
					Headers: map[string]string{
						"Access-Control-Allow-Origin": "https://www.example.com",
					},
				},
			}
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
