// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
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

		t.Run("exhausted user stays limited after more than 20 distinct users", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header"
			headers := map[string]string{"x-user-id": "cache-retention-original"}

			// Exhaust the original user's three-token bucket before creating enough
			// other buckets to evict it from Envoy's default 20-entry cache.
			testRatelimit(t, suite, headers, ns, gwAddr, path)
			for i := range 20 {
				expectedResp := http.ExpectedResponse{
					Request: http.Request{
						Path:    path,
						Headers: map[string]string{"x-user-id": fmt.Sprintf("cache-retention-%d", i)},
					},
					Response:  http.Response{StatusCodes: []int{200}},
					Namespace: ns,
				}
				req := http.MakeRequest(t, &expectedResp, gwAddr, "HTTP", "http")
				cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
				require.NoError(t, err)
				require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cRes, expectedResp))
			}

			// Check the very next response without polling: retries could exhaust a
			// newly recreated bucket and hide the eviction. The manifest uses an
			// hourly limit so the original bucket cannot refill during this test.
			expectedResp := http.ExpectedResponse{
				Request: http.Request{
					Path:    path,
					Headers: headers,
				},
				Response:  http.Response{StatusCodes: []int{429}},
				Namespace: ns,
			}
			req := http.MakeRequest(t, &expectedResp, gwAddr, "HTTP", "http")
			cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
			require.NoError(t, err)
			require.NoError(t, http.CompareRoundTrip(t, &req, cReq, cRes, expectedResp))
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
			expectedResp := http.ExpectedResponse{
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
					StatusCodes: []int{429},
					Headers: map[string]string{
						RatelimitLimitHeaderName:     "10", // this means it hit the default bucket
						RatelimitRemainingHeaderName: "0",  // this means the default bucket is exhausted, with 429 response.
					},
				},
				Namespace: ns,
			}
			MakeRequestAndExpectEventuallyConsistentResponseExceptErrors(t, suite.RoundTripper, &suite.TimeoutConfig, gwAddr, &expectedResp)
		})
	},
}
