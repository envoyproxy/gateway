// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwhttp "sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/roundtripper"
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

type oidcRouteTestCase struct {
	routeName          string
	securityPolicyName string
	clientID           string
	testURL            string
	logoutURL          string
	forwardAccessToken bool
	forwardIDToken     *string
}

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
		ns := "gateway-conformance-infra"
		podInitialized := corev1.PodCondition{Type: corev1.PodInitialized, Status: corev1.ConditionTrue}
		// Wait for the keycloak pod to be configured with the test user and client
		WaitForPods(t, suite.Client, ns, map[string]string{"job-name": "setup-keycloak"}, corev1.PodSucceeded, &podInitialized)
		// Apply the security policy after the keycloak pod is ready, this is because EG will try to fetch the
		// OIDC configuration from the keycloak's well-known endpoint
		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/oidc-securitypolicy.yaml", true)

		urlBackedCases := []oidcRouteTestCase{
			{
				routeName:          "http-with-oidc-foo",
				securityPolicyName: "oidc-test-foo",
				clientID:           "oidctest-foo",
				testURL:            "http://www.example.com/foo",
				logoutURL:          "http://www.example.com/foo/logout",
				forwardAccessToken: true,
				forwardIDToken:     ptr.To("x-eg-id-token"),
			},
			{
				routeName:          "http-with-oidc-bar",
				securityPolicyName: "oidc-test-bar",
				clientID:           "oidctest-bar",
				testURL:            "http://www.example.com/bar",
				logoutURL:          "http://www.example.com/bar/logout",
				forwardIDToken:     ptr.To("Authorization"),
			},
		}

		t.Run("oidc provider represented by a URL", func(t *testing.T) {
			for _, tc := range urlBackedCases {
				t.Run(tc.routeName, func(t *testing.T) {
					testOIDC(t, suite, tc, "testdata/oidc-securitypolicy.yaml")
				})
			}
		})

		t.Run("oidc bypass", func(t *testing.T) {
			routeWithOIDCFooNN := types.NamespacedName{Name: "http-with-oidc-foo", Namespace: ns}
			routeWithOIDCBarNN := types.NamespacedName{Name: "http-with-oidc-bar", Namespace: ns}
			routeWithoutOIDCNN := types.NamespacedName{Name: "http-without-oidc", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeWithOIDCFooNN, routeWithOIDCBarNN, routeWithoutOIDCNN)

			ancestorRef := gwapiv1.ParentReference{
				Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
				Kind:      gatewayapi.KindPtr(resource.KindGateway),
				Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
				Name:      gwapiv1.ObjectName(gwNN.Name),
			}
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-test-foo", Namespace: ns}, suite.ControllerName, ancestorRef)
			SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "oidc-test-bar", Namespace: ns}, suite.ControllerName, ancestorRef)

			testCases := []gwhttp.ExpectedResponse{
				{
					TestCaseName: "http route without oidc authentication",
					Request: gwhttp.Request{
						Host: "www.example.com",
						Path: "/public",
					},
					Response: gwhttp.Response{
						StatusCodes: []int{200},
					},
					Namespace: ns,
				},
				{
					TestCaseName: "oidc with jwt passthrough",
					Request: gwhttp.Request{
						Host: "www.example.com",
						Path: "/foo",
						Headers: map[string]string{
							"Authorization": "Bearer " + v1Token,
						},
					},
					Backend: "infra-backend-v1",
					Response: gwhttp.Response{
						StatusCodes: []int{200},
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

		// Delete the existing policy before applying the BackendCluster variant to avoid flaky test.
		for _, spName := range []string{"oidc-test-foo", "oidc-test-bar"} {
			existingSP := &egv1a1.SecurityPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      spName,
				},
			}
			err := suite.Client.Delete(context.TODO(), existingSP)
			require.Truef(t, err == nil || apierrors.IsNotFound(err), "failed to delete SecurityPolicy %s/%s: %v", ns, existingSP.Name, err)
			SecurityPolicyMustNotExist(t, suite.Client, types.NamespacedName{Name: existingSP.Name, Namespace: ns})
		}

		// Apply the security policy that configures OIDC authentication with BackendCluster.
		suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/oidc-securitypolicy-backendcluster.yaml", true)
		t.Run("oidc provider represented by a BackendCluster", func(t *testing.T) {
			testOIDC(t, suite, oidcRouteTestCase{
				routeName:          "http-with-oidc",
				securityPolicyName: "oidc-test",
				clientID:           "oidctest",
				testURL:            "http://www.example.com/myapp",
				logoutURL:          "http://www.example.com/myapp/logout",
			}, "testdata/oidc-securitypolicy-backendcluster.yaml")
		})
	},
}

