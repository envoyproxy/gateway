// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

func TestGetVersionedCreateStrategyForCRD(t *testing.T) {
	var testCRD apiextensionsv1.CustomResourceDefinition
	crdBytes, err := os.ReadFile("testdata/test_crd.yaml")
	require.NoError(t, err)

	err = yaml.Unmarshal(crdBytes, &testCRD)
	require.NoError(t, err)

	createStrategy, err := getVersionedCreateStrategyForCRD(&testCRD)
	require.NoError(t, err)

	testCases := []struct {
		name string
		obj  map[string]interface{}
		err  []string
	}{
		{
			name: "empty object",
			obj:  map[string]interface{}{},
			err: []string{
				"apiVersion: Invalid value: \"\": must be foobar/v1",
			},
		},
		{
			name: "valid object",
			obj: map[string]interface{}{
				"apiVersion": "foobar/v1",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"testEnum": []string{
						"aaa",
					},
					"endpoints": []interface{}{
						map[string]interface{}{
							"ip": map[string]interface{}{
								"address": "1.1.1.1",
								"port":    1234,
							},
						},
					},
				},
			},
		},
		{
			name: "invalid object",
			obj: map[string]interface{}{
				"apiVersion": "foobar/v1",
				"metadata": map[string]interface{}{
					"name": "test",
				},
				"spec": map[string]interface{}{
					"testEnum": []string{
						"ddd",
					},
					"endpoints": []interface{}{
						map[string]interface{}{
							"ip": map[string]interface{}{
								"address": "1.1.1.1",
								"port":    1234,
							},
							"unix": map[string]interface{}{
								"path": "test.sock",
							},
						},
					},
				},
			},
			err: []string{
				"metadata.namespace: Required value",
				"spec.testEnum[0]: Unsupported value: \"ddd\": supported values: \"aaa\", \"bbb\", \"ccc\"",
				"<nil>: Invalid value: \"null\": some validation rules were not checked because the object was invalid; correct the existing errors to complete validation",
			},
		},
		{
			name: "another invalid object",
			obj: map[string]interface{}{
				"apiVersion": "foobar/v1",
				"metadata": map[string]interface{}{
					"name":      "test",
					"namespace": "default",
				},
				"spec": map[string]interface{}{
					"testEnum": []string{
						"ccc",
					},
					"endpoints": []interface{}{
						map[string]interface{}{
							"ip": map[string]interface{}{
								"address": "1.1.1.1",
								"port":    66666,
							},
							"unix": map[string]interface{}{
								"path": "test.sock",
							},
						},
					},
				},
			},
			err: []string{
				"spec.endpoints[0].ip.port: Invalid value: 66666: spec.endpoints[0].ip.port in body should be less than or equal to 65535",
				"spec.endpoints[0]: Invalid value: \"object\": only one of fqdn, ip or unix can be specified",
			},
		},
	}

	ctx := context.Background()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			errList := createStrategy.Validate(ctx, &unstructured.Unstructured{
				Object: tc.obj,
			})
			if len(tc.err) > 0 {
				require.Error(t, errList.ToAggregate())

				errs := make([]string, len(errList.ToAggregate().Errors()))
				for i, err := range errList.ToAggregate().Errors() {
					errs[i] = err.Error()
				}
				require.Exactly(t, tc.err, errs)
			} else {
				require.Nil(t, errList)
			}
		})
	}
}

func TestGetCreateStrategyMapForCRDs(t *testing.T) {
	createStrategyMap := getCreateStrategyMapForCRDs()
	require.NotEqual(t, 0, len(createStrategyMap))
}
