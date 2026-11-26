// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package migrate

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestInlineDocument(t *testing.T) {
	for _, src := range []string{
		"--- {kind: Example, old: true}\n",
		"--- &root\nkind: Example\nold: true\n",
		"--- !!map\nkind: Example\nold: true\n",
	} {
		t.Run(src, func(t *testing.T) {
			in, err := newFileInput("example.yaml", []byte(src))
			require.NoError(t, err)
			var out bytes.Buffer
			require.NoError(t, in.write(&out, true))
			require.Equal(t, src, out.String(), "unchanged input must retain its bytes")

			d := in.docs[1]
			require.NotNil(t, d.obj)
			require.Equal(t, "Example", d.obj.Kind)
			require.True(t, Replace(d.obj.Root, "old", "new", Bool(true)))
			d.changed = true
			out.Reset()
			require.NoError(t, in.write(&out, true))
			var got map[string]any
			require.NoError(t, yaml.Unmarshal(out.Bytes(), &got))
			require.Equal(t, map[string]any{"kind": "Example", "new": true}, got)
		})
	}
}

func TestWriteFileAtomic(t *testing.T) {
	for _, symlink := range []bool{false, true} {
		name := "regular"
		if symlink {
			name = "symlink"
		}
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			target := filepath.Join(dir, "target.yaml")
			require.NoError(t, os.WriteFile(target, []byte("old: true\n"), 0o600))
			require.NoError(t, os.Chmod(target, 0o640))
			path := target
			if symlink {
				path = filepath.Join(dir, "link.yaml")
				require.NoError(t, os.Symlink("target.yaml", path))
			}
			original, err := os.Open(target)
			require.NoError(t, err)
			defer original.Close()
			before, err := original.Stat()
			require.NoError(t, err)
			in, err := newFileInput(path, []byte("new: true\n"))
			require.NoError(t, err)
			require.NoError(t, in.writeFile())
			got, err := os.ReadFile(path)
			require.NoError(t, err)
			require.Equal(t, "new: true\n", string(got))
			after, err := os.Stat(target)
			require.NoError(t, err)
			require.Equal(t, os.FileMode(0o640), after.Mode().Perm())
			require.False(t, os.SameFile(before, after), "write must replace rather than truncate the original")
			old := make([]byte, len("old: true\n"))
			_, err = original.Read(old)
			require.NoError(t, err)
			require.Equal(t, "old: true\n", string(old))
			if symlink {
				link, err := os.Readlink(path)
				require.NoError(t, err)
				require.Equal(t, "target.yaml", link)
			}
			leftovers, err := filepath.Glob(filepath.Join(dir, ".egctl-migrate-*"))
			require.NoError(t, err)
			require.Empty(t, leftovers)
		})
	}
}

func TestCloneAliases(t *testing.T) {
	var doc yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("root: &root\n  self: *root\ncopy: *root\n"), &doc))
	c := Clone(&doc)
	original := Get(doc.Content[0], "root")
	cloned := Get(c.Content[0], "root")
	require.NotSame(t, original, cloned)
	require.Same(t, cloned, Get(cloned, "self").Alias)
	require.Same(t, cloned, Get(c.Content[0], "copy").Alias)
	Set(cloned, "extra", Bool(true))
	require.Nil(t, Get(original, "extra"))
}

func TestYAMLFilesSymlinkedDir(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "manifests")
	require.NoError(t, os.MkdirAll(filepath.Join(target, "sub"), 0o750))
	for _, f := range []string{"a.yaml", "sub/b.yml", "c.txt"} {
		require.NoError(t, os.WriteFile(filepath.Join(target, f), []byte("kind: A\n"), 0o600))
	}
	link := filepath.Join(dir, "link")
	require.NoError(t, os.Symlink(target, link))

	files, err := yamlFiles(link)
	require.NoError(t, err)
	require.Equal(t, []string{filepath.Join(link, "a.yaml"), filepath.Join(link, "sub", "b.yml")}, files)
}

func TestExpandMergeKeys(t *testing.T) {
	for _, input := range []string{
		"base: &base {a: 1, b: 2}\nvalue: {<<: *base, a: 3}\n",
		"value: {<<: [{a: 1}, {a: 2, b: 3}]}\n",
		"value: {<<: {<<: {a: 1}, b: 2}, b: 3}\n",
		"value: {<<: {a: &a {nested: true}}}\ncopy: *a\n",
		"value: {'<<': literal}\n",
	} {
		t.Run(input, func(t *testing.T) {
			var doc yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(input), &doc))
			var want any
			require.NoError(t, doc.Decode(&want))
			require.NoError(t, ExpandMergeKeys(&doc))
			encoded, err := yaml.Marshal(&doc)
			require.NoError(t, err)
			var got any
			require.NoError(t, yaml.Unmarshal(encoded, &got))
			require.Equal(t, want, got)
		})
	}
	for _, input := range []string{
		"value: {<<: 1}",
		"value: {<<: [1]}",
		"value: &value {<<: *value}",
	} {
		var doc yaml.Node
		require.NoError(t, yaml.Unmarshal([]byte(input), &doc))
		require.Error(t, ExpandMergeKeys(&doc))
	}
}