func testOIDC(t *testing.T, suite *suite.ConformanceTestSuite, tc oidcRouteTestCase, securityPolicyManifest string) {
	const ns = "gateway-conformance-infra"

	routeNN := types.NamespacedName{Name: tc.routeName, Namespace: ns}
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

	SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: tc.securityPolicyName, Namespace: ns}, suite.ControllerName, ancestorRef)

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
			tlog.Logf(t, "sending request to %s", tc.testURL)

			// Send a request to the http route with OIDC configured.
			// It will be redirected to the keycloak login page
			res, err := oidcClient.Get(tc.testURL, true)
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
						Name:      tc.securityPolicyName,
					},
				}
				require.NoError(t, suite.Client.Delete(context.TODO(), existingSP))
				suite.Applier.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, securityPolicyManifest, false)
				SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: tc.securityPolicyName, Namespace: ns}, suite.ControllerName, ancestorRef)
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
	res, err = oidcClient.Get(tc.testURL, false)
	require.NoError(t, err)
	body, err = io.ReadAll(res.Body)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, res.StatusCode)
	require.Contains(t, string(body), "infra-backend-v1", "Expected response from the application")

	// Verify that configured OIDC tokens are forwarded upstream with the expected JWT semantics.
	if tc.forwardAccessToken || tc.forwardIDToken != nil {
		cookies := oidcClient.Cookies()
		expectedResponse := gwhttp.ExpectedResponse{
			Request: gwhttp.Request{
				Host: "www.example.com",
				Path: strings.TrimPrefix(tc.testURL, "http://www.example.com"),
				Headers: map[string]string{
					"Cookie": buildCookieHeader(cookies),
				},
			},
			Backend: "infra-backend-v1",
			ExpectedRequest: &gwhttp.ExpectedRequest{
				Request: gwhttp.Request{
					Path: strings.TrimPrefix(tc.testURL, "http://www.example.com"),
				},
			},
			Response: gwhttp.Response{
				StatusCodes: []int{200},
			},
			Namespace: ns,
		}
		req := gwhttp.MakeRequest(t, &expectedResponse, httpGWAddr, "HTTP", "http")
		waitForForwardedOIDCTokens(t, suite.RoundTripper, req, expectedResponse, tc)
	}

	// Verify that we can logout
	// Note: OAuth2 filter just clears its cookies and does not log out from the IdP.
	res, err = oidcClient.Get(tc.logoutURL, false)
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

func buildCookieHeader(cookies map[string]*http.Cookie) string {
	pairs := make([]string, 0, len(cookies))
	for _, cookie := range cookies {
		pairs = append(pairs, cookie.Name+"="+cookie.Value)
	}
	sort.Strings(pairs)
	return strings.Join(pairs, "; ")
}

type jwtClaims struct {
	Type          string      `json:"typ"`
	AuthorizedAZP string      `json:"azp"`
	Audience      interface{} `json:"aud"`
}

func waitForForwardedOIDCTokens(t *testing.T, r roundtripper.RoundTripper, req roundtripper.Request, expected gwhttp.ExpectedResponse, tc oidcRouteTestCase) {
	t.Helper()

	err := wait.PollUntilContextTimeout(context.Background(), time.Second, 30*time.Second, true, func(_ context.Context) (bool, error) {
		cReq, cRes, err := r.CaptureRoundTrip(req)
		if err != nil {
			tlog.Logf(t, "request failed while verifying forwarded OIDC tokens: %v", err)
			return false, nil
		}
		if err := gwhttp.CompareRoundTrip(t, &req, cReq, cRes, expected); err != nil {
			tlog.Logf(t, "round trip not ready while verifying forwarded OIDC tokens: %v", err)
			return false, nil
		}

		if tc.forwardAccessToken {
			token, err := getForwardedJWTHeader(cReq.Headers, "Authorization", true)
			if err != nil {
				tlog.Logf(t, "access token forwarding not ready: %v", err)
				return false, nil
			}
			if err := assertJWTClaims(token, "Authorization", "Bearer", "oidctest-foo"); err != nil {
				tlog.Logf(t, "access token forwarding not ready: %v", err)
				return false, nil
			}
			t.Logf("forwarded access token: %s", token)
		}
		if tc.forwardIDToken != nil {
			token, err := getForwardedJWTHeader(cReq.Headers, *tc.forwardIDToken, strings.EqualFold(*tc.forwardIDToken, "Authorization"))
			if err != nil {
				tlog.Logf(t, "id token forwarding not ready: %v", err)
				return false, nil
			}
			if err := assertJWTClaims(token, *tc.forwardIDToken, "ID", tc.clientID); err != nil {
				tlog.Logf(t, "id token forwarding not ready: %v", err)
				return false, nil
			}
			t.Logf("forwarded id token in %s: %s", *tc.forwardIDToken, token)
		}

		return true, nil
	})
	require.NoError(t, err)
}

func assertJWTClaims(token string, name string, expectedType string, expectedAZP string) error {
	claims, err := decodeJWTClaims(token)
	if err != nil {
		return fmt.Errorf("failed to decode %s JWT: %w", name, err)
	}
	if claims.Type != expectedType {
		return fmt.Errorf("expected %s token typ %q, got %q", name, expectedType, claims.Type)
	}
	if claims.AuthorizedAZP != expectedAZP {
		return fmt.Errorf("expected %s token azp %q, got %q", name, expectedAZP, claims.AuthorizedAZP)
	}

	return nil
}

func getForwardedJWTHeader(headers map[string][]string, name string, bearer bool) (string, error) {
	if headers == nil {
		return "", fmt.Errorf("no headers captured")
	}

	var value string
	for k, v := range headers {
		if strings.EqualFold(k, name) {
			value = strings.Join(v, ",")
			break
		}
	}
	if value == "" {
		return "", fmt.Errorf("expected %s header to be set", name)
	}
	if !bearer {
		return value, nil
	}
	if !strings.HasPrefix(value, "Bearer ") {
		return "", fmt.Errorf("expected %s header to use Bearer prefix, got %q", name, value)
	}
	return strings.TrimPrefix(value, "Bearer "), nil
}

func decodeJWTClaims(token string) (*jwtClaims, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	claims := new(jwtClaims)
	if err := json.Unmarshal(payload, claims); err != nil {
		return nil, err
	}
	return claims, nil
}
