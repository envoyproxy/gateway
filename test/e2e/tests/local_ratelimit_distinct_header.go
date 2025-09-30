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
	ConformanceTests = append(ConformanceTests, LocalRateLimitDistinctHeaderTest)
}

var LocalRateLimitDistinctHeaderTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimitDistinctHeader",
	Description: "Test that local rate limit filter works with distinct header",
	Manifests:   []string{"testdata/local-ratelimit-distinct-header.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-ratelimit-distinct-header", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		t.Run("requests with x-user-id header should be limited per user", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header"

			testRatelimit(t, suite, map[string]string{
				"x-user-id": "john",
				"x-org-id":  "",
			}, ns, gwAddr, path)
			testRatelimit(t, suite, map[string]string{
				"x-user-id": "alice",
				"x-org-id":  "",
			}, ns, gwAddr, path)
		})

		t.Run("requests with x-user-id header and matching x-org-id header should be limited per user", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header-and-exact-header"

			testRatelimit(t, suite, map[string]string{
				"x-user-id": "john",
				"x-org-id":  "foo",
			}, ns, gwAddr, path)
			testRatelimit(t, suite, map[string]string{
				"x-user-id": "alice",
				"x-org-id":  "foo",
			}, ns, gwAddr, path)
		})

		t.Run("requests with x-user-id header but no matching x-org-id header will hit default bucket", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header-and-exact-header"
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: path,
					Headers: map[string]string{
						"x-user-id": "john",
						"x-org-id":  "bar",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    path,
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCode: 200,
					Headers: map[string]string{
						RatelimitLimitHeaderName:     "10", // this means it hit the default bucket
						RatelimitRemainingHeaderName: "4",
					},
				},
				Namespace: ns,
			})
		})
	},
}
