// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HTTPRouteMixedBackendsTest)
}

var HTTPRouteMixedBackendsTest = suite.ConformanceTest{
	ShortName:   "HTTPRouteMixedBackends",
	Description: "HTTPRoute rule that mixes Service and Backend refs should successfully route to both targets",
	Manifests: []string{
		"testdata/httproute-mixed-backends.yaml",
	},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := ConformanceInfraNamespace
		routeNN := types.NamespacedName{Name: "httproute-mixed-backends", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "backend-mixed-fqdn", Namespace: ns})

		kubernetes.NamespacesMustBeReady(t, suite.Client, suite.TimeoutConfig, []string{ns})

		expectedResponse := http.ExpectedResponse{
			Request: http.Request{
				Host: "mixed.example.com",
				Path: "/mixed",
			},
			Response: http.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}

		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

		req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")

		expectedBackends := map[string]struct{}{
			"infra-backend-v1": {},
			"infra-backend-v2": {},
		}
		seen := make(map[string]struct{})

		const maxRequests = 40
		for i := 0; i < maxRequests; i++ {
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("failed to capture round trip: %v", err)
			}

			if err := http.CompareRoundTrip(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Fatalf("failed to compare request and response: %v", err)
			}

			podName := cReq.Pod
			if podName == "" {
				t.Fatalf("missing pod information in captured request")
			}

			backend := backendFromPodName(podName)
			if backend == "" {
				t.Fatalf("unexpected pod name %q", podName)
			}

			if _, ok := expectedBackends[backend]; ok {
				seen[backend] = struct{}{}
			}

			if len(seen) == len(expectedBackends) {
				break
			}
		}

		if len(seen) != len(expectedBackends) {
			t.Fatalf("expected to see traffic routed to both infra-backend-v1 and infra-backend-v2, got %v", seen)
		}
	},
}

func backendFromPodName(podName string) string {
	switch {
	case strings.HasPrefix(podName, "infra-backend-v1"):
		return "infra-backend-v1"
	case strings.HasPrefix(podName, "infra-backend-v2"):
		return "infra-backend-v2"
	default:
		return ""
	}
}
