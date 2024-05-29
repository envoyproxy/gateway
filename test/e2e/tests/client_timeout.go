// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	httputils "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ClientTimeoutTest)
}

var largeSizeHeader = func() string {
	var b strings.Builder
	for i := 0; i < 5000; i++ {
		b.WriteString("FakeHeaderValue")
	}
	return b.String()
}

var ClientTimeoutTest = suite.ConformanceTest{
	ShortName:   "ClientTimeout",
	Description: "Test that the ClientTrafficPolicy API implementation supports client timeout",
	Manifests:   []string{"testdata/client-timeout.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http client timeout", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-client-timeout", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// Use raw http request to avoid chunked
			req := &http.Request{
				Method: "GET",
				URL:    &url.URL{Scheme: "http", Host: gwAddr, Path: "/request-timeout"},
				Header: http.Header{
					// larger enough to trigger request timeout
					"x-large-size-header": []string{largeSizeHeader()},
				},
			}

			client := &http.Client{}

			httputils.AwaitConvergence(t,
				suite.TimeoutConfig.RequiredConsecutiveSuccesses,
				suite.TimeoutConfig.MaxTimeToConsistency,
				func(elapsed time.Duration) bool {
					resp, err := client.Do(req)
					if err != nil {
						panic(err)
					}
					defer resp.Body.Close()

					// return 408 instead of 400 when request timeout.
					if http.StatusRequestTimeout == resp.StatusCode {
						return true
					} else {
						t.Logf("response status code: %d, (after %v) ", resp.StatusCode, elapsed)
						return false
					}
				})
		})
	},
}
