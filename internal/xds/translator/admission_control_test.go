// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestListenerContainsAdmissionControl(t *testing.T) {
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
							AdmissionControl: &ir.AdmissionControl{},
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
		{
			name: "listener with nil traffic",
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: nil,
					},
				},
			},
			want: false,
		},
		{
			name: "listener with multiple routes, one has admission control",
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: &ir.TrafficFeatures{},
					},
					{
						Traffic: &ir.TrafficFeatures{
							AdmissionControl: &ir.AdmissionControl{},
						},
					},
				},
			},
			want: true,
		},
		{
			name: "listener with empty routes",
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{},
			},
			want: false,
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
				SuccessRateThreshold:    ptr.To(0.90),
				Aggression:              ptr.To(2.0),
				RPSThreshold:            ptr.To(uint32(10)),
				MaxRejectionProbability: ptr.To(0.80),
				SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
					HTTP: &ir.HTTPSuccessCriteria{
						HTTPSuccessStatus: []ir.HTTPStatusRange{
							{Start: 200, End: 299},
							{Start: 300, End: 399},
						},
					},
					GRPC: &ir.GRPCSuccessCriteria{
						GRPCSuccessStatus: []int32{0, 1},
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
						HTTPSuccessStatus: []ir.HTTPStatusRange{
							{Start: 200, End: 299},
						},
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
						GRPCSuccessStatus: []int32{0},
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
			name: "config with zero thresholds",
			config: &ir.AdmissionControl{
				SuccessRateThreshold:    ptr.To(0.0),
				Aggression:              ptr.To(0.0),
				RPSThreshold:            ptr.To(uint32(0)),
				MaxRejectionProbability: ptr.To(0.0),
			},
			wantErr: false,
		},
		{
			name: "config with max thresholds",
			config: &ir.AdmissionControl{
				SuccessRateThreshold:    ptr.To(1.0),
				Aggression:              ptr.To(10.0),
				RPSThreshold:            ptr.To(uint32(1000)),
				MaxRejectionProbability: ptr.To(1.0),
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

			// Verify enabled is always true
			require.NotNil(t, got.Enabled)
			require.NotNil(t, got.Enabled.DefaultValue)
			assert.True(t, got.Enabled.DefaultValue.Value)

			// Verify sampling window is set
			require.NotNil(t, got.SamplingWindow)

			// Verify sr threshold is set
			require.NotNil(t, got.SrThreshold)
			require.NotNil(t, got.SrThreshold.DefaultValue)

			// Verify aggression is set
			require.NotNil(t, got.Aggression)

			// Verify RPS threshold is set
			require.NotNil(t, got.RpsThreshold)

			// Verify max rejection probability is set
			require.NotNil(t, got.MaxRejectionProbability)
			require.NotNil(t, got.MaxRejectionProbability.DefaultValue)

			// Verify evaluation criteria is always set
			require.NotNil(t, got.EvaluationCriteria)
		})
	}
}

func TestBuildAdmissionControlConfigValues(t *testing.T) {
	// Test that specific values are correctly translated
	config := &ir.AdmissionControl{
		SamplingWindow: &metav1.Duration{
			Duration: 45 * time.Second,
		},
		SuccessRateThreshold:    ptr.To(0.85),
		Aggression:              ptr.To(1.5),
		RPSThreshold:            ptr.To(uint32(20)),
		MaxRejectionProbability: ptr.To(0.75),
		SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
			HTTP: &ir.HTTPSuccessCriteria{
				HTTPSuccessStatus: []ir.HTTPStatusRange{
					{Start: 200, End: 299},
				},
			},
			GRPC: &ir.GRPCSuccessCriteria{
				GRPCSuccessStatus: []int32{0, 1, 2},
			},
		},
	}

	got, err := buildAdmissionControlConfig(config)
	require.NoError(t, err)
	require.NotNil(t, got)

	// Verify sampling window (45 seconds)
	assert.Equal(t, int64(45), got.SamplingWindow.Seconds)

	// Verify success rate threshold (0.85 * 100 = 85.0)
	assert.Equal(t, 85.0, got.SrThreshold.DefaultValue.Value)

	// Verify aggression
	assert.Equal(t, 1.5, got.Aggression.DefaultValue)

	// Verify RPS threshold
	assert.Equal(t, uint32(20), got.RpsThreshold.DefaultValue)

	// Verify max rejection probability (0.75 * 100 = 75.0)
	assert.Equal(t, 75.0, got.MaxRejectionProbability.DefaultValue.Value)

	// Verify evaluation criteria is set (detailed verification done via testdata tests)
	require.NotNil(t, got.EvaluationCriteria)
}

func TestBuildAdmissionControlConfigDefaults(t *testing.T) {
	// Test that defaults are correctly applied when no values are set
	config := &ir.AdmissionControl{}

	got, err := buildAdmissionControlConfig(config)
	require.NoError(t, err)
	require.NotNil(t, got)

	// Default sampling window is 60s
	assert.Equal(t, int64(60), got.SamplingWindow.Seconds)

	// Default success rate threshold is 0.95 * 100 = 95.0
	assert.Equal(t, 95.0, got.SrThreshold.DefaultValue.Value)

	// Default aggression is 1.0
	assert.Equal(t, 1.0, got.Aggression.DefaultValue)

	// Default RPS threshold is 1
	assert.Equal(t, uint32(1), got.RpsThreshold.DefaultValue)

	// Default max rejection probability is 0.95 * 100 = 95.0
	assert.Equal(t, 95.0, got.MaxRejectionProbability.DefaultValue.Value)
}

