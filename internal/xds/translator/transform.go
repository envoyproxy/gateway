// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"encoding/json"
	"errors"
	"fmt"

	mutation_rulesv3 "github.com/envoyproxy/go-control-plane/envoy/config/common/mutation_rules/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	transformv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/transform/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/structpb"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&transform{})
}

type transform struct{}

var _ httpFilter = &transform{}

// patchHCM builds and appends the transform filter to the HCM filter chain.
// The filter is added once in disabled state and enabled per-route.
func (*transform) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if hcmContainsFilter(mgr, egv1a1.EnvoyFilterTransform.String()) {
		return nil
	}

	if !listenerContainsTransform(irListener) {
		return nil
	}

	filter, err := buildHCMTransformFilter()
	if err != nil {
		return err
	}
	mgr.HttpFilters = append(mgr.HttpFilters, filter)

	return nil
}

func buildHCMTransformFilter() (*hcmv3.HttpFilter, error) {
	transformProto := &transformv3.TransformConfig{}
	transformAny, err := proto.ToAnyWithValidation(transformProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterTransform.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: transformAny,
		},
		Disabled: true,
	}, nil
}

func (*transform) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func (*transform) patchRoute(route *routev3.Route, irRoute *ir.HTTPRoute, _ *ir.HTTPListener) error {
	if !routeContainsTransform(irRoute) {
		return nil
	}

	transformCfg, err := buildTransformPerRouteConfig(irRoute.Traffic.Transform)
	if err != nil {
		return err
	}

	if route.TypedPerFilterConfig == nil {
		route.TypedPerFilterConfig = make(map[string]*anypb.Any)
	}
	route.TypedPerFilterConfig[egv1a1.EnvoyFilterTransform.String()] = transformCfg
	return nil
}

func buildTransformPerRouteConfig(t *ir.HTTPTransform) (*anypb.Any, error) {
	cfg := &transformv3.TransformConfig{}

	if t.RequestTransformation != nil {
		transformation, err := buildTransformation(t.RequestTransformation)
		if err != nil {
			return nil, fmt.Errorf("request transformation: %w", err)
		}
		cfg.RequestTransformation = transformation
	}

	if t.ResponseTransformation != nil {
		transformation, err := buildTransformation(t.ResponseTransformation)
		if err != nil {
			return nil, fmt.Errorf("response transformation: %w", err)
		}
		cfg.ResponseTransformation = transformation
	}

	return proto.ToAnyWithValidation(cfg)
}

func buildTransformation(t *ir.HTTPTransformation) (*transformv3.Transformation, error) {
	transformation := &transformv3.Transformation{}

	for _, h := range t.SetHeaders {
		transformation.HeadersMutations = append(transformation.HeadersMutations,
			&mutation_rulesv3.HeaderMutation{
				Action: &mutation_rulesv3.HeaderMutation_Append{
					Append: &corev3.HeaderValueOption{
						Header: &corev3.HeaderValue{
							Key:   h.Name,
							Value: h.Value,
						},
						AppendAction: corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD,
					},
				},
			})
	}

	for _, h := range t.AddHeaders {
		transformation.HeadersMutations = append(transformation.HeadersMutations,
			&mutation_rulesv3.HeaderMutation{
				Action: &mutation_rulesv3.HeaderMutation_Append{
					Append: &corev3.HeaderValueOption{
						Header: &corev3.HeaderValue{
							Key:   h.Name,
							Value: h.Value,
						},
						AppendAction: corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD,
					},
				},
			})
	}

	for _, name := range t.RemoveHeaders {
		transformation.HeadersMutations = append(transformation.HeadersMutations,
			&mutation_rulesv3.HeaderMutation{
				Action: &mutation_rulesv3.HeaderMutation_Remove{
					Remove: name,
				},
			})
	}

	if t.Body != nil {
		bodyTransformation, err := buildBodyTransformation(t.Body)
		if err != nil {
			return nil, err
		}
		transformation.BodyTransformation = bodyTransformation
	}

	return transformation, nil
}

func buildBodyTransformation(b *ir.HTTPBodyTransformation) (*transformv3.BodyTransformation, error) {
	bt := &transformv3.BodyTransformation{}

	formatString := &corev3.SubstitutionFormatString{}
	if b.FormatString != nil {
		formatString.Format = &corev3.SubstitutionFormatString_TextFormat{
			TextFormat: *b.FormatString,
		}
	} else if len(b.JSONBody) > 0 {
		jsonStruct := &structpb.Struct{}
		if err := json.Unmarshal(b.JSONBody, jsonStruct); err != nil {
			return nil, fmt.Errorf("failed to parse JSON body template: %w", err)
		}
		formatString.Format = &corev3.SubstitutionFormatString_JsonFormat{
			JsonFormat: jsonStruct,
		}
	}
	bt.BodyFormat = formatString

	switch b.Action {
	case ir.BodyTransformActionReplace:
		bt.Action = transformv3.BodyTransformation_REPLACE
	default:
		bt.Action = transformv3.BodyTransformation_MERGE
	}

	return bt, nil
}

func listenerContainsTransform(irListener *ir.HTTPListener) bool {
	for _, route := range irListener.Routes {
		if routeContainsTransform(route) {
			return true
		}
	}
	return false
}

func routeContainsTransform(irRoute *ir.HTTPRoute) bool {
	return irRoute != nil &&
		irRoute.Traffic != nil &&
		irRoute.Traffic.Transform != nil
}
