// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"
)

// SkipTests is a list of tests that are skipped in the conformance suite.
var SkipTests = []suite.ConformanceTest{
	tests.GatewayStaticAddresses,
	tests.GatewayHTTPListenerIsolation, // https://github.com/envoyproxy/gateway/issues/3352
}

func skipTestsShortNames(skipTests []suite.ConformanceTest) []string {
	shortNames := make([]string, len(skipTests))
	for i, test := range skipTests {
		shortNames[i] = test.ShortName
	}
	return shortNames
}

// EnvoyGatewaySuite is the conformance suite configuration for the Gateway API.
var EnvoyGatewaySuite = suite.ConformanceOptions{
	SupportedFeatures: features.AllFeatures,
	ExemptFeatures: sets.New[features.SupportedFeature]().
		Insert(features.MeshCoreFeatures.UnsortedList()...).
		Insert(features.MeshExtendedFeatures.UnsortedList()...),
	SkipTests: skipTestsShortNames(SkipTests),
}
