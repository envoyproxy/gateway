// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package jsonpatch

import (
	"testing"

	"github.com/stretchr/testify/require"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/ir"
)

const sourceDocument = `
{ 
   "topLevel" : {
      "mapContainer" : {
          "key": "value",
		  "other": "key"
	  },
	  "arrayContainer": [
	     "str1",
		 "str2"
	  ],
	  "mapArray" : [
	     { 
		   "name": "first",
		   "key" : "value"
		 },
		 { 
		   "name": "second",
		   "key" : "other value"
		 }
	  ]
   }
}
`

func TestApplyJSONPatches(t *testing.T) {
	testCases := []struct {
		name           string
		patchOperation []ir.JSONPatchOperation
		errorExpected  bool
	}{
		{
			name: "simple add with single patch",
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:   "add",
					Path: ptr.To("/topLevel/newKey"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("true"),
					},
				},
			},
			errorExpected: false,
		},
		{
			name: "two operations in a set",
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:   "add",
					Path: ptr.To("/topLevel/newKey"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("true"),
					},
				},
				{
					Op:   "remove",
					Path: ptr.To("/topLevel/arrayContainer/1"),
				},
			},
			errorExpected: false,
		},
		{
			name: "invalid operation",
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:   "badbadbad",
					Path: ptr.To("/topLevel/newKey"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("true"),
					},
				},
			},
			errorExpected: true,
		},
		{
			name: "jsonpath affecting two places",
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "remove",
					JSONPath: ptr.To("$.topLevel.mapArray[*].key"),
				},
			},
			errorExpected: false,
		},
		{
			name: "invalid jsonpath",
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "remove",
					JSONPath: ptr.To("i'm not a json path string"),
				},
			},
			errorExpected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ApplyJSONPatches([]byte(sourceDocument), tc.patchOperation...)
			if tc.errorExpected {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
