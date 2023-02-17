// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// KindCorsKind is the name of the KindCors kind.
	KindCors = "Cors"
)

//+kubebuilder:object:root=true

type Cors struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of the CorsFilter type.
	Spec CorsPolicySpec `json:"spec"`

	// Note: The status sub-resource has been excluded but may be added in the future.
}

type CorsPolicySpec struct {
	CorsPolicy CorsPolicy `json:"corsPolicy,omitempty"`
}

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

func init() {
	SchemeBuilder.Register(&Cors{})
}