func TestAdmissionControlPatchHCM(t *testing.T) {
	ac := &admissionControl{}

	tests := []struct {
		name     string
		mgr      *hcmv3.HttpConnectionManager
		listener *ir.HTTPListener
		wantErr  bool
		wantLen  int // expected number of http filters after patching
	}{
		{
			name:     "nil hcm",
			mgr:      nil,
			listener: &ir.HTTPListener{},
			wantErr:  true,
		},
		{
			name:     "nil listener",
			mgr:      &hcmv3.HttpConnectionManager{},
			listener: nil,
			wantErr:  true,
		},
		{
			name: "no routes with admission control",
			mgr:  &hcmv3.HttpConnectionManager{},
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{Traffic: &ir.TrafficFeatures{}},
				},
			},
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "route with admission control",
			mgr:  &hcmv3.HttpConnectionManager{},
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: &ir.TrafficFeatures{
							AdmissionControl: &ir.AdmissionControl{},
						},
					},
				},
			},
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "filter already exists",
			mgr: &hcmv3.HttpConnectionManager{
				HttpFilters: []*hcmv3.HttpFilter{
					{Name: string(egv1a1.EnvoyFilterAdmissionControl)},
				},
			},
			listener: &ir.HTTPListener{
				Routes: []*ir.HTTPRoute{
					{
						Traffic: &ir.TrafficFeatures{
							AdmissionControl: &ir.AdmissionControl{},
						},
					},
				},
			},
			wantErr: false,
			wantLen: 1, // Should not add duplicate
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ac.patchHCM(tt.mgr, tt.listener)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, tt.mgr.HttpFilters, tt.wantLen)
		})
	}
}

func TestAdmissionControlPatchRoute(t *testing.T) {
	ac := &admissionControl{}

	tests := []struct {
		name     string
		route    *routev3.Route
		irRoute  *ir.HTTPRoute
		listener *ir.HTTPListener
		wantErr  bool
		wantCfg  bool // whether typedPerFilterConfig should be set
	}{
		{
			name:     "nil route",
			route:    nil,
			irRoute:  &ir.HTTPRoute{},
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  false,
		},
		{
			name:     "nil ir route",
			route:    &routev3.Route{},
			irRoute:  nil,
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  false,
		},
		{
			name:  "route without traffic features",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Traffic: nil,
			},
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  false,
		},
		{
			name:  "route without admission control",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{},
			},
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  false,
		},
		{
			name:  "route with admission control",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					AdmissionControl: &ir.AdmissionControl{},
				},
			},
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  true,
		},
		{
			name:  "route with full admission control config",
			route: &routev3.Route{},
			irRoute: &ir.HTTPRoute{
				Traffic: &ir.TrafficFeatures{
					AdmissionControl: &ir.AdmissionControl{
						SamplingWindow: &metav1.Duration{
							Duration: 30 * time.Second,
						},
						SuccessRateThreshold:    ptr.To(0.90),
						Aggression:              ptr.To(2.0),
						RPSThreshold:            ptr.To(uint32(10)),
						MaxRejectionProbability: ptr.To(0.80),
						SuccessCriteria: &ir.AdmissionControlSuccessCriteria{
							HTTP: &ir.HTTPSuccessCriteria{
								HTTPSuccessStatus: []ir.HTTPStatusRange{
									{Start: 200, End: 299},
								},
							},
						},
					},
				},
			},
			listener: &ir.HTTPListener{},
			wantErr:  false,
			wantCfg:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ac.patchRoute(tt.route, tt.irRoute, tt.listener)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			if tt.wantCfg {
				require.NotNil(t, tt.route.TypedPerFilterConfig)
				assert.Contains(t, tt.route.TypedPerFilterConfig, string(egv1a1.EnvoyFilterAdmissionControl))
			}
		})
	}
}

func TestAdmissionControlPatchResources(t *testing.T) {
	ac := &admissionControl{}
	tCtx := &types.ResourceVersionTable{}
	routes := []*ir.HTTPRoute{
		{
			Traffic: &ir.TrafficFeatures{
				AdmissionControl: &ir.AdmissionControl{},
			},
		},
	}

	// patchResources should always return nil since admission control doesn't need additional resources
	err := ac.patchResources(tCtx, routes)
	require.NoError(t, err)
}

func TestBuildHCMAdmissionControlFilter(t *testing.T) {
	filter, err := buildHCMAdmissionControlFilter()
	require.NoError(t, err)
	require.NotNil(t, filter)
	assert.Equal(t, string(egv1a1.EnvoyFilterAdmissionControl), filter.Name)
	assert.NotNil(t, filter.ConfigType)
}

func TestParseDuration(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Duration
		wantErr bool
	}{
		{
			name:    "valid duration 60s",
			input:   "60s",
			want:    60 * time.Second,
			wantErr: false,
		},
		{
			name:    "valid duration 30s",
			input:   "30s",
			want:    30 * time.Second,
			wantErr: false,
		},
		{
			name:    "valid duration 1m",
			input:   "1m0s",
			want:    time.Minute,
			wantErr: false,
		},
		{
			name:    "invalid duration",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseDuration(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
