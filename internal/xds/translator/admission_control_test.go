// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestAdmissionControlFilter(t *testing.T) {
	tests := []struct {
		name     string
		listener *ir.HTTPListener
		want     bool
	}{
		{
			name: "listener with admission control",
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: &ir.TrafficFeatures{
							AdmissionControl: &ir.AdmissionControl{
								Enabled: func() *bool { b := true; return &b }(),
							},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "listener without admission control",
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: &ir.TrafficFeatures{},
					},
				},
			},
			want: false,
		},
		{
			name:     "nil listener",
			listener: nil,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := listenerContainsAdmissionControl(tt.listener)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestBuildAdmissionControlConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ir.AdmissionControl
		wantErr bool
	}{
		{
			name: "valid admission control config",
			config: &ir.AdmissionControl{
				Enabled: func() *bool { b := true; return &b }(),
			},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildAdmissionControlConfig(tt.config)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, got)
		})
	}
}
