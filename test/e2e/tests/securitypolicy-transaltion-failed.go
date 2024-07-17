// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

func init() {
	ConformanceTests = append(ConformanceTests, OIDCTest)
}

// OIDCTest tests OIDC authentication for an http route with OIDC configured.
// The http route points to an application to verify that OIDC authentication works on application/http path level.
var OIDCTest = suite.ConformanceTest{
	ShortName:   "SecurityPolicyFailedTranslation",
	Description: "Test direct 500 response when failed to translate SecurityPolicy",
	Manifests:   []string{"testdata/oidc-keycloak.yaml", "testdata/securitypolicy-translation-failed.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with failed SecurityPolicy", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-oidc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwapiv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustFail(t, suite.Client, types.NamespacedName{Name: "oidc-test", Namespace: ns}, suite.ControllerName, ancestorRef, "")

			// TODO: We should wait for the `programmed` condition to be true before sending traffic.
			expectedResponse := http.ExpectedResponse{
				Request: gwhttp.Request{
					Host: "www.example.com",
					Path: "/myapp",
				},
				Response: gwhttp.Response{
					StatusCode: 500,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
