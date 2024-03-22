// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// Wasm defines a wasm extension.
type Wasm struct {
	// Name is a unique name for this Wasm extension. It is used to identify the
	// Wasm extension if multiple extensions are handled by the same vm_id and root_id.
	// It's also used for logging/debugging.
	Name string `json:"name"`

	// VmID is an ID that will be used along with a hash of the wasm code to
	// determine which VM will be used to load the Wasm extension. All extensions
	// that have the same vm_id and code will use the same VM.
	//
	// Note that sharing a VM between plugins can reduce memory utilization and
	// make sharing of data easier, but it may have security implications.
	VmID *string `json:"vmID,omitempty"`

	// RootID is a unique ID for a set of extensions in a VM which will share a
	// RootContext and Contexts if applicable (e.g., an Wasm HttpFilter and an Wasm AccessLog).
	// If left blank, all extensions with a blank root_id with the same vm_id will share Context(s).
	RootID *string `json:"rootID,omitempty"`

	// Code is the wasm code for the extension.
	Code WasmCodeSource `json:"code"`

	// Config is the configuration for the Wasm extension.
	// This configuration will be passed as a JSON string to the Wasm extension.
	Config *apiextensionsv1.JSON `json:"config"`

	// FailOpen is a switch used to control the behavior when a fatal error occurs
	// during the initialization or the execution of the Wasm extension.
	// If FailOpen is set to true, the system bypasses the Wasm extension and
	// allows the traffic to pass through. Otherwise, if it is set to false or
	// not set (defaulting to false), the system blocks the traffic and returns
	// an HTTP 5xx error.
	//
	// +optional
	// +kubebuilder:default=false
	FailOpen *bool `json:"failOpen,omitempty"`

	// InsertBefore is the name of the filter that this Wasm extension should be
	// inserted before.
	// If the specified filter is not found in the filter chain, this Wasm extension
	// will be inserted before the next filter found in the chain, if any. If no
	// any other filters are found in the chain, this Wasm extension will be
	// inserted before the router filter.
	//
	// If not specified, this Wasm extension will be inserted before the router filter.
	//InsertBefore *EnvoyFilter `json:"insertBeforeFilter"`
}

// WasmCodeSource defines the source of the wasm code.
type WasmCodeSource struct {
	// ConfigMap is the name of the ConfigMap containing the wasm code.
	//
	// The key in the ConfigMap should be the name of the Wasm. For example,
	// if the Wasm is named "my-wasm-extension", the ConfigMap should have a key
	// named "my-wasm-extension" and the value should be the wasm code.
	ConfigMap *string `json:"ConfigMap,omitempty"`

	// HTTP is the HTTP URL containing the wasm code.
	//
	// Note that the HTTP server must be accessible from the Envoy proxy.
	HTTP *string `json:"http,omitempty"`

	// Image is the OCI image containing the wasm code.
	// Image *string `json:"image,omitempty"` //TODO: Add support for OCI image in the future.
}
