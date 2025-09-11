// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	mutation_rulesv3 "github.com/envoyproxy/go-control-plane/envoy/config/common/mutation_rules/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	mutationv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/header_mutation/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func init() {
	registerHTTPFilter(&headerMutation{})
}

type headerMutation struct{}

var _ httpFilter = &headerMutation{}

// patchHCM builds and appends the header mutation filter to the HTTP Connection Manager
// if applicable, and it does not already exist.
func (*headerMutation) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	if hcmContainsFilter(mgr, egv1a1.EnvoyFilterHeaderMutation.String()) {
		return nil
	}

	filter, err := buildHeaderMutationFilter(irListener.Headers)
	if err != nil {
		return err
	}
	if filter != nil {
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
	}

	return nil
}

func (*headerMutation) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}

func (*headerMutation) patchRoute(*routev3.Route, *ir.HTTPRoute, *ir.HTTPListener) error {
	return nil
}

func buildHeaderMutationFilter(headers *ir.HeaderSettings) (*hcmv3.HttpFilter, error) {
	if headers == nil {
		return nil, nil
	}

	addHeaders := headers.LateAddResponseHeaders
	removeHeaders := headers.LateRemoveResponseHeaders
	if len(addHeaders) == 0 && len(removeHeaders) == 0 {
		return nil, nil
	}

	mutationProto := &mutationv3.HeaderMutation{
		Mutations: &mutationv3.Mutations{
			ResponseMutations: buildHeaderMutationRules(addHeaders, removeHeaders),
		},
	}

	mutationAny, err := proto.ToAnyWithValidation(mutationProto)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterHeaderMutation.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: mutationAny,
		},
	}, nil
}

func buildHeaderMutationRules(addHeaders []ir.AddHeader, removeHeaders []string) []*mutation_rulesv3.HeaderMutation {
	var mutationRules []*mutation_rulesv3.HeaderMutation

	for _, header := range addHeaders {
		var appendAction corev3.HeaderValueOption_HeaderAppendAction
		if header.Append {
			appendAction = corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		} else {
			appendAction = corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		}
		// Allow empty headers to be set, but don't add the config to do so unless necessary
		if len(header.Value) == 0 {
			mutationRules = append(mutationRules, &mutation_rulesv3.HeaderMutation{
				Action: &mutation_rulesv3.HeaderMutation_Append{
					Append: &corev3.HeaderValueOption{
						Header: &corev3.HeaderValue{
							Key: header.Name,
						},
						AppendAction:   appendAction,
						KeepEmptyValue: true,
					},
				},
			})
		} else {
			for _, val := range header.Value {
				mutationRules = append(mutationRules, &mutation_rulesv3.HeaderMutation{
					Action: &mutation_rulesv3.HeaderMutation_Append{
						Append: &corev3.HeaderValueOption{
							Header: &corev3.HeaderValue{
								Key:   header.Name,
								Value: val,
							},
							AppendAction:   appendAction,
							KeepEmptyValue: val == "",
						},
					},
				})
			}
		}
	}

	for _, header := range removeHeaders {
		mr := &mutation_rulesv3.HeaderMutation{
			Action: &mutation_rulesv3.HeaderMutation_Remove{
				Remove: header,
			},
		}
		mutationRules = append(mutationRules, mr)
	}

	return mutationRules
}
