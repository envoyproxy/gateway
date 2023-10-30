// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const OIDCClientSecretKey = "client_secret"

// OIDC defines the configuration for the OpenID Connect (OIDC) authentication.
type OIDC struct {
	// The OIDC Provider configuration.
	Provider OIDCProvider `json:"provider"`

	// The client ID assigned to this policy to be used in the OIDC
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// +kubebuilder:validation:MinLength=1
	ClientID string `json:"clientID"`

	// The Kubernetes secret which contains the OIDC client secret assigned to
	// the filter to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// This is an Opaque secret. The client secret should be stored in the key
	// "client_secret".
	// +kubebuilder:validation:Required
	ClientSecret gwapiv1b1.SecretObjectReference `json:"clientSecret"`

	// The OIDC scopes to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// The "openid" scope is always added to the list of scopes if not already
	// specified.
	// +optional
	Scopes []string `json:"scopes,omitempty"`
}

// OIDCProvider defines the OIDC Provider configuration.
type OIDCProvider struct {
	// The OIDC Provider's [issuer identifier](https://openid.net/specs/openid-connect-discovery-1_0.html#IssuerDiscovery).
	// Issuer MUST be a URI RFC 3986 [RFC3986] with a scheme component that MUST
	// be https, a host component, and optionally, port and path components and
	// no query or fragment components.
	// +kubebuilder:validation:MinLength=1
	Issuer string `json:"issuer"`

	// TODO zhaohuabing validate the issuer

	// The OIDC Provider's [authorization endpoint](https://openid.net/specs/openid-connect-core-1_0.html#AuthorizationEndpoint).
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	AuthorizationEndpoint string `json:"authorizationEndpoint,omitempty"`

	// The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	TokenEndpoint string `json:"tokenEndpoint,omitempty"`
}
