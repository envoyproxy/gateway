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
