package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEnvoyPodSelector(t *testing.T) {
	cases := []struct {
		name     string
		gcName   string
		expected map[string]string
	}{
		{
			name:   "default",
			gcName: "eg",
			expected: map[string]string{
				"gatewayClass":                   "eg",
				"app.gateway.envoyproxy.io/name": "envoy",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			got := envoyPodSelector(tc.gcName)
			require.Equal(t, tc.expected, got.MatchLabels)
		})
	}
}
