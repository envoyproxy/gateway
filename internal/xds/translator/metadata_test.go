// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestBuildXdsMetadata(t *testing.T) {
	in := &ir.ResourceMetadata{
		Kind:        "HTTPRoute",
		Name:        "my-route",
		Namespace:   "default",
		Annotations: ir.MapToSlice(map[string]string{"foo": "bar"}),
		SectionName: "section-1",
		Policies: []*ir.PolicyMetadata{
			{
				Kind:      "BackendTrafficPolicy",
				Name:      "rl-1",
				Namespace: "pol-ns",
			},
		},
	}

	md := buildXdsMetadata(in, false)
	require.NotNil(t, md, "expected metadata, got nil")

	fm, ok := md.FilterMetadata[envoyGatewayXdsMetadataNamespace]
	require.Truef(t, ok, "expected filter metadata %q to exist", envoyGatewayXdsMetadataNamespace)

	// resources
	resVal, ok := fm.Fields[envoyGatewayMetadataKeyResources]
	require.Truef(t, ok, "expected %q field in filter metadata", envoyGatewayMetadataKeyResources)
	resList := resVal.GetListValue()
	require.NotNil(t, resList, "expected non-empty resources list")
	require.NotEmpty(t, resList.Values, "expected non-empty resources list")
	resStruct := resList.Values[0].GetStructValue()
	require.NotNil(t, resStruct, "expected struct for resource value")

	require.Equal(t, in.Kind, resStruct.Fields[envoyGatewayXdsMetadataKeyKind].GetStringValue(), "kind mismatch")
	require.Equal(t, in.Name, resStruct.Fields[envoyGatewayXdsMetadataKeyName].GetStringValue(), "name mismatch")
	require.Equal(t, in.Namespace, resStruct.Fields[envoyGatewayXdsMetadataKeyNamespace].GetStringValue(), "namespace mismatch")
	require.Equal(t, in.SectionName, resStruct.Fields[envoyGatewayXdsMetadataKeySectionName].GetStringValue(), "sectionName mismatch")

	annStruct := resStruct.Fields[envoyGatewayXdsMetadataKeyAnnotations].GetStructValue()
	require.NotNil(t, annStruct, "expected annotations struct")
	require.Equal(t, "bar", annStruct.Fields["foo"].GetStringValue(), "annotation foo mismatch")

	// policies
	polVal, ok := fm.Fields[envoyGatewayMetadataKeyPolicies]
	require.Truef(t, ok, "expected %q field in filter metadata for policies", envoyGatewayMetadataKeyPolicies)
	polList := polVal.GetListValue()
	require.NotNil(t, polList, "expected policies list")
	require.Len(t, polList.Values, len(in.Policies), "expected policies list length")

	polStruct := polList.Values[0].GetStructValue()
	require.NotNil(t, polStruct, "expected struct for policy value")
	require.Equal(t, in.Policies[0].Kind, polStruct.Fields[envoyGatewayXdsMetadataKeyKind].GetStringValue(), "policy kind mismatch")
	require.Equal(t, in.Policies[0].Name, polStruct.Fields[envoyGatewayXdsMetadataKeyName].GetStringValue(), "policy name mismatch")
	require.Equal(t, in.Policies[0].Namespace, polStruct.Fields[envoyGatewayXdsMetadataKeyNamespace].GetStringValue(), "policy namespace mismatch")
}

func Test_buildResourceMetadataIngressResource(t *testing.T) {
	type args struct {
		metadata        *ir.ResourceMetadata
		ingressResource bool
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "ingress annotations",
			args: args{
				metadata: &ir.ResourceMetadata{
					Kind:               "HTTPRoute",
					Name:               "my-route",
					Namespace:          "default",
					Annotations:        ir.MapToSlice(map[string]string{"foo": "bar"}),
					IngressAnnotations: ir.MapToSlice(map[string]string{"foobar": "barfoo"}),
					SectionName:        "section-1",
				},
				ingressResource: true,
			},
			want: map[string]string{"foobar": "barfoo"},
		},
		{
			name: "ingress resource without ingress annotations",
			args: args{
				metadata: &ir.ResourceMetadata{
					Kind:        "HTTPRoute",
					Name:        "my-route",
					Namespace:   "default",
					Annotations: ir.MapToSlice(map[string]string{"foo": "bar"}),
					SectionName: "section-1",
				},
				ingressResource: true,
			},
			want: map[string]string{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildXdsMetadata(tt.args.metadata, tt.args.ingressResource)

			fm, ok := got.FilterMetadata[envoyGatewayXdsMetadataNamespace]
			require.Truef(t, ok, "expected filter metadata %q to exist", envoyGatewayXdsMetadataNamespace)

			// resources
			resVal, ok := fm.Fields[envoyGatewayMetadataKeyResources]
			require.Truef(t, ok, "expected %q field in filter metadata", envoyGatewayMetadataKeyResources)
			resList := resVal.GetListValue()
			require.NotNil(t, resList, "expected non-empty resources list")
			require.NotEmpty(t, resList.Values, "expected non-empty resources list")
			resStruct := resList.Values[0].GetStructValue()
			require.NotNil(t, resStruct, "expected struct for resource value")

			require.Nil(t, resStruct.Fields[envoyGatewayXdsMetadataKeySectionName], "unexpected sectionName found")

			annStruct := resStruct.Fields[envoyGatewayXdsMetadataKeyAnnotations].GetStructValue()
			if len(tt.want) == 0 {
				require.Nil(t, annStruct, "expected nil annotations struct")
			} else {
				require.NotNil(t, annStruct, "expected annotations struct")
				require.Len(t, annStruct.Fields, len(tt.want), "unexpected annotation list length")
				for k, v := range tt.want {
					require.Equal(t, v, annStruct.Fields[k].GetStringValue(), "annotation mismatch")
				}
			}
		})
	}
}
