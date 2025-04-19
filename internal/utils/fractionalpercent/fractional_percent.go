// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fractionalpercent

import (
	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/shopspring/decimal"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// FromFloat32 translates a float to envoy.type.FractionalPercent.
func FromFloat32(p float32) *xdstype.FractionalPercent {
	return &xdstype.FractionalPercent{
		Numerator:   uint32(p * 10000),
		Denominator: xdstype.FractionalPercent_MILLION,
	}
}

// FromIn32 translates an int32 instance to envoy.type.FractionalPercent.
func FromIn32(p int32) *xdstype.FractionalPercent {
	return &xdstype.FractionalPercent{
		Numerator:   uint32(p),
		Denominator: xdstype.FractionalPercent_HUNDRED,
	}
}

var million = decimal.NewFromInt(1000000)

const (
	Hundred         = 100
	Thousand        = 1000
	TenThousand     = 10000
	HundredThousand = 100000
	Million         = 1000000
)

// FromFraction translates a gwapiv1.Fraction instance to envoy.type.FractionalPercent.
func FromFraction(fraction *gwapiv1.Fraction) *xdstype.FractionalPercent {
	if fraction.Denominator == nil {
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator),
			Denominator: xdstype.FractionalPercent_HUNDRED,
		}
	}

	// envoy only support denominator with one of [100, 10000, 1000000]
	switch *fraction.Denominator {
	case Hundred:
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator),
			Denominator: xdstype.FractionalPercent_HUNDRED,
		}
	case Thousand:
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator * 10),
			Denominator: xdstype.FractionalPercent_TEN_THOUSAND,
		}
	case TenThousand:
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator),
			Denominator: xdstype.FractionalPercent_TEN_THOUSAND,
		}
	case HundredThousand:
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator * 10),
			Denominator: xdstype.FractionalPercent_MILLION,
		}
	case Million:
		return &xdstype.FractionalPercent{
			Numerator:   uint32(fraction.Numerator),
			Denominator: xdstype.FractionalPercent_MILLION,
		}
	}

	// Envoy only supports 100, 10000, and 1000000 as denominator.
	// Convert the fraction to a millionths' representation.
	percent := decimal.NewFromInt32(fraction.Numerator).Mul(million).Div(decimal.NewFromInt32(*fraction.Denominator))
	return &xdstype.FractionalPercent{
		Numerator:   uint32(percent.IntPart()),
		Denominator: xdstype.FractionalPercent_MILLION,
	}
}
