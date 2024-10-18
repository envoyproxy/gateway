// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, OIDCBackendClusterTest)
}

// OIDCTest tests OIDC authentication for an http route with OIDC configured.
// The http route points to an application to verify that OIDC authentication works on application/http path level.
var OIDCBackendClusterTest = suite.ConformanceTest{
	ShortName:   "OIDC with BackendCluster",
	Description: "Test OIDC authentication",
	Manifests:   []string{"testdata/oidc-keycloak.yaml", "testdata/oidc-securitypolicy-backendcluster.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("oidc provider represented by a BackendCluster", func(t *testing.T) {
			// Add a function to dump current cluster status
			t.Cleanup(func() {
				CollectAndDump(t, suite.RestConfig)
			})

			testOIDC(t, suite)
		})
	},
}
