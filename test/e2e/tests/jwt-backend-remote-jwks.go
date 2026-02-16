// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, JWTBackendRemoteJWKSTest)
}

var JWTBackendRemoteJWKSTest = suite.ConformanceTest{
	ShortName:   "JWTBackendRemoteJWKS",
	Description: "JWT with Backend as remote JWKS",
	Manifests:   []string{"testdata/jwt-backend-remote-jwks.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("jwt claim base routing", func(t *testing.T) {
			testClaimBasedRouting(t, suite)
		})
	},
}
