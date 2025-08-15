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
)

// loadFromFilesAndDirs loads resources from specific files and directories.
func loadFromFilesAndDirs(files, dirs []string) (*resource.LoadResources, error) {
	rs := resource.NewLoadResources()

	for _, file := range files {
		r, err := loadFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load resources from file %s: %w", file, err)
		}
		rs.Merge(r)
	}

	for _, dir := range dirs {
		r, err := loadFromDir(dir)
		if err != nil {
			return nil, err
		}
		rs.Merge(r)
	}

	return rs, nil
}

// loadFromFile loads resources from a specific file.
func loadFromFile(path string) (*resource.LoadResources, error) {
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

// loadFromDir loads resources from all the files under a specific directory excluding subdirectories.
func loadFromDir(path string) (*resource.LoadResources, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	rs := resource.NewLoadResources()
	for _, entry := range entries {
		// Ignoring subdirectories and all hidden files and directories.
		if entry.IsDir() || strings.HasPrefix(entry.Name(), ".") {
			continue
		}
		full := filepath.Join(path, entry.Name())
		r, err := loadFromFile(full)
		if err != nil {
			return nil, fmt.Errorf("failed to load resources from file %s: %w", full, err)
		}

		rs.Merge(r)
	}

	return rs, nil
}
