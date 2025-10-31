// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AdmissionControl defines the admission control configuration for backend traffic.
// This feature rejects a portion of requests when the success rate falls below
// a specified threshold to prevent cascading failures.
//
// +optional
type AdmissionControl struct {
	// Enabled enables or disables admission control. Defaults to false.
	//
	// +optional
	// +kubebuilder:default=false
	Enabled *bool `json:"enabled,omitempty"`

	// SuccessCriteria defines the criteria for determining request success.
	//
	// +optional
	SuccessCriteria *AdmissionControlSuccessCriteria `json:"successCriteria,omitempty"`

	// SamplingWindow defines the time window for sampling requests.
	// Defaults to "30s".
	//
	// +optional
	// +kubebuilder:default="30s"
	SamplingWindow *metav1.Duration `json:"samplingWindow,omitempty"`

	// Aggression defines how aggressively to reject requests when success rate is low.
	// Must be between 1.0 and 10.0. Defaults to 1.0.
	//
	// +optional
	// +kubebuilder:default=1.0
	// +kubebuilder:validation:Minimum=1.0
	// +kubebuilder:validation:Maximum=10.0
	Aggression *float64 `json:"aggression,omitempty"`

	// SRThreshold defines the success rate threshold below which admission control activates.
	// Must be between 0.0 and 1.0. Defaults to 0.95.
	//
	// +optional
	// +kubebuilder:default=0.95
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	SRThreshold *float64 `json:"srThreshold,omitempty"`

	// RPSThreshold defines the minimum requests per second required for admission control.
	// Defaults to 5.0.
	//
	// +optional
	// +kubebuilder:default=5.0
	// +kubebuilder:validation:Minimum=0.0
	RPSThreshold *float64 `json:"rpsThreshold,omitempty"`

	// MaxRejectionProbability defines the maximum probability of rejecting a request.
	// Must be between 0.0 and 1.0. Defaults to 0.8.
	//
	// +optional
	// +kubebuilder:default=0.8
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	MaxRejectionProbability *float64 `json:"maxRejectionProbability,omitempty"`
}

// AdmissionControlSuccessCriteria defines the criteria for determining request success.
type AdmissionControlSuccessCriteria struct {
	// HTTP defines success criteria for HTTP requests.
	//
	// +optional
	HTTP *AdmissionControlHTTPSuccessCriteria `json:"http,omitempty"`

	// GRPC defines success criteria for gRPC requests.
	//
	// +optional
	GRPC *AdmissionControlGRPCSuccessCriteria `json:"grpc,omitempty"`
}

// AdmissionControlHTTPSuccessCriteria defines success criteria for HTTP requests.
type AdmissionControlHTTPSuccessCriteria struct {
	// HTTPSuccessStatus defines the HTTP status codes considered successful.
	// If not specified, defaults to 200-299 range.
	//
	// +optional
	HTTPSuccessStatus []AdmissionControlStatusRange `json:"httpSuccessStatus,omitempty"`
}

// AdmissionControlGRPCSuccessCriteria defines success criteria for gRPC requests.
type AdmissionControlGRPCSuccessCriteria struct {
	// GRPCSuccessStatus defines the gRPC status codes considered successful.
	// If not specified, defaults to OK (0).
	//
	// +optional
	GRPCSuccessStatus []AdmissionControlStatusRange `json:"grpcSuccessStatus,omitempty"`
}

// AdmissionControlStatusRange defines a range of status codes.
type AdmissionControlStatusRange struct {
	// Start defines the start of the status code range (inclusive).
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=999
	Start int32 `json:"start"`

	// End defines the end of the status code range (inclusive).
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=999
	End int32 `json:"end"`
}
