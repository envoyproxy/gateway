// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	admissioncontrolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/admission_control/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// buildUpstreamAdmissionControlFilter builds an admission control filter for use
// as an upstream HTTP filter on a cluster. Envoy's admission_control filter is a
// "dual filter" that supports both the downstream (HCM) and upstream (cluster)
// extension categories, but does not support per-route typedPerFilterConfig.
// Placing it as an upstream filter gives per-cluster success-rate tracking.
func buildUpstreamAdmissionControlFilter(ac *ir.AdmissionControl) (*hcmv3.HttpFilter, error) {
	config, err := buildAdmissionControlConfig(ac)
	if err != nil {
		return nil, err
	}

	configAny, err := proto.ToAnyWithValidation(config)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: "envoy.filters.http.admission_control",
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: configAny,
		},
	}, nil
}

// buildAdmissionControlConfig builds the admission control configuration from the IR.
func buildAdmissionControlConfig(admissionControl *ir.AdmissionControl) (*admissioncontrolv3.AdmissionControl, error) {
	if admissionControl == nil {
		return nil, errors.New("admissionControl cannot be nil")
	}

	// The filter is enabled whenever the policy is configured.
	config := &admissioncontrolv3.AdmissionControl{
		Enabled: &corev3.RuntimeFeatureFlag{
			DefaultValue: &wrapperspb.BoolValue{Value: true},
		},
	}

	if admissionControl.SamplingWindow != nil {
		config.SamplingWindow = durationpb.New(admissionControl.SamplingWindow.Duration)
	}

	if admissionControl.MinSuccessRate != nil {
		config.SrThreshold = &corev3.RuntimePercent{
			DefaultValue: &typev3.Percent{Value: float64(*admissionControl.MinSuccessRate)},
		}
	}

	if admissionControl.RejectionAggression != nil {
		config.Aggression = &corev3.RuntimeDouble{
			DefaultValue: float64(*admissionControl.RejectionAggression),
		}
	}

	if admissionControl.MinRequestRate != nil {
		config.RpsThreshold = &corev3.RuntimeUInt32{
			DefaultValue: *admissionControl.MinRequestRate,
		}
	}

	if admissionControl.MaxRejectionPercent != nil {
		config.MaxRejectionProbability = &corev3.RuntimePercent{
			DefaultValue: &typev3.Percent{Value: float64(*admissionControl.MaxRejectionPercent)},
		}
	}

	successCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria{}

	// Set success criteria (part of EvaluationCriteria oneof)
	if admissionControl.SuccessCriteria != nil {
		// HTTP success criteria: each individual status code becomes a single-element range [code, code+1)
		if admissionControl.SuccessCriteria.HTTP != nil && len(admissionControl.SuccessCriteria.HTTP.StatusCodes) > 0 {
			httpCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_HttpCriteria{}
			for _, code := range admissionControl.SuccessCriteria.HTTP.StatusCodes {
				httpCriteria.HttpSuccessStatus = append(httpCriteria.HttpSuccessStatus, &typev3.Int32Range{
					Start: code,
					End:   code + 1,
				})
			}
			successCriteria.HttpCriteria = httpCriteria
		}

		if admissionControl.SuccessCriteria.GRPC != nil && len(admissionControl.SuccessCriteria.GRPC.StatusCodes) > 0 {
			grpcCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_GrpcCriteria{}
			for _, status := range admissionControl.SuccessCriteria.GRPC.StatusCodes {
				if code, ok := grpcStatusCodeToUint32(status); ok {
					grpcCriteria.GrpcSuccessStatus = append(grpcCriteria.GrpcSuccessStatus, code)
				}
			}
			successCriteria.GrpcCriteria = grpcCriteria
		}

	}

	// Always set EvaluationCriteria (required field)
	config.EvaluationCriteria = &admissioncontrolv3.AdmissionControl_SuccessCriteria_{
		SuccessCriteria: successCriteria,
	}

	return config, nil
}

// grpcStatusCodes maps a gRPC status code string name to its numeric value.
// See https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc
var grpcStatusCodes = map[string]uint32{
	"Ok":                 0,
	"Cancelled":          1,
	"Unknown":            2,
	"InvalidArgument":    3,
	"DeadlineExceeded":   4,
	"NotFound":           5,
	"AlreadyExists":      6,
	"PermissionDenied":   7,
	"ResourceExhausted":  8,
	"FailedPrecondition": 9,
	"Aborted":            10,
	"OutOfRange":         11,
	"Unimplemented":      12,
	"Internal":           13,
	"Unavailable":        14,
	"DataLoss":           15,
	"Unauthenticated":    16,
}

// grpcStatusCodeToUint32 maps a gRPC status code string name to its numeric value.
func grpcStatusCodeToUint32(name string) (uint32, bool) {
	code, ok := grpcStatusCodes[name]
	return code, ok
}
