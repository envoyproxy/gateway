// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	nethttp "net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests,
		HTTPExtAuthTest,
		HTTPExtAuthHeadersPrefixTest,
		HTTPExtAuthHeadersSuffixTest,
		HTTPExtAuthHeadersRegexTest,
	)
}

// HTTPExtAuthTest tests ExtAuth authentication for an http route with ExtAuth configured.
// The http route points to an application to verify that ExtAuth authentication works on application/http path level.
// The ExtAuth service is an HTTP service.
var HTTPExtAuthTest = suite.ConformanceTest{
	ShortName:   "HTTPExtAuth",
	Description: "Test HTTP ExtAuth authentication",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-http-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-ext-auth", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		// Wait for the http ext auth service pod to be ready
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("http route with ext auth authentication", func(t *testing.T) {
			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Host: "www.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization": "Bearer token1",
						"x-test-foo":    "bar",
					},
				},
				// Verify that the http headers returned by the ext auth service
				// are added to the original request before sending it to the backend
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"x-current-user": "user1",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			t.Run("verify headersToExtAuthOnMatch - Exact match", func(t *testing.T) {
				require.Eventually(t, func() bool {
					testReq := http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"Authorization": "Bearer token1",
							"x-test-foo":    "bar",
						},
					}
					req := http.MakeRequest(t, &http.ExpectedResponse{Request: testReq, Namespace: ns}, gwAddr, "HTTP", "http")
					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						t.Logf("request failed: %v", err)
						return false
					}

					if cResp.StatusCode != nethttp.StatusOK {
						return false
					}

					currentHeaders, ok := cReq.Headers["X-Current-Headers"]
					if !ok || len(currentHeaders) == 0 {
						t.Logf("X-Current-Headers not found in backend request headers")
						return false
					}

					t.Logf("X-Current-Headers received by backend: %s", currentHeaders[0])

					// Verify that x-test-foo header is included (exact match)
					return strings.Contains(strings.ToLower(currentHeaders[0]), "x-test-foo: bar")
				}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second)
			})

			t.Run("verify headersToExtAuthOnMatch - header not matching exact should not be included", func(t *testing.T) {
				require.Eventually(t, func() bool {
					testReq := http.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"Authorization":    "Bearer token1",
							"x-test-foo":       "bar",
							"x-test-foo-extra": "should-not-be-included",
						},
					}
					req := http.MakeRequest(t, &http.ExpectedResponse{Request: testReq, Namespace: ns}, gwAddr, "HTTP", "http")
					cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
					if err != nil {
						t.Logf("request failed: %v", err)
						return false
					}

					if cResp.StatusCode != nethttp.StatusOK {
						return false
					}

					currentHeaders, ok := cReq.Headers["X-Current-Headers"]
					if !ok || len(currentHeaders) == 0 {
						t.Logf("X-Current-Headers not found in backend request headers")
						return false
					}

					t.Logf("X-Current-Headers received by backend: %s", currentHeaders[0])

					// Verify x-test-foo is included
					if !strings.Contains(strings.ToLower(currentHeaders[0]), "x-test-foo: bar") {
						return false
					}

					// Verify x-test-foo-extra is NOT included (doesn't match exact)
					return !strings.Contains(strings.ToLower(currentHeaders[0]), "x-test-foo-extra")
				}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second)
			})
		})
	},
}

// HTTPExtAuthHeadersPrefixTest tests headersToExtAuthOnMatch with Prefix match type
var HTTPExtAuthHeadersPrefixTest = suite.ConformanceTest{
	ShortName:   "HTTPExtAuthHeadersPrefix",
	Description: "Test headersToExtAuthOnMatch with Prefix match",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-http-headers-prefix.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-ext-auth-prefix", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("prefix match - headers with matching prefix should be included", func(t *testing.T) {
			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-prefix", Namespace: ns}, suite.ControllerName, ancestorRef)

			require.Eventually(t, func() bool {
				testReq := http.Request{
					Host: "prefix.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization":  "Bearer token1",
						"x-prefix-foo":   "value1",
						"x-prefix-bar":   "value2",
						"x-other-header": "should-not-be-included",
					},
				}
				req := http.MakeRequest(t, &http.ExpectedResponse{Request: testReq, Namespace: ns}, gwAddr, "HTTP", "http")
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Logf("request failed: %v", err)
					return false
				}

				if cResp.StatusCode != nethttp.StatusOK {
					return false
				}

				currentHeaders, ok := cReq.Headers["X-Current-Headers"]
				if !ok || len(currentHeaders) == 0 {
					t.Logf("X-Current-Headers not found in backend request headers")
					return false
				}

				t.Logf("X-Current-Headers received by backend: %s", currentHeaders[0])
				headerStr := strings.ToLower(currentHeaders[0])

				// Verify headers with prefix "x-prefix-" are included
				hasPrefix1 := strings.Contains(headerStr, "x-prefix-foo: value1")
				hasPrefix2 := strings.Contains(headerStr, "x-prefix-bar: value2")
				hasOther := strings.Contains(headerStr, "x-other-header")

				return hasPrefix1 && hasPrefix2 && !hasOther
			}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second)
		})
	},
}

