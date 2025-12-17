// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/utils/file"
	"github.com/envoyproxy/gateway/internal/utils/test"
)

func TestIterYAMLBytes(t *testing.T) {
	inputs := `test: foo1
---
test: foo2
---
# This is comment.
test: foo3
---
---
`

	names := make([]string, 0)
	err := IterYAMLBytes([]byte(inputs), func(bytes []byte) error {
		var obj map[string]string
		err := yaml.Unmarshal(bytes, &obj)
		require.NoError(t, err)

		if name, ok := obj["test"]; ok {
			names = append(names, name)
		}
		return nil
	})
	require.NoError(t, err)
	require.ElementsMatch(t, names, []string{"foo1", "foo2", "foo3"})
}

func testName(inputFile string) string {
	_, fileName := filepath.Split(inputFile)
	return strings.TrimSuffix(fileName, ".in.yaml")
}

func TestLoadAllSupportedResourcesFromYAMLBytes(t *testing.T) {
	// list all file names in testdata
	inputFiles, err := filepath.Glob(filepath.Join("testdata", "*.in.yaml"))
	require.NoError(t, err)
	for _, inFile := range inputFiles {
		t.Run(testName(inFile), func(t *testing.T) {
			t.Parallel() // this's used for race detection
			data, err := os.ReadFile(inFile)
			require.NoError(t, err)
			got, err := LoadResourcesFromYAMLBytes(data, true, nil)
			require.NoError(t, err)

			outputFile := strings.Replace(inFile, ".in.yaml", ".out.yaml", 1)
			if test.OverrideTestData() {
				out, err := yaml.Marshal(got)
				require.NoError(t, err)
				require.NoError(t, file.Write(string(out), outputFile))
			}

			want := &Resources{}
			output, err := os.ReadFile(outputFile)
			require.NoError(t, err)
			mustUnmarshal(t, output, want)

			opts := []cmp.Option{
				cmpopts.EquateEmpty(),
			}
			require.Empty(t, cmp.Diff(want, got, opts...))
		})
	}
}

func mustUnmarshal(t *testing.T, val []byte, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict(val, out, yaml.DisallowUnknownFields))
}
