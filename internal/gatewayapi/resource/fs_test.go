// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	envoygateway "github.com/envoyproxy/gateway"
)

func TestOpenAndReadGatewayCRDsFS(t *testing.T) {
	crds, err := gatewayCRDsFS.Open("")
	require.NoError(t, err)
	defer crds.Close()

	buf := make([]byte, len(envoygateway.GatewayCRDs))
	cur, err := crds.Read(buf)
	require.NoError(t, err)
	require.Positive(t, cur)
}

func TestReadGatewayCRDsDirFS(t *testing.T) {
	dirEntries, err := fs.ReadDir(gatewayCRDsFS, ".")
	require.NoError(t, err)
	require.Len(t, dirEntries, 1)

	dirEntry := dirEntries[0]
	require.Equal(t, fs.FileMode(0o444), dirEntry.Type())

	fileInfo, err := dirEntry.Info()
	require.NoError(t, err)
	require.Equal(t, "gateway-crds.yaml", fileInfo.Name())
	require.NotNil(t, fileInfo.ModTime())
	require.Nil(t, fileInfo.Sys())
	require.False(t, fileInfo.IsDir())

	fileBytes, err := fs.ReadFile(gatewayCRDsFS, fileInfo.Name())
	require.NoError(t, err)
	require.Equal(t, fileInfo.Size(), int64(len(fileBytes)))
}
