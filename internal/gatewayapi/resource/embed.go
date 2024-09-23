// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"bytes"
	_ "embed"
	"io"
	"io/fs"
	"time"
)

var (
	//go:embed crd/gateway-crds.yaml
	gatewayCRDs   []byte
	gatewayCRDsFS = memGatewayCRDsFS{}

	_ fs.FS          = memGatewayCRDsFS{}
	_ fs.ReadDirFile = memGatewayCRDsFile{}
	_ fs.FileInfo    = memGatewayCRDsFileInfo{}
	_ fs.DirEntry    = memGatewayCRDsDirEntry{}
)

// memGatewayCRDsFS is a mocked fs.FS for openapi to read gatewayCRDs from.
type memGatewayCRDsFS struct{}

func (m memGatewayCRDsFS) Open(_ string) (fs.File, error) {
	return &memGatewayCRDsFile{}, nil
}

// memGatewayCRDsFile is mocked fs.ReadDirFile for memGatewayCRDsFS.
type memGatewayCRDsFile struct{}

func (m memGatewayCRDsFile) Stat() (fs.FileInfo, error) {
	return &memGatewayCRDsFileInfo{}, nil
}

func (m memGatewayCRDsFile) Close() error {
	return nil
}

func (m memGatewayCRDsFile) Read(b []byte) (int, error) {
	fi, _ := m.Stat()
	if int64(len(b)) >= fi.Size() {
		return bytes.NewReader(gatewayCRDs).Read(b)
	}
	return 0, io.EOF
}

func (m memGatewayCRDsFile) ReadDir(_ int) ([]fs.DirEntry, error) {
	return []fs.DirEntry{&memGatewayCRDsDirEntry{}}, nil
}

// memGatewayCRDsDirEntry is a mocked fs.DirEntry for memGatewayCRDsFile.
type memGatewayCRDsDirEntry struct {
	memGatewayCRDsFileInfo
}

func (m memGatewayCRDsDirEntry) Type() fs.FileMode          { return 0o444 }
func (m memGatewayCRDsDirEntry) Info() (fs.FileInfo, error) { return &memGatewayCRDsFileInfo{}, nil }

// memGatewayCRDsFileInfo is a mocked fs.FileInfo for memGatewayCRDsFile.
type memGatewayCRDsFileInfo struct{}

func (m memGatewayCRDsFileInfo) Name() string       { return "gateway-crds.yaml" }
func (m memGatewayCRDsFileInfo) Size() int64        { return int64(len(gatewayCRDs)) }
func (m memGatewayCRDsFileInfo) Mode() fs.FileMode  { return 0o444 }
func (m memGatewayCRDsFileInfo) ModTime() time.Time { return time.Now() }
func (m memGatewayCRDsFileInfo) IsDir() bool        { return false }
func (m memGatewayCRDsFileInfo) Sys() any           { return nil }
