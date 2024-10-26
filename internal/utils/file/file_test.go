// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWriteDir(t *testing.T) {
	tmpDir := t.TempDir()
	testFilename := "test"
	data := []byte("foobar")

	err := WriteDir(data, tmpDir, testFilename)
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(tmpDir, testFilename))

	got, err := os.ReadFile(filepath.Join(tmpDir, testFilename))
	require.NoError(t, err)
	require.Equal(t, data, got)
}
