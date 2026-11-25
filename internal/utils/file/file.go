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

// Write writes data into a given filepath. The file is readable by the owner
// only, since callers use it for private keys and secrets.
func Write(data, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	defer file.Close()
	// The perm above only applies when the file is created, and O_TRUNC keeps
	// the mode of an existing inode, so tighten it before writing: a file left
	// behind by an older release can still be 0644.
	if err := file.Chmod(0o600); err != nil {
		return err
	}
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
	err := os.MkdirAll(dir, 0o750)
	if err != nil {
		return err
	}

	return Write(string(data), filepath.Join(dir, filename))
}
