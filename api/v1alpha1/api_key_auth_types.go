// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const APIKeysSecretKey = "credentials"

// APIKeyAuth defines the configuration for the API Key Authentication.
type APIKeyAuth struct {
	// CredentialRefs is the Kubernetes secret which contains the API keys.
	// This is an Opaque secret.
	// Each API key is stored in the key representing the client id.
	// If the secrets have a key for a duplicated client, the first one will be used.
	CredentialRefs []gwapiv1.SecretObjectReference `json:"credentialRefs"`

	// ExtractFrom is where to fetch the key from the coming request.
	// The value from the first source that has a key will be used.
	ExtractFrom []*ExtractFrom `json:"extractFrom"`
}

// ExtractFrom is where to fetch the key from the coming request.
// Only one of header, queryParams or cookie is supposed to be specified.
//
// Note: we intentionally don't add the validation for the only one of header, queryParams or cookie is supposed to be specified with +kubebuilder:validation:XValidation:rule.
// Instead, we add the validation in the controller reconciliation.
// Technically we can define CEL, but the CEL estimated cost exceeds the threshold and it wouldn't be accepted.
//
// +kubebuilder:validation:XValidation:rule="(has(self.header) || has(self.queryParams) || has(self.cookie))",message="one of header, queryParams or cookie must be specified"
type ExtractFrom struct {
	// Header is the name of the header to fetch the key from.
	// This field is optional, but only one of header, queryParams or cookie is supposed to be specified.
	//
	// +optional
	Header *string `json:"header,omitempty"`
	// QueryParams is the name of the query parameter to fetch the key from.
	// This field is optional, but only one of header, queryParams or cookie is supposed to be specified.
	//
	// +optional
	QueryParams *string `json:"queryParams,omitempty"`
	// Cookie is the name of the cookie to fetch the key from.
	// This field is optional, but only one of header, queryParams or cookie is supposed to be specified.
	//
	// +optional
	Cookie *string `json:"cookie,omitempty"`
}
