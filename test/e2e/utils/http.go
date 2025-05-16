// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package utils

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

// MakeRequestAndExpectEventuallyConsistentResponse sends a request to the gateway and waits for an eventually consistent response.
// This's a fork of upstream unless https://github.com/kubernetes-sigs/gateway-api/issues/3794 fixed.
func MakeRequestAndExpectEventuallyConsistentResponse(t *testing.T, suite *suite.ConformanceTestSuite, gwAddr string, expected http.ExpectedResponse) {
	t.Helper()

	req := http.MakeRequest(t, &expected, gwAddr, "HTTP", "http")

	http.AwaitConvergence(t, suite.TimeoutConfig.RequiredConsecutiveSuccesses, suite.TimeoutConfig.MaxTimeToConsistency, func(elapsed time.Duration) bool {
		cReq, cRes, err := suite.RoundTripper.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "Request failed, not ready yet: %v (after %v)", err.Error(), elapsed)
			return false
		}

		if err := compareRequest(t, &req, cReq, cRes, expected); err != nil {
			tlog.Logf(t, "Response expectation failed for request: %+v  not ready yet: %v (after %v)", req, err, elapsed)
			return false
		}

		return true
	})
	tlog.Logf(t, "Request passed")
}

// GatewaysMustBeAccepted waits for the Gateways to be accepted and returns the address of the Gateways.
// This is used when a HTTPRoute referenced by multiple Gateways.
// Warning: we didn't check the status of HTTPRoute.
func GatewaysMustBeAccepted(t *testing.T, suite *suite.ConformanceTestSuite, gwRefs []kubernetes.GatewayRef) map[types.NamespacedName]string {
	t.Helper()

	gwAddress := make(map[types.NamespacedName]string)
	requiredListenerConditions := []metav1.Condition{
		{
			Type:   string(gwapiv1.ListenerConditionResolvedRefs),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
		{
			Type:   string(gwapiv1.ListenerConditionAccepted),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
		{
			Type:   string(gwapiv1.ListenerConditionProgrammed),
			Status: metav1.ConditionTrue,
			Reason: "", // any reason
		},
	}

	for _, gw := range gwRefs {
		gwAddr, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, gw)
		require.NoErrorf(t, err, "timed out waiting for Gateway %s address to be assigned", gw.NamespacedName)
		gwAddress[gw.NamespacedName] = gwAddr

		kubernetes.GatewayListenersMustHaveConditions(t, suite.Client, suite.TimeoutConfig, gw.NamespacedName, requiredListenerConditions)
	}

	return gwAddress
}

func compareRequest(t *testing.T, req *roundtripper.Request, cReq *roundtripper.CapturedRequest, cRes *roundtripper.CapturedResponse, expected http.ExpectedResponse) error {
	if roundtripper.IsTimeoutError(cRes.StatusCode) {
		if roundtripper.IsTimeoutError(expected.Response.StatusCode) {
			return nil
		}
	}
	if expected.Response.StatusCode != cRes.StatusCode {
		return fmt.Errorf("expected status code to be %d, got %d", expected.Response.StatusCode, cRes.StatusCode)
	}
	if cRes.StatusCode == 200 {
		// The request expected to arrive at the backend is
		// the same as the request made, unless otherwise
		// specified.
		if expected.ExpectedRequest == nil {
			expected.ExpectedRequest = &http.ExpectedRequest{Request: expected.Request}
		}

		if expected.ExpectedRequest.Method == "" {
			expected.ExpectedRequest.Method = "GET"
		}

		if expected.ExpectedRequest.Host != "" && expected.ExpectedRequest.Host != cReq.Host {
			return fmt.Errorf("expected host to be %s, got %s", expected.ExpectedRequest.Host, cReq.Host)
		}

		if expected.ExpectedRequest.Path != cReq.Path {
			return fmt.Errorf("expected path to be %s, got %s", expected.ExpectedRequest.Path, cReq.Path)
		}
		if expected.ExpectedRequest.Method != cReq.Method {
			return fmt.Errorf("expected method to be %s, got %s", expected.ExpectedRequest.Method, cReq.Method)
		}
		if expected.Namespace != cReq.Namespace {
			return fmt.Errorf("expected namespace to be %s, got %s", expected.Namespace, cReq.Namespace)
		}
		if expected.ExpectedRequest.Headers != nil {
			if cReq.Headers == nil {
				return fmt.Errorf("no headers captured, expected %v", len(expected.ExpectedRequest.Headers))
			}
			for name, val := range cReq.Headers {
				cReq.Headers[strings.ToLower(name)] = val
			}
			for name, expectedVal := range expected.ExpectedRequest.Headers {
				actualVal, ok := cReq.Headers[strings.ToLower(name)]
				if !ok {
					return fmt.Errorf("expected %s header to be set, actual headers: %v", name, cReq.Headers)
				} else if strings.Join(actualVal, ",") != expectedVal {
					return fmt.Errorf("expected %s header to be set to %s, got %s", name, expectedVal, strings.Join(actualVal, ","))
				}
			}
		}

		if expected.Response.Headers != nil {
			if cRes.Headers == nil {
				return fmt.Errorf("no headers captured, expected %v", len(expected.ExpectedRequest.Headers))
			}
			for name, val := range cRes.Headers {
				cRes.Headers[strings.ToLower(name)] = val
			}

			for name, expectedVal := range expected.Response.Headers {
				actualVal, ok := cRes.Headers[strings.ToLower(name)]
				if !ok {
					return fmt.Errorf("expected %s header to be set, actual headers: %v", name, cRes.Headers)
				}

				if expectedVal == "" {
					// If the expected value is empty, we don't care about the actual value.
					// This is useful for headers that are set by the backend, and we don't
					// care about their values.
					continue
				}

				if strings.Join(actualVal, ",") != expectedVal {
					return fmt.Errorf("expected %s header to be set to %s, got %s", name, expectedVal, strings.Join(actualVal, ","))
				}
			}
		}

		if len(expected.Response.AbsentHeaders) > 0 {
			for name, val := range cRes.Headers {
				cRes.Headers[strings.ToLower(name)] = val
			}

			for _, name := range expected.Response.AbsentHeaders {
				val, ok := cRes.Headers[strings.ToLower(name)]
				if ok {
					return fmt.Errorf("expected %s header to not be set, got %s", name, val)
				}
			}
		}

		// Verify that headers expected *not* to be present on the
		// request are actually not present.
		if len(expected.ExpectedRequest.AbsentHeaders) > 0 {
			for name, val := range cReq.Headers {
				cReq.Headers[strings.ToLower(name)] = val
			}

			for _, name := range expected.ExpectedRequest.AbsentHeaders {
				val, ok := cReq.Headers[strings.ToLower(name)]
				if ok {
					return fmt.Errorf("expected %s header to not be set, got %s", name, val)
				}
			}
		}

		if !strings.HasPrefix(cReq.Pod, expected.Backend) {
			return fmt.Errorf("expected pod name to start with %s, got %s", expected.Backend, cReq.Pod)
		}
	} else if roundtripper.IsRedirect(cRes.StatusCode) {
		if expected.RedirectRequest == nil {
			return nil
		}

		setRedirectRequestDefaults(req, cRes, &expected)

		if expected.RedirectRequest.Host != cRes.RedirectRequest.Host {
			return fmt.Errorf("expected redirected hostname to be %q, got %q", expected.RedirectRequest.Host, cRes.RedirectRequest.Host)
		}

		gotPort := cRes.RedirectRequest.Port
		if expected.RedirectRequest.Port == "" {
			// If the test didn't specify any expected redirect port, we'll try to use
			// the scheme to determine sensible defaults for the port. Well known
			// schemes like "http" and "https" MAY skip setting any port.
			if strings.ToLower(cRes.RedirectRequest.Scheme) == "http" && gotPort != "80" && gotPort != "" {
				return fmt.Errorf("for http scheme, expected redirected port to be 80 or not set, got %q", gotPort)
			}
			if strings.ToLower(cRes.RedirectRequest.Scheme) == "https" && gotPort != "443" && gotPort != "" {
				return fmt.Errorf("for https scheme, expected redirected port to be 443 or not set, got %q", gotPort)
			}
			if strings.ToLower(cRes.RedirectRequest.Scheme) != "http" || strings.ToLower(cRes.RedirectRequest.Scheme) != "https" {
				tlog.Logf(t, "Can't validate redirectPort for unrecognized scheme %v", cRes.RedirectRequest.Scheme)
			}
		} else if expected.RedirectRequest.Port != gotPort {
			// An expected port was specified in the tests but it didn't match with
			// gotPort.
			return fmt.Errorf("expected redirected port to be %q, got %q", expected.RedirectRequest.Port, gotPort)
		}

		if expected.RedirectRequest.Scheme != cRes.RedirectRequest.Scheme {
			return fmt.Errorf("expected redirected scheme to be %q, got %q", expected.RedirectRequest.Scheme, cRes.RedirectRequest.Scheme)
		}

		if expected.RedirectRequest.Path != cRes.RedirectRequest.Path {
			return fmt.Errorf("expected redirected path to be %q, got %q", expected.RedirectRequest.Path, cRes.RedirectRequest.Path)
		}
	}
	return nil
}

func setRedirectRequestDefaults(req *roundtripper.Request, cRes *roundtripper.CapturedResponse, expected *http.ExpectedResponse) {
	// If the expected host is nil it means we do not test host redirect.
	// In that case we are setting it to the one we got from the response because we do not know the ip/host of the gateway.
	if expected.RedirectRequest.Host == "" {
		expected.RedirectRequest.Host = cRes.RedirectRequest.Host
	}

	if expected.RedirectRequest.Scheme == "" {
		expected.RedirectRequest.Scheme = req.URL.Scheme
	}

	if expected.RedirectRequest.Path == "" {
		expected.RedirectRequest.Path = req.URL.Path
	}
}
