// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
)

func TestValidateEnvoyProxy(t *testing.T) {
	testCases := []struct {
		name     string
		obj      *egcfgv1a1.EnvoyProxy
		expected bool
	}{
		{
			name:     "nil envoyproxy",
			obj:      nil,
			expected: false,
		},
		{
			name: "nil provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: nil,
				},
			},
			expected: true,
		},
		{
			name: "unsupported provider",
			obj: &egcfgv1a1.EnvoyProxy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
				Spec: egcfgv1a1.EnvoyProxySpec{
					Provider: &egcfgv1a1.ResourceProvider{
						Type: egcfgv1a1.ProviderTypeFile,
					},
				},
			},
			expected: false,
		},
	}

	for i := range testCases {
		tc := testCases[i]
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateEnvoyProxy(tc.obj)
			if tc.expected {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
