// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"bufio"
	"os"
	"path/filepath"
)

// Write writes data into a given filepath.
func Write(data, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer file.Close()
	write := bufio.NewWriter(file)
	_, err = write.WriteString(data)
	if err != nil {
		return err
	}
	write.Flush()

	return nil
}

// WriteDir write data into a given filename under certain directory.
func WriteDir(data []byte, dir, filename string) error {
	err := os.MkdirAll(dir, 0o755)
	if err != nil {
		return err
	}

	return Write(string(data), filepath.Join(dir, filename))
}
