// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteHTTPRouteFilterTest)
}

// GRPCRouteHTTPRouteFilterTest verifies that a GRPCRoute can reference an Envoy Gateway
// HTTPRouteFilter via an extensionRef filter. It uses a credential-injection filter and
// asserts the injected header reaches the upstream, confirming the filter is resolved,
// translated, and enforced end-to-end for gRPC traffic.
var GRPCRouteHTTPRouteFilterTest = suite.ConformanceTest{
	ShortName:   "GRPCRouteHTTPRouteFilter",
	Description: "GRPCRoute referencing an Envoy Gateway HTTPRouteFilter via extensionRef",
	Manifests:   []string{"testdata/grpcroute-httproutefilter.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("inject credential into gRPC request via HTTPRouteFilter", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "grpcroute-httproutefilter", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.GRPCRoute{}, true, routeNN)

			okResp := grpc.ExpectedResponse{
				EchoRequest: &grpcechoserver.EchoRequest{},
				Backend:     "grpc-infra-backend-v1",
				Namespace:   ns,
			}

			// Ensure the route with the HTTPRouteFilter is reachable and returns OK.
			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, okResp)

			// The credential-injection HTTPRouteFilter must have added the x-credential
			// header to the request received by the upstream gRPC server.
			resp, err := suite.GRPCClient.SendRPC(t, gwAddr, okResp, suite.TimeoutConfig.MaxTimeToConsistency)
			require.NoError(t, err, "failed to send gRPC request")

			const expectedHeader = "x-credential"
			const expectedValue = "Basic dXNlcjE6dGVzdDI="
			found := false
			for _, h := range resp.Response.GetAssertions().GetHeaders() {
				if strings.EqualFold(h.GetKey(), expectedHeader) && h.GetValue() == expectedValue {
					found = true
					break
				}
			}
			require.Truef(t, found, "expected upstream request to contain header %q=%q injected by the HTTPRouteFilter, got headers %v",
				expectedHeader, expectedValue, resp.Response.GetAssertions().GetHeaders())
		})
	},
}
