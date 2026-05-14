// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func TestGetAncestorRefForEnvoyPatchPolicyTargetRef_UsesTargetFields(t *testing.T) {
	ref := gwapiv1.LocalPolicyTargetReference{
		Group: gwapiv1.Group("example.io"),
		Kind:  gwapiv1.Kind("ExampleKind"),
		Name:  gwapiv1.ObjectName("example-name"),
	}

	ancestor := getAncestorRefForEnvoyPatchPolicyTargetRef(ref)
	require.NotNil(t, ancestor.Group)
	require.NotNil(t, ancestor.Kind)
	assert.Equal(t, gwapiv1.Group("example.io"), *ancestor.Group)
	assert.Equal(t, gwapiv1.Kind("ExampleKind"), *ancestor.Kind)
	assert.Equal(t, gwapiv1.ObjectName("example-name"), ancestor.Name)
}
