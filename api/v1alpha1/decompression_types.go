// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// DecompressorType defines the types of decompressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Gzip;Brotli;Zstd
type DecompressorType string

const (
	GzipDecompressorType DecompressorType = "Gzip"

	BrotliDecompressorType DecompressorType = "Brotli"

	ZstdDecompressorType DecompressorType = "Zstd"
)

// GzipDecompressor defines the config for the Gzip decompressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/decompressor/v3/gzip.proto#extension-envoy-compression-gzip-decompressor
type GzipDecompressor struct{}

// BrotliDecompressor defines the config for the Brotli decompressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/brotli/decompressor/v3/brotli.proto#extension-envoy-compression-brotli-decompressor
type BrotliDecompressor struct{}

// ZstdDecompressor defines the config for the Zstd decompressor.
// The default values can be found here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/zstd/decompressor/v3/zstd.proto#extension-envoy-compression-zstd-decompressor
type ZstdDecompressor struct{}

// Decompression defines the config of enabling decompression.
// This can help decompress compressed requests from clients and/or compressed responses from backends.
type Decompression struct {
	// DecompressorType defines the decompressor type to use for decompression.
	//
	// +required
	Type DecompressorType `json:"type"`

	// The configuration for Brotli decompressor.
	//
	// +optional
	Brotli *BrotliDecompressor `json:"brotli,omitempty"`

	// The configuration for GZIP decompressor.
	//
	// +optional
	Gzip *GzipDecompressor `json:"gzip,omitempty"`

	// The configuration for Zstd decompressor.
	//
	// +optional
	Zstd *ZstdDecompressor `json:"zstd,omitempty"`
}