func TestUnchangedMergeDocument(t *testing.T) {
	input := "apiVersion: v1\nkind: ConfigMap\ndata:\n  <<: {key: value}\n"
	in, err := newFileInput("config.yaml", []byte(input))
	require.NoError(t, err)
	var out bytes.Buffer
	require.NoError(t, in.write(&out, true))
	require.Equal(t, input, out.String())
}

func TestWritePreservesRemovedAnchors(t *testing.T) {
	for _, operation := range []string{"delete", "replace", "set"} {
		for _, value := range []string{"true", "{nested: &nested true, copy: *nested}"} {
			t.Run(operation+"/"+value, func(t *testing.T) {
				src := "kind: Example\nold: &enabled " + value + "\nother: *enabled\nagain: *enabled\n"
				in, err := newFileInput("example.yaml", []byte(src))
				require.NoError(t, err)
				d := in.docs[0]
				var want any
				require.NoError(t, Get(d.obj.Root, "other").Decode(&want))
				switch operation {
				case "delete":
					Delete(d.obj.Root, "old")
				case "replace":
					Replace(d.obj.Root, "old", "new", Mapping())
				case "set":
					Set(d.obj.Root, "old", Mapping())
				}
				d.changed = true
				var out bytes.Buffer
				require.NoError(t, in.write(&out, true))
				var got map[string]any
				require.NoError(t, yaml.Unmarshal(out.Bytes(), &got))
				require.Equal(t, want, got["other"])
				require.Equal(t, want, got["again"])
				// Rendering must not mutate the migration tree.
				require.Equal(t, yaml.AliasNode, Get(d.obj.Root, "other").Kind)
			})
		}
	}
}

func TestWritePreservesScalarCommentText(t *testing.T) {
	for _, scalar := range []string{"|\n  hello # note\n", ">\n  hello # note\n", "\"hello # note\"\n"} {
		t.Run(scalar, func(t *testing.T) {
			in, err := newFileInput("example.yaml", []byte("kind: Example\ntext: "+scalar+"old: true # note\n"))
			require.NoError(t, err)
			d := in.docs[0]
			want := Scalar(Get(d.obj.Root, "text"))
			Replace(d.obj.Root, "old", "new", Bool(true))
			d.changed = true
			var out bytes.Buffer
			require.NoError(t, in.write(&out, true))
			var got map[string]any
			require.NoError(t, yaml.Unmarshal(out.Bytes(), &got))
			require.Equal(t, want, got["text"])
			require.Contains(t, out.String(), "new: true  # note")
		})
	}
}

func TestPadLineComments(t *testing.T) {
	for _, scalar := range []string{
		"|\n  hello # note\n",
		">\n  hello # note\n",
		"\"hello # note\n  world\"\n",
		"'hello # note\n  world'\n",
	} {
		t.Run(scalar, func(t *testing.T) {
			input := "text: " + scalar + "first: true # longer # note\nsecond: false # note\n"
			want := "text: " + scalar + "first: true  # longer # note\nsecond: false  # note\n"
			require.Equal(t, want, string(padLineComments([]byte(input))))
			require.Equal(t, want, string(padLineComments([]byte(want))), "padding must be idempotent")
		})
	}
}

func TestListInputRejected(t *testing.T) {
	for _, input := range []string{
		"apiVersion: v1\nkind: List\nitems: []\n",
		"kind: SecurityPolicyList\nitems: []\n",
		"identity: &kind List\nkind: *kind\nitems: []\n",
		"<<: {kind: List}\nitems: []\n",
	} {
		_, err := newFileInput("-", []byte(input))
		require.ErrorContains(t, err, "-: document 1: list resource")
	}
}

func TestManagedBy(t *testing.T) {
	for _, tc := range []struct {
		name        string
		labels      map[string]string
		annotations map[string]string
		want        string
	}{
		{name: "unmanaged"},
		{name: "helm label", labels: map[string]string{"app.kubernetes.io/managed-by": "Helm"}, want: "Helm"},
		{name: "helm annotation", annotations: map[string]string{"meta.helm.sh/release-name": "eg"}, want: "Helm"},
		{
			name: "helm takes precedence over its instance label",
			labels: map[string]string{
				"app.kubernetes.io/managed-by": "Helm",
				"app.kubernetes.io/instance":   "eg",
			},
			want: "Helm",
		},
		{name: "argo cd default tracking label", labels: map[string]string{"app.kubernetes.io/instance": "app"}, want: "Argo CD"},
		{name: "argo cd custom label", labels: map[string]string{"argocd.argoproj.io/instance": "app"}, want: "Argo CD"},
		{name: "argo cd tracking annotation", annotations: map[string]string{"argocd.argoproj.io/tracking-id": "app:/Gateway:default/eg"}, want: "Argo CD"},
		{name: "flux kustomization", labels: map[string]string{"kustomize.toolkit.fluxcd.io/name": "apps"}, want: "Flux"},
		{name: "flux helm release", labels: map[string]string{"helm.toolkit.fluxcd.io/name": "eg"}, want: "Flux"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			o := &metav1.PartialObjectMetadata{}
			o.SetLabels(tc.labels)
			o.SetAnnotations(tc.annotations)
			require.Equal(t, tc.want, managedBy(o))
		})
	}
}
