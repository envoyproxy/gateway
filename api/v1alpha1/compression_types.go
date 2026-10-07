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
	// MemoryLevel controls the amount of internal memory used by zlib. Higher values use more
	// memory, but are faster and produce better compression results. Value must be in the range [1, 9].
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=9
	MemoryLevel *uint32 `json:"memoryLevel,omitempty"`

	// CompressionLevel selects the zlib compression level. Higher levels provide better compression
	// at the cost of increased latency and CPU usage. Value must be in the range [1, 9].
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=9
	CompressionLevel *uint32 `json:"compressionLevel,omitempty"`

	// WindowBits is the base two logarithmic of the compressor's window size. Larger window results
	// in better compression at the expense of memory usage. Value must be in the range [9, 15].
	//
	// +optional
	// +kubebuilder:validation:Minimum=9
	// +kubebuilder:validation:Maximum=15
	WindowBits *uint32 `json:"windowBits,omitempty"`
}

// BrotliCompressor defines the config for the Brotli compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/compressor/v3/brotli.proto#extension-envoy-compression-brotli-compressor
type BrotliCompressor struct {
	// Quality controls the main compression speed-density lever. The higher the quality, the slower
	// the compression and the denser the output. Value must be in the range [0, 11].
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=11
	Quality *uint32 `json:"quality,omitempty"`

	// WindowBits is the base two logarithmic of the compressor's window size. Larger window results
	// in better compression at the expense of memory usage. Value must be in the range [10, 24].
	//
	// +optional
	// +kubebuilder:validation:Minimum=10
	// +kubebuilder:validation:Maximum=24
	WindowBits *uint32 `json:"windowBits,omitempty"`
}

// ZstdCompressor defines the config for the Zstd compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/zstd/compressor/v3/zstd.proto#extension-envoy-compression-zstd-compressor
type ZstdCompressor struct {
	// CompressionLevel sets the compression parameters according to a pre-defined compression level
	// table. Higher levels provide better compression at the cost of increased latency and CPU usage.
	// Value 0 means the default level. Value must be in the range [0, 22].
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=22
	CompressionLevel *uint32 `json:"compressionLevel,omitempty"`
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
