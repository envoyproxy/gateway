// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	grpcjsontranscoder "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/grpc_json_transcoder/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&grpcJSONTranscoder{})
}

type grpcJSONTranscoder struct{}

var _ httpFilter = &grpcJSONTranscoder{}

// patchHCM appends one disabled gRPC-JSON transcoder filter per distinct transcoder config
// in the listener. The config is fully populated rather than left to the route, because
// Envoy validates the listener eagerly and rejects a GrpcJsonTranscoder without
// descriptor_set.
func (*grpcJSONTranscoder) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error

	for _, route := range irListener.Routes {
		if route.GRPCJSONTranscoder == nil {
			continue
		}

		transcoder := route.GRPCJSONTranscoder
		if hcmContainsFilter(mgr, grpcJSONTranscoderFilterName(transcoder)) {
			continue
		}

		filter, err := buildHCMGRPCJSONTranscoderFilter(transcoder)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}

		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return errs
}

// buildHCMGRPCJSONTranscoderFilter returns a disabled-by-default gRPC-JSON transcoder filter.
func buildHCMGRPCJSONTranscoderFilter(transcoder *ir.GRPCJSONTranscoder) (*hcmv3.HttpFilter, error) {
	cfg, err := buildGRPCJSONTranscoderConfig(transcoder)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: grpcJSONTranscoderFilterName(transcoder),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: cfg,
		},
		Disabled: true,
	}, nil
}

// buildGRPCJSONTranscoderConfig converts the IR config into a validated xDS Any.
func buildGRPCJSONTranscoderConfig(transcoder *ir.GRPCJSONTranscoder) (*anypb.Any, error) {
	cfg := &grpcjsontranscoder.GrpcJsonTranscoder{
		DescriptorSet: &grpcjsontranscoder.GrpcJsonTranscoder_ProtoDescriptorBin{
			ProtoDescriptorBin: transcoder.ProtoDescriptorBin,
		},
		Services:               transcoder.Services,
		IgnoredQueryParameters: transcoder.IgnoredQueryParameters,
	}

	if o := transcoder.PrintOptions; o != nil {
		cfg.PrintOptions = &grpcjsontranscoder.GrpcJsonTranscoder_PrintOptions{
			AddWhitespace:              derefBool(o.AddWhitespace),
			AlwaysPrintPrimitiveFields: derefBool(o.AlwaysPrintPrimitiveFields),
			AlwaysPrintEnumsAsInts:     derefBool(o.AlwaysPrintEnumsAsInts),
			PreserveProtoFieldNames:    derefBool(o.PreserveProtoFieldNames),
		}
	}

	cfg.MatchIncomingRequestRoute = derefBool(transcoder.MatchIncomingRequestRoute)
	cfg.AutoMapping = derefBool(transcoder.AutoMapping)
	cfg.IgnoreUnknownQueryParameters = derefBool(transcoder.IgnoreUnknownQueryParameters)
	cfg.ConvertGrpcStatus = derefBool(transcoder.ConvertGRPCStatus)

	return proto.ToAnyWithValidation(cfg)
}

func derefBool(b *bool) bool {
	return b != nil && *b
}

func grpcJSONTranscoderFilterName(transcoder *ir.GRPCJSONTranscoder) string {
	return perRouteFilterName(egv1a1.EnvoyFilterGRPCJSONTranscoder, transcoder.Name)
}

// patchRoute enables the route's gRPC-JSON transcoder filter instance.
func (*grpcJSONTranscoder) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	if irRoute.GRPCJSONTranscoder == nil {
		return nil
	}

	// The config lives on the HCM filter; repeating it here would copy the descriptor
	// into every route.
	return enableFilterOnRoute(route, grpcJSONTranscoderFilterName(irRoute.GRPCJSONTranscoder),
		&routev3.FilterConfig{Config: &anypb.Any{}})
}

func (*grpcJSONTranscoder) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}
