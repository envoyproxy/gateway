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
//	  "iss": "https://foo.bar.com",
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
//	  "iss": "https://foo.bar.com",
//		 "sub": "1234567890",
//		 "name": "Alice", // must be John Doe
//		 "admin": true,
//		 "iat": 1516239022,
//		 "roles": "admin, superuser",
//		 "scope": "read add delete modify"
//		}
//
// nolint: gosec
const jwtTokenWithoutRequiredClaimValue = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsIm5hbWUiOiJBbGljZSIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjIsInJvbGVzIjoiYWRtaW4sIHN1cGVydXNlciIsInNjb3BlIjoicmVhZCBhZGQgZGVsZXRlIG1vZGlmeSJ9.kUcA6rE7ScioabJHLITb6NqXQYYHvR1Szx8WQAsT9Dk2D_zWLdupTtiYLdUiaPR8UweZ3GKEo6QmGpa0i8ytfzAMNbqaV32VDupyBkK7TiqSv02uIMbSemMmtoxrQMjPNe-MPsHYxK3M9_eKtkwuaYrg-f0J8-E3ZAJxt5IWaSdjI-PKi4qgttGHnDlVav3QCvnIkr2EOnCLo92c0y-nJz7Vrxhg_QXJmR8LpN5atGUhuypnfqcPgKLVy71LqaqO2Z6QE210ernxTLjUhWCdSA-6rPNGA54jaPZD1I1saR7g0MYvXvGF-34G6DZHGBnBzmoLKEofh4QlfxaKacnG3kubG-zUsJa8AE3kmb2E6YCAiOU6Vv8eQb7GH3m6eMViZQwLujkUZZO7_gUedck4VgW7EAegKdV5cwsjsnnF4T3ogEG12RqJrXNS-Zw993bZBTh8BddhEZe2WqKu7C1LJ8-fHBRsCg0YyrsFsvm8DppOKpy06lUM9TWnEO7QKupT"

func init() {
	ConformanceTests = append(ConformanceTests, AuthorizationJWTTest)
}

var AuthorizationJWTTest = suite.ConformanceTest{
	ShortName:   "Authorization with jwt claims",
	Description: "Authorization with jwt claims",
	Manifests:   []string{"testdata/authorization-jwt.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-jwt-1", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN)

		ancestorRef := gwapiv1a2.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(gatewayapi.KindGateway),
			Namespace: gatewayapi.NamespacePtr(gwNN.Namespace),
			Name:      gwapiv1.ObjectName(gwNN.Name),
		}
		SecurityPolicyMustBeAccepted(t, suite.Client, types.NamespacedName{Name: "authorization-jwt-1", Namespace: ns}, suite.ControllerName, ancestorRef)

		t.Run("allow requests with jwt claims", func(t *testing.T) {
			expectedResponse := http.ExpectedResponse{
				Request: http.Request{
					Path: "/protected1",
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
					Path: "/protected1",
					Headers: map[string]string{
						"Authorization": "Bearer " + jwtTokenWithoutRequiredClaimValue,
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
