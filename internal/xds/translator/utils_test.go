// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
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

func TestAddClusterFromURLWithTraffic(t *testing.T) {
	tCtx := &types.ResourceVersionTable{}
	traffic := &ir.TrafficFeatures{
		Timeout: &ir.Timeout{
			TCP: &ir.TCPTimeout{
				ConnectTimeout: &metav1.Duration{Duration: 2 * time.Second},
			},
		},
	}

	err := addClusterFromURL("https://example.com/jwks.json", traffic, tCtx)
	require.NoError(t, err)

	cluster := findXdsCluster(tCtx, "example_com_443")
	require.NotNil(t, cluster)
	require.Equal(t, durationpb.New(2*time.Second), cluster.ConnectTimeout)
}
