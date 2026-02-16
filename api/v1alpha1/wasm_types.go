// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// WasmEnv defines the environment variables for the VM of a Wasm extension
type WasmEnv struct {
	// HostKeys is a list of keys for environment variables from the host envoy process
	// that should be passed into the Wasm VM. This is useful for passing secrets to to Wasm extensions.
	// +optional
	HostKeys []string `json:"hostKeys,omitempty"`
}

// Wasm defines a Wasm extension.
//
// Note: at the moment, Envoy Gateway does not support configuring Wasm runtime.
// v8 is used as the VM runtime for the Wasm extensions.
type Wasm struct {
	// Name is a unique name for this Wasm extension. It is used to identify the
	// Wasm extension if multiple extensions are handled by the same vm_id and root_id.
	// It's also used for logging/debugging.
	// If not specified, EG will generate a unique name for the Wasm extension.
	//
	// +optional
	Name *string `json:"name,omitempty"`

	// RootID is a unique ID for a set of extensions in a VM which will share a
	// RootContext and Contexts if applicable (e.g., an Wasm HttpFilter and an Wasm AccessLog).
	// If left blank, all extensions with a blank root_id with the same vm_id will share Context(s).
	//
	// Note: RootID must match the root_id parameter used to register the Context in the Wasm code.
	RootID *string `json:"rootID,omitempty"`

	// Code is the Wasm code for the extension.
	Code WasmCodeSource `json:"code"`

	// Config is the configuration for the Wasm extension.
	// This configuration will be passed as a JSON string to the Wasm extension.
	// +optional
	Config *apiextensionsv1.JSON `json:"config,omitempty"`

	// FailOpen is a switch used to control the behavior when a fatal error occurs
	// during the initialization or the execution of the Wasm extension.
	//
	// If FailOpen is set to true, the system bypasses the Wasm extension and
	// allows the traffic to pass through. If it is set to false or
	// not set (defaulting to false), the system blocks the traffic and returns
	// an HTTP 5xx error.
	//
	// If set to true, the Wasm extension will also be bypassed if the configuration is invalid.
	//
	// +optional
	// +kubebuilder:default=false
	FailOpen *bool `json:"failOpen,omitempty"`

	// Priority defines the location of the Wasm extension in the HTTP filter chain.
	// If not specified, the Wasm extension will be inserted before the router filter.
	// Priority *uint32 `json:"priority,omitempty"`

	// Env configures the environment for the Wasm extension
	// +optional
	Env *WasmEnv `json:"env,omitempty"`
}

// WasmCodeSource defines the source of the Wasm code.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'HTTP' ? has(self.http) : !has(self.http)",message="If type is HTTP, http field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Image' ? has(self.image) : !has(self.image)",message="If type is Image, image field needs to be set."
type WasmCodeSource struct {
	// Type is the type of the source of the Wasm code.
	// Valid WasmCodeSourceType values are "HTTP" or "Image".
	//
	// +kubebuilder:validation:Enum=HTTP;Image;ConfigMap
	// +unionDiscriminator
	Type WasmCodeSourceType `json:"type"`

	// HTTP is the HTTP URL containing the Wasm code.
	//
	// Note that the HTTP server must be accessible from the Envoy proxy.
	// +optional
	HTTP *HTTPWasmCodeSource `json:"http,omitempty"`

	// Image is the OCI image containing the Wasm code.
	//
	// Note that the image must be accessible from the Envoy Gateway.
	// +optional
	Image *ImageWasmCodeSource `json:"image,omitempty"`

	// PullPolicy is the policy to use when pulling the Wasm module by either the HTTP or Image source.
	// This field is only applicable when the SHA256 field is not set.
	//
	// If not specified, the default policy is IfNotPresent except for OCI images whose tag is latest.
	//
	// Note: EG does not update the Wasm module every time an Envoy proxy requests
	// the Wasm module even if the pull policy is set to Always.
	// It only updates the Wasm module when the EnvoyExtension resource version changes.
	// +optional
	PullPolicy *ImagePullPolicy `json:"pullPolicy,omitempty"`
}

