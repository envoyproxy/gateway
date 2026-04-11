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
	// Defaults to 30s if not specified.
	//
	// +optional
	SamplingWindow *gwapiv1.Duration `json:"samplingWindow,omitempty"`

	// SuccessRateThreshold is the lowest request success rate, as a percentage in the
	// range [1, 100], at which the filter will not reject requests. Defaults to 95 if
	// not specified. Envoy rejects values below 1%, so values lower than 1 are not allowed.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=100
	SuccessRateThreshold *uint32 `json:"successRateThreshold,omitempty"`

	// Aggression controls the rejection probability curve. A value of 1 means a linear
	// increase in rejection probability as the success rate decreases. Higher values
	// result in more aggressive rejection at higher success rates.
	// Envoy requires aggression to be greater than 0 and clamps values below 1 to 1.
	// Defaults to 1 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=1
	Aggression *uint32 `json:"aggression,omitempty"`

	// RPSThreshold defines the minimum requests per second below which requests will
	// pass through the filter without rejection. Defaults to 0 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	RPSThreshold *uint32 `json:"rpsThreshold,omitempty"`

	// MaxRejectionProbability represents the upper limit of the rejection probability,
	// expressed as a percentage in the range [0, 100]. Defaults to 80 if not specified.
	//
	// +optional
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	MaxRejectionProbability *uint32 `json:"maxRejectionProbability,omitempty"`

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
	// GRPCSuccessStatus defines gRPC status codes that are considered successful.
	// Status codes are defined in https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc.
	//
	// +optional
	GRPCSuccessStatus []GRPCSuccessCode `json:"grpcSuccessStatus,omitempty"`
}
