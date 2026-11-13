// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestDrainStrategyArg(t *testing.T) {
	cases := []struct {
		name     string
		shutdown *egv1a1.ShutdownConfig
		expected string
	}{
		{
			name:     "nil shutdown config",
			expected: "immediate",
		},
		{
			name:     "drain strategy not set",
			shutdown: &egv1a1.ShutdownConfig{},
			expected: "immediate",
		},
		{
			name:     "gradual",
			shutdown: &egv1a1.ShutdownConfig{DrainStrategy: new(egv1a1.DrainStrategyGradual)},
			expected: "gradual",
		},
		{
			name:     "immediate",
			shutdown: &egv1a1.ShutdownConfig{DrainStrategy: new(egv1a1.DrainStrategyImmediate)},
			expected: "immediate",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, drainStrategyArg(tc.shutdown))
		})
	}
}
