package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadFromDir(t *testing.T) {
	// Checks if the function loadFromDir returns the expected subdirectories.
	// TODO: Add test cases for resource loading.

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
		initDir      string
		expectSubDir []string
	}{
		{
			name:    "one level nesting",
			initDir: subDir1,
			expectSubDir: []string{
				subDir1,
				subDir1A,
			},
		},
		{
			name:         "no nested dir",
			initDir:      subDir2,
			expectSubDir: []string{subDir2},
		},
		{
			name:    "two level nesting",
			initDir: dirPath,
			expectSubDir: []string{
				dirPath,
				subDir1,
				subDir1A,
				subDir2,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, subDirs, err := loadFromDir(tc.initDir)
			require.NoError(t, err)
			require.ElementsMatch(t, subDirs.UnsortedList(), tc.expectSubDir)
		})
	}
}
