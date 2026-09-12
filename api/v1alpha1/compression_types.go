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

// GzipCompressionStrategy defines the zlib compression strategies supported by the Gzip compressor.
//
// +kubebuilder:validation:Enum=Default;Filtered;HuffmanOnly;RLE;Fixed
type GzipCompressionStrategy string

const (
	// GzipCompressionStrategyDefault is the default zlib compression strategy,
	// suitable for most of the content.
	GzipCompressionStrategyDefault GzipCompressionStrategy = "Default"

	// GzipCompressionStrategyFiltered is a compression strategy for data produced
	// by a filter or predictor.
	GzipCompressionStrategyFiltered GzipCompressionStrategy = "Filtered"

	// GzipCompressionStrategyHuffmanOnly is a compression strategy that only uses
	// Huffman encoding.
	GzipCompressionStrategyHuffmanOnly GzipCompressionStrategy = "HuffmanOnly"

	// GzipCompressionStrategyRLE is a compression strategy that limits match
	// distances to one (run-length encoding), designed for image data.
	GzipCompressionStrategyRLE GzipCompressionStrategy = "RLE"

	// GzipCompressionStrategyFixed is a compression strategy that prevents the
	// use of dynamic Huffman codes.
	GzipCompressionStrategyFixed GzipCompressionStrategy = "Fixed"
)

// GzipCompressor defines the config for the Gzip compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/compressor/v3/gzip.proto#extension-envoy-compression-gzip-compressor
type GzipCompressor struct {
	// CompressionLevel sets the zlib compression level, from 1 (fastest compression,
	// lowest compression ratio) to 9 (slowest compression, best compression ratio).
	// If not set, Envoy uses the zlib default level, which is equivalent to level 6.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=9
	CompressionLevel *uint32 `json:"compressionLevel,omitempty"`

	// CompressionStrategy selects the zlib compression strategy, which is directly
	// related to the characteristics of the content. Most of the time "Default" is
	// the best choice. If not set, defaults to Default.
	//
	// +optional
	CompressionStrategy *GzipCompressionStrategy `json:"compressionStrategy,omitempty"`

	// MemoryLevel controls the amount of internal memory used by zlib, from 1 to 9.
	// Higher values use more memory, but are faster and produce better compression
	// results. If not set, defaults to 5.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=9
	MemoryLevel *uint32 `json:"memoryLevel,omitempty"`

	// WindowBits represents the base two logarithm of the compressor's window size,
	// from 9 to 15. Larger window results in better compression at the expense of
	// memory usage. If not set, defaults to 12, which will produce a 4096 bytes window.
	//
	// +optional
	// +kubebuilder:validation:Minimum=9
	// +kubebuilder:validation:Maximum=15
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// ChunkSize is the size in bytes of zlib's next output buffer, from 4096 to 65536.
	// If not set, defaults to 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=4096
	// +kubebuilder:validation:Maximum=65536
	ChunkSize *uint32 `json:"chunkSize,omitempty"`
}

// BrotliEncoderMode defines the modes to tune the Brotli encoder for a specific input.
//
// +kubebuilder:validation:Enum=Default;Generic;Text;Font
type BrotliEncoderMode string

const (
	// BrotliEncoderModeDefault is the default encoder mode.
	BrotliEncoderModeDefault BrotliEncoderMode = "Default"

	// BrotliEncoderModeGeneric tunes the encoder for any input.
	BrotliEncoderModeGeneric BrotliEncoderMode = "Generic"

	// BrotliEncoderModeText tunes the encoder for UTF-8 formatted text input.
	BrotliEncoderModeText BrotliEncoderMode = "Text"

	// BrotliEncoderModeFont tunes the encoder for WOFF 2.0 font input.
	BrotliEncoderModeFont BrotliEncoderMode = "Font"
)

