// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from upstream gateway-api, it will be moved to upstream.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/echo-basic/grpcechoserver"
	"sigs.k8s.io/gateway-api/conformance/utils/grpc"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GRPCRouteBackendFQDNTest, GRPCRouteBackendIPTest)
}

var GRPCRouteBackendFQDNTest = suite.ConformanceTest{
	ShortName:   "GRPCRouteBackendFQDNTest",
	Description: "GRPCRoutes with a backend ref to a FQDN Backend",
	Manifests:   []string{"testdata/grpcroute-to-backend-fqdn.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("grpc-route-1", func(t *testing.T) {
			testGRPCRouteWithBackend(t, suite, "backend-fqdn")
		})
	},
}

var GRPCRouteBackendIPTest = suite.ConformanceTest{
	ShortName:   "GRPCRouteBackendIPTest",
	Description: "GRPCRoutes with a backend ref to an IP Backend",
	Manifests:   []string{"testdata/grpcroute-to-backend-ip.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("grpc-route-1", func(t *testing.T) {
			svcNN := types.NamespacedName{
				Name:      "grpc-infra-backend-v1",
				Namespace: "gateway-conformance-infra",
			}
			svc, err := GetService(suite.Client, svcNN)
			if err != nil {
				t.Fatalf("failed to get service %s: %v", svcNN, err)
			}

			backendIPName := "backend-ip"
			ns := "gateway-conformance-infra"
			err = CreateBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}, svc.Spec.ClusterIP, 8080)
			if err != nil {
				t.Fatalf("failed to create backend %s: %v", backendIPName, err)
			}
			t.Cleanup(func() {
				if err := DeleteBackend(suite.Client, types.NamespacedName{Name: backendIPName, Namespace: ns}); err != nil {
					t.Fatalf("failed to delete backend %s: %v", backendIPName, err)
				}
			})
			testGRPCRouteWithBackend(t, suite, backendIPName)
		})
	},
}

func testGRPCRouteWithBackend(t *testing.T, suite *suite.ConformanceTestSuite, backendName string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: "exact-matching", Namespace: ns}
	gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.GRPCRoute{}, false, routeNN)
	BackendMustBeAccepted(t, suite.Client, types.NamespacedName{Name: backendName, Namespace: ns})
	OkResp := grpc.ExpectedResponse{
		EchoRequest: &grpcechoserver.EchoRequest{},
		Backend:     "grpc-infra-backend-v1",
		Namespace:   ns,
	}

	grpc.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.GRPCClient, suite.TimeoutConfig, gwAddr, OkResp)
}
