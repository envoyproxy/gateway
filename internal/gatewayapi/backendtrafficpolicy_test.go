// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"math"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestInt64ToUint32(t *testing.T) {
	type testCase struct {
		Name    string
		In      int64
		Out     uint32
		Success bool
	}

	testCases := []testCase{
		{
			Name:    "valid",
			In:      1024,
			Out:     1024,
			Success: true,
		},
		{
			Name:    "invalid-underflow",
			In:      -1,
			Out:     0,
			Success: false,
		},
		{
			Name:    "invalid-overflow",
			In:      math.MaxUint32 + 1,
			Out:     0,
			Success: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			out, success := int64ToUint32(tc.In)
			require.Equal(t, tc.Out, out)
			require.Equal(t, tc.Success, success)
		})
	}
}

func TestMakeIrStatusSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.HTTPStatus
		want []ir.HTTPStatus
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.HTTPStatus{200, 404},
			want: []ir.HTTPStatus{200, 404},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.HTTPStatus{200, 404, 200},
			want: []ir.HTTPStatus{200, 404},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrStatusSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrStatusSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeIrTriggerSet(t *testing.T) {
	tests := []struct {
		name string
		in   []egv1a1.TriggerEnum
		want []ir.TriggerEnum
	}{
		{
			name: "no duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
		{
			name: "with duplicates",
			in:   []egv1a1.TriggerEnum{"5xx", "reset", "5xx"},
			want: []ir.TriggerEnum{"5xx", "reset"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := makeIrTriggerSet(tt.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("makeIrTriggerSet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsPrimeNumber(t *testing.T) {
	cases := []struct {
		n       uint64
		isPrime bool
	}{
		{n: 0, isPrime: false},
		{n: 1, isPrime: false},
		{n: 2, isPrime: true},
		{n: 3, isPrime: true},
		{n: 4, isPrime: false},
		{n: 5, isPrime: true},
		{n: 6, isPrime: false},
		{n: 7, isPrime: true},
		{n: 8, isPrime: false},
		{n: 9, isPrime: false},
		{n: 10, isPrime: false},
		{n: 11, isPrime: true},
		{n: 12, isPrime: false},
		{n: 13, isPrime: true},
		{n: 14, isPrime: false},
		{n: 15, isPrime: false},
		{n: 16, isPrime: false},
		{n: 17, isPrime: true},
		{n: 18, isPrime: false},
		{n: 19, isPrime: true},
		{n: 20, isPrime: false},
		{n: 5000011, isPrime: true},
	}

	for _, tc := range cases {
		if got := isPrimeNumber(tc.n); got != tc.isPrime {
			t.Errorf("isPrimeNumber(%d) = %v, want %v", tc.n, got, tc.isPrime)
		}
	}
}
