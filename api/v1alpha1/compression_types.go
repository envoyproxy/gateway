// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// CompressorType defines the types of compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=Gzip
type CompressorType string

// GzipCompressor defines the config for the Gzip compressor. There are some configs
// available from the Envoy Proxy. Currently only use the default value for configs.
// For the default value for those configs can be reference here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/compression/gzip/compressor/v3/gzip.proto#extension-envoy-compression-gzip-compressor
type GzipCompressor struct {
}

// Compression defines the config of compression for the http streams. Currently
// only the minial config was added. All configs from the Envoy Proxy are using
// the default value. For the default value can be reference here:
// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/compressor/v3/compressor.proto#extensions-filters-http-compressor-v3-compressor
type Compression struct {
	// CompressorType defines which type compressor wants to use for compression.
	//
	// +required
	Type CompressorType `json:"type,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	Gzip *GzipCompressor `json:"gzip,omitempty"`
}
