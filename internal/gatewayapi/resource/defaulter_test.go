// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/kubectl-validate/pkg/openapiclient"
)

func TestApplyDefault(t *testing.T) {
	defaulter, err := newDefaulter(openapiclient.NewLocalCRDFiles(os.DirFS("testdata/schema")))
	require.NoError(t, err)

	testCases := []struct {
		name   string
		error  bool
		input  map[string]interface{}
		expect map[string]interface{}
	}{
		{
			name: "empty object with nested field",
			input: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"objectField": map[string]interface{}{},
				},
			},
			expect: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"stringField":  "defaultString",
					"integerField": 42.,
					"floatField":   3.14,
					"booleanField": true,
					"enumField":    "option1",
					"objectField": map[string]interface{}{
						"nestedString":  "nestedDefault",
						"nestedInteger": 10.,
					},
					"mapField": map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			error: false,
		},
		{
			name: "empty object without nested field",
			input: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{},
			},
			expect: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"stringField":  "defaultString",
					"integerField": 42.,
					"floatField":   3.14,
					"booleanField": true,
					"enumField":    "option1",
					"mapField": map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			error: false,
		},
		{
			name: "object with few field unset",
			input: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"stringField":  "exampleString",
					"booleanField": false,
					"objectField": map[string]interface{}{
						"nestedString": "nestedExample",
					},
				},
			},
			expect: map[string]interface{}{
				"apiVersion": "example.com/v1",
				"kind":       "TestCR",
				"metadata": map[string]interface{}{
					"name":      "test-cr",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"stringField":  "exampleString",
					"integerField": 42.,
					"floatField":   3.14,
					"booleanField": false,
					"enumField":    "option1",
					"objectField": map[string]interface{}{
						"nestedString":  "nestedExample",
						"nestedInteger": 10.,
					},
					"mapField": map[string]interface{}{
						"key1": "value1",
						"key2": "value2",
					},
				},
			},
			error: false,
		},
		{
			name:  "nil input",
			input: nil,
			error: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := defaulter.ApplyDefault(&unstructured.Unstructured{Object: tc.input})
			if tc.error {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, got.Object)
			}
		})
	}
}
