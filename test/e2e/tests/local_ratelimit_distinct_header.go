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

		t.Run("requests with x-user-id header should be limited per user", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header"
			testDistinctHeaderRatelimit(t, "john", "", ns, gwAddr, path, true, suite)
			testDistinctHeaderRatelimit(t, "alice", "", ns, gwAddr, path, true, suite)
		})

		t.Run("requests with x-user-id header and matching x-org-id header should be limited per user", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header-and-exact-header"
			testDistinctHeaderRatelimit(t, "john", "foo", ns, gwAddr, path, true, suite)
			testDistinctHeaderRatelimit(t, "alice", "foo", ns, gwAddr, path, true, suite)
		})

		t.Run("requests with x-user-id header but no matching x-org-id header should not be limited", func(t *testing.T) {
			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			BackendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-distinct-header-and-exact-header", Namespace: ns}, suite.ControllerName, ancestorRef)
			path := "/ratelimit-distinct-header-and-exact-header"
			testDistinctHeaderRatelimit(t, "john", "bar", ns, gwAddr, path, false, suite)
		})
	},
}

func testDistinctHeaderRatelimit(t *testing.T, user, org, ns, gwAddr, path string, limited bool, suite *suite.ConformanceTestSuite) {
	expectOkResp := http.ExpectedResponse{
		Request: http.Request{
			Path: path,
			Headers: map[string]string{
				"x-user-id": user,
				"x-org-id":  org,
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
				"x-user-id": user,
				"x-org-id":  org,
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
