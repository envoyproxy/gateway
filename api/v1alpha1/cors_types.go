// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// Origin is defined by the scheme (protocol), hostname (domain), and port of
// the URL used to access it. The hostname can be “precise” which is just the
// domain name or “wildcard” which is a domain name prefixed with a single
// wildcard label such as “*.example.com”.
//
// For example, the following are valid origins:
// - https://foo.example.com
// - https://*.example.com
// - http://foo.example.com:8080
// - http://*.example.com:8080
//
// +kubebuilder:validation:MinLength=1
// +kubebuilder:validation:MaxLength=253
// +kubebuilder:validation:Pattern=`^https?:\/\/(\*\.)?[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*(:[0-9]+)?$`
type Origin string

// CORS defines the configuration for Cross-Origin Resource Sharing (CORS).
type CORS struct {
	// AllowOrigins defines the origins that are allowed to make requests.
	// +kubebuilder:validation:MinItems=1
	AllowOrigins []Origin `json:"allowOrigins,omitempty" yaml:"allowOrigins"`
	// AllowMethods defines the methods that are allowed to make requests.
	// +kubebuilder:validation:MinItems=1
	AllowMethods []string `json:"allowMethods,omitempty" yaml:"allowMethods"`
	// AllowHeaders defines the headers that are allowed to be sent with requests.
	AllowHeaders []string `json:"allowHeaders,omitempty" yaml:"allowHeaders,omitempty"`
	// ExposeHeaders defines the headers that can be exposed in the responses.
	ExposeHeaders []string `json:"exposeHeaders,omitempty" yaml:"exposeHeaders,omitempty"`
	// MaxAge defines how long the results of a preflight request can be cached.
	MaxAge *metav1.Duration `json:"maxAge,omitempty" yaml:"maxAge,omitempty"`
	// AllowCredentials indicates whether a request can include user credentials
	// like cookies, authentication headers, or TLS client certificates.
	AllowCredentials *bool `json:"allowCredentials,omitempty" yaml:"allowCredentials,omitempty"`
}
