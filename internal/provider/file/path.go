// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/util/sets"
)

// getDirsAndFilesForWatcher prepares dirs and files for the watcher in notifier.
func getDirsAndFilesForWatcher(paths []string) (
	dirs sets.Set[string], files sets.Set[string], err error,
) {
	dirs, files = sets.New[string](), sets.New[string]()

	// Separate paths by whether is a directory or not.
	paths = sets.NewString(paths...).List()
	for _, path := range paths {
		var p os.FileInfo
		p, err = os.Lstat(path)
		if err != nil {
			return
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
