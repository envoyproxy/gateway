// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package field

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSetValue(t *testing.T) {
	testcases := []struct {
		name        string
		input       any
		fieldName   string
		fieldValue  any
		expect      any
		expectedErr bool
	}{{
		name:        "field name cannot be empty",
		input:       "",
		expectedErr: true,
	},
		{
			name:        "input cannot be a string",
			input:       "",
			fieldName:   "K",
			expectedErr: true,
		}, {
			name:        "input cannot be a struct",
			input:       struct{}{},
			fieldName:   "K",
			expectedErr: true,
		}, {
			name: "field cannot be unexported",
			input: &struct {
				name string
			}{name: "test"},
			expect: &struct {
				name string
			}{name: "test1"},
			expectedErr: true,
			fieldName:   "name",
			fieldValue:  "test1",
		}, {
			name: "simple struct set value",
			input: &struct {
				Name string
			}{Name: "test"},
			expect: &struct {
				Name string
			}{Name: "test1"},
			expectedErr: false,
			fieldName:   "Name",
			fieldValue:  "test1",
		}, {
			name: "simple recursive struct set value",
			input: &struct {
				Name  string
				Child struct {
					Name string
				}
			}{Name: "test", Child: struct{ Name string }{
				Name: "test",
			}},
			expect: &struct {
				Name  string
				Child struct {
					Name string
				}
			}{Name: "test1", Child: struct{ Name string }{
				Name: "test1",
			}},
			expectedErr: false,
			fieldName:   "Name",
			fieldValue:  "test1",
		}, {
			name: "simple recursive child struct in slice set value",
			input: &struct {
				Name  string
				Child []struct {
					Name string
				}
			}{Name: "test", Child: []struct{ Name string }{
				{Name: "test"},
				{Name: "test"},
			}},
			expect: &struct {
				Name  string
				Child []struct {
					Name string
				}
			}{Name: "test1", Child: []struct{ Name string }{
				{Name: "test1"},
				{Name: "test1"},
			}},
			expectedErr: false,
			fieldName:   "Name",
			fieldValue:  "test1",
		}, {
			name: "simple recursive child struct in map set value",
			input: &struct {
				Name  string
				Child map[string]*struct {
					Name string
				}
			}{Name: "test", Child: map[string]*struct{ Name string }{
				"test":  {Name: "test"},
				"test1": {Name: "test"},
			}},
			expect: &struct {
				Name  string
				Child map[string]*struct {
					Name string
				}
			}{Name: "test1", Child: map[string]*struct{ Name string }{
				"test":  {Name: "test1"},
				"test1": {Name: "test1"},
			}},
			expectedErr: false,
			fieldName:   "Name",
			fieldValue:  "test1",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			err := SetValue(tc.input, tc.fieldName, tc.fieldValue)
			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expect, tc.input)
			}
		})
	}
}
