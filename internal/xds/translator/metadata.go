// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	envoyGatewayXdsMetadataNamespace      = "envoy-gateway"
	envoyGatewayXdsMetadataKeyKind        = "kind"
	envoyGatewayXdsMetadataKeyName        = "name"
	envoyGatewayXdsMetadataKeyNamespace   = "namespace"
	envoyGatewayXdsMetadataKeyAnnotations = "annotations"
	envoyGatewayXdsMetadataKeySectionName = "sectionName"
	envoyGatewayMetadataKeyResources      = "resources"
	envoyGatewayMetadataKeyPolicies       = "policies"
	envoyGatewayMetadataKeyBackend        = "backend"
)

func buildXdsMetadata(metadata *ir.ResourceMetadata) *corev3.Metadata {
	if metadata == nil {
		return nil
	}

	resourcesList := &structpb.ListValue{}
	resourcesList.Values = append(resourcesList.Values, buildResourceMetadata(metadata))

	md := &corev3.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			envoyGatewayXdsMetadataNamespace: {
				Fields: map[string]*structpb.Value{
					envoyGatewayMetadataKeyResources: {
						Kind: &structpb.Value_ListValue{
							ListValue: resourcesList,
						},
					},
				},
			},
		},
	}

	policyList := &structpb.ListValue{}

	for _, policy := range metadata.Policies {
		policyList.Values = append(policyList.Values, buildpolicyMetadata(policy))
	}

	if len(policyList.Values) > 0 {
		md.FilterMetadata[envoyGatewayXdsMetadataNamespace].Fields[envoyGatewayMetadataKeyPolicies] = &structpb.Value{
			Kind: &structpb.Value_ListValue{
				ListValue: policyList,
			},
		}
	}

	return md
}

// buildXdsEndpointMetadata builds lb_endpoint-level metadata from a DestinationSetting's ResourceMetadata.
// It emits a nested "backend" struct (matching Gateway API backendRef terminology) so individual fields
// are reachable in Envoy access log format strings, e.g: %METADATA(UPSTREAM_HOST:envoy-gateway:backend:name)%
func buildXdsEndpointMetadata(metadata *ir.ResourceMetadata) *corev3.Metadata {
	if metadata == nil {
		return nil
	}
	backendFields := map[string]*structpb.Value{
		envoyGatewayXdsMetadataKeyKind:        structpb.NewStringValue(metadata.Kind),
		envoyGatewayXdsMetadataKeyName:        structpb.NewStringValue(metadata.Name),
		envoyGatewayXdsMetadataKeyNamespace:   structpb.NewStringValue(metadata.Namespace),
		envoyGatewayXdsMetadataKeySectionName: structpb.NewStringValue(metadata.SectionName),
	}
	return &corev3.Metadata{
		FilterMetadata: map[string]*structpb.Struct{
			envoyGatewayXdsMetadataNamespace: {
				Fields: map[string]*structpb.Value{
					envoyGatewayMetadataKeyBackend: {
						Kind: &structpb.Value_StructValue{
							StructValue: &structpb.Struct{Fields: backendFields},
						},
					},
				},
			},
		},
	}
}

func buildpolicyMetadata(md *ir.PolicyMetadata) *structpb.Value {
	routeResourceFields := map[string]*structpb.Value{
		envoyGatewayXdsMetadataKeyKind: {
			Kind: &structpb.Value_StringValue{
				StringValue: md.Kind,
			},
		},
		envoyGatewayXdsMetadataKeyName: {
			Kind: &structpb.Value_StringValue{
				StringValue: md.Name,
			},
		},
		envoyGatewayXdsMetadataKeyNamespace: {
			Kind: &structpb.Value_StringValue{
				StringValue: md.Namespace,
			},
		},
	}

	return &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: routeResourceFields,
			},
		},
	}
}

func buildResourceMetadata(metadata *ir.ResourceMetadata) *structpb.Value {
	routeResourceFields := map[string]*structpb.Value{
		envoyGatewayXdsMetadataKeyKind: {
			Kind: &structpb.Value_StringValue{
				StringValue: metadata.Kind,
			},
		},
		envoyGatewayXdsMetadataKeyName: {
			Kind: &structpb.Value_StringValue{
				StringValue: metadata.Name,
			},
		},
		envoyGatewayXdsMetadataKeyNamespace: {
			Kind: &structpb.Value_StringValue{
				StringValue: metadata.Namespace,
			},
		},
	}

	if len(metadata.Annotations) > 0 {
		routeResourceFields[envoyGatewayXdsMetadataKeyAnnotations] = &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: mapToStruct(metadata.Annotations),
			},
		}
	}

	if metadata.SectionName != "" {
		routeResourceFields[envoyGatewayXdsMetadataKeySectionName] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: metadata.SectionName,
			},
		}
	}

	routeResourceValue := &structpb.Value{
		Kind: &structpb.Value_StructValue{
			StructValue: &structpb.Struct{
				Fields: routeResourceFields,
			},
		},
	}
	return routeResourceValue
}

func mapToStruct(data []ir.MapEntry) *structpb.Struct {
	fields := make(map[string]*structpb.Value)
	for _, entry := range data {
		fields[entry.Key] = &structpb.Value{
			Kind: &structpb.Value_StringValue{
				StringValue: entry.Value,
			},
		}
	}

	return &structpb.Struct{
		Fields: fields,
	}
}
