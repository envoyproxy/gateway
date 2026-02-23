// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// HTTPTransform defines the configuration for HTTP header and body transformations
// using the Envoy Transform filter. Only JSON body transformation is currently supported.
//
// +kubebuilder:validation:XValidation:rule="has(self.requestTransformation) || has(self.responseTransformation)",message="at least one of requestTransformation or responseTransformation must be specified"
// +k8s:deepcopy-gen=true
type HTTPTransform struct {
	// RequestTransformation defines transformations to apply to the request
	// before forwarding to the backend.
	//
	// +optional
	RequestTransformation *HTTPTransformation `json:"requestTransformation,omitempty"`

	// ResponseTransformation defines transformations to apply to the response
	// before sending to the client.
	//
	// +optional
	ResponseTransformation *HTTPTransformation `json:"responseTransformation,omitempty"`
}

// HTTPTransformation defines header and body transformations.
//
// +k8s:deepcopy-gen=true
type HTTPTransformation struct {
	// SetHeaders specifies headers to set on the request or response.
	// If the header already exists, it will be overwritten.
	// Header values support Envoy command operators including
	// %REQUEST_BODY(KEY*)% and %RESPONSE_BODY(KEY*)% for referencing JSON body fields.
	//
	// +optional
	SetHeaders []HTTPTransformHeader `json:"setHeaders,omitempty"`

	// AddHeaders specifies headers to add (append) to the request or response.
	// If the header already exists, the new value will be appended.
	// Header values support Envoy command operators including
	// %REQUEST_BODY(KEY*)% and %RESPONSE_BODY(KEY*)% for referencing JSON body fields.
	//
	// +optional
	AddHeaders []HTTPTransformHeader `json:"addHeaders,omitempty"`

	// RemoveHeaders specifies headers to remove from the request or response.
	//
	// +optional
	RemoveHeaders []string `json:"removeHeaders,omitempty"`

	// Body specifies the body transformation configuration.
	// Only JSON body transformation is currently supported.
	//
	// +optional
	Body *HTTPBodyTransformation `json:"body,omitempty"`
}

// HTTPTransformHeader defines a header name-value pair for the transform filter.
//
// +k8s:deepcopy-gen=true
type HTTPTransformHeader struct {
	// Name specifies the name of the header.
	//
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MinLength=1
	Name string `json:"name"`

	// Value specifies the value of the header.
	// Supports Envoy substitution format specifiers including
	// %REQUEST_BODY(KEY*)% and %RESPONSE_BODY(KEY*)%.
	//
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// HTTPBodyTransformation defines body transformation configuration.
//
// +kubebuilder:validation:XValidation:rule="(has(self.formatString) && !has(self.json)) || (!has(self.formatString) && has(self.json))",message="exactly one of formatString or json must be specified"
// +k8s:deepcopy-gen=true
type HTTPBodyTransformation struct {
	// FormatString specifies a text format string template for body transformation.
	// Supports Envoy substitution format specifiers.
	// Exactly one of FormatString or JSON must be specified.
	//
	// +optional
	FormatString *string `json:"formatString,omitempty"`

	// JSON specifies a JSON object template for body transformation.
	// The keys and values support Envoy substitution format specifiers.
	// Exactly one of FormatString or JSON must be specified.
	//
	// +optional
	JSON *apiextensionsv1.JSON `json:"json,omitempty"`

	// Action specifies whether the transformed body should merge into or replace the original body.
	// Defaults to Merge.
	//
	// +optional
	// +kubebuilder:default=Merge
	// +kubebuilder:validation:Enum=Merge;Replace
	Action *BodyTransformAction `json:"action,omitempty"`
}

// BodyTransformAction specifies the action to take with the transformed body.
// +kubebuilder:validation:Enum=Merge;Replace
type BodyTransformAction string

const (
	// BodyTransformActionMerge merges the transformed body into the original body.
	BodyTransformActionMerge BodyTransformAction = "Merge"

	// BodyTransformActionReplace replaces the original body with the transformed body.
	BodyTransformActionReplace BodyTransformAction = "Replace"
)
