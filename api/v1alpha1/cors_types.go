// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// Origin is defined by the scheme (protocol), hostname (domain), and port of
// the URL used to access it. The hostname can be "precise" which is just the
// domain name or "wildcard" which is a domain name prefixed with a single
// wildcard label such as "*.example.com".
// In addition to that a single wildcard (with or without scheme) can be
// configured to match any origin.
//
// For example, the following are valid origins:
// - https://foo.example.com
// - https://*.example.com
// - http://foo.example.com:8080
// - http://*.example.com:8080
// - https://*
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^(\*|https?:\/\/(\*|(\*\.)?(([\w-]+\.?)+)?[\w-]+)(:\d{1,5})?)$`
type Origin string

// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
type CORS struct {
	// AllowOrigins defines the origins that are allowed to make requests.
	// It specifies the allowed origins in the Access-Control-Allow-Origin CORS response header.
	// The value "*" allows any origin to make requests.
	//
	// +optional
	AllowOrigins []Origin `json:"allowOrigins,omitempty"`

	// AllowMethods defines the methods that are allowed to make requests.
	// It specifies the allowed methods in the Access-Control-Allow-Methods CORS response header..
	// The value "*" allows any method to be used.
	//
	// +optional
	AllowMethods []string `json:"allowMethods,omitempty"`

	// AllowHeaders defines the headers that are allowed to be sent with requests.
	// It specifies the allowed headers in the Access-Control-Allow-Headers CORS response header..
	// The value "*" allows any header to be sent.
	//
	// +optional
	AllowHeaders []string `json:"allowHeaders,omitempty"`

	// ExposeHeaders defines which response headers should be made accessible to
	// scripts running in the browser.
	// It specifies the headers in the Access-Control-Expose-Headers CORS response header..
	// The value "*" allows any header to be exposed.
	//
	// +optional
	ExposeHeaders []string `json:"exposeHeaders,omitempty"`

	// MaxAge defines how long the results of a preflight request can be cached.
	// It specifies the value in the Access-Control-Max-Age CORS response header..
	//
	// +optional
	MaxAge *gwapiv1.Duration `json:"maxAge,omitempty"`

	// AllowCredentials indicates whether a request can include user credentials
	// like cookies, authentication headers, or TLS client certificates.
	// It specifies the value in the Access-Control-Allow-Credentials CORS response header.
	//
	// +optional
	AllowCredentials *bool `json:"allowCredentials,omitempty"`

	// TODO zhaohuabing: according to CORS spec, wildcard should be treated as a literal value
	// for CORS requests with credentials.
	// This needs to be supported in the Envoy CORS filter.
}
