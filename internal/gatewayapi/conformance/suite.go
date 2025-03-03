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
	tests.GatewayInfrastructure,
}

// SkipFeatures is a list of features that are skipped in the conformance report.
var SkipFeatures = sets.New[features.FeatureName](
	features.GatewayStaticAddressesFeature.Name,
	features.GatewayInfrastructurePropagationFeature.Name,
)

func skipTestsShortNames(skipTests []suite.ConformanceTest) []string {
	shortNames := make([]string, len(skipTests))
	for i, test := range skipTests {
		shortNames[i] = test.ShortName
	}
	return shortNames
}

// EnvoyGatewaySuite is the conformance suite configuration for the Gateway API.
var EnvoyGatewaySuite = suite.ConformanceOptions{
	SupportedFeatures: allFeatures(),
	ExemptFeatures:    meshFeatures(),
	SkipTests:         skipTestsShortNames(SkipTests),
}

func allFeatures() sets.Set[features.FeatureName] {
	allFeatures := sets.New[features.FeatureName]()
	for _, feature := range features.AllFeatures.UnsortedList() {
		// Dont add skipped features in the conformance report.
		if !SkipFeatures.Has(feature.Name) {
			allFeatures.Insert(feature.Name)
		}
	}
	return allFeatures
}

func meshFeatures() sets.Set[features.FeatureName] {
	meshFeatures := sets.New[features.FeatureName]()
	for _, feature := range features.MeshCoreFeatures.UnsortedList() {
		meshFeatures.Insert(feature.Name)
	}
	for _, feature := range features.MeshExtendedFeatures.UnsortedList() {
		meshFeatures.Insert(feature.Name)
	}
	return meshFeatures
}
