// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMergeEnvoyProxyConfigs(t *testing.T) {
	testCases := []struct {
		name              string
		defaultSpec       *egv1a1.EnvoyProxySpec
		gatewayClassProxy *egv1a1.EnvoyProxy
		gatewayProxy      *egv1a1.EnvoyProxy
		expectedSpec      *egv1a1.EnvoyProxySpec
		expectError       bool
	}{
		{
			name:         "no configs provided",
			expectedSpec: nil,
		},
		{
			name: "only default spec - replace mode (default)",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
		},
		{
			name: "replace mode - gatewayclass overrides default",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](8),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](8),
			},
		},
		{
			name: "replace mode - gateway overrides all",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](4),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](8),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](16),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				Concurrency: ptr.To[int32](16),
			},
		},
		{
			name: "strategic merge - combines configs from all levels",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   ptr.To(egv1a1.StrategicMerge),
				Concurrency: ptr.To[int32](4),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentAdmin: egv1a1.LogLevelDebug,
						},
					},
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](16),
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   ptr.To(egv1a1.StrategicMerge),
				Concurrency: ptr.To[int32](16), // Gateway overrides
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,  // From default
						egv1a1.LogComponentAdmin:   egv1a1.LogLevelDebug, // From gatewayclass
					},
				},
			},
		},
		{
			name: "merge type from gateway takes precedence",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				// No MergeType (replace mode)
				Concurrency: ptr.To[int32](4),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](8),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType:   ptr.To(egv1a1.StrategicMerge),
					Concurrency: ptr.To[int32](16),
				},
			},
			// Gateway specifies StrategicMerge, so configs should be merged
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   ptr.To(egv1a1.StrategicMerge),
				Concurrency: ptr.To[int32](16),
			},
		},
		{
			name: "json merge mode",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   ptr.To(egv1a1.JSONMerge),
				Concurrency: ptr.To[int32](4),
				Logging: egv1a1.ProxyLogging{
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
					},
				},
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Logging: egv1a1.ProxyLogging{
						Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
							egv1a1.LogComponentAdmin: egv1a1.LogLevelDebug,
						},
					},
				},
			},
			expectedSpec: &egv1a1.EnvoyProxySpec{
				MergeType:   ptr.To(egv1a1.JSONMerge),
				Concurrency: ptr.To[int32](4),
				Logging: egv1a1.ProxyLogging{
					// JSONMerge (RFC 7396) merges objects recursively, so both keys are present
					Level: map[egv1a1.ProxyLogComponent]egv1a1.LogLevel{
						egv1a1.LogComponentDefault: egv1a1.LogLevelInfo,
						egv1a1.LogComponentAdmin:   egv1a1.LogLevelDebug,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := MergeEnvoyProxyConfigs(tc.defaultSpec, tc.gatewayClassProxy, tc.gatewayProxy)

			if tc.expectError {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.expectedSpec == nil {
				require.Nil(t, result)
				return
			}

			require.NotNil(t, result)
			require.Equal(t, tc.expectedSpec.MergeType, result.Spec.MergeType)
			require.Equal(t, tc.expectedSpec.Concurrency, result.Spec.Concurrency)

			if len(tc.expectedSpec.Logging.Level) > 0 {
				require.Equal(t, tc.expectedSpec.Logging.Level, result.Spec.Logging.Level)
			}
		})
	}
}

func TestDetermineMergeType(t *testing.T) {
	testCases := []struct {
		name              string
		defaultSpec       *egv1a1.EnvoyProxySpec
		gatewayClassProxy *egv1a1.EnvoyProxy
		gatewayProxy      *egv1a1.EnvoyProxy
		expected          egv1a1.MergeType
	}{
		{
			name:     "no configs - returns Replace (default)",
			expected: egv1a1.Replace,
		},
		{
			name: "default spec specifies StrategicMerge",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType: ptr.To(egv1a1.StrategicMerge),
			},
			expected: egv1a1.StrategicMerge,
		},
		{
			name: "gatewayclass overrides default",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType: ptr.To(egv1a1.StrategicMerge),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: ptr.To(egv1a1.JSONMerge),
				},
			},
			expected: egv1a1.JSONMerge,
		},
		{
			name: "gateway overrides gatewayclass and default",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType: ptr.To(egv1a1.StrategicMerge),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: ptr.To(egv1a1.JSONMerge),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: ptr.To(egv1a1.StrategicMerge),
				},
			},
			expected: egv1a1.StrategicMerge,
		},
		{
			name: "gateway with nil MergeType falls back to gatewayclass",
			defaultSpec: &egv1a1.EnvoyProxySpec{
				MergeType: ptr.To(egv1a1.StrategicMerge),
			},
			gatewayClassProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					MergeType: ptr.To(egv1a1.JSONMerge),
				},
			},
			gatewayProxy: &egv1a1.EnvoyProxy{
				Spec: egv1a1.EnvoyProxySpec{
					Concurrency: ptr.To[int32](4), // No MergeType specified
				},
			},
			expected: egv1a1.JSONMerge,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := determineMergeType(tc.defaultSpec, tc.gatewayClassProxy, tc.gatewayProxy)
			require.Equal(t, tc.expected, result)
		})
	}
}
