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
	"github.com/envoyproxy/gateway/test/e2e/utils"
)

func init() {
	ConformanceTests = append(ConformanceTests, LocalRateLimitDistinctCIDRTest)
}

var LocalRateLimitDistinctCIDRTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimitDistinctCIDR",
	Description: "Test that local rate limit filter works with distinct cidr",
	Manifests:   []string{"testdata/local-ratelimit-distinct-cidr.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-ratelimit-distinct-cidr", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}

		t.Run("requests with x-forwarded-for header should be limited per IP", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr"
			testRatelimit(t, suite, map[string]string{
				"X-Forwarded-For": "192.168.1.1",
				"x-org-id":        "",
			}, ns, gwAddr, path)
			testRatelimit(t, suite, map[string]string{
				"X-Forwarded-For": "192.168.1.2",
				"x-org-id":        "",
			}, ns, gwAddr, path)
		})

		t.Run("requests with x-forwarded-for header and matching x-org-id header should be limited per IP", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr-and-exact-header"
			testRatelimit(t, suite, map[string]string{
				"X-Forwarded-For": "192.168.1.1",
				"x-org-id":        "foo",
			}, ns, gwAddr, path)
			testRatelimit(t, suite, map[string]string{
				"X-Forwarded-For": "192.168.1.2",
				"x-org-id":        "foo",
			}, ns, gwAddr, path)
		})

		t.Run("requests with with x-forwarded-for header but no matching x-org-id header will hit default bucket", func(t *testing.T) {
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr-and-exact-header"

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: path,
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.1.1",
						"x-org-id":        "bar",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path:    path,
						Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
					},
				},
				Response: http.Response{
					StatusCode: 429,
					Headers: map[string]string{
						RatelimitLimitHeaderName:     "10", // this means it hit the default bucket
						RatelimitRemainingHeaderName: "0",
					},
				},
				Namespace: ns,
			})
		})
	},
}

func testRatelimit(t *testing.T, suite *suite.ConformanceTestSuite, headers map[string]string, ns, gwAddr, path string) {
	utils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite, gwAddr, http.ExpectedResponse{
		Request: http.Request{
			Path:    path,
			Headers: headers,
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
				RatelimitLimitHeaderName:     "3",
				RatelimitRemainingHeaderName: "", // empty string means we don't care about the value
			},
		},
		Namespace: ns,
	})

	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
		Request: http.Request{
			Path:    path,
			Headers: headers,
		},
		ExpectedRequest: &http.ExpectedRequest{
			Request: http.Request{
				Path:    path,
				Headers: nil, // don't check headers since Envoy will append the client IP to the X-Forwarded-For header
			},
		},
		Response: http.Response{
			StatusCode: 429,
			Headers: map[string]string{
				RatelimitLimitHeaderName:     "3",
				RatelimitRemainingHeaderName: "0", // at the end the remaining should be 0
			},
		},
		Namespace: ns,
	})
}
