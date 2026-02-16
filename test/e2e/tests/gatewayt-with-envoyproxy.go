// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, GatewayWithEnvoyProxy)
}

var GatewayWithEnvoyProxy = suite.ConformanceTest{
	ShortName:   "GatewayWithEnvoyProxy",
	Description: "Attach an EnvoyProxy to a Gateway",
	Manifests:   []string{"testdata/gateway-with-envoyproxy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Attach an EnvoyProxy to a Gateway and set RoutingType to Service", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "gateway-with-envoyproxy", Namespace: ns}
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			backendNN := types.NamespacedName{Name: "infra-backend-v1", Namespace: ns}
			svc := corev1.Service{}
			require.NoError(t, suite.Client.Get(context.Background(), backendNN, &svc))

			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/basic-auth-1",
				},
				Response: http.Response{
					StatusCodes: []int{200},

					// Verify that the RouteType is set to Service by the attached EnvoyProxy
					Headers: map[string]string{
						"upstream-host": net.JoinHostPort(svc.Spec.ClusterIP, "8080"),
					},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
