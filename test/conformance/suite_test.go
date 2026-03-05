// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Add go build to avoid this to be run as part of make run-conformance/run-experimental-conformance
//go:build conformance_unit_test

package conformance

import (
	"testing"

	"sigs.k8s.io/gateway-api/pkg/features"
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
				if !got.SupportedFeatures.Has(in) {
					t.Fatalf("%s should be included", in)
				}
			}

			for _, in := range tc.excludedFeatures {
				if got.SupportedFeatures.Has(in) {
					t.Fatalf("%s should be excluded", in)
				}
			}
		})
	}
}
