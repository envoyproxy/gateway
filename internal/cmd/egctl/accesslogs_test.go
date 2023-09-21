// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	kapisv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

func newTestClient() (client.Client, error) {
	scheme, err := newScheme()
	if err != nil {
		return nil, err
	}

	cli := fakeclient.NewClientBuilder().WithScheme(scheme).Build()
	return cli, nil
}

func findEnvoyPatchPolicy(ctx context.Context, cli client.Client, key types.NamespacedName) bool {
	policy := &v1alpha1.EnvoyPatchPolicy{}
	if err := cli.Get(ctx, key, policy); err != nil {
		return false
	}
	return true
}

func TestExpectedEnvoyPatchPolicy(t *testing.T) {
	actual, err := expectedEnvoyPatchPolicy("abc", "eg", "http", "1.2.3.4", "test", 9001)
	assert.NoError(t, err)

	c, err := os.ReadFile("testdata/accesslogs/test-access-logs.yaml")
	assert.NoError(t, err)

	var expect *v1alpha1.EnvoyPatchPolicy
	err = yaml.Unmarshal(c, &expect)
	assert.NoError(t, err)

	// Compare the values of envoy patch policy separately.
	assert.Equal(t, len(expect.Spec.JSONPatches), len(actual.Spec.JSONPatches))
	n := len(actual.Spec.JSONPatches)
	for i := 0; i < n; i++ {
		actualPatchValues := actual.Spec.JSONPatches[i].Operation.Value
		expectPatchValues := expect.Spec.JSONPatches[i].Operation.Value

		actualJSON, err := actualPatchValues.MarshalJSON()
		assert.NoError(t, err)
		expectJSON, err := expectPatchValues.MarshalJSON()
		assert.NoError(t, err)

		var actualMap, expectMap map[string]interface{}
		err = json.Unmarshal(actualJSON, &actualMap)
		assert.NoError(t, err)
		err = json.Unmarshal(expectJSON, &expectMap)
		assert.NoError(t, err)

		assert.Equal(t, expectMap, actualMap)

		// Compare other field except this field.
		actual.Spec.JSONPatches[i].Operation.Value = kapisv1.JSON{}
		expect.Spec.JSONPatches[i].Operation.Value = kapisv1.JSON{}
	}

	assert.Equal(t, expect, actual)
}

func TestEnvoyPatchPolicyOperations(t *testing.T) {
	cli, err := newTestClient()
	assert.NoError(t, err)

	policy, err := expectedEnvoyPatchPolicy("abc", "eg", "http", "foo", "test", 9001)
	assert.NoError(t, err)

	ctx := context.Background()
	err = createOrUpdateEnvoyPatchPolicy(ctx, cli, policy)
	assert.NoError(t, err)
	assert.Equal(t, true, findEnvoyPatchPolicy(ctx, cli, types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Name,
	}))

	err = deleteEnvoyPatchPolicy(ctx, cli, policy)
	assert.NoError(t, err)
	assert.Equal(t, false, findEnvoyPatchPolicy(ctx, cli, types.NamespacedName{
		Namespace: policy.Namespace,
		Name:      policy.Name,
	}))
}
