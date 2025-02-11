// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/api/resource"
)

// ClientStreamBufferSettings allows users to configure request and response buffering
//
// +notImplementedHide
type ClientStreamBufferSettings struct {
	// +kubebuilder:validation:XIntOrString
	// +kubebuilder:validation:Pattern="^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$"
	//
	// +optional
	// +notImplementedHide
	// MaxRequestBytes provides configuration for the maximum request size for each incoming request.
	MaxRequestBytes *resource.Quantity `json:"maxRequestBytes,omitempty"`

	FileSystem FileSystemBuffers `json:"fileSystem,omitempty"`
}

// FileSystemBuffers allows users to configure a file system buffer http filter
//
// +notImplementedHide
type FileSystemBuffers struct {
	// Manger provides the AsyncFileManager configuration
	//
	// +notImplementedHide
	Manager FileManagerConfig `json:"manager,omitempty"`
	// StoragePath is an optional path to which the unlinked files should be written - this may determine which physical storage device will be used.
	// If unset in route, vhost and listener, will use the environment variable TMPDIR, or, if that’s also unset, will use /tmp.
	//
	// +notImplementedHide
	StoragePath string `json:"storagePath,omitempty"`
	// Request provides the request stream configuration
	//
	// +notImplementedHide
	Request BufferStreamConfig `json:"request,omitempty"`
	// Request provides the request stream configuration
	//
	// +notImplementedHide
	Response BufferStreamConfig `json:"response,omitempty"`
}

// FileManagerConfig allows a user to configure an AsyncFileManager
//
// +notImplementedHide
type FileManagerConfig struct {
	ID         string                `json:"id,omitempty"`
	ThreadPool FileManagerThreadPool `json:"threadPool,omitempty"`
}

// FileManagerThreadPool is the user configuration for a thread-pool based async file manager.
//
// +notImplementedHide
type FileManagerThreadPool struct {
	ThreadCount int `json:"threadCount,omitempty"`
}

// BufferStreamConfig is the stream configuration for one direction of the filter behavior
//
// +notImplementedHide
type BufferStreamConfig struct {
	Behavior                        BufferStreamBehavior `json:"behavior,omitempty"`
	MemoryBufferLimit               int64                `json:"memoryBufferLimit,omitempty"`
	StorageBufferLimit              int64                `json:"storageBufferLimit,omitempty"`
	StorageBufferQueueHighWatermark int64                `json:"storageBufferQueueHighWatermark,omitempty"`
}

// BufferStreamBehavior configures the behavior of the filter for a stream
//
// +kubebuilder:validation:Enum=StreamWhenPossible;Bypass;InjectContentLengthIfNecessary;FullyBufferAndAlwaysInject;FullyBuffer
// +notImplementedHide
type BufferStreamBehavior string

const (
	// BufferStreamBehaviorStreamWhenPossible StreamWhenPossible will not inject content-length header. Output immediately, buffer only if output is slower than input.
	//
	// +notImplementedHide
	BufferStreamBehaviorStreamWhenPossible BufferStreamBehavior = "StreamWhenPossible"
	// BufferStreamBehaviorBypass Bypass will never buffer, effectively do nothing.
	//
	// +notImplementedHide
	BufferStreamBehaviorBypass BufferStreamBehavior = "Bypass"
	// BufferStreamBehaviorInjectContLenIfNecessary If content-length is not present, buffer the entire input,
	// inject content-length header, then output. If content-length is already present, act like stream_when_possible.
	//
	// +notImplementedHide
	BufferStreamBehaviorInjectContLenIfNecessary BufferStreamBehavior = "InjectContentLengthIfNecessary"
	// BufferStreamBehaviorFullyBufferAndAlwaysInject Always buffer the entire input, and inject content-length, overwriting any provided content-length header.
	//
	// +notImplementedHide
	BufferStreamBehaviorFullyBufferAndAlwaysInject BufferStreamBehavior = "FullyBufferAndAlwaysInject"
	// BufferStreamBehaviorFullyBuffer Always buffer the entire input, do not modify content-length.
	//
	// +notImplementedHide
	BufferStreamBehaviorFullyBuffer BufferStreamBehavior = "FullyBuffer"
)
