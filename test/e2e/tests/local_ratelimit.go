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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, LocalRateLimitTest)
}

var LocalRateLimitTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimit",
	Description: "Make sure local rate limit works",
	Manifests:   []string{"testdata/local-ratelimit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		// let make sure the gateway and http route are accepted
		// and there's no rate limit on this route
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-no-ratelimit", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		expectOkResp := http.ExpectedResponse{
			Request: http.Request{
				Path: "/no-ratelimit",
			},
			Response: http.Response{
				StatusCode: 200,
			},
			Namespace: ns,
		}
		expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

		// keep sending requests till get 200 first, that will cost one 200
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

		// the requests should not be limited because there is no rate limit on this route
		if err := GotExactExpectedResponse(t, 10, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
			t.Errorf("fail to get expected response at last fourth request: %v", err)
		}

		t.Run("SpecificUser", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-specific-user", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-specific-user", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-specific-user",
					Headers: map[string]string{
						"x-user-id": "john",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest request
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}

			// this request should be limited because the user is john
			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-specific-user",
					Headers: map[string]string{
						"x-user-id": "john",
					},
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectLimitResp)

			// this request should not be limited because the user is not john
			expectOkResp = http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-specific-user",
					Headers: map[string]string{
						"x-user-id": "mike",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq = http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")
			// the requests should not be limited because the user is mike
			if err := GotExactExpectedResponse(t, 4, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}
		})

		t.Run("AllTraffic", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-all-traffic", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-all-traffic", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-all-traffic",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest request
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}

			// this request should be limited at the end
			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-all-traffic",
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectLimitResp)
		})

		t.Run("HeaderInvertMatch", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-invert-match", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-invert-match", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectOkResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-invert-match",
					Headers: map[string]string{
						"x-user-id": "one",
						"x-org-id":  "org1",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest request
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}

			// this request should be limited because the user is one and org is not test and the limit is 3
			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-invert-match",
					Headers: map[string]string{
						"x-user-id": "one",
						"x-org-id":  "org1",
					},
				},
				Response: http.Response{
					StatusCode: 429,
				},
				Namespace: ns,
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectLimitResp)

			// with test org
			expectOkResp = http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-invert-match",
					Headers: map[string]string{
						"x-user-id": "one",
						"x-org-id":  "test",
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}
			expectOkReq = http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")
			// the requests should not be limited because the user is one but org is test
			if err := GotExactExpectedResponse(t, 4, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}
		})
	},
}
