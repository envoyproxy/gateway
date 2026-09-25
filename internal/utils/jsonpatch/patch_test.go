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
					Path: new("/topLevel/newKey"),
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
					Path: new("/topLevel/newKey"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("true"),
					},
				},
				{
					Op:   "remove",
					Path: new("/topLevel/arrayContainer/1"),
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
					Path: new("/topLevel/newKey"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("true"),
					},
				},
			},
			errorExpected: true,
			errorContains: new("unsupported JSONPatch operation"),
		},
		{
			name: "jsonpath affecting two places",
			doc:  sourceDocument,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "remove",
					JSONPath: new("$.topLevel.mapArray[*].key"),
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
					JSONPath: new("i'm not a json path string"),
				},
			},
			errorExpected: true,
			errorContains: new("unable to convert jsonPath"),
		},
		{
			name: "dot escaped json path",
			doc:  sourceDotEscape,
			patchOperation: []ir.JSONPatchOperation{
				{
					Op:       "replace",
					JSONPath: new("$.otherLevel['dot.key']"),
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
					Path:     new("dot.key"),
					JSONPath: new("$.otherLevel"),
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
					JSONPath: new("$.otherLevel['~my']"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"folder\""),
					},
				},
				{
					Op:       "replace",
					JSONPath: new("$.otherLevel['/other/']"),
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
					JSONPath: new("$.secondLevel.doesNotExist"),
					Value: &apiextensionsv1.JSON{
						Raw: []byte("\"folder\""),
					},
				},
			},
			errorExpected: true,
			errorContains: new("no jsonPointers were found"),
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

					require.JSONEq(t, expectedJSON, resultJSON)
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

func TestApplyJSONPatchesMultipleMatches(t *testing.T) {
	for _, tc := range []struct {
		name          string
		doc           string
		patches       []ir.JSONPatchOperation
		want          string
		errorContains string
	}{
		{
			name: "later JSONPath sees earlier changes",
			doc:  `{"items":[{},{}]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "add", JSONPath: new("$.items[*]"), Path: new("/enabled"), Value: &apiextensionsv1.JSON{Raw: []byte("true")}},
				{Op: "add", JSONPath: new("$.items[?(@.enabled == true)]"), Path: new("/selected"), Value: &apiextensionsv1.JSON{Raw: []byte("true")}},
			},
			want: `{"items":[{"enabled":true,"selected":true},{"enabled":true,"selected":true}]}`,
		},
		{
			name: "array removals preserve match order",
			doc:  `{"values":[0,1,2,3]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "remove", JSONPath: new("$.values[0,1]")},
			},
			want: `{"values":[1,3]}`,
		},
		{
			name: "array insertions preserve match order",
			doc:  `{"values":[0,1]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "add", JSONPath: new("$.values[0,1]"), Value: &apiextensionsv1.JSON{Raw: []byte("9")}},
			},
			want: `{"values":[9,9,0,1]}`,
		},
		{
			name: "move reads the source after each operation",
			doc:  `{"values":[0,1,2,3]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "move", JSONPath: new("$.values[2,3]"), From: new("/values/0")},
			},
			want: `{"values":[2,0,3,1]}`,
		},
		{
			name: "copy into multiple objects",
			doc:  `{"source":{"value":1},"items":[{},{}]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "copy", JSONPath: new("$.items[*]"), Path: new("/copy"), From: new("/source")},
			},
			want: `{"source":{"value":1},"items":[{"copy":{"value":1}},{"copy":{"value":1}}]}`,
		},
		{
			name: "validation failure does not skip later patches",
			doc:  `{"items":[{},{}]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "invalid", Path: new("/unused")},
				{Op: "add", JSONPath: new("$.items[*]"), Path: new("/value"), Value: &apiextensionsv1.JSON{Raw: []byte("1")}},
			},
			want:          `{"items":[{"value":1},{"value":1}]}`,
			errorContains: "unsupported JSONPatch operation",
		},
		{
			name: "remove failure after an earlier match succeeds",
			doc:  `{"items":[{"value":1},{},{"value":3}]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "remove", JSONPath: new("$.items[*]"), Path: new("/value")},
			},
			errorContains: "/items/1/value",
		},
		{
			name: "test failure after an earlier match succeeds",
			doc:  `{"values":[1,2,1]}`,
			patches: []ir.JSONPatchOperation{
				{Op: "test", JSONPath: new("$.values[*]"), Value: &apiextensionsv1.JSON{Raw: []byte("1")}},
			},
			errorContains: "/values/1",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.doc)
			got, err := ApplyJSONPatches(input, tc.patches...)
			if tc.errorContains != "" {
				require.ErrorContains(t, err, tc.errorContains)
			} else {
				require.NoError(t, err)
			}
			if tc.want == "" {
				require.Empty(t, got)
			} else {
				require.JSONEq(t, tc.want, string(got))
			}
			require.Equal(t, tc.doc, string(input), "input must remain unchanged")
		})
	}
}
