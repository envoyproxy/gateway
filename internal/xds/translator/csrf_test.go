// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

// Envoy only lets an invalid request through in shadow mode if that request wasn't also
// selected by FilterEnabled, and it draws the two fractions independently. FilterEnabled
// must therefore be the complement of the shadow fraction, and ShadowEnabled must stay at
// 100% so it claims the whole non-enforced remainder rather than a fraction of it.
func TestBuildXdsCSRFPolicyFractions(t *testing.T) {
	tests := []struct {
		name           string
		shadowFraction *gwapiv1.Fraction
		expectedFilter *xdstype.FractionalPercent
		expectedShadow *xdstype.FractionalPercent
	}{
		{
			name:           "no shadow fraction enforces everything",
			shadowFraction: nil,
			expectedFilter: &xdstype.FractionalPercent{Numerator: 100, Denominator: xdstype.FractionalPercent_HUNDRED},
		},
		{
			name:           "full shadow fraction enforces nothing",
			shadowFraction: &gwapiv1.Fraction{Numerator: 100},
			expectedFilter: &xdstype.FractionalPercent{Numerator: 0, Denominator: xdstype.FractionalPercent_HUNDRED},
			expectedShadow: &xdstype.FractionalPercent{Numerator: 100, Denominator: xdstype.FractionalPercent_HUNDRED},
		},
		{
			name:           "partial shadow fraction enforces the remainder",
			shadowFraction: &gwapiv1.Fraction{Numerator: 20},
			expectedFilter: &xdstype.FractionalPercent{Numerator: 80, Denominator: xdstype.FractionalPercent_HUNDRED},
			expectedShadow: &xdstype.FractionalPercent{Numerator: 100, Denominator: xdstype.FractionalPercent_HUNDRED},
		},
		{
			name:           "custom denominator is preserved in the complement",
			shadowFraction: &gwapiv1.Fraction{Numerator: 25, Denominator: new(int32(1000))},
			expectedFilter: &xdstype.FractionalPercent{Numerator: 9750, Denominator: xdstype.FractionalPercent_TEN_THOUSAND},
			expectedShadow: &xdstype.FractionalPercent{Numerator: 100, Denominator: xdstype.FractionalPercent_HUNDRED},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			policy := buildXdsCSRFPolicy(&ir.CSRF{ShadowFraction: tc.shadowFraction})

			assert.Equal(t, tc.expectedFilter.GetNumerator(), policy.GetFilterEnabled().GetDefaultValue().GetNumerator())
			assert.Equal(t, tc.expectedFilter.GetDenominator(), policy.GetFilterEnabled().GetDefaultValue().GetDenominator())

			if tc.expectedShadow == nil {
				assert.Nil(t, policy.GetShadowEnabled())
				return
			}
			assert.Equal(t, tc.expectedShadow.GetNumerator(), policy.GetShadowEnabled().GetDefaultValue().GetNumerator())
			assert.Equal(t, tc.expectedShadow.GetDenominator(), policy.GetShadowEnabled().GetDefaultValue().GetDenominator())
		})
	}
}
