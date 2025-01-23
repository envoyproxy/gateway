// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"k8s.io/apimachinery/pkg/util/sets"
)

// loadFromFilesAndDirs loads resources from specific files and directories.
// It returns a slice of resources loaded, as well as a set of all the subdirectories
// that were traversed recursively (including the provided directories).
func loadFromFilesAndDirs(files, dirs []string) ([]*resource.Resources, sets.Set[string], error) {
	rs := make([]*resource.Resources, 0)
	subDirs := sets.New[string]()

	for _, file := range files {
		r, err := loadFromFile(file)
		if err != nil {
			return nil, nil, err
		}
		rs = append(rs, r)
	}

	for _, dir := range dirs {
		r, s, err := loadFromDir(dir)
		if err != nil {
			return nil, nil, err
		}
		rs = append(rs, r...)
		subDirs = subDirs.Union(s)
	}

	return rs, subDirs, nil
}

// loadFromFile loads resources from a specific file.
func loadFromFile(path string) (*resource.Resources, error) {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s is not exist", path)
		}
		return nil, err
	}

	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	return resource.LoadResourcesFromYAMLBytes(bytes, false)
}

func loadFromDir(path string) ([]*resource.Resources, sets.Set[string], error) {
	rs := make([]*resource.Resources, 0)
	subDirs := sets.New[string]()

	err := traverseDirectory(path, &rs, &subDirs)
	if err != nil {
		return nil, nil, err
	}

	return rs, subDirs, nil
}

// traverseDirectory is a helper function that recursively traverses the directory
// and loads resources from all files while skipping hidden files and directories.
func traverseDirectory(dirPath string, rs *[]*resource.Resources, subDirs *sets.Set[string]) error {
	subDirs.Insert(dirPath)
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())

		// Skip hidden files and directories.
		if strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		if entry.IsDir() {
			// Recursively process subdirectories.
			if err := traverseDirectory(fullPath, rs, subDirs); err != nil {
				return err
			}
		} else {
			// Load resources from files.
			r, err := loadFromFile(fullPath)
			if err != nil {
				return err
			}
			*rs = append(*rs, r)
		}
	}

	return nil
}
