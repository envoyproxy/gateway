// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package slice

import (
	"reflect"
	"testing"
)

func TestContainsString(t *testing.T) {
	src := []string{"aa", "bb", "cc"}

	if !ContainsString(src, "bb") {
		t.Errorf("ContainsString didn't find the string as expected")
	}

	src = make([]string, 0)
	if ContainsString(src, "") {
		t.Errorf("The result returned is not the expected result")
	}
}

func TestRemoveString(t *testing.T) {
	tests := []struct {
		testName string
		input    []string
		remove   string
		want     []string
	}{
		{
			testName: "Nil input slice",
			input:    nil,
			remove:   "",
			want:     nil,
		},
		{
			testName: "Remove a string from input slice",
			input:    []string{"a", "ab", "cdef"},
			remove:   "ab",
			want:     []string{"a", "cdef"},
		},
		{
			testName: "Slice doesn't contain the string",
			input:    []string{"a", "ab", "cdef"},
			remove:   "NotPresentInSlice",
			want:     []string{"a", "ab", "cdef"},
		},
		{
			testName: "All strings removed, result is nil",
			input:    []string{"a"},
			remove:   "a",
			want:     nil,
		},
	}
	for _, tt := range tests {
		if got := RemoveString(tt.input, tt.remove); !reflect.DeepEqual(got, tt.want) {
			t.Errorf("%v: RemoveString(%v, %q) = %v WANT %v", tt.testName, tt.input, tt.remove, got, tt.want)
		}
	}
}
