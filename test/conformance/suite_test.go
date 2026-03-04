// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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
		includeFeatures      []features.FeatureName
		excludeFeatures      []features.FeatureName
	}{
		{
			name:                 "TLSRouteModeMixed should be excluded when gatewayNamespaceMode is true and standardChannel is true",
			gatewayNamespaceMode: true,
			standardChannel:      true,
			excludeFeatures: []features.FeatureName{
				features.SupportTLSRouteModeMixed,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(*testing.T) {
			got := EnvoyGatewaySuite(tc.gatewayNamespaceMode, tc.standardChannel)
			for _, in := range tc.includeFeatures {
				if !got.SupportedFeatures.Has(in) {
					t.Fatalf("%s should be included", in)
				}
			}

			for _, in := range tc.excludeFeatures {
				if got.SupportedFeatures.Has(in) {
					t.Fatalf("%s should be excluded", in)
				}
			}
		})
	}
}
