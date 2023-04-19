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
				"foo":                            "bar",
				"app.gateway.envoyproxy.io/name": "envoy",
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run("", func(t *testing.T) {
			got := EnvoyLabels(tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}
