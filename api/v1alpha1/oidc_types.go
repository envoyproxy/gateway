// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

const OIDCClientSecretKey = "client-secret"

// OIDC defines the configuration for the OpenID Connect (OIDC) authentication.
type OIDC struct {
	// The OIDC Provider configuration.
	Provider OIDCProvider `json:"provider"`

	// The client ID to be used in the OIDC
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// +kubebuilder:validation:MinLength=1
	ClientID string `json:"clientID"`

	// The Kubernetes secret which contains the OIDC client secret to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	//
	// This is an Opaque secret. The client secret should be stored in the key
	// "client-secret".
	// +kubebuilder:validation:Required
	ClientSecret gwapiv1b1.SecretObjectReference `json:"clientSecret"`

	// The OIDC scopes to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// The "openid" scope is always added to the list of scopes if not already
	// specified.
	// +optional
	Scopes []string `json:"scopes,omitempty"`

	// The OIDC resources to be used in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// +optional
	Resources []string `json:"resources,omitempty"`

	// The redirect URL to be used in the OIDC
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// If not specified, uses the default redirect URI "%REQ(x-forwarded-proto)%://%REQ(:authority)%/oauth2/callback"
	RedirectURL *string `json:"redirectURL,omitempty"`

	// The path to log a user out, clearing their credential cookies.
	//
	// If not specified, uses a default logout path "/logout"
	LogoutPath *string `json:"logoutPath,omitempty"`

	// ForwardAccessTokenAsBearerToken indicates whether the Envoy should forward
	// the access token as a bearer token in the "Authorization" header to the backend.
	//
	// If not specified, defaults to false.
	// +optional
	// +notImplementedHide
	ForwardAccessTokenAsBearerToken *bool `json:"forwardAccessTokenAsBearerToken,omitempty"`

	// The default lifetime of the ID token and access token.
	// Please note that Envoy will always use the expiry time from the response
	// of the authorization server if it is provided. This field is only used when
	// the expiry time is not provided by the authorization.
	//
	// If not specified, defaults to 0. In this case, the "expires_in" field in
	// the authorization response must be set by the authorization server, or the
	// OAuth flow will fail.
	//
	// +optional
	// +notImplementedHide
	DefaultTokenExpireTime *metav1.Duration `json:"defaultTokenExpireTime,omitempty"`

	// RefreshToken indicates whether the Envoy should automatically refresh the
	// id token and access token when they expire.
	// When set to true, the Envoy will use the refresh token to get a new id token
	// and access token when they expire.
	//
	// If not specified, defaults to false.
	// +optional
	// +notImplementedHide
	RefreshToken *bool `json:"refreshToken,omitempty"`

	// The default lifetime of the refresh token.
	// This field is only used when the exp (expiration time) claim is omitted in
	// the refresh token or the refresh token is not JWT.
	//
	// If not specified, defaults to 604800s (one week).
	// Note: this field is only used when RefreshToken is set to true.
	// +optional
	// +notImplementedHide
	DefaultRefreshTokenExpireTime *metav1.Duration `json:"defaultRefreshTokenExpireTime,omitempty"`
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
	AuthorizationEndpoint *string `json:"authorizationEndpoint,omitempty"`

	// The OIDC Provider's [token endpoint](https://openid.net/specs/openid-connect-core-1_0.html#TokenEndpoint).
	// If not provided, EG will try to discover it from the provider's [Well-Known Configuration Endpoint](https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderConfigurationResponse).
	//
	// +optional
	TokenEndpoint *string `json:"tokenEndpoint,omitempty"`
}
