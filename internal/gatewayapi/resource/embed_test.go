// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"io/fs"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOpenAndReadGatewayCRDsFS(t *testing.T) {
	crds, err := gatewayCRDsFS.Open("")
	require.NoError(t, err)
	defer crds.Close()

	buf := make([]byte, len(gatewayCRDs))
	_, err = crds.Read(buf)
	require.NoError(t, err)

	expect, err := os.ReadFile(path.Join("crd", "gateway-crds.yaml"))
	require.NoError(t, err)

	require.Equal(t, expect, buf)
}

func TestReadGatewayCRDsDirFS(t *testing.T) {
	dirEntries, err := fs.ReadDir(gatewayCRDsFS, ".")
	require.NoError(t, err)
	require.Len(t, dirEntries, 1)

	dirEntry := dirEntries[0]
	fileInfo, err := dirEntry.Info()
	require.NoError(t, err)
	require.Equal(t, "gateway-crds.yaml", fileInfo.Name())
	require.False(t, fileInfo.IsDir())

	fileBytes, err := fs.ReadFile(gatewayCRDsFS, fileInfo.Name())
	require.NoError(t, err)
	require.Equal(t, fileInfo.Size(), int64(len(fileBytes)))
}
