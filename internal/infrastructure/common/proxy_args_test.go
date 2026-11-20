// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package common

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func TestBuildProxyArgsMaxConnections(t *testing.T) {
	cases := []struct {
		name     string
		spec     egv1a1.EnvoyProxySpec
		expected string
	}{
		{
			name:     "overload manager not set",
			expected: "max_active_downstream_connections: 50000",
		},
		{
			name:     "max connections not set",
			spec:     egv1a1.EnvoyProxySpec{OverloadManager: &egv1a1.OverloadManager{}},
			expected: "max_active_downstream_connections: 50000",
		},
		{
			name:     "max connections set",
			spec:     egv1a1.EnvoyProxySpec{OverloadManager: &egv1a1.OverloadManager{MaxConnections: new(int64(100000))}},
			expected: "max_active_downstream_connections: 100000",
		},
		{
			name: "user bootstrap applies after max connections",
			spec: egv1a1.EnvoyProxySpec{
				OverloadManager: &egv1a1.OverloadManager{MaxConnections: new(int64(100000))},
				Bootstrap: &egv1a1.ProxyBootstrap{
					Type: new(egv1a1.BootstrapTypeJSONPatch),
					JSONPatches: []egv1a1.JSONPatchOperation{{
						Op:    "replace",
						Path:  new("/overload_manager/resource_monitors/0/typed_config/max_active_downstream_connections"),
						Value: &apiextensionsv1.JSON{Raw: []byte("200000")},
					}},
				},
			},
			expected: "max_active_downstream_connections: 200000",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			infra := &ir.ProxyInfra{Name: "proxy", Config: &egv1a1.EnvoyProxy{Spec: tc.spec}}
			args, err := BuildProxyArgs(infra, nil, &bootstrap.RenderBootstrapConfigOptions{}, "node", false)
			require.NoError(t, err)

			i := slices.Index(args, "--config-yaml")
			require.GreaterOrEqual(t, i, 0)
			require.Contains(t, args[i+1], tc.expected)
		})
	}
}

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
