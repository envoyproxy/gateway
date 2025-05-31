// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// DynamicModule defines a Dynamic Module extension.
type DynamicModule struct {
	// ExtensionName is a unique name for this Dynamic Module extension. It is used to identify the
	// Dynamic Module extension if multiple extensions are loaded.
	// It's also used for logging/debugging.
	// If not specified, EG will generate a unique name for the Dynamic Module extension.
	//
	// +optional
	ExtensionName *string `json:"extensionName,omitempty"`

	// Module is the name of the dynamic module to load.
	// The module name is used to search for the shared library file in the search path.
	// The search path is configured by the environment variable ENVOY_DYNAMIC_MODULES_SEARCH_PATH.
	// The actual search path is ${ENVOY_DYNAMIC_MODULES_SEARCH_PATH}/lib${name}.so.
	//
	// +kubebuilder:validation:Required
	Module string `json:"module"`

	// ExtensionConfig is the configuration for the Dynamic Module extension.
	// This configuration will be passed to the Dynamic Module extension.
	// +optional
	ExtensionConfig *string `json:"extensionConfig,omitempty"`

	// DoNotClose prevents the module from being unloaded with dlclose.
	// This is useful for modules that have global state that should not be unloaded.
	// A module is closed when no more references to it exist in the process.
	// For example, no HTTP filters are using the module (e.g. after configuration update).
	//
	// +optional
	DoNotClose *bool `json:"doNotClose,omitempty"`
}
