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
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
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

// Extension-managed resources are loaded as unstructured objects into their own category. As in
// standalone mode, where the offline controller registers them, the kinds are registered in the
// scheme as Unstructured before loading.
func TestLoadCustomExtensionKinds(t *testing.T) {
	for _, gvk := range []schema.GroupVersionKind{
		{Group: "example.extensions.io", Version: "v1alpha1", Kind: "ListenerContextExample"},
		{Group: "cert.example.io", Version: "v1alpha1", Kind: "ExampleCertificate"},
	} {
		envoygateway.GetScheme().AddKnownTypeWithName(gvk, &unstructured.Unstructured{})
	}

	eg := &egv1a1.EnvoyGateway{
		EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
			ExtensionManager: &egv1a1.ExtensionManager{
				PolicyResources: []egv1a1.GroupVersionKind{
					{Group: "example.extensions.io", Version: "v1alpha1", Kind: "ListenerContextExample"},
				},
				CertificateResources: []egv1a1.GroupVersionKind{
					{Group: "cert.example.io", Version: "v1alpha1", Kind: "ExampleCertificate"},
				},
			},
		},
	}

	in := []byte(`
apiVersion: gateway.networking.k8s.io/v1
kind: GatewayClass
metadata:
  name: eg
spec:
  controllerName: gateway.envoyproxy.io/gatewayclass-controller
---
apiVersion: example.extensions.io/v1alpha1
kind: ListenerContextExample
metadata:
  name: some-policy
  namespace: default
spec:
  username: user
---
apiVersion: cert.example.io/v1alpha1
kind: ExampleCertificate
metadata:
  name: app-cert
  namespace: default
spec:
  certificateId: example-certificate-1
status:
  conditions:
  - type: Ready
    status: "True"
    reason: Ready
`)

	got, err := LoadResourcesFromYAMLBytes(in, true, eg)
	require.NoError(t, err)

	require.Len(t, got.ExtensionServerPolicies, 1)
	require.Equal(t, "ListenerContextExample", got.ExtensionServerPolicies[0].GetKind())

	require.Len(t, got.ExtensionCertificates, 1)
	cert := got.ExtensionCertificates[0]
	require.Equal(t, "ExampleCertificate", cert.GetKind())
	require.Equal(t, "cert.example.io", cert.GroupVersionKind().Group)
	require.Equal(t, "app-cert", cert.GetName())
	require.Equal(t, "default", cert.GetNamespace())

	// The spec and status must survive intact -- the identifier comes from the spec and
	// admission gates on the Ready condition.
	id, found, err := unstructured.NestedString(cert.Object, "spec", "certificateId")
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, "example-certificate-1", id)

	conds, found, err := unstructured.NestedSlice(cert.Object, "status", "conditions")
	require.NoError(t, err)
	require.True(t, found)
	require.Len(t, conds, 1)
}
