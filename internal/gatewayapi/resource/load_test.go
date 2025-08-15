// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"fmt"
	"os"
	"path/filepath"
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

func TestLoadAllSupportedResourcesFromYAMLBytes(t *testing.T) {
	inFile := requireTestDataFile(t, "all-resources", "in")
	got, err := LoadResourcesFromYAMLBytes(inFile, true)
	require.NoError(t, err)

	if test.OverrideTestData() {
		out, err := yaml.Marshal(got)
		require.NoError(t, err)
		require.NoError(t, file.Write(string(out), filepath.Join("testdata", "all-resources.out.yaml")))
	}

	want := &LoadResources{}
	outFile := requireTestDataFile(t, "all-resources", "out")
	mustUnmarshal(t, outFile, want)

	opts := []cmp.Option{
		cmpopts.IgnoreFields(LoadResources{}, "Resources.serviceMap"),
		cmpopts.EquateEmpty(),
	}
	require.Empty(t, cmp.Diff(want, got, opts...))
}

func requireTestDataFile(t *testing.T, name, ioType string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("testdata", fmt.Sprintf("%s.%s.yaml", name, ioType)))
	require.NoError(t, err)
	return content
}

func mustUnmarshal(t *testing.T, val []byte, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict(val, out, yaml.DisallowUnknownFields))
}
