// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"encoding/json"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/api/v1alpha1"
)

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

		actualJson, err := actualPatchValues.MarshalJSON()
		assert.NoError(t, err)
		expectJson, err := expectPatchValues.MarshalJSON()
		assert.NoError(t, err)

		var actualMap, expectMap map[string]interface{}
		err = json.Unmarshal(actualJson, &actualMap)
		assert.NoError(t, err)
		err = json.Unmarshal(expectJson, &expectMap)
		assert.NoError(t, err)

		assert.Equal(t, expectMap, actualMap)

		// Compare other field except this field.
		actual.Spec.JSONPatches[i].Operation.Value = v1.JSON{}
		expect.Spec.JSONPatches[i].Operation.Value = v1.JSON{}
	}

	assert.Equal(t, expect, actual)
}
