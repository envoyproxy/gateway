// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCJSONTranscoderTest)
}

// The method EchoTwo maps to, as reported back by the backend.
const grpcEchoMethod = "/gateway_api_conformance.echo_basic.grpcecho.GrpcEcho/EchoTwo"

// The conformance echo helpers can't be used here: they parse the plain HTTP echo body,
// whereas a transcoded response is a gRPC EchoResponse rendered as JSON.
type grpcEchoResponse struct {
	Assertions struct {
		FullyQualifiedMethod string `json:"fullyQualifiedMethod"`
		Headers              []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"headers"`
	} `json:"assertions"`
}

var GRPCJSONTranscoderTest = suite.ConformanceTest{
	ShortName:   "GRPCJSONTranscoder",
	Description: "Transcode a JSON/HTTP request into a gRPC call using the gRPC-JSON transcoder",
	Manifests:   []string{"testdata/grpc-json-transcoder.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}

		// By default the transcoder rewrites :path to the gRPC method and Envoy matches
		// routes again, so the rewritten path needs a route of its own -- both ways of
		// providing one are exercised here. matchIncomingRequestRoute opts out of the
		// re-match and needs no second route at all.
		for _, tc := range []struct {
			name          string
			route         string
			host          string
			postTranscode string
			// grpcRoute, when set, serves the rewritten path and must be programmed
			// before the transcoded request can succeed.
			grpcRoute string
		}{
			{
				name:          "second HTTPRoute rule serves the rewritten path",
				route:         "grpc-json-transcoder-httproute",
				host:          "httproute.transcoder.example.com",
				postTranscode: "HTTPRoute",
			},
			{
				name:          "GRPCRoute serves the rewritten path",
				route:         "grpc-json-transcoder-grpcroute-httproute",
				host:          "grpcroute.transcoder.example.com",
				postTranscode: "GRPCRoute",
				grpcRoute:     "grpc-json-transcoder-grpcroute",
			},
			{
				name:          "matchIncomingRequestRoute keeps the original route",
				route:         "grpc-json-transcoder-single-route",
				host:          "single.transcoder.example.com",
				postTranscode: "none, the matched route is kept",
			},
		} {
			t.Run(tc.name, func(t *testing.T) {
				routeNN := types.NamespacedName{Name: tc.route, Namespace: ns}
				gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
					suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

				// Without this the transcoded request races the GRPCRoute being
				// programmed and is answered 404 until it catches up.
				if tc.grpcRoute != "" {
					kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
						suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.GRPCRoute{}, false,
						types.NamespacedName{Name: tc.grpcRoute, Namespace: ns})
				}

				// From EchoTwo's google.api.http option.
				url := fmt.Sprintf("http://%s/v1/grpc-echo/echo-two", gwAddr)

				var body grpcEchoResponse
				var lastErr error
				err := wait.PollUntilContextTimeout(context.Background(), time.Second,
					suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
						req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
						if err != nil {
							return false, err
						}
						req.Host = tc.host

						res, err := http.DefaultClient.Do(req)
						if err != nil {
							lastErr = err
							return false, nil
						}
						defer res.Body.Close()

						raw, err := io.ReadAll(res.Body)
						if err != nil {
							lastErr = err
							return false, nil
						}

						if res.StatusCode != http.StatusOK {
							lastErr = fmt.Errorf("expected 200, got %d: %s", res.StatusCode, raw)
							return false, nil
						}
						if ct := res.Header.Get("content-type"); !strings.HasPrefix(ct, "application/json") {
							lastErr = fmt.Errorf("expected a JSON content-type, got %q", ct)
							return false, nil
						}
						if gs := res.Header.Get("grpc-status"); gs != "0" {
							lastErr = fmt.Errorf("expected grpc-status 0, got %q", gs)
							return false, nil
						}

						body = grpcEchoResponse{}
						if err := json.Unmarshal(raw, &body); err != nil {
							lastErr = fmt.Errorf("response is not the transcoded EchoResponse: %w: %s", err, raw)
							return false, nil
						}
						return true, nil
					})
				if err != nil {
					t.Fatalf("never got a transcoded response from %s (host %s, %s post-transcode route): %v (last error: %v)",
						url, tc.host, tc.postTranscode, err, lastErr)
				}

				if got := body.Assertions.FullyQualifiedMethod; got != grpcEchoMethod {
					t.Errorf("expected the backend to see method %s, got %s", grpcEchoMethod, got)
				}

				var upstreamContentType string
				for _, h := range body.Assertions.Headers {
					if strings.EqualFold(h.Key, "content-type") {
						upstreamContentType = h.Value
						break
					}
				}
				if !strings.HasPrefix(upstreamContentType, "application/grpc") {
					t.Errorf("expected the backend to receive an application/grpc content-type, got %q", upstreamContentType)
				}
			})
		}
	},
}
