// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestMergeBackendTrafficPolicy(t *testing.T) {
	r := resource.MustParse("100m")

	cases := []struct {
		name     string
		original *egv1a1.BackendTrafficPolicy
		patch    *egv1a1.BackendTrafficPolicy

		expected          *egv1a1.BackendTrafficPolicy
		jsonMergeExpected *egv1a1.BackendTrafficPolicy
	}{
		{
			name: "merge",
			original: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: &r,
						},
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](2),
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type: "original",
						},
					},
				},
			},
			patch: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](3),
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type: "patched",
						},
					},
				},
			},
			expected: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: &r,
						},
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](3),
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type: "patched",
						},
						{
							Type: "original",
						},
					},
				},
			},
			jsonMergeExpected: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: &r,
						},
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](3),
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type: "patched",
						},
					},
				},
			},
		},
		{
			name: "override",
			original: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](13),
						},
					},
				},
			},
			patch: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](3),
						},
					},
				},
			},
			expected: &egv1a1.BackendTrafficPolicy{
				Spec: egv1a1.BackendTrafficPolicySpec{
					ClusterSettings: egv1a1.ClusterSettings{
						Retry: &egv1a1.Retry{
							NumRetries: ptr.To[int32](3),
						},
					},
				},
			},
		},
	}
	for _, tc := range cases {
		for _, mergeType := range []egv1a1.MergeType{egv1a1.StrategicMerge, egv1a1.JSONMerge} {
			t.Run(fmt.Sprintf("%s/%s", mergeType, tc.name), func(t *testing.T) {
				got, err := Merge[*egv1a1.BackendTrafficPolicy](tc.original, tc.patch, mergeType)
				require.NoError(t, err)

				switch mergeType {
				case egv1a1.StrategicMerge:
					require.Equal(t, tc.expected, got)
				case egv1a1.JSONMerge:
					if tc.jsonMergeExpected != nil {
						require.Equal(t, tc.jsonMergeExpected, got)
					} else {
						require.Equal(t, tc.expected, got)
					}
				}
			})
		}
	}
}
