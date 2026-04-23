// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package fraction

import (
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func Deref(fraction *gwapiv1.Fraction, defaultValue float64) float64 {
	if fraction != nil {
		numerator := float64(fraction.Numerator)
		denominator := float64(ptr.Deref(fraction.Denominator, 100))
		return numerator / denominator
	}
	return defaultValue
}
