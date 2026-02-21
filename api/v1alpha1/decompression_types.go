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
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Gzip' ? !has(self.brotli) && !has(self.zstd) : true",message="If decompressor type is Gzip, brotli and zstd fields must not be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Brotli' ? !has(self.gzip) && !has(self.zstd) : true",message="If decompressor type is Brotli, gzip and zstd fields must not be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Zstd' ? !has(self.gzip) && !has(self.brotli) : true",message="If decompressor type is Zstd, gzip and brotli fields must not be set."
type Decompression struct {
	// Type defines the decompressor type to use for decompression.
	//
	// +kubebuilder:validation:Enum=Gzip;Brotli;Zstd
	// +unionDiscriminator
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
