// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	admissioncontrolv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/admission_control/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&admissionControl{})
}

type admissionControl struct{}

var _ httpFilter = &admissionControl{}

// patchHCM builds and appends the admission control filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*admissionControl) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}

	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if !listenerContainsAdmissionControl(irListener) {
		return nil
	}

	// Return early if the admission control filter already exists.
	for _, existingFilter := range mgr.HttpFilters {
		if existingFilter.Name == string(egv1a1.EnvoyFilterAdmissionControl) {
			return nil
		}
	}

	admissionControlFilter, err := buildHCMAdmissionControlFilter()
	if err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, admissionControlFilter)

	return nil
}

// buildHCMAdmissionControlFilter returns a basic admission control HTTP filter.
func buildHCMAdmissionControlFilter() (*hcmv3.HttpFilter, error) {
	// Create a basic admission control configuration
	admissionControlProto := &admissioncontrolv3.AdmissionControl{}

	admissionControlAny, err := proto.ToAnyWithValidation(admissionControlProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: string(egv1a1.EnvoyFilterAdmissionControl),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: admissionControlAny,
		},
	}, nil
}

// patchRoute patches the provided route with the admission control config if applicable.
func (*admissionControl) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil || irRoute == nil {
		return nil
	}

	// Check if admission control is configured for this route
	if irRoute.Traffic == nil || irRoute.Traffic.AdmissionControl == nil {
		return nil
	}

	admissionControlConfig := irRoute.Traffic.AdmissionControl

	// Skip if admission control is explicitly disabled
	if admissionControlConfig.Enabled != nil && !*admissionControlConfig.Enabled {
		return nil
	}

	// Build the admission control configuration
	routeCfgProto, err := buildAdmissionControlConfig(admissionControlConfig)
	if err != nil {
		return err
	}

	// Add the admission control filter to the route
	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}

	routeCfgAny, err := proto.ToAnyWithValidation(routeCfgProto)
	if err != nil {
		return err
	}

	route.TypedPerFilterConfig[string(egv1a1.EnvoyFilterAdmissionControl)] = routeCfgAny

	return nil
}

// buildAdmissionControlConfig builds the admission control configuration from the IR.
func buildAdmissionControlConfig(admissionControl *ir.AdmissionControl) (*admissioncontrolv3.AdmissionControl, error) {
	if admissionControl == nil {
		return nil, errors.New("admissionControl cannot be nil")
	}

	config := &admissioncontrolv3.AdmissionControl{}

	// Set enabled (defaults to true if not specified)
	enabled := true
	if admissionControl.Enabled != nil {
		enabled = *admissionControl.Enabled
	}
	config.Enabled = &corev3.RuntimeFeatureFlag{
		DefaultValue: &wrapperspb.BoolValue{Value: enabled},
	}

	// Only set fields the user explicitly configured; Envoy applies its own defaults
	// (sampling_window=30s, sr_threshold=95%, aggression=1.0, rps_threshold=0, max_rejection_probability=80%).
	if admissionControl.SamplingWindow != nil {
		duration, err := parseDuration(admissionControl.SamplingWindow.Duration.String())
		if err != nil {
			return nil, fmt.Errorf("invalid samplingWindow: %w", err)
		}
		config.SamplingWindow = durationpb.New(duration)
	}

	if admissionControl.SuccessRateThreshold != nil {
		config.SrThreshold = &corev3.RuntimePercent{
			DefaultValue: &typev3.Percent{Value: *admissionControl.SuccessRateThreshold * 100.0},
		}
	}

	if admissionControl.Aggression != nil {
		config.Aggression = &corev3.RuntimeDouble{
			DefaultValue: *admissionControl.Aggression,
		}
	}

	if admissionControl.RPSThreshold != nil {
		config.RpsThreshold = &corev3.RuntimeUInt32{
			DefaultValue: *admissionControl.RPSThreshold,
		}
	}

	if admissionControl.MaxRejectionProbability != nil {
		config.MaxRejectionProbability = &corev3.RuntimePercent{
			DefaultValue: &typev3.Percent{Value: *admissionControl.MaxRejectionProbability * 100.0},
		}
	}

	// Set success criteria (part of EvaluationCriteria oneof)
	if admissionControl.SuccessCriteria != nil {
		successCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria{}

		// HTTP success criteria: each individual status code becomes a single-element range [code, code+1)
		if admissionControl.SuccessCriteria.HTTP != nil && len(admissionControl.SuccessCriteria.HTTP.HTTPSuccessStatus) > 0 {
			httpCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_HttpCriteria{}
			for _, code := range admissionControl.SuccessCriteria.HTTP.HTTPSuccessStatus {
				httpCriteria.HttpSuccessStatus = append(httpCriteria.HttpSuccessStatus, &typev3.Int32Range{
					Start: code,
					End:   code + 1,
				})
			}
			successCriteria.HttpCriteria = httpCriteria
		}

		// gRPC success criteria: map string enum names to numeric codes
		if admissionControl.SuccessCriteria.GRPC != nil && len(admissionControl.SuccessCriteria.GRPC.GRPCSuccessStatus) > 0 {
			grpcCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_GrpcCriteria{}
			for _, status := range admissionControl.SuccessCriteria.GRPC.GRPCSuccessStatus {
				if code, ok := grpcStatusCodeToUint32(status); ok {
					grpcCriteria.GrpcSuccessStatus = append(grpcCriteria.GrpcSuccessStatus, code)
				}
			}
			successCriteria.GrpcCriteria = grpcCriteria
		}

		// Set as EvaluationCriteria (oneof field)
		config.EvaluationCriteria = &admissioncontrolv3.AdmissionControl_SuccessCriteria_{
			SuccessCriteria: successCriteria,
		}
	}

	return config, nil
}

// parseDuration parses a duration string and returns a time.Duration.
func parseDuration(s string) (time.Duration, error) {
	return time.ParseDuration(s)
}

// grpcStatusCodeToUint32 maps a gRPC status code string name to its numeric value.
// See https://github.com/grpc/grpc/blob/master/doc/statuscodes.md#status-codes-and-their-use-in-grpc
func grpcStatusCodeToUint32(name string) (uint32, bool) {
	codes := map[string]uint32{
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
	code, ok := codes[name]
	return code, ok
}

// patchResources adds all the other needed resources referenced by this filter.
func (*admissionControl) patchResources(_ *types.ResourceVersionTable, _ []*ir.HTTPRoute) error {
	// Admission control filter doesn't require additional resources
	return nil
}

// listenerContainsAdmissionControl returns true if the provided listener contains
// any route with admission control configured.
func listenerContainsAdmissionControl(irListener *ir.HTTPListener) bool {
	if irListener == nil {
		return false
	}

	for _, route := range irListener.Routes {
		if route.Traffic != nil && route.Traffic.AdmissionControl != nil {
			// Check if enabled (defaults to true)
			if route.Traffic.AdmissionControl.Enabled == nil || *route.Traffic.AdmissionControl.Enabled {
				return true
			}
		}
	}

	return false
}
