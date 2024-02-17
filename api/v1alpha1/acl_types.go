// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// +kubebuilder:validation:XValidation:rule="(has(self.allow) || has(self.deny))",message="one of allow or deny must be specified"
//
// ACL defines the IP deny/allow configuration.
type ACL struct {
	Allow []IPSpec `json:"allow,omitempty"`
	Deny  []IPSpec `json:"deny,omitempty"`
}

// IPSpec defines the configuration for IP.
type IPSpec struct {
	// Prefix contains the IP prefix.
	// Example: 1.2.3.0
	//
	Prefix string `json:"prefix,omitempty"`
	// Length contains the length of the IP network prefix.
	// Example: 24
	Length uint32 `json:"length,omitempty"`
}
