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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildAdmissionControlConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *ir.AdmissionControl
		wantErr bool
	}{
		{
			name:    "valid admission control config",
			config:  &ir.AdmissionControl{},
			wantErr: false,
		},
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name:    "empty config with defaults",
			config:  &ir.AdmissionControl{},
			wantErr: false,
		},
		{
			name: "full config with all fields",
			config: &ir.AdmissionControl{
				SamplingWindow: &metav1.Duration{
					Duration: 30 * time.Second,
				},
				MinSuccessRate:      new(uint32(90)),
				RejectionAggression: new(uint32(2)),
				MinRequestRate:      new(uint32(10)),
				MaxRejectionPercent: new(uint32(80)),
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
					HTTP: &ir.HTTPSuccessCriteria{
						StatusCodes: []int32{200, 201, 300, 301},
					},
					GRPC: &ir.GRPCSuccessCriteria{
						StatusCodes: []string{"Ok", "Cancelled"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "config with only HTTP success criteria",
			config: &ir.AdmissionControl{
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
					HTTP: &ir.HTTPSuccessCriteria{
						StatusCodes: []int32{200, 201, 202},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "config with only gRPC success criteria",
			config: &ir.AdmissionControl{
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
					GRPC: &ir.GRPCSuccessCriteria{
						StatusCodes: []string{"Ok"},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "config with empty success criteria",
			config: &ir.AdmissionControl{
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{},
			},
			wantErr: false,
		},
		{
			name: "config with custom sampling window",
			config: &ir.AdmissionControl{
				SamplingWindow: &metav1.Duration{
					Duration: 120 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "config with min thresholds",
			config: &ir.AdmissionControl{
				MinSuccessRate:      new(uint32(1)),
				RejectionAggression: new(uint32(1)),
				MinRequestRate:      new(uint32(0)),
				MaxRejectionPercent: new(uint32(0)),
			},
			wantErr: false,
		},
		{
			name: "config with max thresholds",
			config: &ir.AdmissionControl{
				MinSuccessRate:      new(uint32(100)),
				RejectionAggression: new(uint32(10)),
				MinRequestRate:      new(uint32(1000)),
				MaxRejectionPercent: new(uint32(100)),
			},
			wantErr: false,
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
			require.NotNil(t, got)

			require.NotNil(t, got.Enabled)
			require.NotNil(t, got.Enabled.DefaultValue)
			assert.True(t, got.Enabled.DefaultValue.Value)

			require.NotNil(t, got.EvaluationCriteria)
		})
	}
}

func TestBuildAdmissionControlConfigValues(t *testing.T) {
	config := &ir.AdmissionControl{
		SamplingWindow: &metav1.Duration{
			Duration: 45 * time.Second,
		},
		MinSuccessRate:      new(uint32(85)),
		RejectionAggression: new(uint32(2)),
		MinRequestRate:      new(uint32(20)),
		MaxRejectionPercent: new(uint32(75)),
		SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
			HTTP: &ir.HTTPSuccessCriteria{
				StatusCodes: []int32{200, 201, 202},
			},
			GRPC: &ir.GRPCSuccessCriteria{
				StatusCodes: []string{"Ok", "Cancelled", "Unknown"},
			},
		},
	}

	got, err := buildAdmissionControlConfig(config)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Equal(t, int64(45), got.SamplingWindow.Seconds)
	assert.Equal(t, 85.0, got.SrThreshold.DefaultValue.Value)
	assert.Equal(t, 2.0, got.Aggression.DefaultValue)
	assert.Equal(t, uint32(20), got.RpsThreshold.DefaultValue)
	assert.Equal(t, 75.0, got.MaxRejectionProbability.DefaultValue.Value)
	require.NotNil(t, got.EvaluationCriteria)
}

func TestBuildAdmissionControlConfigDefaults(t *testing.T) {
	config := &ir.AdmissionControl{}

	got, err := buildAdmissionControlConfig(config)
	require.NoError(t, err)
	require.NotNil(t, got)

	assert.Nil(t, got.SamplingWindow)
	assert.Nil(t, got.SrThreshold)
	assert.Nil(t, got.Aggression)
	assert.Nil(t, got.RpsThreshold)
	assert.Nil(t, got.MaxRejectionProbability)

	require.NotNil(t, got.Enabled)
	require.NotNil(t, got.EvaluationCriteria)
}

func TestBuildUpstreamAdmissionControlFilter(t *testing.T) {
	tests := []struct {
		name    string
		ac      *ir.AdmissionControl
		wantErr bool
	}{
		{
			name:    "nil config",
			ac:      nil,
			wantErr: true,
		},
		{
			name:    "empty config",
			ac:      &ir.AdmissionControl{},
			wantErr: false,
		},
		{
			name: "full config",
			ac: &ir.AdmissionControl{
				SamplingWindow: &metav1.Duration{
					Duration: 30 * time.Second,
				},
				MinSuccessRate:      new(uint32(95)),
				RejectionAggression: new(uint32(1)),
				MinRequestRate:      new(uint32(5)),
				MaxRejectionPercent: new(uint32(80)),
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
					HTTP: &ir.HTTPSuccessCriteria{
						StatusCodes: []int32{200, 201},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filter, err := buildUpstreamAdmissionControlFilter(tt.ac)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, filter)
			assert.Equal(t, "envoy.filters.http.admission_control", filter.Name)
			assert.NotNil(t, filter.ConfigType)
		})
	}
}

func TestGRPCStatusCodeToUint32(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		want   uint32
		wantOk bool
	}{
		{name: "Ok", input: "Ok", want: 0, wantOk: true},
		{name: "Cancelled", input: "Cancelled", want: 1, wantOk: true},
		{name: "Unavailable", input: "Unavailable", want: 14, wantOk: true},
		{name: "Unauthenticated", input: "Unauthenticated", want: 16, wantOk: true},
		{name: "invalid code", input: "Invalid", want: 0, wantOk: false},
		{name: "wrong case OK", input: "OK", want: 0, wantOk: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := grpcStatusCodeToUint32(tt.input)
			assert.Equal(t, tt.wantOk, ok)
			if tt.wantOk {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
