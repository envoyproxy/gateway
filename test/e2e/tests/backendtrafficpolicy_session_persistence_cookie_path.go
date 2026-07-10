// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"fmt"
	stdhttp "net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTrafficPolicySessionPersistenceCookiePathTest)
}

var BackendTrafficPolicySessionPersistenceCookiePathTest = suite.ConformanceTest{
	ShortName:   "BackendTrafficPolicySessionPersistenceCookiePath",
	Description: "Test that BackendTrafficPolicy sessionPersistence.cookie.path overrides the Set-Cookie Path the client receives",
	Manifests:   []string{"testdata/backendtrafficpolicy-session-persistence-cookie-path.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "session-persistence-cookie-path", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(
			t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
			kubernetes.NewGatewayRef(gwNN), routeNN,
		)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		BackendTrafficPolicyMustBeAccepted(t, suite.Client,
			types.NamespacedName{Name: "session-persistence-cookie-path-override", Namespace: ns},
			suite.ControllerName, ancestorRef)

		requestPath := "/cookie-session-persistence-cookie-path"
		res := &http.ExpectedResponse{
			Request: http.Request{
				Path: requestPath,
			},
			Response: http.Response{
				StatusCodes: []int{stdhttp.StatusOK},
			},
			Namespace: ns,
			Backend:   "infra-backend-v1",
		}
		req := http.MakeRequest(t, res, gwAddr, "HTTP", "http")

		// Make sure backend is ready.
		http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, *res)

		t.Run("cookie path is overridden to /", func(t *testing.T) {
			pod := ""
			// We make 10 requests to the gateway and expect them to be routed to the same pod.
			for i := range 10 {
				captReq, resp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Fatalf("failed to make request: %v", err)
				}

				if i == 0 {
					// First request, capture the pod name and cookie.
					if captReq.Pod == "" {
						t.Fatalf("expected pod to be set")
					}

					cookie, err := parseCookie(resp.Headers, "Session-A")
					if err != nil {
						t.Fatalf("failed to parse cookie: %v", err)
					}

					// Without the override the cookie Path would be derived from the
					// matched route path (requestPath). The BackendTrafficPolicy
					// sessionPersistence.cookie.path override must pin it to "/".
					if diff := cmp.Diff(cookie, &stdhttp.Cookie{
						Name:     "Session-A",
						MaxAge:   10,
						Path:     "/",
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
	},
}
