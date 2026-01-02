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
type GzipCompressor struct{}

// BrotliCompressor defines the config for the Brotli compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/compressor/v3/brotli.proto#extension-envoy-compression-brotli-compressor
type BrotliCompressor struct{}

// ZstdCompressor defines the config for the Zstd compressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/zstd/compressor/v3/zstd.proto#extension-envoy-compression-zstd-compressor
type ZstdCompressor struct{}

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
