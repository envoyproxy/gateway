// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package jsonpatch

import (
	"sort"
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

const case4Escaping string = `{
    "values": [{
		"name": "test1",
		"dotted.key": "Hello"
	},
	{
		"name": "test2",
		"dotted.key": "there"
	},
	{
		"name": "test3",
		"~abc": "tilde"
	},
	{
		"name": "test4",
		"//abc": "slash"
	},
	{
		"name": "test5",
		"~/abc/~": "mixed"
	}]
}`

func Test(t *testing.T) {
	testCases := []struct {
		name string

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
			name:     "TestCase-01",
			doc:      case1Simple,
			jsonPath: "$.xyz",
			expected: []string{},
		},
		{
			name:     "TestCase-02",
			doc:      case1Simple,
			jsonPath: "$.xyz",
			path:     "doesnotexist",
			expected: []string{},
		},
		{
			name:     "TestCase-03",
			doc:      case1Simple,
			jsonPath: "$.a",
			expected: []string{
				"/a",
			},
		},
		{
			name:     "TestCase-04",
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test2')]",
			expected: []string{
				"/v/1",
			},
		},
		{
			name:     "TestCase-05",
			doc:      case2Nested,
			jsonPath: "..v[?(@.x=='test1')].y",
			expected: []string{
				"/v/0/y",
			},
		},
		{
			name:     "TestCase-06",
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test2')].y",
			expected: []string{
				"/v/1/y",
			},
		},
		{
			name:     "TestCase-07",
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='test1')].y",
			expected: []string{
				"/v/0/y",
			},
		},
		{
			name:     "TestCase-08",
			doc:      case2Nested,
			jsonPath: "$.v[*].y",
			expected: []string{
				"/v/0/y",
				"/v/1/y",
			},
		},
		{
			name:     "TestCase-09",
			doc:      case2Nested,
			jsonPath: "$.v[?(@.x=='UNKNOWN')].y",
			expected: []string{},
		},
		{
			name:     "TestCase-10",
			doc:      case1Simple,
			jsonPath: ".a",
			expected: []string{
				"/a",
			},
		},
		{
			name:     "TestCase-11",
			doc:      case1Simple,
			jsonPath: "a",
			expected: []string{
				"/a",
			},
		},
		{
			name:     "TestCase-12",
			doc:      case2Nested,
			jsonPath: "f.w",
			expected: []string{
				"/f/w",
			},
		},
		{
			name:     "TestCase-13",
			doc:      case2Nested,
			jsonPath: "f.*",
			expected: []string{
				"/f/w",
				"/f/q",
				"/f/y",
			},
		},
		{
			name:     "TestCase-14",
			doc:      case2Nested,
			jsonPath: "v.*",
			expected: []string{
				"/v/0",
				"/v/1",
			},
		},
		{
			name:     "TestCase-15",
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
			name:     "TestCase-16",
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
			name:     "TestCase-17",
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
			name:     "TestCase-18",
			doc:      case2Nested,
			jsonPath: "**.y",
			expected: []string{
				"/v/0/y",
				"/v/1/y",
			},
		},
		{
			name:     "TestCase-19",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www_example_com')]",
			expected: []string{
				"/virtual_hosts/1/routes/0",
			},
		},
		{
			name:     "TestCase-20",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www_test_com')]",
			expected: []string{
				"/virtual_hosts/0/routes/0",
			},
		},
		{
			name:     "TestCase-21",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			expected: []string{
				"/virtual_hosts/0/routes/0",
				"/virtual_hosts/1/routes/0",
			},
		},
		{
			name:     "TestCase-22",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')].route.cluster",
			expected: []string{
				"/virtual_hosts/0/routes/0/route/cluster",
				"/virtual_hosts/1/routes/0/route/cluster",
			},
		},
		{
			name:     "TestCase-23",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]['route']['cluster']",
			expected: []string{
				"/virtual_hosts/0/routes/0/route/cluster",
				"/virtual_hosts/1/routes/0/route/cluster",
			},
		},
		{
			name:     "TestCase-24",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name=='httproute/default/backend/rule/1/match/1/www_example_com')].route.upgrade_configs",
			expected: []string{
				"/virtual_hosts/1/routes/0/route/upgrade_configs",
			},
		},
		{
			name:     "TestCase-25",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "/abc",
			expected: []string{
				"/virtual_hosts/0/routes/0/abc",
				"/virtual_hosts/1/routes/0/abc",
			},
		},
		{
			name:     "TestCase-26",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "abc",
			expected: []string{
				"/virtual_hosts/0/routes/0/abc",
				"/virtual_hosts/1/routes/0/abc",
			},
		},
		{
			name:     "TestCase-27",
			doc:      case3Route,
			jsonPath: "..routes[?(@.name =~ 'www')]",
			path:     "/",
			expected: []string{
				"/virtual_hosts/0/routes/0/",
				"/virtual_hosts/1/routes/0/",
			},
		},
		{
			name:     "TestCase-28",
			doc:      case4Escaping,
			jsonPath: "$.values[?(@.name =~ 'test2')]",
			path:     "dotted.key",
			expected: []string{
				"/values/1/dotted.key",
			},
		},
		{
			name:     "TestCase-29",
			doc:      case4Escaping,
			jsonPath: "$.values[?(@.name =~ 'test2')]['dotted.key']",
			expected: []string{
				"/values/1/dotted.key",
			},
		},
		{
			name:     "TestCase-30",
			doc:      case4Escaping,
			jsonPath: "$.values[?(@.name =~ 'test3')].~abc",
			expected: []string{
				"/values/2/~0abc",
			},
		},
		{
			name:     "TestCase-31",
			doc:      case4Escaping,
			jsonPath: "$.values[?(@.name =~ 'test4')]['//abc']",
			expected: []string{
				"/values/3/~1~1abc",
			},
		},
		{
			name:     "TestCase-32",
			doc:      case4Escaping,
			jsonPath: "$.values[?(@.name =~ 'test5')]['~/abc/~']",
			expected: []string{
				"/values/4/~0~1abc~1~0",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pointers, err := ConvertPathToPointers([]byte(tc.doc), tc.jsonPath, tc.path)
			if err != nil {
				require.NoError(t, err)
			}

			expectedAsString := asString(tc.expected)
			pointersAsString := asString(pointers)

			require.Equal(t, expectedAsString, pointersAsString)
		})
	}
}

func TestException(t *testing.T) {
	tests := []struct {
		name string

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
			name:     "TestCaseEx-01",
			doc:      case1Simple,
			jsonPath: ".$",
			expected: "Error during parsing jpath",
		},
		{
			name:     "TestCaseEx-02",
			doc:      case1Simple,
			jsonPath: "$",
			expected: "only Root",
		},
		{
			name:     "TestCaseEx-03",
			doc:      "{",
			jsonPath: ".$",
			expected: "Error during parsing json",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := ConvertPathToPointers([]byte(test.doc), test.jsonPath, test.path)
			if err == nil {
				require.Error(t, err)
			}
			require.ErrorContains(t, err, test.expected)
		})
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
