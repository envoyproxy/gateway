// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e"
	egtests "github.com/envoyproxy/gateway/test/e2e/tests"
)

func conformanceOpts(t *testing.T) suite.ConformanceOptions {
	gatewayNamespaceMode := egtests.IsGatewayNamespaceMode()
	internalSuite := EnvoyGatewaySuite(gatewayNamespaceMode, egtests.UseStandardChannel())

	opts := conformance.DefaultOptions(t)
	opts.SkipTests = internalSuite.SkipTests
	opts.SupportedFeatures = internalSuite.SupportedFeatures
	opts.ExemptFeatures = internalSuite.ExemptFeatures
	if egtests.IPFamily == "dual" {
		// I don't know why this happens, but the UDPRoute test failed on dual stack
		// because on some VM(e.g. Ubuntu 22.04), the ipv4 address for UDP gateway is not
		// reachable. There's a same test in our e2e test fixtures that passed, it's so odd.
		// So we skip this test on dual stack for now.
		opts.SkipTests = append(opts.SkipTests,
			tests.UDPRouteTest.ShortName,
		)
	}

	opts.Hook = e2e.Hook
	opts.FailFast = true

	return opts
}

// SkipTests is a list of tests that are skipped in the conformance suite.
func SkipTests(gatewayNamespaceMode bool) []suite.ConformanceTest {
	skipTests := make([]suite.ConformanceTest, 0, 4)
	skipTests = append(skipTests,
		// TODO: fix following conformance tests
		tests.ListenerSetHostnameConflict,
		tests.ListenerSetProtocolConflict,
	)

	if gatewayNamespaceMode {
		return skipTests
	}

	skipTests = append(skipTests, tests.GatewayInfrastructure)

	return skipTests
}

// SkipFeatures is a list of features that are skipped in the conformance report.
func SkipFeatures(gatewayNamespaceMode bool) sets.Set[features.FeatureName] {
	if gatewayNamespaceMode {
		return sets.New[features.FeatureName](
			features.GatewayHTTPSListenerDetectMisdirectedRequestsFeature.Name,
		)
	}

	return sets.New(
		features.GatewayHTTPSListenerDetectMisdirectedRequestsFeature.Name,
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
func EnvoyGatewaySuite(gatewayNamespaceMode, standardChannel bool) suite.ConformanceOptions {
	return suite.ConformanceOptions{
		SupportedFeatures: allFeatures(gatewayNamespaceMode, standardChannel),
		ExemptFeatures:    meshFeatures(),
		SkipTests:         skipTestsShortNames(SkipTests(gatewayNamespaceMode)),
	}
}

func allFeatures(gatewayNamespaceMode, standardChannel bool) sets.Set[features.FeatureName] {
	result := sets.New[features.FeatureName]()
	skipped := SkipFeatures(gatewayNamespaceMode)
	for _, feature := range features.AllFeatures.UnsortedList() {
		if standardChannel && feature.Channel != features.FeatureChannelStandard {
			continue
		}

		// Don't add skipped features in the conformance report.
		if !skipped.Has(feature.Name) {
			result.Insert(feature.Name)
		}
	}
	for _, feature := range features.UDPRouteFeatures {
		if standardChannel && feature.Channel != features.FeatureChannelStandard {
			continue
		}

		result.Insert(feature.Name)
	}

	// this's used to skip TLSRouteListenerMixedTerminationNotSupported
	result.Insert(features.TLSRouteModeMixedFeature.Name)
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
