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
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

// testOIDCWithAuthorizationDeny verifies that an authentication-independent
// authorization deny rule (clientCIDRs) is enforced before the OIDC filter on
// the same route. A client in the denied CIDR must receive a 403, not a 302
// redirect to the identity provider, while an allowed client is still
// redirected to the IdP (proving OIDC remains in effect after the pre-auth
// deny). See https://github.com/envoyproxy/gateway/issues/8913.
//
// It is invoked from OIDCTest so that it can reuse the already-running Keycloak
// instance instead of standing up its own identity provider.
func testOIDCWithAuthorizationDeny(t *testing.T, suite *suite.ConformanceTestSuite) {
	t.Run("authorization deny is enforced before the OIDC redirect (#8913)", func(t *testing.T) {
		const ns = "gateway-conformance-infra"

		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/oidc-authorization.yaml", true)

		routeNN := types.NamespacedName{Name: "http-with-oidc-authz", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-authz", Namespace: ns}, suite.ControllerName, ancestorRef)

		// A client in the denied CIDR must be rejected with 403 before the OIDC
		// filter runs. Before the fix this returned a 302 redirect to Keycloak.
		t.Run("denied client gets 403 instead of an OIDC redirect", func(t *testing.T) {
			expectedResponse := gwhttp.ExpectedResponse{
				Request: gwhttp.Request{
					Host:             "www.example.com",
					Path:             "/protected-oidc",
					UnfollowRedirect: true,
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.1.1", // in the denied CIDR
					},
				},
				Response: gwhttp.Response{
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}
			gwhttp.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		// A client outside the denied CIDR is not short-circuited by the deny and
		// is redirected to the identity provider by the OIDC filter, confirming
		// OIDC still applies after the pre-auth authorization stage.
		t.Run("allowed client is still redirected to the IdP", func(t *testing.T) {
			expectedResponse := gwhttp.ExpectedResponse{
				Request: gwhttp.Request{
					Host:             "www.example.com",
					Path:             "/protected-oidc",
					UnfollowRedirect: true,
					Headers: map[string]string{
						"X-Forwarded-For": "192.168.3.1", // outside the denied CIDR
					},
				},
				Response: gwhttp.Response{
					StatusCodes: []int{302},
				},
				Namespace: ns,
			}
			gwhttp.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	})
}
