// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDirsAndFilesForWatcher(t *testing.T) {
	testPath := path.Join("testdata", "paths")
	testCases := []struct {
		name        string
		paths       []string
		expectDirs  []string
		expectFiles []string
	}{
		{
			name: "get file and dir path",
			paths: []string{
				path.Join(testPath, "dir"), path.Join(testPath, "foo"),
			},
			expectDirs: []string{
				path.Join(testPath, "dir"),
			},
			expectFiles: []string{
				path.Join(testPath, "foo"),
			},
		},
		{
			name: "overlap file path will be ignored",
			paths: []string{
				path.Join(testPath, "dir"), path.Join(testPath, "dir", "bar"),
			},
			expectDirs: []string{
				path.Join(testPath, "dir"),
			},
			expectFiles: []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dirs, paths, _ := getDirsAndFilesForWatcher(tc.paths)
			require.ElementsMatch(t, dirs.UnsortedList(), tc.expectDirs)
			require.ElementsMatch(t, paths.UnsortedList(), tc.expectFiles)
		})
	}
}
