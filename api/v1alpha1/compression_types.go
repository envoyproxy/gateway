// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type GzipCompressor struct {
}

type Compression struct {
	// A compressor library to use for compression
	//
	// +required
	CompressorLibrary *CompressorLibrary `json:"compressorLibrary,omitempty"`
}

type CompressorLibrary struct {
	// LibraryType defines which library want to use for compression.
	//
	// +required
	CompressorLibraryType CompressorLibraryType `json:"compressorLibraryType,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	GzipCompressor *GzipCompressor `json:"gzipCompressor,omitempty"`
}
