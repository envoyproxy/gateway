// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindCorsFilterKind is the name of the CorsFilter kind.
	KindCorsFilter = "CorsFilter"
)

//+kubebuilder:object:root=true

type CorsFilter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the CorsFilter type.
	Spec CorsFilterSpec `json:"spec"`

	// Note: The status sub-resource has been excluded but may be added in the future.
}

// CorsFilterSpec defines the desired state of the CorsFilter type.
// +union

type CorsFilterSpec struct {
	// Type decides the scope for the RateLimits.
	// Valid CorsFilterType values are "Global".
	//
	// +unionDiscriminator
	Type CorsType `json:"type"`
	// Global defines global cors configuration.
	//
	// +optional
	CorsPolicy CorsPolicy `json:"corsPolicy,omitempty"`
}

// CorsType specifies the types of Cors.
// +kubebuilder:validation:Enum=Global;PerRoute
type CorsType string

const (
	// GlobalCorsType allows the cors to be applied across all Envoy proxy instances.
	GlobalCorsType CorsType = "Global"
)

type CorsPolicy struct {
	AllowOrigins []*StringMatch `json:"allowOrigins,omitempty"`

	AllowMethods []string `json:"allowMethods,omitempty"`

	AllowHeaders []string `json:"allowHeaders,omitempty"`

	ExposeHeaders []string `json:"exposeHeaders,omitempty"`

	MaxAge int64 `json:"maxAge,omitempty"`

	AllowCredentials bool `json:"allowCredentials,omitempty"`
}
type StringMatch struct {
	Exact  *string `json:"exact,omitempty"`
	Prefix *string `json:"prefix,omitempty"`
	Regex  *string `json:"regex,omitempty"`
}

// +kubebuilder:object:root=true
type CorsFilterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CorsFilter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CorsFilter{}, &CorsFilterList{})
}
