// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package path

import (
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"
)

// ValidateOutputPath takes an output file path and returns it as an absolute path.
// It returns an error if the absolute path cannot be determined or if the parent directory does not exist.
func ValidateOutputPath(outputPath string) (string, error) {
	outputPath, err := filepath.Abs(outputPath)
	if err != nil {
		return "", err
	}
	if _, err := os.Stat(filepath.Dir(outputPath)); err != nil {
		return "", err
	}
	return outputPath, nil
}

// ListDirsAndFiles return a list of directories and files from a list of paths recursively.
func ListDirsAndFiles(paths []string) (dirs, files sets.Set[string]) {
	dirs, files = sets.New[string](), sets.New[string]()
	// Separate paths by whether is a directory or not.
	paths = sets.NewString(paths...).UnsortedList()
	for _, path := range paths {
		var p os.FileInfo
		p, err := os.Lstat(path)
		if err != nil {
			// skip
			continue
		}

		if p.IsDir() {
			dirs.Insert(path)
		} else {
			files.Insert(path)
		}
	}

	// Ignore filepath if its parent directory is also be watched.
	var ignoreFiles []string
	for fp := range files {
		if dirs.Has(filepath.Dir(fp)) {
			ignoreFiles = append(ignoreFiles, fp)
		}
	}
	files.Delete(ignoreFiles...)

	return
}

// GetParentDirs returns all the parent directories of given files.
func GetParentDirs(files []string) sets.Set[string] {
	parents := sets.New[string]()
	for _, f := range files {
		parents.Insert(filepath.Dir(f))
	}
	return parents
}
