// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Add go build to avoid this to be run as part of make run-conformance/run-experimental-conformance
//go:build conformance_unit_test

package conformance

import (
	"slices"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func TestEnvoyGatewaySuite(t *testing.T) {
	cases := []struct {
		name                 string
		gatewayNamespaceMode bool
		standardChannel      bool
		includedFeatures     []features.FeatureName
		excludedFeatures     []features.FeatureName
	}{
		{
			name:                 "TLSRouteModeMixed should be included when gatewayNamespaceMode is true and standardChannel is true",
			gatewayNamespaceMode: true,
			standardChannel:      true,
			includedFeatures: []features.FeatureName{
				// need this feature to skip TLSRouteListenerMixedTerminationNotSupported test.
				features.SupportTLSRouteModeMixed,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(*testing.T) {
			got := EnvoyGatewaySuite(tc.gatewayNamespaceMode, tc.standardChannel)
			for _, in := range tc.includedFeatures {
				if !slices.Contains(got.SupportedFeatures, in) {
					t.Fatalf("%s should be included", in)
				}
			}

			for _, in := range tc.excludedFeatures {
				if slices.Contains(got.SupportedFeatures, in) {
					t.Fatalf("%s should be excluded", in)
				}
			}
		})
	}
}

// TestAdvertisedFeaturesAreTested makes sure that every feature reported in the GatewayClass status
// is run by the experimental conformance suite, in every mode, and has at least one test.
func TestAdvertisedFeaturesAreTested(t *testing.T) {
	covered := sets.New[features.FeatureName]()
	for _, test := range tests.ConformanceTests {
		covered.Insert(test.Features...)
	}
	for _, feature := range status.GatewaySupportedFeatures {
		if !covered.Has(features.FeatureName(feature.Name)) {
			t.Errorf("%s is reported as supported, but no conformance test requires it", feature.Name)
		}
	}

	for _, gatewayNamespaceMode := range []bool{false, true} {
		got := EnvoyGatewaySuite(gatewayNamespaceMode, false).SupportedFeatures
		for _, feature := range status.GatewaySupportedFeatures {
			if !slices.Contains(got, features.FeatureName(feature.Name)) {
				t.Errorf("%s is reported as supported, but not tested (gatewayNamespaceMode=%t)",
					feature.Name, gatewayNamespaceMode)
			}
		}
	}
}
