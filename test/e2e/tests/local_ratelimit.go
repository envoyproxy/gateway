// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func init() {
	ConformanceTests = append(ConformanceTests, LocalRateLimitSpecificUserTest)
	ConformanceTests = append(ConformanceTests, LocalRateLimitAllTrafficTest)
	ConformanceTests = append(ConformanceTests, LocalRateLimitNoLimitRouteTest)
}

var LocalRateLimitSpecificUserTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimitSpecificUser",
	Description: "Limit a specific user",
	Manifests:   []string{"testdata/local-ratelimit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("limit a specific user", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-specific-user", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			backendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-specific-user", Namespace: ns})

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
			expectLimitReq := http.MakeRequest(t, &expectLimitResp, gwAddr, "HTTP", "http")

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// fire the rest request
			if err := GotExactExpectedResponse(t, 2, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at first three request: %v", err)
			}

			// this request should be limited because the user is john and the limit is 3
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
				t.Errorf("fail to get expected response at last fourth request: %v", err)
			}

			// test another user
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
	},
}

var LocalRateLimitAllTrafficTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimitAllTraffic",
	Description: "Limit all traffic on a route",
	Manifests:   []string{"testdata/local-ratelimit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("limit all traffic on a route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-ratelimit-all-traffic", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			backendTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ratelimit-all-traffic", Namespace: ns})

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

			expectLimitResp := http.ExpectedResponse{
				Request: http.Request{
					Path: "/ratelimit-all-traffic",
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

			// this request should be limited because the limit is 3
			if err := GotExactExpectedResponse(t, 1, suite.RoundTripper, expectLimitReq, expectLimitResp); err != nil {
				t.Errorf("fail to get expected response at last fourth request: %v", err)
			}
		})
	},
}

var LocalRateLimitNoLimitRouteTest = suite.ConformanceTest{
	ShortName:   "LocalRateLimitNoLimitRoute",
	Description: "No rate limit on this route",
	Manifests:   []string{"testdata/local-ratelimit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("no rate limit on this route", func(t *testing.T) {
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

			// should just send exactly 4 requests, and expect 429

			// keep sending requests till get 200 first, that will cost one 200
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectOkResp)

			// the requests should not be limited because there is no rate limit on this route
			if err := GotExactExpectedResponse(t, 3, suite.RoundTripper, expectOkReq, expectOkResp); err != nil {
				t.Errorf("fail to get expected response at last fourth request: %v", err)
			}
		})
	},
}

// backendTrafficPolicyMustBeAccepted waits for the specified BackendTrafficPolicy to be accepted.
func backendTrafficPolicyMustBeAccepted(
	t *testing.T,
	client client.Client,
	policyName types.NamespacedName) {
	t.Helper()

	waitErr := wait.PollUntilContextTimeout(context.Background(), 1*time.Second, 60*time.Second, true, func(ctx context.Context) (bool, error) {
		policy := &egv1a1.BackendTrafficPolicy{}
		err := client.Get(ctx, policyName, policy)
		if err != nil {
			return false, fmt.Errorf("error fetching BackendTrafficPolicy: %w", err)
		}

		for _, condition := range policy.Status.Conditions {
			if condition.Type == string(gwv1a2.PolicyConditionAccepted) && condition.Status == metav1.ConditionTrue {
				return true, nil
			}
		}
		t.Logf("BackendTrafficPolicy not yet accepted: %v", policy)
		return false, nil
	})
	require.NoErrorf(t, waitErr, "error waiting for BackendTrafficPolicy to be accepted")
}
