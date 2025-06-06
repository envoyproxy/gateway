// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"math"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDetermineIPFamily(t *testing.T) {
	tests := []struct {
		name     string
		settings []*ir.DestinationSetting
		want     *egv1a1.IPFamily
	}{
		{
			name:     "nil settings should return nil",
			settings: nil,
			want:     nil,
		},
		{
			name:     "empty settings should return nil",
			settings: []*ir.DestinationSetting{},
			want:     nil,
		},
		{
			name: "single IPv4 setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv4)},
			},
			want: ptr.To(egv1a1.IPv4),
		},
		{
			name: "single IPv6 setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv6)},
			},
			want: ptr.To(egv1a1.IPv6),
		},
		{
			name: "single DualStack setting",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.DualStack)},
			},
			want: ptr.To(egv1a1.DualStack),
		},
		{
			name: "mixed IPv4 and IPv6 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv4)},
				{IPFamily: ptr.To(egv1a1.IPv6)},
			},
			want: ptr.To(egv1a1.DualStack),
		},
		{
			name: "DualStack with IPv4 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.DualStack)},
				{IPFamily: ptr.To(egv1a1.IPv4)},
			},
			want: ptr.To(egv1a1.DualStack),
		},
		{
			name: "DualStack with IPv6 should return DualStack",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.DualStack)},
				{IPFamily: ptr.To(egv1a1.IPv6)},
			},
			want: ptr.To(egv1a1.DualStack),
		},
		{
			name: "mixed with nil IPFamily should be ignored",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv4)},
				{IPFamily: nil},
				{IPFamily: ptr.To(egv1a1.IPv6)},
			},
			want: ptr.To(egv1a1.DualStack),
		},
		{
			name: "multiple IPv4 settings should return IPv4",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv4)},
				{IPFamily: ptr.To(egv1a1.IPv4)},
			},
			want: ptr.To(egv1a1.IPv4),
		},
		{
			name: "multiple IPv6 settings should return IPv6",
			settings: []*ir.DestinationSetting{
				{IPFamily: ptr.To(egv1a1.IPv6)},
				{IPFamily: ptr.To(egv1a1.IPv6)},
			},
			want: ptr.To(egv1a1.IPv6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineIPFamily(tt.settings)
			assert.Equal(t, tt.want, got)
		})
	}
}

func makeRats(pairs ...[2]int64) []*big.Rat {
	rats := make([]*big.Rat, len(pairs))
	for i, pair := range pairs {
		rats[i] = big.NewRat(pair[0], pair[1])
	}
	return rats
}

func TestNormalizeRatiosToUint32_Basic(t *testing.T) {
	rats := makeRats([2]int64{1, 3}, [2]int64{2, 3})

	result, err := normalizeRatiosToUint32(rats)
	require.NoError(t, err)
	require.Len(t, result, 2)
	require.Equal(t, result[0]*2, result[1])
}

func TestNormalizeRatiosToUint32_SingleValue(t *testing.T) {
	rats := makeRats([2]int64{42, 7})

	result, err := normalizeRatiosToUint32(rats)
	require.NoError(t, err)
	require.Len(t, result, 1)
	require.Equal(t, uint32(6), result[0])
}

func TestNormalizeRatiosToUint32_Overflow(t *testing.T) {
	rats := makeRats([2]int64{math.MaxInt64, 1})

	result, err := normalizeRatiosToUint32(rats)
	require.Nil(t, result)
	require.Error(t, err)
	require.Contains(t, err.Error(), "scaled ratios is too large")
}

func TestNormalizeRatiosToUint32_MaxUint32(t *testing.T) {
	r := big.NewRat(int64(math.MaxUint32), 1)
	rats := []*big.Rat{r}

	result, err := normalizeRatiosToUint32(rats)
	require.NoError(t, err)
	require.Equal(t, []uint32{math.MaxUint32}, result)
}
