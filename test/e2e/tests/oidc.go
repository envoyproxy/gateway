// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"io"
	"net/http"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

const (
	testURL             = "http://www.example.com/myapp"
	logoutURL           = "http://www.example.com/myapp/logout"
	keyCloakLoginFormID = "kc-form-login"
	username            = "oidcuser"
	password            = "oidcpassword"
)

func init() {
	ConformanceTests = append(ConformanceTests, OIDCTest)
}

// OIDCTest tests OIDC authentication for an http route with OIDC configured.
// The http route points to an application to verify that OIDC authentication works on application/http path level.
var OIDCTest = suite.ConformanceTest{
	ShortName:   "OIDC",
	Description: "Test OIDC authentication",
	Manifests:   []string{"testdata/oidc-keycloak.yaml", "testdata/oidc-securitypolicy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http route with oidc authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-with-oidc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podInitialized := corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}

			// Wait for the keycloak pod to be configured with the test user and client
			WaitForPods(t, suite.Client, ns, map[string]string{"job-name": "setup-keycloak"}, corev1.PodSucceeded, podInitialized)

			// Initialize the test OIDC client that will keep track of the state of the OIDC login process
			client, err := NewOIDCTestClient(
				WithLoggingOptions(t.Log, true),
				// Map the application and keycloak cluster DNS name to the gateway address
				WithCustomAddressMappings(map[string]string{
					"www.example.com:80":                    gwAddr,
					"keycloak.gateway-conformance-infra:80": gwAddr,
				}),
			)
			require.NoError(t, err)

			// Send a request to the http route with OIDC configured.
			// It will be redirected to the keycloak login page
			res, err := client.Get(testURL, true)
			require.NoError(t, err)
			require.Equal(t, 200, res.StatusCode, "Expected 200 OK")

			// Parse the response body to get the URL where the login page would post the user-entered credentials
			require.NoError(t, client.ParseLoginForm(res.Body, keyCloakLoginFormID), "Failed to parse login form")

			// Submit the login form to the IdP.
			// This will authenticate and redirect back to the application
			res, err = client.Login(map[string]string{"username": username, "password": password, "credentialId": ""})
			require.NoError(t, err, "Failed to login to the IdP")

			// Verify that we get the expected response from the application
			body, err := io.ReadAll(res.Body)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Contains(t, string(body), "infra-backend-v1", "Expected response from the application")

			// Verify that we can access the application without logging in again
			res, err = client.Get(testURL, false)
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, res.StatusCode)
			require.Contains(t, string(body), "infra-backend-v1", "Expected response from the application")

			// Verify that we can logout
			// Note: OAuth2 filter just clears its cookies and does not log out from the IdP.
			res, err = client.Get(logoutURL, false)
			require.NoError(t, err)
			require.Equal(t, http.StatusFound, res.StatusCode)

			// After logout, OAuth2 filter will redirect back to the root of the host, e.g, "www.example.com".
			// Ideally, this should redirect to the application's root, e.g, "www.example.com/myapp",
			// but Envoy OAuth2 filter does not support this yet.
			require.Equal(t, "http://www.example.com/", res.Header.Get("Location"), "Expected redirect to the root of the host")

			// Verify that the oauth2 cookies have been deleted
			var cookieDeleted bool
			deletedCookies := res.Header.Values("Set-Cookie")
			regx := regexp.MustCompile("^IdToken-.+=deleted.+")
			for _, cookie := range deletedCookies {
				if regx.Match([]byte(cookie)) {
					cookieDeleted = true
				}
			}
			require.True(t, cookieDeleted, "IdToken cookie not deleted")
		})

		t.Run("http route without oidc authentication", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "http-without-oidc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			ancestorRef := gwv1a2.ParentReference{
				Group:     gatewayapi.GroupPtr(gwv1.GroupName),
				Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			podInitialized := corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}
			WaitForPods(t, suite.Client, ns, map[string]string{"job-name": "setup-keycloak"}, corev1.PodSucceeded, podInitialized)

			expectedResponse := gwhttp.ExpectedResponse{
				Request: gwhttp.Request{
					Host: "www.example.com",
					Path: "/public",
				},
				Response: gwhttp.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			req := gwhttp.MakeRequest(t, &expectedResponse, gwAddr, "HTTP", "http")
			cReq, cResp, err := suite.RoundTripper.CaptureRoundTrip(req)
			if err != nil {
				t.Errorf("failed to get expected response: %v", err)
			}

			if err := gwhttp.CompareRequest(t, &req, cReq, cResp, expectedResponse); err != nil {
				t.Errorf("failed to compare request and response: %v", err)
			}
		})
	},
}
