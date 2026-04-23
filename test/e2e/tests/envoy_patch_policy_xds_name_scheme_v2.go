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
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyXDSNameSchemeV2Test)
}

var EnvoyPatchPolicyXDSNameSchemeV2Test = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicyXDSNameSchemeV2",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy-xds-name-scheme-v2.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("envoy patch policy", func(t *testing.T) {
			testEnvoyPatchPolicy(t, suite)
		})
	},
}
