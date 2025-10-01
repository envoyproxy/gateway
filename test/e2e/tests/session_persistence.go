// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/types"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, HeaderBasedSessionPersistenceTest, CookieBasedSessionPersistenceTest)
}

var HeaderBasedSessionPersistenceTest = suite.ConformanceTest{
	ShortName:   "HeaderBasedSessionPersistence",
	Description: "Test that the session persistence filter is correctly configured with header based session persistence",
	Manifests:   []string{"testdata/header-based-session-persistence.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("traffic is routed based on header based session persistence", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "header-based-session-persistence", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			req := httputils.MakeRequest(t, &httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v2",
				},
			}, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v2",
				},
				Response: httputils.Response{
					StatusCodes: []int{http.StatusOK},
				},
				Namespace: ns,
			})

			pod := ""
			// We make 10 requests to the gateway and expect them to be routed to the same pod.
			for i := range 10 {
				captReq, res, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}

				if i == 0 {
					// First request, capture the pod name and header.
					sessionHeader, ok := res.Headers["Session-A"]
					if !ok {
						t.Fatalf("expected header Session-A to be set: %v", res.Headers)
					}

					if captReq.Pod == "" {
						t.Fatalf("expected pod to be set")
					}
					pod = captReq.Pod
					req.Headers["Session-A"] = sessionHeader
					continue
				}

				t.Logf("request is received from pod %s", captReq.Pod)

				if captReq.Pod != pod {
					t.Fatalf("expected pod to be the same as previous requests")
				}
			}
		})
		t.Run("session persistence is configured per route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "header-based-session-persistence", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			req := httputils.MakeRequest(t, &httputils.ExpectedResponse{
				Request: httputils.Request{
					// /v1 path does not have the session persistence.
					Path: "/v1",
				},
			}, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v1",
				},
				Response: httputils.Response{
					StatusCodes: []int{http.StatusOK},
				},
				Namespace: ns,
			})

			_, res, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}

			if h, ok := res.Headers["Session-A"]; ok {
				t.Fatalf("expected header Session-A to not be set: %v", h)
			}
		})
	},
}

var CookieBasedSessionPersistenceTest = suite.ConformanceTest{
	ShortName:   "CookieBasedSessionPersistence",
	Description: "Test that the session persistence filter is correctly configured with cookie based session persistence",
	Manifests:   []string{"testdata/cookie-based-session-persistence.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("traffic is routed based on cookie based session persistence", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "cookie-based-session-persistence", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			req := httputils.MakeRequest(t, &httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v2",
				},
			}, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v2",
				},
				Response: httputils.Response{
					StatusCodes: []int{http.StatusOK},
				},
				Namespace: ns,
			})

			pod := ""
			// We make 10 requests to the gateway and expect them to be routed to the same pod.
			for i := range 10 {
				captReq, res, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}

				if i == 0 {
					// First request, capture the pod name and cookie.
					if captReq.Pod == "" {
						t.Fatalf("expected pod to be set")
					}

					cookie, err := parseCookie(res.Headers, "Session-A")
					if err != nil {
						t.Fatalf("failed to parse cookie: %v", err)
					}

					// Check the cookie is set correctly.
					if diff := cmp.Diff(cookie, &http.Cookie{
						Name:     "Session-A",
						MaxAge:   10,
						Path:     "/v2",
						HttpOnly: true,
						Quoted:   true,
					}, cmpopts.IgnoreFields(http.Cookie{}, "Value", "Raw"), // Ignore the value as it is random.
					); diff != "" {
						t.Fatalf("unexpected cookie: %v", diff)
					}

					pod = captReq.Pod
					req.Headers["Cookie"] = []string{fmt.Sprintf("Session-A=%s", cookie.Value)}
					continue
				}

				t.Logf("request is received from pod %s", captReq.Pod)

				if captReq.Pod != pod {
					t.Fatalf("expected pod to be the same as previous requests")
				}
			}
		})
		t.Run("session persistence is configured per route", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "cookie-based-session-persistence", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
			req := httputils.MakeRequest(t, &httputils.ExpectedResponse{
				Request: httputils.Request{
					// /v1 path does not have the session persistence.
					Path: "/v1",
				},
			}, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			httputils.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, httputils.ExpectedResponse{
				Request: httputils.Request{
					Path: "/v1",
				},
				Response: httputils.Response{
					StatusCodes: []int{http.StatusOK},
				},
				Namespace: ns,
			})

			_, res, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Fatalf("failed to make request: %v", err)
			}

			if _, ok := res.Headers["Set-Cookie"]; ok {
				t.Fatal("expected the envoy not to response set-cookie back")
			}
		})
	},
}

func parseCookie(headers map[string][]string, cookieName string) (*http.Cookie, error) {
	parser := &http.Response{Header: headers}
	for _, c := range parser.Cookies() {
		if c.Name == cookieName {
			return c, nil
		}
	}
	return nil, fmt.Errorf("cookie %s not found: headers: %v", cookieName, headers)
}
