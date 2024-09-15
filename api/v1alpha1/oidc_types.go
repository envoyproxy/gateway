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

	// The optional cookie name overrides to be used for Bearer and IdToken cookies in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// If not specified, uses a randomly generated suffix
	// +optional
	CookieNames *OIDCCookieNames `json:"cookieNames,omitempty"`

	// The optional domain to set the access and ID token cookies on.
	// If not set, the cookies will default to the host of the request, not including the subdomains.
	// If set, the cookies will be set on the specified domain and all subdomains.
	// This means that requests to any subdomain will not require reauthentication after users log in to the parent domain.
	// +optional
	// +notImplementedHide
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9]))*$`
	CookieDomain *string `json:"cookieDomain,omitempty"`

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

	// ForwardAccessToken indicates whether the Envoy should forward the access token
	// via the Authorization header Bearer scheme to the upstream.
	// If not specified, defaults to false.
	// +optional
	ForwardAccessToken *bool `json:"forwardAccessToken,omitempty"`

	// DefaultTokenTTL is the default lifetime of the id token and access token.
	// Please note that Envoy will always use the expiry time from the response
	// of the authorization server if it is provided. This field is only used when
	// the expiry time is not provided by the authorization.
	//
	// If not specified, defaults to 0. In this case, the "expires_in" field in
	// the authorization response must be set by the authorization server, or the
	// OAuth flow will fail.
	//
	// +optional
	DefaultTokenTTL *metav1.Duration `json:"defaultTokenTTL,omitempty"`

	// RefreshToken indicates whether the Envoy should automatically refresh the
	// id token and access token when they expire.
	// When set to true, the Envoy will use the refresh token to get a new id token
	// and access token when they expire.
	//
	// If not specified, defaults to false.
	// +optional
	RefreshToken *bool `json:"refreshToken,omitempty"`

	// DefaultRefreshTokenTTL is the default lifetime of the refresh token.
	// This field is only used when the exp (expiration time) claim is omitted in
	// the refresh token or the refresh token is not JWT.
	//
	// If not specified, defaults to 604800s (one week).
	// Note: this field is only applicable when the "refreshToken" field is set to true.
	// +optional
	DefaultRefreshTokenTTL *metav1.Duration `json:"defaultRefreshTokenTTL,omitempty"`
}

// OIDCProvider defines the OIDC Provider configuration.
// +kubebuilder:validation:XValidation:rule="!has(self.backendRef)",message="BackendRefs must be used, backendRef is not supported."
// +kubebuilder:validation:XValidation:rule="has(self.backendSettings)? (has(self.backendSettings.retry)?(has(self.backendSettings.retry.perRetry)? !has(self.backendSettings.retry.perRetry.timeout):true):true):true",message="Retry timeout is not supported."
// +kubebuilder:validation:XValidation:rule="has(self.backendSettings)? (has(self.backendSettings.retry)?(has(self.backendSettings.retry.retryOn)? !has(self.backendSettings.retry.retryOn.httpStatusCodes):true):true):true",message="HTTPStatusCodes is not supported."
type OIDCProvider struct {
	// BackendRefs is used to specify the address of the OIDC Provider.
	// If the BackendRefs is not specified, The host and port of the OIDC Provider's token endpoint
	// will be used as the address of the OIDC Provider.
	//
	// TLS configuration can be specified in a BackendTLSConfig resource and target the BackendRefs.
	//
	// Other settings for the connection to the OIDC Provider can be specified in the BackendSettings resource.
	//
	// +optional
	// +notImplementedHide
	BackendCluster `json:",inline"`

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

// OIDCCookieNames defines the names of cookies to use in the Envoy OIDC filter.
type OIDCCookieNames struct {
	// The name of the cookie used to store the AccessToken in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// If not specified, defaults to "AccessToken-(randomly generated uid)"
	// +optional
	AccessToken *string `json:"accessToken,omitempty"`
	// The name of the cookie used to store the IdToken in the
	// [Authentication Request](https://openid.net/specs/openid-connect-core-1_0.html#AuthRequest).
	// If not specified, defaults to "IdToken-(randomly generated uid)"
	// +optional
	IDToken *string `json:"idToken,omitempty"`
}
