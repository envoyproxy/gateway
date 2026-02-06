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
)

// Creates XDS metadata from IR metadata
// When creating metadata for an XDS ingress resource (e.g. listener or filter chain), we should consider
// that changes in XDS metadata will lead to a drain and possible disruption to long-lived connections.
// Generally, most K8s resource metadata is stable since resource kind, name and namespace are immutable.
// SectionNames and Annotations are mutable, and so we treat them with caution:
// a dedicated class of annotations is used, with the user understanding that changes to these will lead to a drain
// and section names are not propagated at all, to avoid unintended drains from renames.
func buildXdsMetadata(metadata *ir.ResourceMetadata, ingressResource bool) *corev3.Metadata {
	if metadata == nil {
		return nil
	}

	resourcesList := &structpb.ListValue{}
	resourcesList.Values = append(resourcesList.Values, buildResourceMetadata(metadata, ingressResource))

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

func buildResourceMetadata(metadata *ir.ResourceMetadata, ingressResource bool) *structpb.Value {
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

	// ingress annotations are used iff metadata is created for a resource where metadata changes trigger drains
	if ingressResource {
		if len(metadata.IngressAnnotations) > 0 {
			routeResourceFields[envoyGatewayXdsMetadataKeyAnnotations] = &structpb.Value{
				Kind: &structpb.Value_StructValue{
					StructValue: mapToStruct(metadata.IngressAnnotations),
				},
			}
		}
	} else if len(metadata.Annotations) > 0 {
		routeResourceFields[envoyGatewayXdsMetadataKeyAnnotations] = &structpb.Value{
			Kind: &structpb.Value_StructValue{
				StructValue: mapToStruct(metadata.Annotations),
			},
		}
	}

	// Section names are not propagated as they are unstable and can cause drains
	if !ingressResource && metadata.SectionName != "" {
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
