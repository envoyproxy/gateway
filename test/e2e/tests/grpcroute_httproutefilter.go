// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"google.golang.org/grpc/codes"
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
// HTTPRouteFilter via an extensionRef filter. It uses an HTTPRouteFilter that contributes
// a cookie match, which is ANDed with the rule's method match, and asserts that routing is
// enforced end-to-end for gRPC traffic: a request carrying the cookie is routed to the
// backend, while a request without it does not match the route.
var GRPCRouteHTTPRouteFilterTest = suite.ConformanceTest{
	ShortName:   "GRPCRouteHTTPRouteFilter",
	Description: "GRPCRoute referencing an Envoy Gateway HTTPRouteFilter via extensionRef",
	Manifests:   []string{"testdata/grpcroute-httproutefilter.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("cookie match contributed by HTTPRouteFilter is enforced for gRPC", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "grpcroute-httproutefilter", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.GRPCRoute{}, true, routeNN)

			// A request that carries the "canary=true" cookie matches the rule and is
			// routed to the backend.
			matchResp := grpc.ExpectedResponse{
				EchoRequest: &grpcechoserver.EchoRequest{},
				RequestMetadata: &grpc.RequestMetadata{
					Metadata: map[string]string{"cookie": "canary=true"},
				},
				Backend:   "grpc-infra-backend-v1",
				Namespace: ns,
			}
			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, matchResp)

			// A request without the cookie does not match the route. gRPC surfaces an
			// unmatched route (HTTP 404) as the Unimplemented status code.
			noMatchResp := grpc.ExpectedResponse{
				EchoRequest: &grpcechoserver.EchoRequest{},
				Response:    grpc.Response{Code: codes.Unimplemented},
			}
			grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, noMatchResp)
		})
	},
}
