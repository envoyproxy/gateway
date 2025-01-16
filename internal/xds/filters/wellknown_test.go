// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	"fmt"
	"os"
	"path"
	"strings"
	"testing"

	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

func TestGenerateCompressorFilter(t *testing.T) {
	cases := []struct {
		compressorType egv1a1.CompressorType
	}{
		{
			compressorType: egv1a1.BrotliCompressorType,
		},
		{
			compressorType: egv1a1.GzipCompressorType,
		},
	}

	for _, tc := range cases {
		t.Run(string(tc.compressorType), func(t *testing.T) {
			got, err := GenerateCompressorFilter(tc.compressorType)
			require.NoError(t, err)

			b := readTestData(t, strings.ToLower(fmt.Sprintf("%s.yaml", tc.compressorType)))
			expected := &hcm.HttpFilter{}
			err = proto.FromYAML(b, expected)
			require.NoError(t, err)

			require.Empty(t, cmp.Diff(expected, got, protocmp.Transform()))
		})
	}
}

func TestGenerateHealthCheckFilter(t *testing.T) {
	got, err := GenerateHealthCheckFilter("test_path")
	require.NoError(t, err)

	b := readTestData(t, "healthcheck.yaml")
	expected := &hcm.HttpFilter{}
	err = proto.FromYAML(b, expected)
	require.NoError(t, err)

	require.Empty(t, cmp.Diff(expected, got, protocmp.Transform()))
}

func readTestData(t *testing.T, filename string) []byte {
	data, err := os.ReadFile(path.Join("testdata", filename))
	require.NoError(t, err)
	return data
}
