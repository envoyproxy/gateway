// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// GzipCompressionLevel defines the compression level of zip compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=DEFAULT_COMPRESSION;BEST_SPEED;COMPRESSION_LEVEL_1;COMPRESSION_LEVEL_2;COMPRESSION_LEVEL_3;COMPRESSION_LEVEL_4;COMPRESSION_LEVEL_5;COMPRESSION_LEVEL_6;COMPRESSION_LEVEL_7;COMPRESSION_LEVEL_8;COMPRESSION_LEVEL_9;BEST_COMPRESSION
type GzipCompressionLevel string

// GzipCompressionStrategy defines the compression strategy of zip compressor library supported by Envoy Gateway.
//
// +kubebuilder:validation:Enum=DEFAULT_STRATEGY;FILTERED;HUFFMAN_ONLY;RLE;FIXED
type GzipCompressionStrategy string

type GzipCompressor struct {
	// Value from 1 to 9 that controls the amount of internal memory used by zlib. Higher
	// values use more memory, but are faster and produce better compression results. The default
	// value is 5
	//
	// +optional
	MemoryLevel *uint32 `json:"memoryLevel,omitempty"`

	// A value used for selecting the zlib compression level. This setting will affect speed and
	// amount of compression applied to the content. “BEST_COMPRESSION” provides higher compression
	// at the cost of higher latency and is equal to “COMPRESSION_LEVEL_9”. “BEST_SPEED” provides lower
	// compression with minimum impact on response time, the same as “COMPRESSION_LEVEL_1”. “DEFAULT_COMPRESSION”
	// provides an optimal result between speed and compression. According to zlib’s manual this level
	// gives the same result as “COMPRESSION_LEVEL_6”. This field will be set to “DEFAULT_COMPRESSION”
	// if not specified.
	//
	// +optional
	CompressionLevel GzipCompressionLevel `json:"compressionLevel,omitempty"`

	// A value used for selecting the zlib compression strategy which is directly related to the
	// characteristics of the content. Most of the time “DEFAULT_STRATEGY” will be the best choice,
	// which is also the default value for the parameter, though there are situations when changing
	// this parameter might produce better results. For example, run-length encoding (RLE) is typically
	// used when the content is known for having sequences which same data occurs many consecutive times.
	// For more information about each strategy, please refer to zlib manual.
	//
	// +optional
	CompressionStrategy GzipCompressionStrategy `json:"compressionStrategy,omitempty"`

	// Value from 9 to 15 that represents the base two logarithmic of the compressor’s window size.
	// Larger window results in better compression at the expense of memory usage. The default is 12
	// which will produce a 4096 bytes window. For more details about this parameter, please refer to
	// zlib manual > deflateInit2.
	//
	// +optional
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// Value for Zlib’s next output buffer. If not set, defaults to 4096
	//
	// +optional
	ChunkSize *uint32 `json:"chunkSize,omitempty"`
}

type GzipDecompressor struct {
	// Value from 9 to 15 that represents the base two logarithmic of the compressor’s window size.
	// Larger window results in better compression at the expense of memory usage. The default is 12
	// which will produce a 4096 bytes window. For more details about this parameter, please refer to
	// zlib manual > deflateInit2.
	//
	// +optional
	WindowBits *uint32 `json:"windowBits,omitempty"`

	// Value for zlib’s decompressor output buffer. If not set, defaults to 4096.
	//
	// +optional
	ChunkSize *uint32 `json:"chunkSize,omitempty"`

	// An upper bound to the number of times the output buffer is allowed to be bigger than the size
	// of the accumulated input. This value is used to prevent decompression bombs. If not set,
	// defaults to 100.
	//
	// +optional
	MaxInflateRatio *uint32 `json:"maxInflateRatio,omitempty"`
}

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

	// If true, disables compression when the response contains an etag header.
	// When it is false, weak etags will be preserve and remove the ones that require
	// strong validation.
	//
	// +optional
	DisableOnEtagHeader bool `json:"disableOnEtagHeader,omitempty"`

	// If true, removes accept-encoding from the request headers before dispatching
	// it to the upstream so that responses do not get compressed before reaching
	// the filter.
	//
	// +optional
	RemoveAcceptEncodingHeader bool `json:"removeAcceptEncodingHeader,omitempty"`

	// If true, chooses this compressor first to do compression when the q-values in
	// Accept-Encoding are same. The last compressor which enables choose_first will
	// be chosen if multiple compressor in the policy have choose_first as true.
	//
	// +optional
	ChooseFirst bool `json:"chooseFirst,omitempty"`

	// A compressor library to use for compression
	//
	// +required
	CompressorLibrary *CompressorLibrary `json:"compressorLibrary,omitempty"`
}

type Decompression struct {
	// A decompressor library to use for decompression
	//
	// +required
	DecompressiorLibrary *DecompressorLibrary `json:"decompressorLibrary,omitempty"`

	// The configuration for GZIP decompressor.
	//
	// +optional
	GzipDecompressor *GzipDecompressor `json:"gzipDecompressor,omitempty"`
}

type CompressorLibrary struct {
	// LibraryType defines which library want to use for compression.
	//
	// +required
	CompressorLibraryType CompressorLibraryType `json:"compressorLibraryType,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	GzipCompressor *GzipCompressor `json:"gzipCompressor,omitempty"`
}

type DecompressorLibrary struct {
	// LibraryType defines which library want to use for decompression.
	//
	// +required
	DecompressorLibraryType DecompressorLibraryType `json:"decompressorLibraryType,omitempty"`

	// The configuration for GZIP compressor.
	//
	// +optional
	GzipDecompressor *GzipDecompressor `json:"gzipDecompressor,omitempty"`
}
