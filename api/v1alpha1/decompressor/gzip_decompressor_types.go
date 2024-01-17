// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type GzipDecompressor struct {
	// Value from 9 to 15 that represents the base two logarithmic of the decompressor’s window size.
	//
	// +optional
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// Value for zlib’s decompressor output buffer. If not set, defaults to 4096.
	//
	// +optional
	ChunkSize *uint32 `json:"chunkSize,omitempty"`

	// An upper bound to the number of times the output buffer is allowed to be bigger than the size of the accumulated input.
	//
	// +optional
	MaxInflateRatio *uint32 `json:maxInflateRatio,omitempty`
}