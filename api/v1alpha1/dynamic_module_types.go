// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// DynamicModuleSourceType specifies the types of sources for dynamic module code.
// +kubebuilder:validation:Enum=Local;Remote;Image
type DynamicModuleSourceType string

const (
	// LocalDynamicModuleSourceType specifies a module loaded from the local filesystem.
	LocalDynamicModuleSourceType DynamicModuleSourceType = "Local"

	// RemoteDynamicModuleSourceType specifies a module fetched from a remote source.
	RemoteDynamicModuleSourceType DynamicModuleSourceType = "Remote"

	// ImageDynamicModuleSourceType specifies a module extracted from an OCI image.
	ImageDynamicModuleSourceType DynamicModuleSourceType = "Image"
)

// DynamicModuleSource defines the source of the dynamic module code.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Remote' ? has(self.remote) : !has(self.remote)",message="If type is Remote, remote field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type != 'Local' || has(self.local)",message="If type is Local, local field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Remote' ? !has(self.local) : true",message="If type is Remote, local field must not be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Image' ? has(self.image) : !has(self.image)",message="If type is Image, image field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Image' ? !has(self.local) : true",message="If type is Image, local field must not be set."
type DynamicModuleSource struct {
	// Type is the type of the source of the dynamic module code.
	// Defaults to Local.
	//
	// +kubebuilder:default=Local
	// +unionDiscriminator
	// +optional
	Type *DynamicModuleSourceType `json:"type,omitempty"`

	// Local specifies a module loaded from the proxy's local filesystem
	// by absolute path.
	//
	// +optional
	Local *LocalDynamicModuleSource `json:"local,omitempty"`

	// Remote specifies a module fetched from a remote source.
	// The module binary is downloaded and cached by Envoy.
	//
	// +optional
	Remote *RemoteDynamicModuleSource `json:"remote,omitempty"`

	// Image specifies a module extracted from an OCI image.
	// Envoy Gateway pulls the image, extracts the shared library, and serves it
	// to Envoy from its built-in HTTP server, the same way it serves Wasm modules
	// from OCI images.
	//
	// +optional
	Image *ImageDynamicModuleSource `json:"image,omitempty"`
}

// LocalDynamicModuleSource defines a dynamic module loaded from the local filesystem.
type LocalDynamicModuleSource struct {
	// Path is the absolute filesystem path to the dynamic module shared library (.so file).
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	Path string `json:"path"`
}

// RemoteDynamicModuleSource defines a dynamic module fetched from a remote HTTP source.
type RemoteDynamicModuleSource struct {
	// URL is the HTTP or HTTPS URL of the dynamic module shared library (.so file).
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +kubebuilder:validation:Pattern=`^https?://[^/?#]+(?:[/?#].*)?$`
	URL string `json:"url"`

	// SHA256 checksum that Envoy will use to verify the downloaded module binary.
	//
	// +kubebuilder:validation:Pattern=`^[a-f0-9]{64}$`
	SHA256 string `json:"sha256"`
}

// ImageDynamicModuleSource defines a dynamic module extracted from an OCI image.
//
// Both container images, with the shared library as a file in one of the image
// layers, and OCI artifacts, with the shared library as a layer annotated with
// its file name (as pushed by `oras push`), are supported.
//
// For a multi-platform image, the linux image for the CPU architecture of the
// Envoy Gateway controller is used, which must match the architecture of the
// Envoy proxies.
type ImageDynamicModuleSource struct {
	// URL is the URL of the OCI image.
	// URL can be in the format of `registry/image:tag` or `registry/image@sha256:digest`.
	// If neither a tag nor a digest is specified, the `latest` tag is used.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	URL string `json:"url"`

	// SHA256 checksum that will be used to verify the OCI image.
	//
	// It must match the digest of the OCI image, which is the digest of the
	// image index for a multi-platform image.
	//
	// If not specified, Envoy Gateway will not verify the downloaded OCI image.
	//
	// +kubebuilder:validation:Pattern=`^[a-f0-9]{64}$`
	// +optional
	SHA256 *string `json:"sha256,omitempty"`

	// Path is the path of the dynamic module shared library (.so file) inside the image,
	// for example `/usr/lib/envoy/libmy_module.so`.
	//
	// If not specified, the image must contain exactly one file with the `.so` extension.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=4096
	// +optional
	Path *string `json:"path,omitempty"`

	// PullPolicy is the policy to use when pulling the image.
	// If not specified, the default policy is IfNotPresent, except for OCI images
	// whose tag is latest, which default to Always.
	//
	// +optional
	PullPolicy *ImagePullPolicy `json:"pullPolicy,omitempty"`

	// PullSecretRef is a reference to the secret containing the credentials to pull the image.
	// The secret must be of type `kubernetes.io/dockerconfigjson`.
	// A secret in another namespace than the EnvoyProxy requires a ReferenceGrant.
	//
	// +kubebuilder:validation:XValidation:message="only support Secret kind.",rule="self.kind == 'Secret'"
	// +optional
	PullSecretRef *gwapiv1.SecretObjectReference `json:"pullSecretRef,omitempty"`
}

// DynamicModuleEntry defines a dynamic module that is registered and allowed
// for use by EnvoyExtensionPolicy resources.
type DynamicModuleEntry struct {
	// Name is the logical name for this module. EnvoyExtensionPolicy resources
	// reference modules by this name.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`
	Name string `json:"name"`

	// Source defines where the dynamic module code is loaded from.
	Source DynamicModuleSource `json:"source"`

	// DoNotClose prevents the module from being unloaded with dlclose when no
	// more references exist. This is useful for modules that maintain global
	// state that should not be destroyed on configuration updates.
	// Defaults to false.
	//
	// +optional
	// +kubebuilder:default=false
	DoNotClose *bool `json:"doNotClose,omitempty"`

	// LoadGlobally loads the dynamic module with the RTLD_GLOBAL flag.
	// By default, modules are loaded with RTLD_LOCAL to avoid symbol conflicts.
	// Set this to true when the module needs to share symbols with other
	// dynamic libraries it loads.
	// Defaults to false.
	//
	// +optional
	// +kubebuilder:default=false
	LoadGlobally *bool `json:"loadGlobally,omitempty"`
}

// DynamicModule defines a dynamic module HTTP filter to be loaded by Envoy.
// The module must be registered in the EnvoyProxy resource's dynamicModules
// allowlist by the infrastructure operator.
type DynamicModule struct {
	// Name references a dynamic module registered in the EnvoyProxy resource's
	// dynamicModules list. The referenced module must exist in the registry;
	// otherwise, the policy will be rejected.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`
	Name string `json:"name"`

	// FilterName identifies a specific filter implementation within the dynamic
	// module. A single shared library can contain multiple filter implementations.
	// This value is passed to the module's HTTP filter config init function to
	// select the appropriate implementation.
	// If not specified, defaults to an empty string.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	FilterName *string `json:"filterName,omitempty"`

	// Config is the configuration for the dynamic module filter.
	// This is serialized as JSON and passed to the module's initialization function.
	//
	// +optional
	Config *apiextensionsv1.JSON `json:"config,omitempty"`

	// TerminalFilter indicates that this dynamic module handles requests without
	// requiring an upstream backend. The module is responsible for generating and
	// sending the response to downstream directly.
	// Defaults to false.
	//
	// +optional
	// +kubebuilder:default=false
	TerminalFilter *bool `json:"terminalFilter,omitempty"`
}
