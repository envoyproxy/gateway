// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTranslate(t *testing.T) {
	testCases := []struct {
		name string
		from string
		to   string
	}{
		{
			name: "from-gateway-api-to-xds",
			from: "gateway-api",
			to:   "xds",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			b := bytes.NewBufferString("")
			root := NewTranslateCommand()
			root.SetOut(b)
			root.SetErr(b)
			args := []string{
				"translate",
				"--from",
				tc.from,
				"--to",
				tc.to,
				"--file",
				"testdata/in/" + tc.name + ".yaml",
			}

			root.SetArgs(args)
			assert.NoError(t, root.ExecuteContext(context.Background()))
			out, err := io.ReadAll(b)
			assert.NoError(t, err)
			require.Equal(t, requireTestDataOutFile(t, tc.name+".out"), string(out))
		})
	}
}

func requireTestDataOutFile(t *testing.T, name ...string) string {
	t.Helper()
	elems := append([]string{"testdata", "out"}, name...)
	content, err := os.ReadFile(filepath.Join(elems...))
	require.NoError(t, err)
	return string(content)
}
