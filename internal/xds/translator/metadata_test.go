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

	md := buildXdsMetadata(in)
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
