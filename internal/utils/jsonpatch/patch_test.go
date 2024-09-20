// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package jsonpatch

import (
	"encoding/json"
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

const sourceDotEscape = `
   {
		"otherLevel": {
			"dot.key": "oldValue",
			"~my": "file",
			"/other/": "zip"
		}
   }
`

var expectedDotEscapeCase1 = `{
	"otherLevel": {
		"dot.key": "newValue",
		"~my": "file",
		"/other/": "zip"
	}
}`

var expectedDotEscapeCase2 = `{
	"otherLevel": {
		"dot.key": "oldValue",
		"~my": "folder",
		"/other/": "tar"
	}
}`

func TestApplyJSONPatches(t *testing.T) {
	testCases := []struct {
		doc            string
		name           string
		patchOperation []ir.JSONPatchOperation
		errorExpected  bool
		errorContains  *string
		expectedDoc    *string
	}{
		{
			name: "simple add with single patch",
			doc:  sourceDocument,
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
			doc:  sourceDocument,
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
			doc:  sourceDocument,
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
			errorContains: ptr.To("unsupported JSONPatch operation"),
		},
		{
			name: "jsonpath affecting two places",
			doc:  sourceDocument,
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
			doc:  sourceDocument,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "remove",
					JSONPath: ptr.To("i'm not a json path string"),
				},
			},
			errorExpected: true,
			errorContains: ptr.To("unable to convert jsonPath"),
		},
		{
			name: "dot escaped json path",
			doc:  sourceDotEscape,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "replace",
					JSONPath: ptr.To("$.otherLevel['dot.key']"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"newValue\""),
					},
				},
			},
			expectedDoc:   &expectedDotEscapeCase1,
			errorExpected: false,
		},
		{
			name: "dot escaped json path combined with path",
			doc:  sourceDotEscape,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "replace",
					Path:     ptr.To("dot.key"),
					JSONPath: ptr.To("$.otherLevel"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"newValue\""),
					},
				},
			},
			expectedDoc:   &expectedDotEscapeCase1,
			errorExpected: false,
		},
		{
			name: "json pointer chars which need to be escaped",
			doc:  sourceDotEscape,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "replace",
					JSONPath: ptr.To("$.otherLevel['~my']"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"folder\""),
					},
				},
				{
					Op:       "replace",
					JSONPath: ptr.To("$.otherLevel['/other/']"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"tar\""),
					},
				},
			},
			expectedDoc:   &expectedDotEscapeCase2,
			errorExpected: false,
		},
		{
			name: "jsonPath returns no jsonPointer",
			doc:  sourceDocument,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "replace",
					JSONPath: ptr.To("$.secondLevel.doesNotExist"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"folder\""),
					},
				},
			},
			errorExpected: true,
			errorContains: ptr.To("no jsonPointers were found"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			jDoc, err := ApplyJSONPatches([]byte(tc.doc), tc.patchOperation...)
			if tc.errorExpected {
				require.Error(t, err)
				if tc.errorContains != nil {
					require.ErrorContains(t, err, *tc.errorContains)
				}
			} else {
				if tc.expectedDoc != nil {
					resultData, err := jDoc.MarshalJSON()
					if err != nil {
						t.Error(err)
					}

					resultJSON, err := formatJSON(resultData)
					if err != nil {
						t.Error(err)
					}

					expectedJSON, err := formatJSON([]byte(*tc.expectedDoc))
					if err != nil {
						t.Error(err)
					}

					require.Equal(t, expectedJSON, resultJSON)
				}
				require.NoError(t, err)
			}
		})
	}
}

func formatJSON(s []byte) (string, error) {
	var obj map[string]interface{}
	err := json.Unmarshal(s, &obj)
	if err != nil {
		return "", err
	}
	buf, err := json.MarshalIndent(obj, "", "    ")
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
