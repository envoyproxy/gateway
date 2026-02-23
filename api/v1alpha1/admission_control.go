// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// AdmissionControl defines the admission control policy to be applied.
// This configuration probabilistically rejects requests based on the success rate
// of previous requests in a configurable sliding time window.
// All fields are optional and will use Envoy's defaults when not specified.
type AdmissionControl struct {
	// SamplingWindow defines the time window over which request success rates are calculated.
	// Defaults to 60s if not specified.
	//
	// +optional
	SamplingWindow *gwapiv1.Duration `json:"samplingWindow,omitempty"`

	// SuccessRateThreshold defines the lowest request success rate at which the filter
	// will not reject requests. The value should be in the range [0.0, 1.0].
	// Defaults to 0.95 (95%) if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	SuccessRateThreshold *float64 `json:"successRateThreshold,omitempty"`

	// Aggression controls the rejection probability curve. A value of 1.0 means a linear
	// increase in rejection probability as the success rate decreases. Higher values
	// result in more aggressive rejection at higher success rates.
	// Defaults to 1.0 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	Aggression *float64 `json:"aggression,omitempty"`

	// RPSThreshold defines the minimum requests per second below which requests will
	// pass through the filter without rejection. Defaults to 1 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	RPSThreshold *uint32 `json:"rpsThreshold,omitempty"`

	// MaxRejectionProbability represents the upper limit of the rejection probability.
	// The value should be in the range [0.0, 1.0]. Defaults to 0.95 (95%) if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0.0
	// +kubebuilder:validation:Maximum=1.0
	MaxRejectionProbability *float64 `json:"maxRejectionProbability,omitempty"`

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
	// HTTPSuccessStatus defines HTTP status codes that are considered successful.
	//
	// +optional
	HTTPSuccessStatus []HTTPStatus `json:"httpSuccessStatus,omitempty"`
}

// GRPCSuccessCode defines gRPC status codes as defined in
// https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc.
// +kubebuilder:validation:Enum=OK;CANCELLED;UNKNOWN;INVALID_ARGUMENT;DEADLINE_EXCEEDED;NOT_FOUND;ALREADY_EXISTS;PERMISSION_DENIED;RESOURCE_EXHAUSTED;FAILED_PRECONDITION;ABORTED;OUT_OF_RANGE;UNIMPLEMENTED;INTERNAL;UNAVAILABLE;DATA_LOSS;UNAUTHENTICATED
type GRPCSuccessCode string

const (
	GRPCSuccessCodeOK                 GRPCSuccessCode = "OK"
	GRPCSuccessCodeCancelled          GRPCSuccessCode = "CANCELLED"
	GRPCSuccessCodeUnknown            GRPCSuccessCode = "UNKNOWN"
	GRPCSuccessCodeInvalidArgument    GRPCSuccessCode = "INVALID_ARGUMENT"
	GRPCSuccessCodeDeadlineExceeded   GRPCSuccessCode = "DEADLINE_EXCEEDED"
	GRPCSuccessCodeNotFound           GRPCSuccessCode = "NOT_FOUND"
	GRPCSuccessCodeAlreadyExists      GRPCSuccessCode = "ALREADY_EXISTS"
	GRPCSuccessCodePermissionDenied   GRPCSuccessCode = "PERMISSION_DENIED"
	GRPCSuccessCodeResourceExhausted  GRPCSuccessCode = "RESOURCE_EXHAUSTED"
	GRPCSuccessCodeFailedPrecondition GRPCSuccessCode = "FAILED_PRECONDITION"
	GRPCSuccessCodeAborted            GRPCSuccessCode = "ABORTED"
	GRPCSuccessCodeOutOfRange         GRPCSuccessCode = "OUT_OF_RANGE"
	GRPCSuccessCodeUnimplemented      GRPCSuccessCode = "UNIMPLEMENTED"
	GRPCSuccessCodeInternal           GRPCSuccessCode = "INTERNAL"
	GRPCSuccessCodeUnavailable        GRPCSuccessCode = "UNAVAILABLE"
	GRPCSuccessCodeDataLoss           GRPCSuccessCode = "DATA_LOSS"
	GRPCSuccessCodeUnauthenticated    GRPCSuccessCode = "UNAUTHENTICATED"
)

// GRPCSuccessCriteria defines success criteria for gRPC requests.
type GRPCSuccessCriteria struct {
	// GRPCSuccessStatus defines gRPC status codes that are considered successful.
	// Status codes are defined in https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc.
	//
	// +optional
	GRPCSuccessStatus []GRPCSuccessCode `json:"grpcSuccessStatus,omitempty"`
}
