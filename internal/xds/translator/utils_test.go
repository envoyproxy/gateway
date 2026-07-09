// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestDetermineIPFamily(t *testing.T) {
	tests := []struct {
		name     string
		backends []*ir.BackendCluster
		want     *egv1a1.IPFamily
	}{
		{
			name:     "nil backends should return nil",
			backends: nil,
			want:     nil,
		},
		{
			name:     "empty backends should return nil",
			backends: []*ir.BackendCluster{},
			want:     nil,
		},
		{
			name:     "single IPv4 setting",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{{IPFamily: new(egv1a1.IPv4)}}}},
			want:     new(egv1a1.IPv4),
		},
		{
			name:     "single IPv6 setting",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{{IPFamily: new(egv1a1.IPv6)}}}},
			want:     new(egv1a1.IPv6),
		},
		{
			name:     "single DualStack setting",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{{IPFamily: new(egv1a1.DualStack)}}}},
			want:     new(egv1a1.DualStack),
		},
		{
			name: "mixed IPv4 and IPv6 should return DualStack",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.IPv4)},
				{IPFamily: new(egv1a1.IPv6)},
			}}},
			want: new(egv1a1.DualStack),
		},
		{
			name: "DualStack with IPv4 should return DualStack",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.DualStack)},
				{IPFamily: new(egv1a1.IPv4)},
			}}},
			want: new(egv1a1.DualStack),
		},
		{
			name: "DualStack with IPv6 should return DualStack",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.DualStack)},
				{IPFamily: new(egv1a1.IPv6)},
			}}},
			want: new(egv1a1.DualStack),
		},
		{
			name: "mixed with nil IPFamily should be ignored",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.IPv4)},
				{IPFamily: nil},
				{IPFamily: new(egv1a1.IPv6)},
			}}},
			want: new(egv1a1.DualStack),
		},
		{
			name: "multiple IPv4 settings should return IPv4",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.IPv4)},
				{IPFamily: new(egv1a1.IPv4)},
			}}},
			want: new(egv1a1.IPv4),
		},
		{
			name: "multiple IPv6 settings should return IPv6",
			backends: []*ir.BackendCluster{{Settings: []*ir.DestinationSetting{
				{IPFamily: new(egv1a1.IPv6)},
				{IPFamily: new(egv1a1.IPv6)},
			}}},
			want: new(egv1a1.IPv6),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineIPFamily(tt.backends)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCreateExtServiceXDSCluster(t *testing.T) {
	tests := []struct {
		name         string
		rd           *ir.RouteDestination
		backendIndex map[string]*ir.BackendCluster
		want         error
	}{
		{
			name: "success with single backend cluster",
			rd: &ir.RouteDestination{
				Name: "ext-svc",
				BackendClusterRefs: []*ir.BackendClusterRef{{
					Name: "ext-svc",
				}},
			},
			backendIndex: map[string]*ir.BackendCluster{
				"ext-svc": {
					Name: "ext-svc",
					Settings: []*ir.DestinationSetting{{
						Endpoints:   []*ir.DestinationEndpoint{{Host: "10.0.0.1", Port: 8080}},
						AddressType: new(ir.IP),
					}},
				},
			},
			want: nil,
		},
		{
			name: "error with multiple backend clusters",
			rd: &ir.RouteDestination{
				Name: "ext-svc",
				BackendClusterRefs: []*ir.BackendClusterRef{
					{Name: "bc-1"},
					{Name: "bc-2"},
				},
			},
			backendIndex: map[string]*ir.BackendCluster{
				"bc-1": {Name: "bc-1", Settings: []*ir.DestinationSetting{{Endpoints: []*ir.DestinationEndpoint{{Host: "10.0.0.1", Port: 8080}}}}},
				"bc-2": {Name: "bc-2", Settings: []*ir.DestinationSetting{{Endpoints: []*ir.DestinationEndpoint{{Host: "10.0.0.2", Port: 8080}}}}},
			},
			want: fmt.Errorf("ext service destination ext-svc must have exactly one backend cluster, got 2"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tCtx := &types.ResourceVersionTable{BackendIndex: tt.backendIndex}
			err := createExtServiceXDSCluster(tt.rd, nil, tCtx)
			if tt.want == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.want.Error())
			}
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
