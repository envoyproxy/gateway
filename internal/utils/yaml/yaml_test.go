// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package yaml

import (
	"reflect"
	"testing"
)

func TestMergeYAML(t *testing.T) {
	tests := []struct {
		name  string
		yaml1 string
		yaml2 string
		want  string
	}{
		{
			name: "test1",
			yaml1: `
a: a
b: 
  c:
    d: d
e:
  f:
  - g
k:
  l: l
`,
			yaml2: `
a: a1
b: 
  c:
    d: d1
e:
  f:
  - h
i:
  j: j
`,
			want: `a: a1
b:
  c:
    d: d1
e:
  f:
  - g
  - h
i:
  j: j
k:
  l: l
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := MergeYAML(tt.yaml1, tt.yaml2)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MergeYAML() got = %v, want %v", got, tt.want)
			}
		})
	}
}
