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
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

//		{
//	     "iss": "https://foo.bar.com",
//		 "sub": "1234567890",
//		 "name": "John Doe",
//		 "admin": true,
//		 "iat": 1516239022,
//		 "roles": "admin, superuser",
//		 "scope": "read add delete modify"
//		}
//
// nolint: gosec
const jwtToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsIm5hbWUiOiJKb2huIERvZSIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjIsInJvbGVzIjoiYWRtaW4sIHN1cGVydXNlciIsInNjb3BlIjoicmVhZCBhZGQgZGVsZXRlIG1vZGlmeSJ9.KLL_-9NGDZSDr12SQiw4R-MaVp9jGJzT5xWHjBOSqQMr6SAm3QK6wSUJfKWxdnLR6QAYHl5rDRs_89qa96J-QkA5NQHjoXXNO36OEa7G2x-KXzeHRl8vBpsKk55ls48ua2V9CHlR0bSREE_Eq_RTKXcjox71fl2vzC6sGgbFQTi6QFFIlR1O9dK-87PE-D_aoujNcYtuoYQGrouzQ9WDQ5xoKVU4Si7bBzv1kzUOziA0J7SFrEv07Yj_p5nZZwZ3JmSQUrYfjQvXEW9FKI0hhajuWkILeAXUp2Kt5raYJliGhD_qMeFKp2aUGhDDpHj-vJuzDKo8CyF5iv-Jv-NKY_3sDp1fPOH9WoUe9ieujRusrdltfxZPOGFEST4dQreVVdOX8zB3Q0L7OScYZ5m-MdsODH0RGQrGg78iJT6Tj-Aluh9KRVlXvPbOdp7YSkaTMjf2dwY0QhillisS-IdjMjL7A3-gzdBbvU2cJh2NRAAHk9YQylgBdCnn-hmHXy_t"

//		{
//	     "iss": "https://foo.bar.com",
//		 "sub": "1234567890",
//		 "name": "Alice",
//		 "admin": true,
//		 "iat": 1516239022,
//		 "roles": "developer",
//		 "scope": "read add delete"
//		}
//
// nolint: gosec
const jwtTokenWithoutRequiredValues = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsIm5hbWUiOiJBbGljZSIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjIsInJvbGVzIjoiZGV2ZWxvcGVyIiwic2NvcGUiOiJyZWFkIGFkZCBkZWxldGUifQ.MjEN1i3dYAiuOyHqdLI42Zulqtu37J1_OXhp0GzXwyes1nRtdyc4YHMWcPzmX2cbWEj87YuTUDz2aYITzG2PB94qQ4ec4VyAcdERH4LhSKf-lPonhFbIpYqefbjmZmnrPh0wG372AEH6jL5RBxhkuoh2TROFYkcM6f1VlAxRVF8gBRpmnDNjw_rtnNju108dcJUnAL5qjzCYwqkzT_uubg0XPApMTNEsjRNJNniIqa3Ae-3BWp1qWBCEV2rpuG5oaBmYuUKtmcaVmBzTUHfpJicxw3k3UcobuCQczTZjbzWpGVXYJ252iLT_6mbiAz2ON2unZCeOWwAEGS-zCH5O62Z6S3QjnJzr1Si0tYdEBoCxHULbXCDOTdMPVKWgY_lSiR1AjwW9NNbeeBaVR0YUnQg8N35WoJvMjVNECnu83_nIE4c7mdEnFSUWeaRr36GYhMX07Vtc5gQdTHIf82NmXSL8wC0jUOq_qI-DuHV-rczvhn2YsUDN3WyyPsuQ3E0L"
func init() {
	ConformanceTests = append(ConformanceTests, AuthorizationJWTTest)
}

var AuthorizationJWTTest = suite.ConformanceTest{
	ShortName:   "Authorization with jwt claims and scopes",
	Description: "Authorization with jwt claims and scopes",
	Manifests:   []string{"testdata/authorization-jwt.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-jwt-claim", Namespace: ns}
		route2NN := types.NamespacedName{Name: "http-with-authorization-jwt-scope", Namespace: ns}
		route3NN := types.NamespacedName{Name: "http-with-authorization-jwt-combined", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN,route2NN,route3NN)

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-jwt-claim", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-jwt-scope", Namespace: ns}, suite.ControllerName, ancestorRef)
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-jwt-combined", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("allow requests with jwt claims", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/claim-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtToken,
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("deny requests with jwt claims that do not match the required claim value", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/claim-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtTokenWithoutRequiredValues,
					},
				},
				Response: http.Response{
					StatusCode: 403,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("allow requests with jwt scopes", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/scope-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtToken,
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("deny requests with jwt scopes that do not match the required scope value", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/scope-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtTokenWithoutRequiredValues,
					},
				},
				Response: http.Response{
					StatusCode: 403,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("allow requests with jwt claims and scopes", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/combined-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtToken,
					},
				},
				Response: http.Response{
					StatusCode: 200,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})

		t.Run("deny requests with jwt scopes and claims that do not match the required scope and claim values", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/combined-test",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtTokenWithoutRequiredValues,
					},
				},
				Response: http.Response{
					StatusCode: 403,
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
