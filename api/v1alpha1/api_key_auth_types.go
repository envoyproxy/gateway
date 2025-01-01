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
	// Credentials is the Kubernetes secret which contains the API keys.
	// This is an Opaque secret.
	// Each API key is stored in the key representing the client id,
	// which can be used in AllowedClients to authorize the client in a simple way.
	Credentials gwapiv1.SecretObjectReference `json:"credentials"`

	// KeySources is where to fetch the key from the coming request.
	// The value from the first source that has a key will be used.
	KeySources []*KeySource `json:"keySources"`

	// AllowedClients is a list of clients that are allowed to access the route or vhost.
	// The clients listed here should be subset of the clients listed in the `Credentials` to provide authorization control
	// after the authentication is successful. If the list is empty, then all authenticated clients
	// are allowed. This provides very limited but simple authorization.
	//
	// +optional
	AllowedClients []string `json:"allowedClients,omitempty"`
}

// KeySource is where to fetch the key from the coming request.
// Only one of header, query or cookie is supposed to be specified.
//
// Note: we intentionally don't add the validation for the only one of header, query or cookie is supposed to be specified with +kubebuilder:validation:XValidation:rule.
// Instead, we add the validation in the controller reconciliation.
// Technically we can define CEL, but the CEL estimated cost exceeds the threshold and it wouldn't be accepted.
//
// +kubebuilder:validation:XValidation:rule="(has(self.header) || has(self.query) || has(self.cookie))",message="one of header, query or cookie must be specified"
type KeySource struct {
	// Header is the name of the header to fetch the key from.
	// This field is optional, but only one of header, query or cookie is supposed to be specified.
	//
	// +optional
	Header *string `json:"header,omitempty"`
	// Query is the name of the query parameter to fetch the key from.
	// This field is optional, but only one of header, query or cookie is supposed to be specified.
	//
	// +optional
	Query *string `json:"query,omitempty"`
	// Cookie is the name of the cookie to fetch the key from.
	// This field is optional, but only one of header, query or cookie is supposed to be specified.
	//
	// +optional
	Cookie *string `json:"cookie,omitempty"`
}
