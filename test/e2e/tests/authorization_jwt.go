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

//			{
//		     "iss": "https://foo.bar.com",
//			 "sub": "1234567890",
//	      "user": {
//	         "name": "John Doe",
//	         "email": "john.doe@example.com",
//	         "roles": ["admin", "editor"]
//	      },
//	      "premium_user": true,
//			 "iat": 1516239022,
//			 "scope": "read add delete modify"
//			}
//
// nolint: gosec
const (
	jwtToken = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsInVzZXIiOnsibmFtZSI6IkpvaG4gRG9lIiwiZW1haWwiOiJqb2huLmRvZUBleGFtcGxlLmNvbSIsInJvbGVzIjpbImFkbWluIiwiZWRpdG9yIl19LCJwcmVtaXVtX3VzZXIiOnRydWUsImlhdCI6MTUxNjIzOTAyMiwic2NvcGUiOiJyZWFkIGFkZCBkZWxldGUgbW9kaWZ5In0.P36iAlmiRCC79OiB3vstF5Q_9OqUYAMGF3a3H492GlojbV6DcuOz8YIEYGsRSWc-BNJaBKlyvUKsKsGVPtYbbF8ajwZTs64wyO-zhd2R8riPkg_HsW7iwGswV12f5iVRpfQ4AG2owmdOToIaoch0aym89He1ZzEjcShr9olgqlAbbmhnk-namd1rP-xpzPnWhhIVI3mCz5hYYgDTMcM7qbokM5FzFttTRXAn5_Luor23U1062Ct_K53QArwxBvwJ-QYiqcBycHf-hh6sMx_941cUswrZucCpa-EwA3piATf9PKAyeeWHfHV9X-y8ipGOFg3mYMMVBuUZ1lBkJCik9f9kboRY6QzpOISARQj9PKMXfxZdIPNuGmA7msSNAXQgqkvbx04jMwb9U7eCEdGZztH4C8LhlRjgj0ZdD7eNbRjeH2F6zrWyMUpGWaWyq6rMuP98W2DWM5ZflK6qvT1c7FuFsWPvWLkgxQwTWQKrHdKwdbsu32Sj8VtUBJ0-ddEb"
	//		{
	//	     "iss": "https://foo.bar.com",
	//		 "sub": "1234567890",
	//       "user": {
	//          "name": "Alice Smith",
	//          "email": "alice.smith@example.com",
	//          "roles": ["developer"]
	//       },
	//       "premium_user": false,
	//		 "iat": 1516239022,
	//		 "scope": "read add delete"
	//		}
	//
	// nolint: gosec
	jwtTokenWithoutRequiredValues = "eyJ0eXAiOiJKV1QiLCJhbGciOiJSUzI1NiIsImtpZCI6ImI1MjBiM2MyYzRiZDc1YTEwZTljZWJjOTU3NjkzM2RjIn0.eyJpc3MiOiJodHRwczovL2Zvby5iYXIuY29tIiwic3ViIjoiMTIzNDU2Nzg5MCIsInVzZXIiOnsibmFtZSI6IkFsaWNlIFNtaXRoIiwiZW1haWwiOiJhbGljZS5zbWl0aEBleGFtcGxlLmNvbSIsInJvbGVzIjpbImRldmVsb3BlciJdfSwicHJlbWl1bV91c2VyIjpmYWxzZSwiaWF0IjoxNTE2MjM5MDIyLCJzY29wZSI6InJlYWQgYWRkIGRlbGV0ZSJ9.Da547nNXzuQXm5E7LuLAiyFswXsW4RDhuitD_rpadtR7PTwzzOsJoqrVWJ_u1jJDaOTWIpLF4gwxDoY-Aoz_couzXzlAbECLs45ZFoc_UdffpfIbGKqTZx8VtwKuDLFsAeDDDqqx1flxFhvXHftJJdZYr1FgFz9u-absMmRU90DLmEZX3Hnyc8k8eBgeiu6vsWUD0-aNy8cWkFRbwRggkGmucFyUTG8Z1MY3iyH5E66W-ISoX8G9bzE9PTxVAAPDTvefD5iLJPSDJ8qV69OuMCJ8Dczq0L9Dd_w0sF-D1s9MTvexmGg4zBWluJ3r-pU9NHEdhqBypehp_yH8xF5Rt9AE7stZ4oPFZNyfrtkE-4IOnSEkMmzcC65g_rscn0ycerv4N5ZNpkr0x2IYYM4iGuo-ULv5Htnli3rffST45kx1XA8cdsrT1D0K3aPxdIxDIk8sTJf5-WVqRyo-bwxXXltwQLB9jCM_7QbTWQBYAJwUpi-0RW4jCl44-42gZnXf"
)

func init() {
	ConformanceTests = append(ConformanceTests, AuthorizationJWTTest)
}

var AuthorizationJWTTest = suite.ConformanceTest{
	ShortName:   "AuthzWithJWTClaimsScopes",
	Description: "Authorization with jwt claims and scopes",
	Manifests:   []string{"testdata/authorization-jwt.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ns := "gateway-conformance-infra"
		route1NN := types.NamespacedName{Name: "http-with-authorization-jwt-claim", Namespace: ns}
		route2NN := types.NamespacedName{Name: "http-with-authorization-jwt-scope", Namespace: ns}
		route3NN := types.NamespacedName{Name: "http-with-authorization-jwt-combined", Namespace: ns}
		gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
		gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), route1NN, route2NN, route3NN)

		ancestorRef := gwapiv1.ParentReference{
			Group:     gatewayapi.GroupPtr(gwapiv1.GroupName),
			Kind:      gatewayapi.KindPtr(resource.KindGateway),
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
					StatusCodes: []int{200},
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
					StatusCodes: []int{403},
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
					StatusCodes: []int{200},
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
					StatusCodes: []int{403},
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
					StatusCodes: []int{200},
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
					StatusCodes: []int{403},
				},
				Namespace: ns,
			}

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
