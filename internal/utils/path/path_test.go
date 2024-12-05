// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package path

import (
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListDirsAndFiles(t *testing.T) {
	basePath, _ := os.MkdirTemp(os.TempDir(), "list-test")
	defer func() {
		_ = os.RemoveAll(basePath)
	}()
	paths, err := os.MkdirTemp(basePath, "paths")
	require.NoError(t, err)
	dirPath, err := os.MkdirTemp(paths, "dir")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path.Join(paths, "foo"), []byte("foo"), 0o700))   // nolint: gosec
	require.NoError(t, os.WriteFile(path.Join(dirPath, "bar"), []byte("bar"), 0o700)) // nolint: gosec

	testCases := []struct {
		name        string
		paths       []string
		expectDirs  []string
		expectFiles []string
	}{
		{
			name: "get file and dir path",
			paths: []string{
				dirPath,
				path.Join(paths, "foo"),
			},
			expectDirs: []string{
				dirPath,
			},
			expectFiles: []string{
				path.Join(paths, "foo"),
			},
		},
		{
			name: "overlap file path will be ignored",
			paths: []string{
				dirPath, path.Join(dirPath, "bar"),
			},
			expectDirs: []string{
				dirPath,
			},
			expectFiles: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirs, files := ListDirsAndFiles(tc.paths)
			require.ElementsMatch(t, dirs.UnsortedList(), tc.expectDirs)
			require.ElementsMatch(t, files.UnsortedList(), tc.expectFiles)
		})
	}
}
