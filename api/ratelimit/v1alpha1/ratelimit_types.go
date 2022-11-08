// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

// Ratelimit allows the user to limit the number of incoming requests
// to a predefined value based on attributes within the traffic flow.
type Ratelimit struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// Spec defines the desired state of HTTPRoute.
	Spec RatelimitSpec `json:"spec"`

	// Status defines the current state of Ratelimit.
	Status RatelimitStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RatelimitList contains a list of Ratelimit resources.
type RatelimitList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Ratelimit `json:"items"`
}

// RatelimitSpec defines the desired state of Ratelimit
type RatelimitSpec struct {
	// Type decides the scope for the Ratelimits.
	Type RatelimitType `json:"type"`
	// Rules are a list of Ratelimit matchers and limits.
	//
	// +kubebuilder:validation:MaxItems=16
	Rules []RatelimitRule `json:"rules"`
}

// RatelimitType specifies the types of Ratelimiting.
// Valid RatelimitType values are:
//
// * "Local"
// * "Global"
//
// +kubebuilder:validation:Enum=Local;Global
type RatelimitType string

const (
	// In this mode, the ratelimits are applied per Envoy proxy instance.
	RatelimitTypeLocal RatelimitType = "Local"

	// In this mode, the ratelimits are applied across all Envoy proxy instances.
	RatelimitTypeGlobal RatelimitType = "Global"
)

// RatelimitRule defines the semantics for matching attributes
// from the incoming requests, and setting limits for them.
type RatelimitRule struct {
	// +optional
	// +kubebuilder:validation:MaxItems=8
	Matches []RatelimitMatch `json:"matches,omitempty"`
	Limit   RatelimitValue   `json:"limit"`
}

// RatelimitMatch specifies the attributes within the traffic flow that can
// be matched on.
type RatelimitMatch struct {
	// ClientAddress defines the semantics for matching on the IP Address
	// of the client making the request.
	// +optional
	ClientAddress *ClientAddressMatch `json:"clientAddress,omitempty"`
	// +listType=map
	// +listMapKey=name
	// +optional
	// +kubebuilder:validation:MaxItems=16
	Headers []HeaderMatch `json:"headers,omitempty"`
}

// ClientAddressMatch defines the match attributes based on the source IP Address
// originated from the client.
type ClientAddressMatch struct {
	// Cidr specifies the subnet range of source IP addresses
	// to be matched on.
	// Not setting this field, implies matching on all source/client IP addresses.
	// +optional
	Cidr *string `json:"cidr,omitempty"`
}

// HeaderMatch defines the match attributes within the HTTP Headers of the request.
type HeaderMatch struct {
	// Type specifies how to match against the value of the header.
	//
	// +optional
	// +kubebuilder:default=Exact
	Type *HeaderMatchType `json:"type,omitempty"`

	// Name of the HTTP header.
	Name string `json:"name"`

	// Value within the HTTP header.
	// Not setting this field, implies matching on all unique values within the header.
	// +optional
	Value *string `json:"value,omitempty"`
}

// HeaderMatchType specifies the semantics of how HTTP header values should be
// compared. Valid HeaderMatchType values are:
//
// * "Exact"
// * "RegularExpression"
//
// +kubebuilder:validation:Enum=Exact;RegularExpression
type HeaderMatchType string

// HeaderMatchType constants.
const (
	HeaderMatchExact             HeaderMatchType = "Exact"
	HeaderMatchRegularExpression HeaderMatchType = "RegularExpression"
)

// RatelimitValue defines the limits for ratelimiting.
type RatelimitValue struct {
	Requests uint          `json:"requests"`
	Unit     RatelimitUnit `json:"unit"`
}

// RatelimitUnit specifies the intervals for setting rate limits.
// Valid RatelimitUnit values are:
//
// * "Second"
// * "Minute"
// * "Hour"
// * "Day"
//
// +kubebuilder:validation:Enum=Second;Minute;Hour;Day
type RatelimitUnit string

// RatelimitStatus is used to define the state of Ratelimit.
type RatelimitStatus struct{}