// WasmCodeSourceType specifies the types of sources for the Wasm code.
// +kubebuilder:validation:Enum=HTTP;Image
type WasmCodeSourceType string

const (
	// HTTPWasmCodeSourceType allows the user to specify the Wasm code in an HTTP URL.
	HTTPWasmCodeSourceType WasmCodeSourceType = "HTTP"

	// ImageWasmCodeSourceType allows the user to specify the Wasm code in an OCI image.
	ImageWasmCodeSourceType WasmCodeSourceType = "Image"
)

// HTTPWasmCodeSource defines the HTTP URL containing the Wasm code.
type HTTPWasmCodeSource struct {
	// URL is the URL containing the Wasm code.
	// +kubebuilder:validation:Pattern=`^((https?:)(\/\/\/?)([\w]*(?::[\w]*)?@)?([\d\w\.-]+)(?::(\d+))?)?([\/\\\w\.()-]*)?(?:([?][^#]*)?(#.*)?)*`
	URL string `json:"url"`

	// SHA256 checksum that will be used to verify the Wasm code.
	//
	// If not specified, Envoy Gateway will not verify the downloaded Wasm code.
	// kubebuilder:validation:Pattern=`^[a-f0-9]{64}$`
	// +optional
	SHA256 *string `json:"sha256"`

	// TLS configuration when connecting to the Wasm code source.
	// +optional
	// +notImplementedHide
	TLS *WasmCodeSourceTLSConfig `json:"tls,omitempty"`
}

// ImageWasmCodeSource defines the OCI image containing the Wasm code.
type ImageWasmCodeSource struct {
	// URL is the URL of the OCI image.
	// URL can be in the format of `registry/image:tag` or `registry/image@sha256:digest`.
	URL string `json:"url"`

	// SHA256 checksum that will be used to verify the OCI image.
	//
	// It must match the digest of the OCI image.
	//
	// If not specified, Envoy Gateway will not verify the downloaded OCI image.
	// kubebuilder:validation:Pattern=`^[a-f0-9]{64}$`
	// +optional
	SHA256 *string `json:"sha256"`

	// PullSecretRef is a reference to the secret containing the credentials to pull the image.
	// Only support Kubernetes Secret resource from the same namespace.
	// +kubebuilder:validation:XValidation:message="only support Secret kind.",rule="self.kind == 'Secret'"
	// +optional
	PullSecretRef *gwapiv1.SecretObjectReference `json:"pullSecretRef,omitempty"`

	// TLS configuration when connecting to the Wasm code source.
	// +optional
	// +notImplementedHide
	TLS *WasmCodeSourceTLSConfig `json:"tls,omitempty"`
}

// ImagePullPolicy defines the policy to use when pulling an OIC image.
// +kubebuilder:validation:Enum=IfNotPresent;Always
type ImagePullPolicy string

const (
	// ImagePullPolicyIfNotPresent will only pull the image if it does not already exist in the EG cache.
	ImagePullPolicyIfNotPresent ImagePullPolicy = "IfNotPresent"

	// ImagePullPolicyAlways will pull the image when the EnvoyExtension resource version changes.
	// Note: EG does not update the Wasm module every time an Envoy proxy requests the Wasm module.
	ImagePullPolicyAlways ImagePullPolicy = "Always"
)

// WasmCodeSourceTLSConfig defines the TLS configuration when connecting to the Wasm code source.
type WasmCodeSourceTLSConfig struct {
	// CACertificateRef contains a references to
	// Kubernetes objects that contain TLS certificates of
	// the Certificate Authorities that can be used
	// as a trust anchor to validate the certificates presented by the Wasm code source.
	//
	// Kubernetes ConfigMap and Kubernetes Secret are supported.
	// Note: The ConfigMap or Secret must be in the same namespace as the EnvoyExtensionPolicy.
	CACertificateRef gwapiv1.SecretObjectReference `json:"caCertificateRef"`
}
