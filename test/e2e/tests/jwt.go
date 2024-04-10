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
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, JWTTest)
}

const (
	// from examples/kubernetes/jwt/test.jwt
	// nolint: gosec
	v1Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.NHVaYe26MbtOYhSKkoKYdFVomg4i8ZJd8_-RU8VNbftc4TSMb4bXP3l3YlNWACwyXPGffz5aXHc6lty1Y2t4SWRqGteragsVdZufDn5BlnJl9pdR_kdVFUsra2rWKEofkZeIC4yWytE58sMIihvo9H1ScmmVwBcQP6XETqYd0aSHp1gOa9RdUPDvoXQ5oqygTqVtxaDr6wUFKrKItgBMzWIdNZ6y7O9E0DhEPTbE9rfBo6KTFsHAZnMg4k68CDp2woYIaXbmYTWcvbzIuHO7_37GT79XdIwkm95QJ7hYC9RiwrV7mesbY4PAahERJawntho0my942XheVLmGwLMBkQ"
	// from examples/kubernetes/jwt/with-different-claim.jwt
	// nolint: gosec
	v2Token = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IlRvbSIsImFkbWluIjp0cnVlLCJpYXQiOjE1MTYyMzkwMjJ9.kyzDDSo7XpweSPU1lxoI9IHzhTBrRNlnmcW9lmCbloZELShg-8isBx4AFoM4unXZTHpS_Y24y0gmd4nDQxgUE-CgjVSnGCb0Xhy3WO1gm9iChoKDyyQ3kHp98EmKxTyxKG2X9GyKcDFNBDjH12OBD7TcJUaBEvLf6Jw1SG2A7FakUPWeK04DQ916-ROylzI6qKyaZ0OpfYIbijvyAQxlQRxxs2XHlAkLdJhfVcUqJBwsFTbwHYARC-WNgd2_etAk1GWdwwZ_NoTmRzZAMryrYJpHY9KPlbnZ93Ye3o9h2viBQ_XRb7JBkWnAGYO4_KswpJWE_7ROUVj8iOJo2jfY6w"
	// nolint: gosec
	anotherToken = "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkplcnJ5IiwiYWRtaW4iOnRydWUsImlhdCI6MTUxNjIzOTAyMn0.VKLURpaPLWanwE5xoGTfuYKqT9a91Fg1tRBAOyFzNa5t9SbtK8As7-3iJg4f_VlBHj13OeKjfpDEvgLerIt5TKnU708YKERB45di_7TNURoiVZayq3_gFznMqoSarP3irLDzh0YKUjc7Vuh3MX99fueTdbeA-c4pMhG_nwiFeRJhZNQQDzzKtmL9C_L2uwP4bDupmcYz6FAA2EN_r67WoXCjPWQoRQmE435EVQ-FYKgAR7qZ5TdjoSN91ByRQ7Ior9srPl7gOvjuaRbu7fjC-LT7wRE26v2vu-BCM2PveJf2NMobNb8q0pcmpB1TWhSXp1MIZs9yxbqEAZLOumYfUw"
)

var JWTTest = suite.ConformanceTest{
	ShortName:   "JWT",
	Description: "JWT Claim",
	Manifests:   []string{"testdata/jwt.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("jwt claim base routing", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "jwt-claim-routing", Namespace: ns}
			gwNN := types.NamespacedName{Name: "same-namespace", Namespace: ns}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), routeNN)

			testCases := []http.ExpectedResponse{
				{
					Request: http.Request{
						Path: "/get",
						Headers: map[string]string{
							"Authorization": "Bearer " + v1Token,
						},
					},
					Backend: "infra-backend-v1",
					Response: http.Response{
						StatusCode: 200,
					},
					Namespace: ns,
				},
				{
					Request: http.Request{
						Path: "/get",
						Headers: map[string]string{
							"Authorization": "Bearer " + v2Token,
						},
					},
					Backend: "infra-backend-v2",
					Response: http.Response{
						StatusCode: 200,
					},
					Namespace: ns,
				},
				{
					Request: http.Request{
						Path: "/get",
						Headers: map[string]string{
							"Authorization": "Bearer " + anotherToken,
						},
					},
					Backend: "infra-backend-v1",
					Response: http.Response{
						StatusCode: 500,
					},
					Namespace: ns,
				},
				{
					Request: http.Request{
						Path: "/get",
						Headers: map[string]string{
							"x-name": "Tom",
						},
					},
					Backend: "infra-backend-v2",
					Response: http.Response{
						StatusCode: 401,
					},
					Namespace: ns,
				},
			}

			for i := range testCases {
				tc := testCases[i]
				t.Run(tc.GetTestCaseName(i), func(t *testing.T) {
					t.Parallel()
					http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, tc)
				})
			}
		})
	},
}
