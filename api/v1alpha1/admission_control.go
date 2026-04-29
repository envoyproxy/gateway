// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// AdmissionControl configures health-based load shedding for upstream backends.
//
// Envoy tracks recent upstream responses over a sliding sampling window. When the
// observed success rate drops below the configured threshold, Envoy
// probabilistically rejects new requests before forwarding them upstream. This can
// reduce pressure on degraded backends and give them time to recover.
//
// All fields are optional. When omitted, Envoy's admission control defaults are used.
//
// +kubebuilder:validation:XValidation:rule="!has(self.minSuccessRate) || (self.minSuccessRate >= 1 && self.minSuccessRate <= 100)",message="minSuccessRate must be between 1 and 100"
// +kubebuilder:validation:XValidation:rule="!has(self.maxRejectionPercent) || (self.maxRejectionPercent >= 0 && self.maxRejectionPercent <= 100)",message="maxRejectionPercent must be between 0 and 100"
// +kubebuilder:validation:XValidation:rule="!has(self.samplingWindow) || duration(self.samplingWindow) >= duration('1s')",message="samplingWindow must be at least 1s"
type AdmissionControl struct {
	// SamplingWindow defines the time window over which request success rates are calculated.
	// Must be at least 1s; Envoy truncates the window to whole seconds and uses it as the
	// denominator in RPS calculations, so sub-second values would produce a zero denominator.
	// Defaults to 30s if not specified.
	//
	// +optional
	SamplingWindow *gwapiv1.Duration `json:"samplingWindow,omitempty"`

	// MinSuccessRate is the lowest request success rate, as a percentage in the
	// range [1, 100], at which the filter will not reject requests. Defaults to 95 if
	// not specified. Envoy rejects values below 1%, so values lower than 1 are not allowed.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	MinSuccessRate *uint32 `json:"minSuccessRate,omitempty"`

	// RejectionAggression controls how steeply the rejection probability rises
	// as the observed success rate falls below MinSuccessRate. A value of 1
	// produces a linear curve; higher values reject more aggressively for a
	// given drop in success rate. Must be greater than 0; values below 1 are
	// clamped to 1. Defaults to 1.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	RejectionAggression *uint32 `json:"rejectionAggression,omitempty"`

	// MinRequestRate defines the minimum requests per second below which requests will
	// pass through the filter without rejection. Defaults to 0 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	MinRequestRate *uint32 `json:"minRequestRate,omitempty"`

	// MaxRejectionPercent represents the upper limit of the rejection probability,
	// expressed as a percentage in the range [0, 100]. Defaults to 80 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxRejectionPercent *uint32 `json:"maxRejectionPercent,omitempty"`

	// SuccessCriteria defines what constitutes a successful request for both HTTP and gRPC.
	//
	// +optional
	SuccessCriteria *AdmissionControlSuccessCriteria `json:"successCriteria,omitempty"`
}

// AdmissionControlSuccessCriteria defines the criteria for determining successful requests.
type AdmissionControlSuccessCriteria struct {
	// HTTP defines success criteria for HTTP requests.
	//
	// +optional
	HTTP *HTTPSuccessCriteria `json:"http,omitempty"`

	// GRPC defines success criteria for gRPC requests.
	//
	// +optional
	GRPC *GRPCSuccessCriteria `json:"grpc,omitempty"`
}

// HTTPSuccessCriteria defines success criteria for HTTP requests.
type HTTPSuccessCriteria struct {
	// StatusCodes defines HTTP status codes that are considered successful.
	//
	// +optional
	StatusCodes []HTTPStatus `json:"statusCodes,omitempty"`
}

// GRPCSuccessCode defines gRPC status codes as defined in
// https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc.
// +kubebuilder:validation:Enum=Ok;Cancelled;Unknown;InvalidArgument;DeadlineExceeded;NotFound;AlreadyExists;PermissionDenied;ResourceExhausted;FailedPrecondition;Aborted;OutOfRange;Unimplemented;Internal;Unavailable;DataLoss;Unauthenticated
type GRPCSuccessCode string

const (
	GRPCSuccessCodeOk                 GRPCSuccessCode = "Ok"
	GRPCSuccessCodeCancelled          GRPCSuccessCode = "Cancelled"
	GRPCSuccessCodeUnknown            GRPCSuccessCode = "Unknown"
	GRPCSuccessCodeInvalidArgument    GRPCSuccessCode = "InvalidArgument"
	GRPCSuccessCodeDeadlineExceeded   GRPCSuccessCode = "DeadlineExceeded"
	GRPCSuccessCodeNotFound           GRPCSuccessCode = "NotFound"
	GRPCSuccessCodeAlreadyExists      GRPCSuccessCode = "AlreadyExists"
	GRPCSuccessCodePermissionDenied   GRPCSuccessCode = "PermissionDenied"
	GRPCSuccessCodeResourceExhausted  GRPCSuccessCode = "ResourceExhausted"
	GRPCSuccessCodeFailedPrecondition GRPCSuccessCode = "FailedPrecondition"
	GRPCSuccessCodeAborted            GRPCSuccessCode = "Aborted"
	GRPCSuccessCodeOutOfRange         GRPCSuccessCode = "OutOfRange"
	GRPCSuccessCodeUnimplemented      GRPCSuccessCode = "Unimplemented"
	GRPCSuccessCodeInternal           GRPCSuccessCode = "Internal"
	GRPCSuccessCodeUnavailable        GRPCSuccessCode = "Unavailable"
	GRPCSuccessCodeDataLoss           GRPCSuccessCode = "DataLoss"
	GRPCSuccessCodeUnauthenticated    GRPCSuccessCode = "Unauthenticated"
)

// GRPCSuccessCriteria defines success criteria for gRPC requests.
type GRPCSuccessCriteria struct {
	// StatusCodes defines gRPC status codes that are considered successful.
	// Status codes are defined in https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc.
	//
	// +optional
	StatusCodes []GRPCSuccessCode `json:"statusCodes,omitempty"`
}
