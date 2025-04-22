// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"testing"

	xdstype "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/golang/protobuf/proto"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestFromFraction(t *testing.T) {
	tests := []struct {
		name     string
		fraction *gwapiv1.Fraction
		want     *xdstype.Percent
	}{
		{
			name:     "nil fraction",
			fraction: nil,
			want:     nil,
		},
		{
			name: "valid fraction",
			fraction: &gwapiv1.Fraction{
				Numerator:   1,
				Denominator: ptr.To[int32](100),
			},
			want: &xdstype.Percent{
				Value: 0.01,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FromFraction(tt.fraction)
			if !proto.Equal(got, tt.want) {
				t.Errorf("FromFraction() = %v, want %v", got, tt.want)
			}
		})
	}
}
