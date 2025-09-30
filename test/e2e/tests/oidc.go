// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"io"
	"net"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

const (
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
	Manifests:   []string{"testdata/oidc-keycloak.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("oidc provider represented by a URL", func(t *testing.T) {
			testOIDC(t, suite, "testdata/oidc-securitypolicy.yaml")
		})

		t.Run("oidc bypass", func(t *testing.T) {
			ns := "gateway-conformance-infra"

			podInitialized := corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}
			// Wait for the keycloak pod to be configured with the test user and client
			WaitForPods(t, suite.Client, ns, map[string]string{"job-name": "setup-keycloak"}, corev1.PodSucceeded, &podInitialized)

			// Apply the security policy that configures OIDC authentication
			suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/oidc-securitypolicy.yaml", true)

			routeWithOIDCNN := types.NamespacedName{Name: "http-with-oidc", Namespace: ns}
			routeWithoutOIDCNN := types.NamespacedName{Name: "http-without-oidc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeWithOIDCNN, routeWithoutOIDCNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-test", Namespace: ns}, suite.ControllerName, ancestorRef)

			testCases := []gwhttp.ExpectedResponse{
				{
					TestCaseName: "http route without oidc authentication",
					Request: gwhttp.Request{
						Host: "www.example.com",
						Path: "/public",
					},
					Response: gwhttp.Response{
						StatusCode: 200,
					},
					Namespace: ns,
				},
				{
					TestCaseName: "oidc with jwt passthrough",
					Request: gwhttp.Request{
						Host: "www.example.com",
						Path: "/myapp",
						Headers: map[string]string{
							"Authorization": "Bearer " + v1Token,
						},
					},
					Backend: "infra-backend-v1",
					Response: gwhttp.Response{
						StatusCode: 200,
					},
					Namespace: ns,
				},
			}

			for i := range testCases {
				tc := testCases[i]
				t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
					t.Parallel()

					gwhttp.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
				})
			}
		})
	},
}

func testOIDC(t *testing.T, suite *suite.ConformanceTestSuite, securityPolicyManifest string) {
	var (
		testURL   = "http://www.example.com/myapp"
		logoutURL = "http://www.example.com/myapp/logout"
		route     = "http-with-oidc"
		sp        = "oidc-test"
		ns        = "gateway-conformance-infra"
	)

	podInitialized := corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}
	// Wait for the keycloak pod to be configured with the test user and client
	WaitForPods(t, suite.Client, ns, map[string]string{"job-name": "setup-keycloak"}, corev1.PodSucceeded, &podInitialized)

	// Apply the security policy that configures OIDC authentication
	suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, securityPolicyManifest, true)

	routeNN := types.NamespacedName{Name: route, Namespace: ns}
	gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
	httpGWAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN, "http"), routeNN)
	host, _, _ := net.SplitHostPort(httpGWAddr)
	tlsGWAddr := net.JoinHostPort(host, "443")
	ancestorRef := gwapiv1.ParentReference{
		Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
		Kind:      gatewayapi.KindPtr(resource.KindGateway),
		Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
		Name:      gwapiv1.ObjectName(gwNN.Name),
	}

	SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: sp, Namespace: ns}, suite.ControllerName, ancestorRef)

	// Initialize the test OIDC client that will keep track of the state of the OIDC login process
	oidcClient, err := NewOIDCTestClient(
		WithLoggingOptions(t.Log, true),
		// Map the application and keycloak cluster DNS name to the gateway address
		WithCustomAddressMappings(map[string]string{
			"www.example.com:80":                     httpGWAddr,
			"keycloak.gateway-conformance-infra:80":  httpGWAddr,
			"keycloak.gateway-conformance-infra:443": tlsGWAddr,
		}),
	)
	require.NoError(t, err)

	if err := wait.PollUntilContextTimeout(context.TODO(), time.Second, 5*time.Minute, true,
		func(_ context.Context) (done bool, err error) {
			tlog.Logf(t, "sending request to %s", testURL)

			// Send a request to the http route with OIDC configured.
			// It will be redirected to the keycloak login page
			res, err := oidcClient.Get(testURL, true)
			if err != nil {
				tlog.Logf(t, "failed to get the login page: %v", err)
				return false, nil
			}
			if res.StatusCode != http.StatusOK {
				tlog.Logf(t, "Failed to get the login page, expected 200 OK, got %d", res.StatusCode)
				return false, nil
			}

			// Parse the response body to get the URL where the login page would post the user-entered credentials
			if err := oidcClient.ParseLoginForm(res.Body, keyCloakLoginFormID); err != nil {
				tlog.Logf(t, "failed to parse login form: %v", err)
				// recreate the security policy to force repushing the configuration to the envoy proxy to recover from the error.
				// This is a workaround for the flaky test: https://github.com/envoyproxy/gateway/issues/3898
				// TODO: we should investigate the root cause of the flakiness and remove this workaround
				existingSP := &egv1a1.SecurityPolicy{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: ns,
						Name:      sp,
					},
				}
				require.NoError(t, suite.Client.Delete(context.TODO(), existingSP))
				suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, securityPolicyManifest, false)
				SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: sp, Namespace: ns}, suite.ControllerName, ancestorRef)
				return false, nil
			}

			t.Log("successfully parsed login form")
			return true, nil
		}); err != nil {
		t.Errorf("failed to parse login form: %v", err)
	}

	// Submit the login form to the IdP.
	// This will authenticate and redirect back to the application
	res, err := oidcClient.Login(map[string]string{"username": username, "password": password, "credentialId": ""})
	require.NoError(t, err, "Failed to login to the IdP")

	// Verify that we get the expected response from the application
	body, err := io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "infra-backend-v1", "Expected response from the application")

	// Verify that we can access the application without logging in again
	res, err = oidcClient.Get(testURL, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "infra-backend-v1", "Expected response from the application")

	// Verify that we can logout
	// Note: OAuth2 filter just clears its cookies and does not log out from the IdP.
	res, err = oidcClient.Get(logoutURL, false)
	require.NoError(t, err)
	require.Equal(t, http.StatusFound, res.StatusCode)

	// After logout, OAuth2 filter will redirect to the IdP end session endpoint.
	require.Contains(t, res.Header.Get("Location"), "https://keycloak.gateway-conformance-infra/realms/master/protocol/openid-connect/logout", "Expected redirect to the root of the host")

	// Verify that the oauth2 cookies have been deleted
	var cookieDeleted bool
	deletedCookies := res.Header.Values("Set-Cookie")
	regx := regexp.MustCompile("^IdToken-.+=deleted.+")
	for _, cookie := range deletedCookies {
		if regx.MatchString(cookie) {
			cookieDeleted = true
		}
	}
	require.True(t, cookieDeleted, "IdToken cookie not deleted")
}
