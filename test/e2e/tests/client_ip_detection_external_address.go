// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, ClientIPDetectionExternalAddressTest)
}

var ClientIPDetectionExternalAddressTest = suite.ConformanceTest{
	ShortName:   "ClientIPDetectionExternalAddress",
	Description: "numTrustedHops client IP detection sets x-envoy-external-address to the trusted client address",
	Manifests:   []string{"testdata/client-ip-detection-external-address.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		routeNN := types.NamespacedName{Name: "http-with-client-ip-detection-external-address", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		ClientTrafficPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "client-ip-detection-external-address-ctp", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("x-envoy-external-address is the trusted X-Forwarded-For entry", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/client-ip-detection-external-address",
					Headers: map[string]string{
						// With numTrustedHops=1 the rightmost entry is the trusted client.
						"X-Forwarded-For": "1.2.3.4",
						// Sent by the client, so Envoy must replace it rather than forward it.
						"X-Envoy-External-Address": "9.9.9.9",
					},
				},
				ExpectedRequest: &http.ExpectedRequest{
					Request: http.Request{
						Path: "/client-ip-detection-external-address",
						Headers: map[string]string{
							"X-Envoy-External-Address": "1.2.3.4",
						},
					},
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
