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

		t.Run("requests with x-forwarded-for header should be limited per IP", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr"
			testDistinctCIDRRatelimit(t, "192.168.1.1", "", ns, gwAddr, path, true, suite)
			testDistinctCIDRRatelimit(t, "192.168.1.2", "", ns, gwAddr, path, true, suite)
		})

		t.Run("requests with x-forwarded-for header and matching x-org-id header should be limited per IP", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr-and-exact-header"
			testDistinctCIDRRatelimit(t, "192.168.1.1", "foo", ns, gwAddr, path, true, suite)
			testDistinctCIDRRatelimit(t, "192.168.1.2", "foo", ns, gwAddr, path, true, suite)
		})

		t.Run("requests with with x-forwarded-for header but no matching x-org-id header should not be limited", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-cidr-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-cidr-and-exact-header"
			testDistinctCIDRRatelimit(t, "192.168.1.1", "bar", ns, gwAddr, path, false, suite)
		})
	},
}

func testDistinctCIDRRatelimit(t *testing.T, clientIP, org, ns, gwAddr, path string, limited bool, suite *suite.ConformanceTestSuite) {
	expectOkResp := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
			Headers: map[string]string{
				"X-Forwarded-For": clientIP,
				"x-org-id":        org,
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
		},
		Namespace: ns,
	}

	expectOkReq := http.MakeRequest(t, &expectOkResp, gwAddr, "HTTP", "http")

	expectLimitResp := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
			Headers: map[string]string{
				"X-Forwarded-For": clientIP,
				"x-org-id":        org,
			},
		},
		Response: http.Response{
			StatusCode: 429,
		},
		Namespace: ns,
	}
	expectLimitReq := http.MakeRequest(t, &expectLimitResp, gwAddr, "HTTP", "http")

	// should just send exactly 4 requests, and expect 429

	// keep sending requests till get 200 first, that will cost one 200
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

	// fire the rest request
	if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
		t.Errorf("fail to get expected response at first three request: %v", err)
	}

	if limited {
		// this request should be limited because the limit is 3
		if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
			t.Errorf("fail to get expected response at the fourth request: %v", err)
		}
	} else {
		if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectOkResp); err != nil {
			t.Errorf("fail to get expected response at the fourth request: %v", err)
		}
	}
}
