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

	// ForwardClientIDHeader is the name of the header to forward the client identity to the backend
	// service. The header will be added to the request with the client id as the value.
	//
	// +optional
	ForwardClientIDHeader *string `json:"forwardClientIDHeader,omitempty"`

	// SanitizeAPIKey indicates whether to remove the API key from the request
	// before forwarding it to the backend service.
	//
	// +optional
	SanitizeAPIKey *bool `json:"sanitizeAPIKey,omitempty"`
}

// ExtractFrom is where to fetch the key from the coming request.
// Only one of header, param or cookie is supposed to be specified.
type ExtractFrom struct {
	// Headers is the names of the header to fetch the key from.
	// If multiple headers are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Headers []string `json:"headers,omitempty"`
	// Params is the names of the query parameter to fetch the key from.
	// If multiple params are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Params []string `json:"params,omitempty"`
	// Cookies is the names of the cookie to fetch the key from.
	// If multiple cookies are specified, envoy will look for the api key in the order of the list.
	// This field is optional, but only one of headers, params or cookies is supposed to be specified.
	//
	// +optional
	Cookies []string `json:"cookies,omitempty"`
}

type APIKeyAuthHeaderForwarding struct {
	// ClientIdentityHeader is the name of the header to forward the client identity to the backend
	// service. The header will be added to the request with the client id as the value.
	//
	// +optional
	ClientIdentityHeader *string `json:"clientIdentityHeader,omitempty"`

	// SuppressCredentials indicates whether to remove the API key credential from the request
	// before forwarding it to the backend service.
	//
	// +optional
	SuppressCredentials *bool `json:"suppressCredentials,omitempty"`
}
