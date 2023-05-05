// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvoyPodSelector(t *testing.T) {
	cases := []struct {
		name     string
		in       map[string]string
		expected map[string]string
	}{
		{
			name: "default",
			in:   map[string]string{"foo": "bar"},
			expected: map[string]string{
				"foo":                          "bar",
				"app.kubernetes.io/name":       "envoy",
				"app.kubernetes.io/component":  "proxy",
				"app.kubernetes.io/managed-by": "envoy-gateway",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			got := envoyLabels(tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}