// BrotliCompressor defines the config for the Brotli compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/compressor/v3/brotli.proto#extension-envoy-compression-brotli-compressor
type BrotliCompressor struct {
	// Quality controls the main compression speed-density lever, from 0 to 11.
	// The higher the quality, the slower the compression. If not set, defaults to 3.
	//
	// +optional
	// +kubebuilder:validation:Maximum=11
	Quality *uint32 `json:"quality,omitempty"`

	// EncoderMode tunes the encoder for a specific input.
	// If not set, defaults to Default.
	//
	// +optional
	EncoderMode *BrotliEncoderMode `json:"encoderMode,omitempty"`

	// WindowBits represents the base two logarithm of the compressor's window size,
	// from 10 to 24. Larger window results in better compression at the expense of
	// memory usage. If not set, defaults to 18.
	//
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=24
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// InputBlockBits represents the base two logarithm of the compressor's input block
	// size, from 16 to 24. Larger input block results in better compression at the
	// expense of memory usage. If not set, defaults to 24.
	//
	// +optional
	// +kubebuilder:validation:Minimum=16
	// +kubebuilder:validation:Maximum=24
	InputBlockBits *uint32 `json:"inputBlockBits,omitempty"`

	// ChunkSize is the size in bytes of the compressor's next output buffer, from
	// 4096 to 65536. If not set, defaults to 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=4096
	// +kubebuilder:validation:Maximum=65536
	ChunkSize *uint32 `json:"chunkSize,omitempty"`

	// DisableLiteralContextModeling disables the "literal context modeling" format
	// feature. This flag is a "decoding-speed vs compression ratio" trade-off.
	// If not set, defaults to false.
	//
	// +optional
	DisableLiteralContextModeling *bool `json:"disableLiteralContextModeling,omitempty"`
}

// ZstdCompressionStrategy defines the compression strategies supported by the Zstd compressor.
// The higher the value of the selected strategy, the more complex it is, resulting in stronger
// and slower compression.
//
// +kubebuilder:validation:Enum=Default;Fast;DFast;Greedy;Lazy;Lazy2;BTLazy2;BTOpt;BTUltra;BTUltra2
type ZstdCompressionStrategy string

const (
	ZstdCompressionStrategyDefault  ZstdCompressionStrategy = "Default"
	ZstdCompressionStrategyFast     ZstdCompressionStrategy = "Fast"
	ZstdCompressionStrategyDFast    ZstdCompressionStrategy = "DFast"
	ZstdCompressionStrategyGreedy   ZstdCompressionStrategy = "Greedy"
	ZstdCompressionStrategyLazy     ZstdCompressionStrategy = "Lazy"
	ZstdCompressionStrategyLazy2    ZstdCompressionStrategy = "Lazy2"
	ZstdCompressionStrategyBTLazy2  ZstdCompressionStrategy = "BTLazy2"
	ZstdCompressionStrategyBTOpt    ZstdCompressionStrategy = "BTOpt"
	ZstdCompressionStrategyBTUltra  ZstdCompressionStrategy = "BTUltra"
	ZstdCompressionStrategyBTUltra2 ZstdCompressionStrategy = "BTUltra2"
)

// ZstdCompressor defines the config for the Zstd compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/zstd/compressor/v3/zstd.proto#extension-envoy-compression-zstd-compressor
type ZstdCompressor struct {
	// CompressionLevel sets the compression parameters according to the pre-defined
	// zstd compression level table, from 1 to 22. Note that the exact compression
	// parameters are dynamically determined, depending on both compression level and
	// source content size (when known). If not set, defaults to 3.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=22
	CompressionLevel *uint32 `json:"compressionLevel,omitempty"`

	// EnableChecksum specifies whether a 32-bits checksum of the content is written
	// at the end of the frame. If not set, defaults to false.
	//
	// +optional
	EnableChecksum *bool `json:"enableChecksum,omitempty"`

	// Strategy selects the zstd compression strategy. The higher the value of the
	// selected strategy, the more complex it is, resulting in stronger and slower
	// compression. If not set, defaults to Default.
	//
	// +optional
	Strategy *ZstdCompressionStrategy `json:"strategy,omitempty"`

	// ChunkSize is the size in bytes of the compressor's next output buffer, from
	// 4096 to 65536. If not set, defaults to 4096.
	//
	// +optional
	// +kubebuilder:validation:Minimum=4096
	// +kubebuilder:validation:Maximum=65536
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
	Brotli *BrotliCompressor `json:"brotli,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	Gzip *GzipCompressor `json:"gzip,omitempty"`

	// The configuration for Zstd compressor.
	//
	// +optional
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
