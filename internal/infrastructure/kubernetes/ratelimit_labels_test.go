// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRateLimitLabels(t *testing.T) {
	cases := []struct {
		name     string
		expected map[string]string
	}{
		{
			name: "ratelimit-labels",
			expected: map[string]string{
				"app.gateway.envoyproxy.io/name": rateLimitInfraName,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			got := rateLimitLabels()
			require.Equal(t, tc.expected, got)
		})
	}
}
