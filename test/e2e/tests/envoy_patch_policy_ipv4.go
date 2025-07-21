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
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyTest)
}

var EnvoyPatchPolicyIPv4Test = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicyIPv4",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy-ipv4.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("envoy patch policy", func(t *testing.T) {
			testEnvoyPatchPolicy(t, suite)
		})
	},
}
