// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fraction

import (
	"testing"

	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestDeref(t *testing.T) {
	testCases := []struct {
		name         string
		fraction     *gwapiv1.Fraction
		defaultValue float64
		expected     float64
	}{
		{
			name:         "nil fraction",
			fraction:     nil,
			defaultValue: 0.5,
			expected:     0.5,
		},
		{
			name: "fraction with default denominator",
			fraction: &gwapiv1.Fraction{
				Numerator: 50,
			},
			defaultValue: 0.1,
			expected:     0.5, // 50/100
		},
		{
			name: "fraction with explicit denominator",
			fraction: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To(int32(10)),
			},
			defaultValue: 0.1,
			expected:     0.1, // 1/10
		},
		{
			name: "fraction with zero numerator",
			fraction: &gwapiv1.Fraction{
				Numerator: 0,
			},
			defaultValue: 1.0,
			expected:     0.0,
		},
		{
			name: "fraction with large values",
			fraction: &gwapiv1.Fraction{
				Numerator:   1000,
				Denominator: new(int32(1000)),
			},
			defaultValue: 1.0,
			expected:     1.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Deref(tc.fraction, tc.defaultValue)
			if result != tc.expected {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
