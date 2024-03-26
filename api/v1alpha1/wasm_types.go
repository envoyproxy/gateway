// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

// Wasm defines a wasm extension.
//
// Note: at the moment, Envoy Gateway does not support configuring Wasm runtime.
// v8 is used as the VM runtime for the Wasm extensions.
type Wasm struct {
	// Name is a unique name for this Wasm extension. It is used to identify the
	// Wasm extension if multiple extensions are handled by the same vm_id and root_id.
	// It's also used for logging/debugging.
	Name string `json:"name"`

	// VMID is an ID that will be used along with a hash of the wasm code to
	// determine which VM will be used to load the Wasm extension. All extensions
	// that have the same vm_id and code will use the same VM.
	//
	// Note that sharing a VM between plugins can reduce memory utilization and
	// make sharing of data easier, but it may have security implications.
	VMID *string `json:"vmID,omitempty"`

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
	// InsertBefore *EnvoyFilter `json:"insertBeforeFilter"`
}

// WasmCodeSource defines the source of the wasm code.
// TODO: zhaohuabing CEL validation" one of the HTTP or Image field must be set
type WasmCodeSource struct {
	// HTTP is the HTTP URL containing the wasm code.
	//
	// Note that the HTTP server must be accessible from the Envoy proxy.
	// +optional
	HTTP *string `json:"http,omitempty"`

	// Image is the OCI image containing the wasm code.
	//
	// Note that the image must be accessible from the Envoy Gateway.
	// +optional
	Image *WasmImage `json:"image,omitempty"`

	// SHA256 checksum that will be used to verify the wasm code.
	SHA256 string `json:"sha256,omitempty"`
}

// WasmImage defines the OCI image containing the wasm code.
type WasmImage struct {
	// URL is the URL of the OCI image.
	URL string `json:"url"`

	// PullSecret is a reference to the secret containing the credentials to pull the image.
	PullSecret gwapiv1b1.SecretObjectReference `json:"pullSecret"`

	// PullPolicy is the policy to use when pulling the image.
	// If not specified, the default policy is IfNotPresent for images whose tag is not latest,
	// and Always for images whose tag is latest.
	// +optional
	PullPolicy *PullPolicy `json:"pullPolicy,omitempty"`
}

// PullPolicy defines the policy to use when pulling an OIC image.
type PullPolicy string

const (
	// PullPolicyIfNotPresent will only pull the image if it does not already exist.
	PullPolicyIfNotPresent PullPolicy = "IfNotPresent"

	// PullPolicyAlways will always pull the image.
	PullPolicyAlways PullPolicy = "Always"
)
