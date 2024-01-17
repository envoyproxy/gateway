// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

type Compression struct {
	// Minimum value of Content-Length header of request or response messages.
	// The default value is 30.
	//
	// + optional
	ContentLength *uint32 `json:"contentLength,omitempty"`

	// Set of strings that allows specifying which mime-types yield compression;
	// e.g., application/json, text/html, etc. When this field is not defined,
	// compression will be applied to the following mime-types: “application/javascript”,
	// “application/json”, “application/xhtml+xml”, “image/svg+xml”, “text/css”,
	// “text/html”, “text/plain”, “text/xml” and their synonyms.
	//
	// + optional
	ContentType []string `json:"contentType,omitempty"`

	// If true, removes accept-encoding from the request headers before dispatching
	// it to the upstream so that responses do not get compressed before reaching
	// the filter.
	//
	// +optional
	RemoveAcceptEncodingHeader bool `json:"removeAcceptEncodingHeader,omitempty"`

	// A compressor library to use for compression
	//
	// +required
	CompressorLibrary *CompressorLibrary `json:"compressorLibrary,omitempty"`
}

type Decompression struct {
	// A decompressor library to use for decompression
	//
	// +required
	DecompressiorLibrary *DecompressiorLibrary `json:"decompressorLibrary,omitempty"`
}

type CompressorLibrary struct {
	// LibraryType defines which library want to use for compression.
	//
	// +required
	CompressorLibraryType LibraryType `json:"compressorLibraryType,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	GzipCompressor *GzipCompressor `json:"gzipCompressor,omitempty"`
}

type DecompressorLibrary struct {
	// LibraryType defines which library want to use for decompression.
	//
	// +required
	DecompressorLibraryType LibraryType `json:"decompressorLibraryType,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	GzipDecompressor *GzipDecompressor `json:"gzipDecompressor,omitempty"`
}

type GzipDecompressor struct {
}
