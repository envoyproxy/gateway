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

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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
		routeNN := types.NamespacedName{Name: "grpc-json-transcoder", Namespace: ns}

		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
			suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)
		// The GRPCRoute has to be programmed for the no-re-match assertion to mean anything.
		kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig,
			suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.GRPCRoute{}, false, routeNN)
		SecurityPolicyMustBeAccepted(t, suite.Client, routeNN, suite.ControllerName, gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		})

		// From EchoTwo's google.api.http option.
		url := fmt.Sprintf("http://%s/v1/grpc-echo/echo-two", gwAddr)

		// Bounded per attempt: without a timeout one hung request consumes the whole poll.
		client := &http.Client{Timeout: 5 * time.Second}
		// pollUntil sends the JSON request until check accepts the response.
		pollUntil := func(t *testing.T, authorization string, check func(*http.Response, []byte) error) {
			t.Helper()
			var lastErr error
			err := wait.PollUntilContextTimeout(context.Background(), time.Second,
				suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
					req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
					if err != nil {
						return false, err
					}
					req.Host = "transcoder.example.com"
					if authorization != "" {
						req.Header.Set("Authorization", authorization)
					}

					res, err := client.Do(req)
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
					if lastErr = check(res, raw); lastErr != nil {
						return false, nil
					}
					return true, nil
				})
			if err != nil {
				t.Fatalf("%s: %v (last error: %v)", url, err, lastErr)
			}
		}

		// Basic auth runs before the transcoder, so this holds with or without a re-match; the
		// x-matched-grpcroute check below is what catches one.
		t.Run("SecurityPolicy on a transcoding HTTPRoute is enforced", func(t *testing.T) {
			pollUntil(t, "", func(res *http.Response, raw []byte) error {
				if res.StatusCode != http.StatusUnauthorized {
					return fmt.Errorf("expected 401, got %d: %s", res.StatusCode, raw)
				}
				return nil
			})
		})

		t.Run("transcoded request stays on the HTTPRoute", func(t *testing.T) {
			var body grpcEchoResponse
			pollUntil(t, "Basic dXNlcjE6dGVzdDE=", func(res *http.Response, raw []byte) error { // user1:test1
				if res.StatusCode != http.StatusOK {
					return fmt.Errorf("expected 200, got %d: %s", res.StatusCode, raw)
				}
				if ct := res.Header.Get("content-type"); !strings.HasPrefix(ct, "application/json") {
					return fmt.Errorf("expected a JSON content-type, got %q", ct)
				}
				if gs := res.Header.Get("grpc-status"); gs != "0" {
					return fmt.Errorf("expected grpc-status 0, got %q", gs)
				}
				body = grpcEchoResponse{}
				if err := json.Unmarshal(raw, &body); err != nil {
					return fmt.Errorf("response is not the transcoded EchoResponse: %w: %s", err, raw)
				}
				return nil
			})

			if got := body.Assertions.FullyQualifiedMethod; got != grpcEchoMethod {
				t.Errorf("expected the backend to see method %s, got %s", grpcEchoMethod, got)
			}

			var upstreamContentType string
			for _, h := range body.Assertions.Headers {
				switch strings.ToLower(h.Key) {
				case "content-type":
					upstreamContentType = h.Value
				case "x-matched-grpcroute":
					t.Errorf("transcoded request was re-matched onto the GRPCRoute")
				}
			}
			if !strings.HasPrefix(upstreamContentType, "application/grpc") {
				t.Errorf("expected the backend to receive an application/grpc content-type, got %q", upstreamContentType)
			}
		})
	},
}
