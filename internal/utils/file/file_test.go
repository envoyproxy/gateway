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

func TestWriteDirPermissions(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "certs")

	err := WriteDir([]byte("secret"), dir, "tls.key")
	require.NoError(t, err)

	fi, err := os.Stat(dir)
	require.NoError(t, err)
	require.Zero(t, fi.Mode().Perm()&0o007, "directory must not be accessible by others, got %v", fi.Mode().Perm())

	key := filepath.Join(dir, "tls.key")
	fi, err = os.Stat(key)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), fi.Mode().Perm())

	// A file left behind by an older release keeps its mode through O_TRUNC,
	// so overwriting an existing 0644 secret has to tighten it as well.
	require.NoError(t, os.Chmod(key, 0o644))
	require.NoError(t, WriteDir([]byte("secret"), dir, "tls.key"))

	fi, err = os.Stat(key)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0o600), fi.Mode().Perm())
}
