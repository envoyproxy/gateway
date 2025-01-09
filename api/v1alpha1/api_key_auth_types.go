// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

const APIKeysSecretKey = "credentials"

type ExtractFromType string

const (
	ExtractFromHeader     ExtractFromType = "Header"
	ExtractFromQueryParam ExtractFromType = "QueryParam"
	ExtractFromCookie     ExtractFromType = "Cookie"
)

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
// Only one of header, queryParam or cookie is supposed to be specified.
//
// +kubebuilder:validation:XValidation:rule="(self.type == 'Header' && has(self.header))",message="When 'type' is 'Header', 'header' must be set."
// +kubebuilder:validation:XValidation:rule="(self.type == 'QueryParam' && has(self.queryParam))",message="When 'type' is 'QueryParam', 'queryParam' must be set."
// +kubebuilder:validation:XValidation:rule="(self.type == 'Cookie' && has(self.cookie))",message="When 'type' is 'Cookie', 'cookie' must be set."
type ExtractFrom struct {
	// Type is the type of the source to fetch the key from.
	// It can be either Header, QueryParam or Cookie, and the corresponding field must be specified.
	//
	// +kubebuilder:validation:Enum=Header;QueryParam;Cookie
	Type ExtractFromType `json:"type"`
	// Header is the name of the header to fetch the key from.
	// This field is marked as optional, but should be specified if the type is Header.
	//
	// +optional
	Header *string `json:"header,omitempty"`
	// QueryParam is the name of the query parameter to fetch the key from.
	// This field is marked as optional, but should be specified if the type is QueryParam.
	//
	// +optional
	QueryParam *string `json:"queryParam,omitempty"`
	// Cookie is the name of the cookie to fetch the key from.
	// This field is marked as optional, but should be specified if the type is Cookie.
	//
	// +optional
	Cookie *string `json:"cookie,omitempty"`
}
