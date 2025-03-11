// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"os"
	"path/filepath"
	"testing"
)

func Test_loadFromFilesAndDirs(t *testing.T) {
	t.Run("non-existent file", func(t *testing.T) {
		_, err := loadFromFilesAndDirs([]string{"non-existent-file"}, nil)
		require.ErrorContains(t, err, "file non-existent-file is not exist")
	})
	t.Run("invalid content in a file", func(t *testing.T) {
		tmpfile := t.TempDir() + "/invalid.yaml"
		err := os.WriteFile(tmpfile, []byte("invalid"), 0644)
		require.NoError(t, err)
		_, err = loadFromFilesAndDirs([]string{tmpfile}, nil)
		require.ErrorContains(t, err,
			fmt.Sprintf("failed to load resources from file %s", tmpfile))
	})
	t.Run("non-existent directory", func(t *testing.T) {
		_, err := loadFromFilesAndDirs(nil, []string{"non-existent-directory"})
		require.ErrorContains(t, err, "no such file or directory")
	})
	t.Run("invalid content in a file in a directory", func(t *testing.T) {
		tmpdir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpdir, "invalid.yaml"), []byte("invalid"), 0644)
		require.NoError(t, err)
		_, err = loadFromFilesAndDirs(nil, []string{tmpdir})
		require.ErrorContains(t, err,
			fmt.Sprintf("failed to load resources from file %s", filepath.Join(tmpdir, "invalid.yaml")))
	})
	t.Run("ok", func(t *testing.T) {
		tmpfile := t.TempDir() + "/valid.yaml"
		err := os.WriteFile(tmpfile, []byte(`
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: aigw-run
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
`), 0644)
		require.NoError(t, err)
		rs, err := loadFromFilesAndDirs([]string{tmpfile}, nil)
		require.NoError(t, err)
		require.Len(t, rs, 1)
	})
}

func Test_loadFromDir(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		_, err := loadFromDir("non-existent-directory")
		require.ErrorContains(t, err, "no such file or directory")
	})
	t.Run("invalid content in a file", func(t *testing.T) {
		tmpdir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpdir, "invalid.yaml"), []byte("invalid"), 0644)
		require.NoError(t, err)
		_, err = loadFromDir(tmpdir)
		require.ErrorContains(t, err,
			fmt.Sprintf("failed to load resources from file %s", filepath.Join(tmpdir, "invalid.yaml")))
	})
	t.Run("ok", func(t *testing.T) {
		tmpdir := t.TempDir()
		err := os.WriteFile(filepath.Join(tmpdir, "valid.yaml"), []byte(`
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: aigw-run
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
`), 0644)
		require.NoError(t, err)
		rs, err := loadFromDir(tmpdir)
		require.NoError(t, err)
		require.Len(t, rs, 1)
	})
}
