// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecated

import (
	"bytes"
	"errors"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	sigsyaml "sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/deprecation"
)

// TestDetectParity checks that the YAML detection finds the deprecated fields that
// internal/deprecation, which sets the DeprecatedField status condition, finds on every golden input.
// The YAML detection also reports fields set to a value the status ignores, see presenceOnly.
// Remove it together with the deprecated fields.
func TestDetectParity(t *testing.T) {
	inputs, err := filepath.Glob("../../testdata/migrate/deprecated/*.in.yaml")
	require.NoError(t, err)
	require.NotEmpty(t, inputs)

	for _, in := range inputs {
		data, err := os.ReadFile(in)
		require.NoError(t, err)
		dec := yaml.NewDecoder(bytes.NewReader(data))
		for i := 0; ; i++ {
			var doc yaml.Node
			if err := dec.Decode(&doc); errors.Is(err, io.EOF) {
				break
			} else {
				require.NoError(t, err)
			}
			root := doc.Content[0]
			kind := migrateKind(root)
			typed := typedDetect(t, kind, root)
			if typed == nil {
				continue
			}
			got := detect(kind, root)
			for _, path := range typed {
				require.Contains(t, got, path, "%s: document %d (%s)", in, i+1, kind)
			}
			for path := range got {
				if !slices.Contains(typed, path) {
					require.True(t, presenceOnly(path), "%s: document %d (%s): unexpected %s", in, i+1, kind, path)
				}
			}
		}
	}
}

// presenceOnly reports whether path is a field the YAML detection reports whenever it is set, and
// the status only for some values: clientValidation.optional (only true), and the telemetry port
// (not 4317, its CRD default).
func presenceOnly(path string) bool {
	return strings.HasSuffix(path, ".clientValidation.optional") || strings.HasSuffix(path, ".port")
}

func migrateKind(root *yaml.Node) string {
	for i := 0; i+1 < len(root.Content); i += 2 {
		if root.Content[i].Value == "kind" {
			return root.Content[i+1].Value
		}
	}
	return ""
}

// typedDetect returns the sorted paths internal/deprecation reports for root, or nil if kind isn't
// one of the migrated kinds.
func typedDetect(t *testing.T, kind string, root *yaml.Node) []string {
	t.Helper()
	data, err := yaml.Marshal(root)
	require.NoError(t, err)

	var fields map[string]deprecation.DeprecatedField
	switch kind {
	case egv1a1.KindClientTrafficPolicy:
		fields = deprecation.ClientTrafficPolicy(decode[egv1a1.ClientTrafficPolicy](t, data))
	case egv1a1.KindBackendTrafficPolicy:
		fields = deprecation.BackendTrafficPolicy(decode[egv1a1.BackendTrafficPolicy](t, data))
	case egv1a1.KindSecurityPolicy:
		fields = deprecation.SecurityPolicy(decode[egv1a1.SecurityPolicy](t, data))
	case egv1a1.KindEnvoyExtensionPolicy:
		fields = deprecation.EnvoyExtensionPolicy(decode[egv1a1.EnvoyExtensionPolicy](t, data))
	case egv1a1.KindEnvoyProxy:
		fields = deprecation.EnvoyProxy(decode[egv1a1.EnvoyProxy](t, data))
	default:
		return nil
	}
	return slices.Sorted(maps.Keys(fields))
}

func decode[T any](t *testing.T, data []byte) *T {
	t.Helper()
	obj := new(T)
	require.NoError(t, sigsyaml.Unmarshal(data, obj))
	return obj
}
