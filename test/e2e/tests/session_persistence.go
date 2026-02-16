// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, SessionPersistenceTest)
}

var SessionPersistenceTest = suite.ConformanceTest{
	ShortName:   "SessionPersistence",
	Description: "Test that the session persistence filter is correctly configured with header/cookie based session persistence",
	Manifests:   []string{"testdata/httproute-session-persistence.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "session-persistence", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		t.Run("session persistence is configured per route", func(t *testing.T) {
			// Make sure backend is ready
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Path: "/no-session-persistence",
				},
				Response: http.Response{
					AbsentHeaders: []string{"Session-A", "Set-Cookie"},
					StatusCodes:   []int{stdhttp.StatusOK},
				},
				Namespace: ns,
				Backend:   "infra-backend-v1",
			})
		})
		t.Run("header based session persistence", func(t *testing.T) {
			res := &http.ExpectedResponse{
				Request: http.Request{
					Path: "/header-session-persistence",
				},
				Namespace: ns,
				Backend:   "infra-backend-v2",
			}
			req := http.MakeRequest(t, res, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, *res)

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
		t.Run("cookie based session persistence", func(t *testing.T) {
			requestPath := "/cookie-session-persistence"
			res := &http.ExpectedResponse{
				Request: http.Request{
					Path: requestPath,
				},
				Response: http.Response{
					StatusCodes: []int{stdhttp.StatusOK},
				},
				Namespace: ns,
				Backend:   "infra-backend-v3",
			}
			req := http.MakeRequest(t, res, gwAddr, "HTTP", "http")

			// Make sure backend is ready
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, *res)

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
					if diff := cmp.Diff(cookie, &stdhttp.Cookie{
						Name:     "Session-A",
						MaxAge:   10,
						Path:     requestPath,
						HttpOnly: true,
						Quoted:   true,
					}, cmpopts.IgnoreFields(stdhttp.Cookie{}, "Value", "Raw"), // Ignore the value as it is random.
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

		// TODO: remove this after https://github.com/kubernetes-sigs/gateway-api/pull/4350 merged.
		t.Run("cookie based session persistence multiple backends", func(t *testing.T) {
			expected := http.ExpectedResponse{
				Request: http.Request{
					Path: "/cookie-session-persistence-multiple-backends",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expected)

			backend := runCookieBasedSessionPersistenceTest(t, suite, gwAddr, &expected)
			// the default lifetimecycle of cookie based session persistennce is Session,
			// we need to retry until the cookie expires and a new backend is selected.
			err := wait.PollUntilContextTimeout(t.Context(), time.Second, time.Minute, true, func(_ context.Context) (done bool, err error) {
				newBackend := runCookieBasedSessionPersistenceTest(t, suite, gwAddr, &expected)
				if newBackend != backend {
					t.Logf("backend changed from %s to %s, as expected", backend, newBackend)
					return true, nil
				}
				t.Logf("backend %s is the same as before, retrying...", newBackend)
				return false, nil
			})
			if err != nil {
				tlog.Errorf(t, "failed to check stats encoding: %v", err)
			}
		})
	},
}

// runCookieBasedSessionPersistenceTest makes 10 requests with the same cookie and expects them to be routed to the same backend/pod.
// which take precedence the weight when there are multiple backends, and returns the backend name.
func runCookieBasedSessionPersistenceTest(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expected *http.ExpectedResponse) string {
	req := http.MakeRequest(t, expected, gwAddr, "HTTP", "http")
	initialPod := ""
	backend := ""
	for i := range 10 {
		cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			t.Fatalf("request %d with cookie failed: %v", i+1, err)
		}
		if err := http.CompareRoundTrip(t, &req, cReq, cRes, *expected); err != nil {
			t.Fatalf("request %d with cookie failed expectations: %v", i+1, err)
		}

		if i == 0 {
			if cReq.Pod == "" {
				t.Fatalf("expected pod to be set")
			}
			cookie, err := parseCookie(cRes.Headers, "Session-A")
			if err != nil {
				t.Fatalf("failed to parse session persistence cookie: %v", err)
			}
			t.Logf("session persistence cookie: %s=%s pod: %s", cookie.Name, cookie.Value, cReq.Pod)
			req.Headers["Cookie"] = []string{fmt.Sprintf("%s=%s", cookie.Name, cookie.Value)}
			initialPod = cReq.Pod
			continue
		}
		if cReq.Pod != initialPod {
			t.Fatalf("expected session persistence to keep routing to pod %q, got %q", initialPod, cReq.Pod)
		}
		// The backend name is in the format of "infra-backend-v1-xxxx", "infra-backend-v2-xxxx",etc.
		// We only care about the prefix "infra-backend-v1", "infra-backend-v2", etc.
		backend = strings.Join(strings.Split(cReq.Pod, "-")[:3], "-")
	}
	return backend
}

func parseCookie(headers map[string][]string, cookieName string) (*stdhttp.Cookie, error) {
	parser := &stdhttp.Response{Header: headers}
	for _, c := range parser.Cookies() {
		if c.Name == cookieName {
			return c, nil
		}
	}
	return nil, fmt.Errorf("cookie %s not found: headers: %v", cookieName, headers)
}
