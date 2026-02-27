// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// DynamicModuleSourceType specifies the types of sources for dynamic module code.
// +kubebuilder:validation:Enum=Local;Remote
type DynamicModuleSourceType string

const (
	// LocalDynamicModuleSourceType specifies a module loaded from the local filesystem.
	LocalDynamicModuleSourceType DynamicModuleSourceType = "Local"

	// RemoteDynamicModuleSourceType specifies a module fetched from a remote source.
	RemoteDynamicModuleSourceType DynamicModuleSourceType = "Remote"
)

// DynamicModuleSource defines the source of the dynamic module code.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Remote' ? has(self.remote) : !has(self.remote)",message="If type is Remote, remote field needs to be set."
type DynamicModuleSource struct {
	// Type is the type of the source of the dynamic module code.
	// Defaults to Local.
	//
	// +kubebuilder:default=Local
	// +unionDiscriminator
	// +optional
	Type *DynamicModuleSourceType `json:"type,omitempty"`

	// Local specifies a module loaded from the proxy's local filesystem.
	// Envoy searches for lib${libraryName}.so in the path specified by the
	// ENVOY_DYNAMIC_MODULES_SEARCH_PATH environment variable.
	//
	// +optional
	Local *LocalDynamicModuleSource `json:"local,omitempty"`

	// Remote specifies a module fetched from a remote source.
	// The module binary is downloaded and cached by Envoy.
	//
	// +optional
	// +notImplementedHide
	Remote *RemoteDynamicModuleSource `json:"remote,omitempty"`
}

// LocalDynamicModuleSource defines a dynamic module loaded from the local filesystem.
type LocalDynamicModuleSource struct {
	// LibraryName is the name of the shared library file that Envoy will load.
	// Envoy searches for lib${libraryName}.so in the path specified by the
	// ENVOY_DYNAMIC_MODULES_SEARCH_PATH environment variable.
	// If not specified, defaults to the value of the DynamicModuleEntry's Name.
	//
	// +optional
	// +kubebuilder:validation:MaxLength=253
	// +kubebuilder:validation:Pattern=`^[a-zA-Z0-9_]([a-zA-Z0-9_.-]*[a-zA-Z0-9_])?$`
	LibraryName *string `json:"libraryName,omitempty"`
}

// RemoteDynamicModuleSource defines a dynamic module fetched from a remote source.
//
// +notImplementedHide
type RemoteDynamicModuleSource struct {}

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
