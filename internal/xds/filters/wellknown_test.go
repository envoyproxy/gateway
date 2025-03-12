// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package filters

import (
	"os"
	"path"
	"testing"

	hcm "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/testing/protocmp"

	"github.com/envoyproxy/gateway/internal/utils/proto"
)

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
