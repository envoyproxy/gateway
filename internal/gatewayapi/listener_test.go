// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestProxySamplingRate(t *testing.T) {
	cases := []struct {
		name     string
		tracing  *egv1a1.ProxyTracing
		expected float64
	}{
		{
			name:     "default",
			tracing:  &egv1a1.ProxyTracing{},
			expected: 100.0,
		},
		{
			name: "rate",
			tracing: &egv1a1.ProxyTracing{
				SamplingRate: ptr.To[uint32](10),
			},
			expected: 10.0,
		},
		{
			name: "fraction numerator only",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator: 100,
				},
			},
			expected: 1.0,
		},
		{
			name: "fraction",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   1,
					Denominator: ptr.To[int32](10),
				},
			},
			expected: 0.1,
		},
		{
			name: "less than zero",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   1,
					Denominator: ptr.To[int32](-1),
				},
			},
			expected: 0,
		},
		{
			name: "greater than 100",
			tracing: &egv1a1.ProxyTracing{
				SamplingFraction: &gwapiv1.Fraction{
					Numerator:   101,
					Denominator: ptr.To[int32](1),
				},
			},
			expected: 100,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual := proxySamplingRate(tc.tracing)
			if actual != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, actual)
			}
		})
	}
}
