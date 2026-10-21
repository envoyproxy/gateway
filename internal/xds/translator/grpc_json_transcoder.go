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
	"k8s.io/utils/ptr"

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

// patchHCM builds and appends the gRPC-JSON transcoder filters to the HTTP Connection Manager.
// Note: this method creates a transcoder filter for each transcoder config used by the routes.
// The filter is disabled by default and enabled on the route level.
func (*grpcJSONTranscoder) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	var errs error

	for _, route := range irListener.Routes {
		transcoder := routeGRPCJSONTranscoder(route)
		if transcoder == nil {
			continue
		}

		// Only generates one transcoder filter for each unique name.
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

func buildGRPCJSONTranscoderConfig(transcoder *ir.GRPCJSONTranscoder) (*anypb.Any, error) {
	cfg := &grpcjsontranscoder.GrpcJsonTranscoder{
		DescriptorSet: &grpcjsontranscoder.GrpcJsonTranscoder_ProtoDescriptorBin{
			ProtoDescriptorBin: transcoder.ProtoDescriptorBin,
		},
		Services:                     transcoder.Services,
		IgnoreUnknownQueryParameters: ptr.Deref(transcoder.IgnoreUnknownQueryParameters, false),
		ConvertGrpcStatus:            ptr.Deref(transcoder.ConvertGRPCStatus, false),
		// The filters before the transcoder, such as JWT and ext auth, have already checked
		// the policies of this route. Matching the transcoded request to another route would
		// skip the policies of that route.
		MatchIncomingRequestRoute: true,
	}

	if o := transcoder.PrintOptions; o != nil {
		cfg.PrintOptions = &grpcjsontranscoder.GrpcJsonTranscoder_PrintOptions{
			AddWhitespace:              ptr.Deref(o.AddWhitespace, false),
			AlwaysPrintPrimitiveFields: ptr.Deref(o.AlwaysPrintPrimitiveFields, false),
			PreserveProtoFieldNames:    ptr.Deref(o.PreserveProtoFieldNames, false),
		}
	}

	return proto.ToAnyWithValidation(cfg)
}

func grpcJSONTranscoderFilterName(transcoder *ir.GRPCJSONTranscoder) string {
	return perRouteFilterName(egv1a1.EnvoyFilterGRPCJSONTranscoder, transcoder.Name)
}

// patchRoute enables the transcoder filter for the route.
func (*grpcJSONTranscoder) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if route == nil {
		return errors.New("xds route is nil")
	}
	if irRoute == nil {
		return errors.New("ir route is nil")
	}
	transcoder := routeGRPCJSONTranscoder(irRoute)
	if transcoder == nil {
		return nil
	}

	// The config is on the HCM filter, so the descriptor is not copied to every route.
	return enableFilterOnRoute(route, grpcJSONTranscoderFilterName(transcoder),
		&routev3.FilterConfig{Config: &anypb.Any{}})
}

// routeGRPCJSONTranscoder returns the transcoder of the route, or nil if the route
// doesn't need one.
// The transcoder is skipped for a direct response or a redirect, because Envoy runs the
// HTTP filters before them and the transcoder would answer first, for example with a 400
// for an invalid JSON body.
func routeGRPCJSONTranscoder(route *ir.HTTPRoute) *ir.GRPCJSONTranscoder {
	if route.DirectResponse != nil || route.Redirect != nil {
		return nil
	}
	return route.GRPCJSONTranscoder
}

func (*grpcJSONTranscoder) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}
