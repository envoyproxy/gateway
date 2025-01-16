// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"fmt"
	"os"
	"path"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

func TestBuildReadyListener(t *testing.T) {
	cases := []struct {
		name  string
		ready *ir.ReadyListener
	}{
		{
			name: "ipv4",
			ready: &ir.ReadyListener{
				IPFamily:          egv1a1.IPv4,
				Address:           "0.0.0.0",
				Port:              19001,
				Path:              "/ready",
				PrometheusEnabled: true,
			},
		},
		{
			name: "ipv4-brotli",
			ready: &ir.ReadyListener{
				IPFamily:                 egv1a1.IPv4,
				Address:                  "0.0.0.0",
				Port:                     19001,
				Path:                     "/ready",
				PrometheusEnabled:        true,
				PrometheusCompressorType: ptr.To(egv1a1.BrotliCompressorType),
			},
		},
		{
			name: "ipv4-gzip",
			ready: &ir.ReadyListener{
				IPFamily:                 egv1a1.IPv4,
				Address:                  "0.0.0.0",
				Port:                     19001,
				Path:                     "/ready",
				PrometheusEnabled:        true,
				PrometheusCompressorType: ptr.To(egv1a1.GzipCompressorType),
			},
		},
		{
			name: "ipv6",
			ready: &ir.ReadyListener{
				IPFamily:          egv1a1.IPv6,
				Address:           "::",
				Port:              19001,
				Path:              "/ready",
				PrometheusEnabled: true,
			},
		},
		{
			name: "dual",
			ready: &ir.ReadyListener{
				IPFamily:          egv1a1.DualStack,
				Address:           "::",
				Port:              19001,
				Path:              "/ready",
				PrometheusEnabled: true,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildReadyListener(tc.ready)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if *overrideTestData {
				data, err := proto.ToYAML(got)
				require.NoError(t, err)
				err = os.WriteFile(path.Join("testdata", "readylistener", tc.name+".yaml"), data, 0o600)
				require.NoError(t, err)
				return
			}

			data := readTestData(t, "readylistener", tc.name)
			expected := &routev3.RouteConfiguration{}
			err = proto.FromYAML(data, expected)
			require.NoError(t, err)

			require.Empty(t, cmp.Diff(expected, got, protocmp.Transform()))
		})
	}
}

func TestBuildReadyRouteConfig(t *testing.T) {
	cases := []struct {
		compressorEnabled bool
	}{
		{
			compressorEnabled: true,
		},
		{
			compressorEnabled: false,
		},
	}

	for _, c := range cases {
		caseName := fmt.Sprintf("compressor-%v", c.compressorEnabled)
		t.Run(caseName, func(t *testing.T) {
			got, err := buildReadyRouteConfig(true, c.compressorEnabled)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if *overrideTestData {
				data, err := proto.ToYAML(got)
				require.NoError(t, err)
				err = os.WriteFile(path.Join("testdata", "readyroute", caseName+".yaml"), data, 0o600)
				require.NoError(t, err)
				return
			}

			data := readTestData(t, "readyroute", caseName)
			expected := &routev3.RouteConfiguration{}
			err = proto.FromYAML(data, expected)
			require.NoError(t, err)

			require.Empty(t, cmp.Diff(expected, got, protocmp.Transform()))
		})
	}
}

func readTestData(t *testing.T, testName, caseName string) []byte {
	data, err := os.ReadFile(path.Join("testdata", testName, caseName+".yaml"))
	require.NoError(t, err)
	return data
}
