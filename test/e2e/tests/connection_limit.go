// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"errors"
	"net"
	"net/url"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, ConnectionLimitTest)
}

var ConnectionLimitTest = suite.ConformanceTest{
	ShortName:   "ConnectionLimit",
	Description: "Deny All Requests",
	Manifests:   []string{"testdata/connection-limit.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Close connections over limit", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-connection-limit", Namespace: ns}
			gwNN := types.NamespacedName{Name: "connection-limit-gateway", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			// make a request
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/",
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := http.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}

			// open some connections
			for i := 0; i < 10; i++ {
				conn, err := net.DialTimeout("tcp", gwAddr, 100*time.Millisecond)
				if err == nil {
					defer conn.Close()
				} else {
					t.Errorf("failed to open connection: %v", err)
				}
			}

			// new requests now fail
			req = http.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			_, _, err = suite.RoundTripper.CaptureRoundTrip(req)

			// expect error
			if err != nil {
				urlError := &url.Error{}
				if errors.As(err, &urlError) {
					if urlError.Err.Error() != "EOF" {
						t.Errorf("expected EOF when connection limit is reached")
					}
				} else {
					t.Errorf("expected net/url error when connection limit is reached")
				}
			} else {
				t.Errorf("expected error when connection limit is reached")
			}

		})
	},
}
