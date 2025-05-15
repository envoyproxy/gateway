// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ir

// DynamicModule holds the information associated with the Dynamic Module extensions.
// +k8s:deepcopy-gen=true
type DynamicModule struct {
	// ExtensionName is a unique name for a Dynamic Module configuration.
	// The xds translator only generates one Dynamic Module filter for each unique name.
	ExtensionName string `json:"extensionName"`

	// Module is the name of the dynamic module to load.
	Module string `json:"module"`

	// ExtensionConfig is the configuration for the Dynamic Module extension.
	// This configuration will be passed to the Dynamic Module extension.
	ExtensionConfig *string `json:"extensionConfig,omitempty"`

	// DoNotClose prevents the module from being unloaded with dlclose.
	DoNotClose bool `json:"doNotClose,omitempty"`
}
