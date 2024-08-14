// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"sort"
	"strconv"
	"testing"

	"github.com/ohler55/ojg/jp"
	"github.com/stretchr/testify/require"
)

const case1Simple string = `{
			"a": "b"
		}`

const case2Nested string = `{
			"a": "b",
			"v": [{
				"x": "test1",
				"y": "hello"
			},
			{
				"x": "test2",
				"y": "world"
			}],
			"f":{
				"w": "hi",
				"q": "welcome",
				"y": "ciao"
			},
			"y": "c"
		}`

const case3Route string = `{
  "name": "default/eg/http",
  "virtual_hosts": [
   {
      "name": "default/eg/http/www_test_com",
      "domains": [
        "www.test.com"
      ],
      "routes": [
        {
          "name": "httproute/default/backend/rule/0/match/0/www_test_com",
          "match": {
            "prefix": "/"
          },
          "route": {
            "cluster": "httproute/default/backend/rule/0",
            "upgrade_configs": [
              {
                "upgrade_type": "websocket"
              }
            ]
          }
        }
      ]
    },
    {
      "name": "default/eg/http/www_example_com",
      "domains": [
        "www.example.com"
      ],
      "routes": [
        {
          "name": "httproute/default/backend/rule/1/match/1/www_example_com",
          "match": {
            "prefix": "/"
          },
          "route": {
            "cluster": "httproute/default/backend/rule/1",
            "upgrade_configs": [
              {
                "upgrade_type": "websocket"
              }
            ]
          }
        }
      ]
    }
  ],
  "ignore_port_in_host_matching": true
}`

func Test(t *testing.T) {
	tests := []struct {
		// Json Document
		doc string

		// JSONPath
		jsonPath string

		// path
		path string

		// List of expected pointers
		expected []string
	}{
		{
			doc:      case1Simple,
			jsonPath: "$.a",
			expected: []string{
				"/a",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test2')]",
			expected: []string{
				"/v/1",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "..v[?(@.x=='test1')].y",
			expected: []string{
				"/v/0/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test2')].y",
			expected: []string{
				"/v/1/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test1')].y",
			expected: []string{
				"/v/0/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$.v[*].y",
			expected: []string{
				"/v/0/y",
				"/v/1/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='UNKNOWN')].y",
			expected: []string{},
		},
		{
			doc:      case1Simple,
			jsonPath: ".a",
			expected: []string{
				"/a",
			},
		},
		{
			doc:      case1Simple,
			jsonPath: "a",
			expected: []string{
				"/a",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "f.w",
			expected: []string{
				"/f/w",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "f.*",
			expected: []string{
				"/f/w",
				"/f/q",
				"/f/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "v.*",
			expected: []string{
				"/v/0",
				"/v/1",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "v.**",
			expected: []string{
				"/v/0/x",
				"/v/0/y",
				"/v/1/x",
				"/v/1/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "$..y",
			expected: []string{
				"/f/y",
				"/v/0/y",
				"/v/1/y",
				"/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "..y",
			expected: []string{
				"/f/y",
				"/v/0/y",
				"/v/1/y",
				"/y",
			},
		},
		{
			doc:      case2Nested,
			jsonPath: "**.y",
			expected: []string{
				"/v/0/y",
				"/v/1/y",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www_example_com')]",
			expected: []string{
				"/virtual_hosts/1/routes/0",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www_test_com')]",
			expected: []string{
				"/virtual_hosts/0/routes/0",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			expected: []string{
				"/virtual_hosts/0/routes/0",
				"/virtual_hosts/1/routes/0",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')].route.cluster",
			expected: []string{
				"/virtual_hosts/0/routes/0/route/cluster",
				"/virtual_hosts/1/routes/0/route/cluster",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]['route']['cluster']",
			expected: []string{
				"/virtual_hosts/0/routes/0/route/cluster",
				"/virtual_hosts/1/routes/0/route/cluster",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name=='httproute/default/backend/rule/1/match/1/www_example_com')].route.upgrade_configs",
			expected: []string{
				"/virtual_hosts/1/routes/0/route/upgrade_configs",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "/abc",
			expected: []string{
				"/virtual_hosts/0/routes/0/abc",
				"/virtual_hosts/1/routes/0/abc",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "abc",
			expected: []string{
				"/virtual_hosts/0/routes/0/abc",
				"/virtual_hosts/1/routes/0/abc",
			},
		},
		{
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "/",
			expected: []string{
				"/virtual_hosts/0/routes/0/",
				"/virtual_hosts/1/routes/0/",
			},
		},
	}

	for i, test := range tests {

		testCasePrefix := "TestCase " + strconv.Itoa(i+1)
		pointers, err := ConvertPathToPointers([]byte(test.doc), test.jsonPath, test.path)
		if err != nil {
			t.Error(testCasePrefix + ": Error during conversion:\n" + err.Error())
			continue
		}

		expectedAsString := asString(test.expected)
		pointersAsString := asString(pointers)

		require.Equal(t, expectedAsString, pointersAsString)
	}
}

func TestException(t *testing.T) {
	tests := []struct {
		// Json Document
		doc string

		// JSONPath
		jsonPath string

		// path
		path string

		// expected exception
		expected string
	}{
		{
			doc:      case1Simple,
			jsonPath: ".$",
			expected: "Error during parsing jpath",
		},
		{
			doc:      case1Simple,
			jsonPath: "$",
			expected: "only Root",
		},
		{
			doc:      "{",
			jsonPath: ".$",
			expected: "Error during parsing json",
		},
	}

	for i, test := range tests {

		testCasePrefix := "TestCase " + strconv.Itoa(i+1)
		_, err := ConvertPathToPointers([]byte(test.doc), test.jsonPath, test.path)
		if err == nil {
			t.Error(testCasePrefix + ": Error expected, but no error found!")
			continue
		}

		require.ErrorContains(t, err, test.expected)
	}
}

func TestUnexpectedFrag(t *testing.T) {
	expr := jp.Expr{}
	expr = append(expr, jp.Union{})

	_, err := expToPointer(expr)
	if err == nil {
		t.Error("Error expected, but no error found!")
	}

	require.ErrorContains(t, err, "There is no conversion implemented for Union")
}

func TestNegativeNth(t *testing.T) {
	result, err := nthToPointer(jp.Nth(-1))
	if err != nil {
		t.Error(err)
	}
	test := string(result)
	if test != "-1" {
		t.Error("expected -1, but was " + test + "!")
	}
}

func asString(values []string) string {
	var buf []byte

	sort.Strings(values)
	for _, v := range values {
		buf = append(buf, []byte(v)...)
		buf = append(buf, []byte("\n")...)
	}

	return string(buf)
}
