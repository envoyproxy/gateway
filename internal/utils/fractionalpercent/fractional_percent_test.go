// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fractionalpercent

import (
	"fmt"
	"testing"

	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestFromIn32(t *testing.T) {
	cases := []struct {
		input    int32
		expected *xdstype.FractionalPercent
	}{
		{
			input: 0,
			expected: &xdstype.FractionalPercent{
				Numerator:   0,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
		{
			input: 50,
			expected: &xdstype.FractionalPercent{
				Numerator:   50,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
		{
			input: 100,
			expected: &xdstype.FractionalPercent{
				Numerator:   100,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%d", tc.input), func(t *testing.T) {
			result := FromIn32(tc.input)
			if result.Numerator != tc.expected.Numerator || result.Denominator != tc.expected.Denominator {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestFromFloat32(t *testing.T) {
	cases := []struct {
		input    float32
		expected *xdstype.FractionalPercent
	}{
		{
			input: 0,
			expected: &xdstype.FractionalPercent{
				Numerator:   0,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
		{
			input: 12.34,
			expected: &xdstype.FractionalPercent{
				Numerator:   123400,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
		{
			input: 100.0,
			expected: &xdstype.FractionalPercent{
				Numerator:   1000000,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.input), func(t *testing.T) {
			result := FromFloat32(tc.input)
			if result.Numerator != tc.expected.Numerator || result.Denominator != tc.expected.Denominator {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestFromFraction(t *testing.T) {
	cases := []struct {
		input    *gwapiv1.Fraction
		expected *xdstype.FractionalPercent
	}{
		{
			input: &gwapiv1.Fraction{
				Numerator: 1,
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   1,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](100),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   1,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator: 50,
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   50,
				Denominator: xdstype.FractionalPercent_HUNDRED,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   50,
				Denominator: ptr.To[int32](1000000),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   50,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](1000),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   10,
				Denominator: xdstype.FractionalPercent_TEN_THOUSAND,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](10000),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   1,
				Denominator: xdstype.FractionalPercent_TEN_THOUSAND,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](100000),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   10,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
		{
			input: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](1000000),
			},
			expected: &xdstype.FractionalPercent{
				Numerator:   1,
				Denominator: xdstype.FractionalPercent_MILLION,
			},
		},
	}

	for _, tc := range cases {
		denominator := 100
		if tc.input.Denominator != nil {
			denominator = int(*tc.input.Denominator)
		}
		caseName := fmt.Sprintf("%v/%v", tc.input.Numerator, denominator)
		t.Run(caseName, func(t *testing.T) {
			result := FromFraction(tc.input)
			if result.Numerator != tc.expected.Numerator || result.Denominator != tc.expected.Denominator {
				t.Errorf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}
