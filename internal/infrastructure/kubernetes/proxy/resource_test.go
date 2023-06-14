// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"testing"

	"github.com/stretchr/testify/require"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
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

func TestComponentLogLevel(t *testing.T) {
	cases := []struct {
		levels    map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel
		component egcfgv1a1.LogComponent
		level     egcfgv1a1.LogLevel

		expected egcfgv1a1.LogLevel
	}{
		{
			component: egcfgv1a1.LogComponentDefault,
			level:     egcfgv1a1.LogLevelInfo,
			expected:  egcfgv1a1.LogLevelInfo,
		},
		{
			component: egcfgv1a1.LogComponentDefault,
			level:     egcfgv1a1.LogLevelWarn,
			expected:  egcfgv1a1.LogLevelWarn,
		},
		{
			component: egcfgv1a1.LogComponentDefault,
			levels: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
				egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
			},
			level:    egcfgv1a1.LogLevelWarn,
			expected: egcfgv1a1.LogLevelInfo,
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := componentLogLevel(tc.levels, tc.component, tc.level)
			require.Equal(t, tc.expected, got)
		})
	}
}

func TestComponentLogLevelArgs(t *testing.T) {
	cases := []struct {
		levels   map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel
		expected string
	}{
		{
			expected: "",
		},
		{
			levels: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
				egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
			},
			expected: "",
		},
		{
			levels: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
				egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
				egcfgv1a1.LogComponentAdmin:   egcfgv1a1.LogLevelWarn,
			},
			expected: "admin:warn",
		},
		{
			levels: map[egcfgv1a1.LogComponent]egcfgv1a1.LogLevel{
				egcfgv1a1.LogComponentDefault: egcfgv1a1.LogLevelInfo,
				egcfgv1a1.LogComponentAdmin:   egcfgv1a1.LogLevelWarn,
				egcfgv1a1.LogComponentFilter:  egcfgv1a1.LogLevelDebug,
			},
			expected: "admin:warn,filter:debug",
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := componentLogLevelArgs(tc.levels)
			require.Equal(t, tc.expected, got)
		})
	}
}