// HTTPExtAuthHeadersSuffixTest tests headersToExtAuthOnMatch with Suffix match type
var HTTPExtAuthHeadersSuffixTest = suite.ConformanceTest{
	ShortName:   "HTTPExtAuthHeadersSuffix",
	Description: "Test headersToExtAuthOnMatch with Suffix match",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-http-headers-suffix.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-ext-auth-suffix", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("suffix match - headers with matching suffix should be included", func(t *testing.T) {
			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-suffix", Namespace: ns}, suite.ControllerName, ancestorRef)

			require.Eventually(t, func() bool {
				testReq := http.Request{
					Host: "suffix.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization":      "Bearer token1",
						"x-user-metadata":    "user123",
						"x-session-metadata": "session456",
						"x-other-header":     "should-not-be-included",
					},
				}
				req := http.MakeRequest(t, &http.ExpectedResponse{Request: testReq, Namespace: ns}, gwAddr, "HTTP", "http")
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Logf("request failed: %v", err)
					return false
				}

				if cResp.StatusCode != nethttp.StatusOK {
					return false
				}

				currentHeaders, ok := cReq.Headers["X-Current-Headers"]
				if !ok || len(currentHeaders) == 0 {
					t.Logf("X-Current-Headers not found in backend request headers")
					return false
				}

				t.Logf("X-Current-Headers received by backend: %s", currentHeaders[0])
				headerStr := strings.ToLower(currentHeaders[0])

				// Verify headers with suffix "-metadata" are included
				hasSuffix1 := strings.Contains(headerStr, "x-user-metadata: user123")
				hasSuffix2 := strings.Contains(headerStr, "x-session-metadata: session456")
				hasOther := strings.Contains(headerStr, "x-other-header")

				return hasSuffix1 && hasSuffix2 && !hasOther
			}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second)
		})
	},
}

// HTTPExtAuthHeadersRegexTest tests headersToExtAuthOnMatch with RegularExpression match type
var HTTPExtAuthHeadersRegexTest = suite.ConformanceTest{
	ShortName:   "HTTPExtAuthHeadersRegex",
	Description: "Test headersToExtAuthOnMatch with RegularExpression match",
	Manifests:   []string{"testdata/ext-auth-service.yaml", "testdata/ext-auth-http-headers-regex.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-ext-auth-regex", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
		WaitForPods(t, suite.Client, ns, map[string]string{"app": "envoy-ext-auth"}, corev1.PodRunning, &PodReady)

		t.Run("regex match - headers matching pattern should be included", func(t *testing.T) {
			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "ext-auth-regex", Namespace: ns}, suite.ControllerName, ancestorRef)

			require.Eventually(t, func() bool {
				testReq := http.Request{
					Host: "regex.example.com",
					Path: "/myapp",
					Headers: map[string]string{
						"Authorization":           "Bearer token1",
						"x-custom-user-header":    "value1", // matches ^x-custom-[a-z]+-header$
						"x-custom-trace-header":   "value2", // matches ^x-custom-[a-z]+-header$
						"x-custom-123-header":     "value3", // does NOT match (has digits)
						"x-custom-header":         "value4", // does NOT match (missing middle part)
						"x-custom-multi-word-hdr": "value5", // does NOT match (has hyphen in middle, ends with hdr not header)
					},
				}
				req := http.MakeRequest(t, &http.ExpectedResponse{Request: testReq, Namespace: ns}, gwAddr, "HTTP", "http")
				cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
				if err != nil {
					t.Logf("request failed: %v", err)
					return false
				}

				if cResp.StatusCode != nethttp.StatusOK {
					return false
				}

				currentHeaders, ok := cReq.Headers["X-Current-Headers"]
				if !ok || len(currentHeaders) == 0 {
					t.Logf("X-Current-Headers not found in backend request headers")
					return false
				}

				t.Logf("X-Current-Headers received by backend: %s", currentHeaders[0])
				headerStr := strings.ToLower(currentHeaders[0])

				// Verify headers matching regex are included
				hasMatch1 := strings.Contains(headerStr, "x-custom-user-header: value1")
				hasMatch2 := strings.Contains(headerStr, "x-custom-trace-header: value2")

				// Verify headers NOT matching regex are excluded
				hasNoMatch1 := !strings.Contains(headerStr, "x-custom-123-header")
				hasNoMatch2 := !strings.Contains(headerStr, "x-custom-header: value4")
				hasNoMatch3 := !strings.Contains(headerStr, "x-custom-multi-word-hdr")

				return hasMatch1 && hasMatch2 && hasNoMatch1 && hasNoMatch2 && hasNoMatch3
			}, suite.TimeoutConfig.MaxTimeToConsistency, time.Second)
		})
	},
}
