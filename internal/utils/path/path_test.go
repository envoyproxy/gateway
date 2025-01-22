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

func TestGetSubDirs(t *testing.T) {
	basePath, _ := os.MkdirTemp(os.TempDir(), "sub-dir-test")
	defer func() {
		_ = os.RemoveAll(basePath)
	}()

	// |- dir/
	// 	   |- sub1/
	// 	      |- sub1a/
	// 	   |- sub2/
	dirPath, err := os.MkdirTemp(basePath, "dir")
	require.NoError(t, err)
	subDir1, err := os.MkdirTemp(dirPath, "sub1")
	require.NoError(t, err)
	subDir2, err := os.MkdirTemp(dirPath, "sub2")
	require.NoError(t, err)
	subDir1A, err := os.MkdirTemp(subDir1, "sub1a")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		initDirs     []string
		expectSubDir []string
	}{
		{
			name:     "one level nesting",
			initDirs: []string{subDir1},
			expectSubDir: []string{
				subDir1,
				subDir1A,
			},
		},
		{
			name:         "no nested dir",
			initDirs:     []string{subDir2},
			expectSubDir: []string{subDir2},
		},
		{
			name:     "two level nesting",
			initDirs: []string{dirPath},
			expectSubDir: []string{
				dirPath,
				subDir1,
				subDir1A,
				subDir2,
			},
		},
		{
			name:     "overlapping directories",
			initDirs: []string{subDir1, dirPath},
			expectSubDir: []string{
				subDir1,
				subDir1A,
				dirPath,
				subDir2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subDirs := GetSubDirs(tc.initDirs)
			require.ElementsMatch(t, subDirs.UnsortedList(), tc.expectSubDir)
		})
	}
}

func TestGetParentDirs(t *testing.T) {
	aPaths := path.Join("a")
	bPaths := path.Join("a", "b")
	cPaths := path.Join("a", "b", "c")

	testCases := []struct {
		name             string
		paths            []string
		expectParentDirs []string
	}{
		{
			name: "all files",
			paths: []string{
				path.Join(cPaths, "foo"),
				path.Join(bPaths, "bar"),
			},
			expectParentDirs: []string{
				cPaths,
				bPaths,
			},
		},
		{
			name: "all dirs",
			paths: []string{
				bPaths + "/",
				cPaths + "/",
			},
			expectParentDirs: []string{
				bPaths,
				cPaths,
			},
		},
		{
			name: "mixed files and dirs",
			paths: []string{
				path.Join(cPaths, "foo"),
				path.Join(cPaths, "bar"),
				path.Join(bPaths, "foo"),
				path.Join(bPaths, "bar"),
				aPaths + "/",
				bPaths + "/",
				cPaths + "/",
			},
			expectParentDirs: []string{
				cPaths,
				bPaths,
				aPaths,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parents := GetParentDirs(tc.paths)
			require.ElementsMatch(t, parents.UnsortedList(), tc.expectParentDirs)
		})
	}
}
