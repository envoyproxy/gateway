// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import "k8s.io/apimachinery/pkg/api/resource"

// CompressorType defines the types of compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Gzip;Brotli;Zstd
type CompressorType string

const (
	GzipCompressorType CompressorType = "Gzip"

	BrotliCompressorType CompressorType = "Brotli"

	ZstdCompressorType CompressorType = "Zstd"
)

// GzipCompressor defines the config for the Gzip compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/compressor/v3/gzip.proto#extension-envoy-compression-gzip-compressor
type GzipCompressor struct {
	// MemoryLevel controls the amount of internal memory used by zlib. Higher values
	// use more memory, but are faster and produce better compression results.
	// Valid values are 1-9. Default is 5.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=9
	MemoryLevel *uint32 `json:"memoryLevel,omitempty"`

	// CompressionLevel controls the speed and amount of compression.
	// Valid values: DEFAULT_COMPRESSION, BEST_SPEED, BEST_COMPRESSION, or COMPRESSION_LEVEL_1 through COMPRESSION_LEVEL_9.
	// Default is DEFAULT_COMPRESSION.
	//
	// +optional
	// +kubebuilder:validation:Enum=DEFAULT_COMPRESSION;BEST_SPEED;BEST_COMPRESSION;COMPRESSION_LEVEL_1;COMPRESSION_LEVEL_2;COMPRESSION_LEVEL_3;COMPRESSION_LEVEL_4;COMPRESSION_LEVEL_5;COMPRESSION_LEVEL_6;COMPRESSION_LEVEL_7;COMPRESSION_LEVEL_8;COMPRESSION_LEVEL_9
	CompressionLevel *string `json:"compressionLevel,omitempty"`

	// CompressionStrategy selects the zlib compression strategy.
	// Valid values: DEFAULT_STRATEGY, FILTERED, HUFFMAN_ONLY, RLE, FIXED.
	// Default is DEFAULT_STRATEGY.
	//
	// +optional
	// +kubebuilder:validation:Enum=DEFAULT_STRATEGY;FILTERED;HUFFMAN_ONLY;RLE;FIXED
	CompressionStrategy *string `json:"compressionStrategy,omitempty"`

	// WindowBits represents the base two logarithmic of the compressor's window size.
	// Larger window results in better compression at the expense of memory usage.
	// Valid values are 9-15. Default is 12 (4096 bytes window).
	//
	// +optional
	// +kubebuilder:validation:Minimum=9
	// +kubebuilder:validation:Maximum=15
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// ChunkSize is the size of the output buffer.
	// Default is 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	ChunkSize *uint32 `json:"chunkSize,omitempty"`
}

// BrotliCompressor defines the config for the Brotli compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/compressor/v3/brotli.proto#extension-envoy-compression-brotli-compressor
type BrotliCompressor struct {
	// Quality controls the compression speed-density lever.
	// Higher quality means slower compression but better compression ratio.
	// Valid values are 0-11. Default is 3.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=11
	Quality *uint32 `json:"quality,omitempty"`

	// EncoderMode tunes encoder for specific input.
	// Valid values: DEFAULT, GENERIC, TEXT, FONT.
	// Default is DEFAULT.
	//
	// +optional
	// +kubebuilder:validation:Enum=DEFAULT;GENERIC;TEXT;FONT
	EncoderMode *string `json:"encoderMode,omitempty"`

	// WindowBits represents the base two logarithmic of the compressor's window size.
	// Larger window results in better compression at the expense of memory usage.
	// Valid values are 10-24. Default is 18.
	//
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=24
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// InputBlockBits represents the base two logarithmic of the compressor's input block size.
	// Larger input block results in better compression at the expense of memory usage.
	// Valid values are 16-24. Default is 24.
	//
	// +optional
	// +kubebuilder:validation:Minimum=16
	// +kubebuilder:validation:Maximum=24
	InputBlockBits *uint32 `json:"inputBlockBits,omitempty"`

	// ChunkSize is the size of the output buffer.
	// Default is 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	ChunkSize *uint32 `json:"chunkSize,omitempty"`

	// DisableLiteralContextModeling disables literal context modeling format feature.
	// This is a decoding-speed vs compression ratio trade-off.
	// Default is false.
	//
	// +optional
	DisableLiteralContextModeling *bool `json:"disableLiteralContextModeling,omitempty"`
}

// ZstdCompressor defines the config for the Zstd compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/zstd/compressor/v3/zstd.proto#extension-envoy-compression-zstd-compressor
type ZstdCompressor struct {
	// CompressionLevel sets compression parameters according to pre-defined compression level table.
	// Valid values start from 1. Default is 3.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	CompressionLevel *uint32 `json:"compressionLevel,omitempty"`

	// EnableChecksum controls whether a 32-bit checksum is written at end of frame.
	// Default is false.
	//
	// +optional
	EnableChecksum *bool `json:"enableChecksum,omitempty"`

	// Strategy affects compression ratio and speed. Higher values result in stronger but slower compression.
	// Valid values: DEFAULT, FAST, DFAST, GREEDY, LAZY, LAZY2, BTLAZY2, BTOPT, BTULTRA, BTULTRA2.
	// Default is DEFAULT.
	//
	// +optional
	// +kubebuilder:validation:Enum=DEFAULT;FAST;DFAST;GREEDY;LAZY;LAZY2;BTLAZY2;BTOPT;BTULTRA;BTULTRA2
	Strategy *string `json:"strategy,omitempty"`

	// ChunkSize is the size of the output buffer.
	// Default is 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	ChunkSize *uint32 `json:"chunkSize,omitempty"`
}

// Compression defines the config of enabling compression.
// This can help reduce the bandwidth at the expense of higher CPU.
type Compression struct {
	// CompressorType defines the compressor type to use for compression.
	//
	// +required
	Type CompressorType `json:"type"`

	// The configuration for Brotli compressor.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Brotli *BrotliCompressor `json:"brotli,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Gzip *GzipCompressor `json:"gzip,omitempty"`

	// The configuration for Zstd compressor.
	//
	// +optional
	// +kubebuilder:pruning:PreserveUnknownFields
	Zstd *ZstdCompressor `json:"zstd,omitempty"`

	// MinContentLength defines the minimum response size in bytes to apply compression.
	// Responses smaller than this threshold will not be compressed.
	// Must be at least 30 bytes as enforced by Envoy Proxy.
	// Note that when the suffix is not provided, the value is interpreted as bytes.
	// Default: 30 bytes
	//
	// +optional
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	MinContentLength *resource.Quantity `json:"minContentLength,omitempty"`
}
