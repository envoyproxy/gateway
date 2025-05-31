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
func SkipTests(gatewayNamespaceMode bool) []suite.ConformanceTest {
	if gatewayNamespaceMode {
		return []suite.ConformanceTest{
			tests.GatewayStaticAddresses,
		}
	}

	return []suite.ConformanceTest{
		tests.GatewayStaticAddresses,
		tests.GatewayInfrastructure,
	}
}

// SkipFeatures is a list of features that are skipped in the conformance report.
func SkipFeatures(gatewayNamespaceMode bool) sets.Set[features.FeatureName] {
	if gatewayNamespaceMode {
		return sets.New(features.GatewayStaticAddressesFeature.Name)
	}

	return sets.New(
		features.GatewayStaticAddressesFeature.Name,
		features.GatewayInfrastructurePropagationFeature.Name,
	)
}

func skipTestsShortNames(skipTests []suite.ConformanceTest) []string {
	shortNames := make([]string, len(skipTests))
	for i, test := range skipTests {
		shortNames[i] = test.ShortName
	}
	return shortNames
}

// EnvoyGatewaySuite is the conformance suite configuration for the Gateway API.
func EnvoyGatewaySuite(gatewayNamespaceMode bool) suite.ConformanceOptions {
	return suite.ConformanceOptions{
		SupportedFeatures: allFeatures(gatewayNamespaceMode),
		ExemptFeatures:    meshFeatures(),
		SkipTests:         skipTestsShortNames(SkipTests(gatewayNamespaceMode)),
	}
}

func allFeatures(gatewayNamespaceMode bool) sets.Set[features.FeatureName] {
	result := sets.New[features.FeatureName]()
	skipped := SkipFeatures(gatewayNamespaceMode)
	for _, feature := range features.AllFeatures.UnsortedList() {
		// Don't add skipped features in the conformance report.
		if !skipped.Has(feature.Name) {
			result.Insert(feature.Name)
		}
	}
	return result
}

func meshFeatures() sets.Set[features.FeatureName] {
	result := sets.New[features.FeatureName]()
	for _, feature := range features.MeshCoreFeatures.UnsortedList() {
		result.Insert(feature.Name)
	}
	for _, feature := range features.MeshExtendedFeatures.UnsortedList() {
		result.Insert(feature.Name)
	}
	return result
}
