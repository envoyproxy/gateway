// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package resource

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/yaml"
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
