// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gzip

// CompressionLevel defines the compression level of zip compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=DEFAULT_COMPRESSION;BEST_SPEED;COMPRESSION_LEVEL_1;COMPRESSION_LEVEL_2;COMPRESSION_LEVEL_3;COMPRESSION_LEVEL_4;COMPRESSION_LEVEL_5;COMPRESSION_LEVEL_6;COMPRESSION_LEVEL_7;COMPRESSION_LEVEL_8;COMPRESSION_LEVEL_9;BEST_COMPRESSION
type CompressionLevel string

// CompressionStrategy defines the compression strategy of zip compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=DEFAULT_STRATEGY;FILTERED;HUFFMAN_ONLY;RLE;FIXED
type CompressionStrategy string

type GzipCompressor struct {
	// Value from 1 to 9 that controls the amount of internal memory used by zlib. Higher
	// values use more memory, but are faster and produce better compression results. The default
	// value is 5
	//
	// +optional
	MemoryLevel *uint32 `json:"memoryLevel,omitempty"`

	// A value used for selecting the zlib compression level. This field will be set to “DEFAULT_COMPRESSION”
	// if not specified.
	//
	// +optional
	CompressionLevel CompressionLevel `json:"compressionLevel,omitempty"`

	// A value used for selecting the zlib compression strategy which is directly related to the characteristics of the content.
	//
	// +optional
	CompressionStrategy CompressionStrategy `json:"compressionStrategy,omitempty"`

	// Value from 9 to 15 that represents the base two logarithmic of the compressor’s window size.
	//
	// +optional
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// Value for Zlib’s next output buffer. If not set, defaults to 4096
	//
	// +optional
	ChunkSize *uint32 `json:"chunkSize,omitempty"`
}
