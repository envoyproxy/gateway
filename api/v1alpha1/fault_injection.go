// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// FaultInjection defines the fault injection policy to be applied. This configuration can be used to
// inject delays and abort requests to mimic failure scenarios such as service failures and overloads
// +union
//
// +kubebuilder:validation:XValidation:rule=" has(self.delay) || has(self.abort) ",message="Delay and abort faults are set at least one."
type FaultInjection struct {

	// If specified, a delay will be injected into the request.
	//
	// +optional
	Delay *FaultInjectionDelay `json:"delay,omitempty"`

	// If specified, the request will be aborted if it meets the configuration criteria.
	//
	// +optional
	Abort *FaultInjectionAbort `json:"abort,omitempty"`
}

// FaultInjectionDelay defines the delay fault injection configuration
type FaultInjectionDelay struct {
	// FixedDelay specifies the fixed delay duration
	//
	// +required
	FixedDelay *metav1.Duration `json:"fixedDelay"`

	// Percentage specifies the percentage of requests to be delayed. Default 100%, if set 0, no requests will be delayed. Accuracy to 0.0001%.
	// +optional
	// +kubebuilder:default=100
	Percentage *float32 `json:"percentage,omitempty"`
}

// FaultInjectionAbort defines the abort fault injection configuration
// +union
//
// +kubebuilder:validation:XValidation:rule=" !(has(self.httpStatus) && has(self.grpcStatus)) ",message="httpStatus and grpcStatus cannot be simultaneously defined."
// +kubebuilder:validation:XValidation:rule=" has(self.httpStatus) || has(self.grpcStatus) ",message="httpStatus and grpcStatus are set at least one."
type FaultInjectionAbort struct {
	// StatusCode specifies the HTTP status code to be returned
	//
	// +optional
	// +kubebuilder:validation:Minimum=200
	// +kubebuilder:validation:Maximum=600
	HTTPStatus *int32 `json:"httpStatus,omitempty"`

	// GrpcStatus specifies the GRPC status code to be returned
	//
	// +optional
	GrpcStatus *int32 `json:"grpcStatus,omitempty"`

	// Percentage specifies the percentage of requests to be aborted. Default 100%, if set 0, no requests will be aborted. Accuracy to 0.0001%.
	// +optional
	// +kubebuilder:default=100
	Percentage *float32 `json:"percentage,omitempty"`
}
