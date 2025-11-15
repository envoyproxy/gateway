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
func (*admissionControl) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, httpListener *ir.HTTPListener) error {
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

	// Set sampling window (defaults to 60s if not specified)
	samplingWindow := "60s"
	if admissionControl.SamplingWindow != nil {
		samplingWindow = admissionControl.SamplingWindow.Duration.String()
	}
	duration, err := parseDuration(samplingWindow)
	if err != nil {
		return nil, fmt.Errorf("invalid samplingWindow: %w", err)
	}
	config.SamplingWindow = durationpb.New(duration)

	// Set success rate threshold (defaults to 0.95 if not specified)
	// Note: srThreshold is in range [0.0, 1.0], but Percent expects [0.0, 100.0]
	srThreshold := 0.95
	if admissionControl.SuccessRateThreshold != nil {
		srThreshold = *admissionControl.SuccessRateThreshold
	}
	config.SrThreshold = &corev3.RuntimePercent{
		DefaultValue: &typev3.Percent{Value: srThreshold * 100.0},
	}

	// Set aggression (defaults to 1.0 if not specified)
	aggression := 1.0
	if admissionControl.Aggression != nil {
		aggression = *admissionControl.Aggression
	}
	config.Aggression = &corev3.RuntimeDouble{
		DefaultValue: aggression,
	}

	// Set RPS threshold (defaults to 1 if not specified)
	rpsThreshold := uint32(1)
	if admissionControl.RPSThreshold != nil {
		rpsThreshold = *admissionControl.RPSThreshold
	}
	config.RpsThreshold = &corev3.RuntimeUInt32{
		DefaultValue: rpsThreshold,
	}

	// Set max rejection probability (defaults to 0.95 if not specified)
	// Note: maxRejectionProbability is in range [0.0, 1.0], but Percent expects [0.0, 100.0]
	maxRejectionProbability := 0.95
	if admissionControl.MaxRejectionProbability != nil {
		maxRejectionProbability = *admissionControl.MaxRejectionProbability
	}
	config.MaxRejectionProbability = &corev3.RuntimePercent{
		DefaultValue: &typev3.Percent{Value: maxRejectionProbability * 100.0},
	}

	// Set success criteria (part of EvaluationCriteria oneof)
	if admissionControl.SuccessCriteria != nil {
		successCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria{}

		// HTTP success criteria
		if admissionControl.SuccessCriteria.HTTP != nil && len(admissionControl.SuccessCriteria.HTTP.HTTPSuccessStatus) > 0 {
			httpCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_HttpCriteria{}
			for _, statusRange := range admissionControl.SuccessCriteria.HTTP.HTTPSuccessStatus {
				httpCriteria.HttpSuccessStatus = append(httpCriteria.HttpSuccessStatus, &typev3.Int32Range{
					Start: statusRange.Start,
					End:   statusRange.End,
				})
			}
			successCriteria.HttpCriteria = httpCriteria
		}

		// gRPC success criteria
		if admissionControl.SuccessCriteria.GRPC != nil && len(admissionControl.SuccessCriteria.GRPC.GRPCSuccessStatus) > 0 {
			grpcCriteria := &admissioncontrolv3.AdmissionControl_SuccessCriteria_GrpcCriteria{}
			for _, status := range admissionControl.SuccessCriteria.GRPC.GRPCSuccessStatus {
				grpcCriteria.GrpcSuccessStatus = append(grpcCriteria.GrpcSuccessStatus, uint32(status))
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

// patchResources adds all the other needed resources referenced by this filter.
func (*admissionControl) patchResources(tCtx *types.ResourceVersionTable, routes []*ir.HTTPRoute) error {
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
