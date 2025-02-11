// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// ClientStreamBufferSettings configures request and response buffering.
//
// +notImplementedHide
type ClientStreamBufferSettings struct {
	// MaxRequestBytes specifies the maximum allowed size for each incoming request.
	// If exceeded, the request will be rejected.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki").
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	MaxRequestBytes *resource.Quantity `json:"maxRequestBytes,omitempty"`

	// FileSystem configures filesystem-based buffering for request and response streams.
	//
	// +optional
	// +notImplementedHide
	FileSystem *FileSystemBuffers `json:"fileSystem,omitempty"`
}

// FileSystemBuffers configures filesystem-based buffering for HTTP streams.
//
// +notImplementedHide
type FileSystemBuffers struct {
	// Manager defines the configuration for the Envoy AsyncFileManager.
	// If unset and the behavior is not bypass in both directions, an Internal Server Error response will be sent.
	//
	// +optional
	// +notImplementedHide
	Manager *FileManagerConfig `json:"manager,omitempty"`

	// StoragePath specifies an optional directory for storing unlinked temporary files.
	// This determines the physical storage device used for buffering.
	//
	// If unset, the default is the TMPDIR environment variable. If TMPDIR is unset, it defaults to "/tmp".
	//
	// +optional
	// +notImplementedHide
	StoragePath *string `json:"storagePath,omitempty"`

	// Request defines buffering behavior for incoming request streams.
	//
	// +optional
	// +notImplementedHide
	Request *BufferStreamConfig `json:"request,omitempty"`

	// Response defines buffering behavior for outgoing response streams.
	//
	// +optional
	// +notImplementedHide
	Response *BufferStreamConfig `json:"response,omitempty"`
}

// FileManagerConfig configures the asynchronous file manager responsible for buffered I/O.
//
// +notImplementedHide
type FileManagerConfig struct {
	// ID  provides an optional unique identifier the file manager instance.
	//
	// +optional
	// +notImplementedHide
	ID *string `json:"id,omitempty"`

	// ThreadPool defines the thread pool configuration for the file manager.
	//
	// +notImplementedHide
	ThreadPool FileManagerThreadPool `json:"threadPool,omitempty"`
}

// FileManagerThreadPool configures the thread pool used by the asynchronous file manager.
//
// +notImplementedHide
type FileManagerThreadPool struct {
	// ThreadCount specifies the number of worker threads dedicated to file operations.
	//
	// If unset or zero, will default to the number of concurrent threads the hardware supports
	//
	// +optional
	// +notImplementedHide
	ThreadCount *int `json:"threadCount,omitempty"`
}

// BufferStreamConfig defines buffering behavior for a single HTTP stream direction (request or response).
//
// +notImplementedHide
type BufferStreamConfig struct {
	// Behavior specifies how the stream should be buffered and when data should be written.
	// Controls whether to bypass / stream / fully buffer / etc. If unset in route the default is stream_when_possible.
	//
	// +optional
	// +notImplementedHide
	Behavior *BufferStreamBehavior `json:"behavior,omitempty"`

	// MemoryBufferLimit defines the maximum amount of data stored in memory before buffering to disk.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki") and defaults to 1Mi.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	MemoryBufferLimit *resource.Quantity `json:"memoryBufferLimit,omitempty"`

	// StorageBufferLimit sets the maximum amount of data (excluding memory) that can be written
	// to the filesystem buffer before further writes are blocked.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki") and defaults to 32MiB.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	StorageBufferLimit *resource.Quantity `json:"storageBufferLimit,omitempty"`

	// StorageBufferQueueHighWatermark specifies the maximum amount that can be queued
	// for writing to storage, above which the source is requested to pause.
	//
	// Accepts values in resource.Quantity format (e.g., "10Mi", "500Ki") and defaults to the same value as memoryBufferLimit.
	//
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	// +optional
	// +notImplementedHide
	StorageBufferQueueHighWatermark *resource.Quantity `json:"storageBufferQueueHighWatermark,omitempty"`
}

// BufferStreamBehavior defines buffering behavior for an HTTP stream.
//
// +kubebuilder:validation:Enum=StreamWhenPossible;Bypass;InjectContentLengthIfNecessary;FullyBufferAndAlwaysInject;FullyBuffer
// +notImplementedHide
type BufferStreamBehavior string

const (
	// BufferStreamBehaviorStreamWhenPossible buffers only when output is slower than input.
	// Does not modify the Content-Length header.
	//
	// +notImplementedHide
	BufferStreamBehaviorStreamWhenPossible BufferStreamBehavior = "StreamWhenPossible"

	// BufferStreamBehaviorBypass disables buffering, effectively making this filter a no-op.
	//
	// +notImplementedHide
	BufferStreamBehaviorBypass BufferStreamBehavior = "Bypass"

	// BufferStreamBehaviorInjectContLenIfNecessary buffers the entire input only if the Content-Length
	// header is missing. If present, behaves like StreamWhenPossible.
	//
	// +notImplementedHide
	BufferStreamBehaviorInjectContLenIfNecessary BufferStreamBehavior = "InjectContentLengthIfNecessary"

	// BufferStreamBehaviorFullyBufferAndAlwaysInject buffers the entire input and overwrites any existing
	// Content-Length header with the correct value.
	//
	// +notImplementedHide
	BufferStreamBehaviorFullyBufferAndAlwaysInject BufferStreamBehavior = "FullyBufferAndAlwaysInject"

	// BufferStreamBehaviorFullyBuffer buffers the entire input but does not modify the Content-Length header.
	//
	// +notImplementedHide
	BufferStreamBehaviorFullyBuffer BufferStreamBehavior = "FullyBuffer"
)
