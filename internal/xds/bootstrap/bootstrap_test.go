// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetRenderedBootstrapConfig(t *testing.T) {
	cases := []struct {
		name             string
		enablePrometheus bool
	}{
		{
			name: "default",
		},
		{
			name:             "enable-prometheus",
			enablePrometheus: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := GetRenderedBootstrapConfig(tc.enablePrometheus)
			assert.NoError(t, err)
			expected, err := readTestData(tc.name)
			assert.NoError(t, err)
			assert.Equal(t, expected, got)
		})
	}
}

func readTestData(caseName string) (string, error) {
	filename := path.Join("testdata", fmt.Sprintf("%s.yaml", caseName))

	b, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return string(b), nil
}
