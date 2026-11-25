// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package migrate

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"go.yaml.in/yaml/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	sigsyaml "sigs.k8s.io/yaml"
)

// input is one source of YAML documents: a file, stdin, or the cluster.
type input struct {
	// path is the file path, "-" for stdin, or empty for the cluster.
	path string
	docs []*document
}

// document is one YAML document of an input.
type document struct {
	// sep is the "---" line that starts the document, including any inline content.
	sep string
	// raw is the original text, written back as is unless the document changed.
	raw []byte
	// node is the parsed document node, nil if the document is empty or not a mapping.
	node *yaml.Node
	obj  *Object
	// managedBy names the tool that manages the object in the cluster, if any.
	managedBy string
	changed   bool
	// compactSeq reports whether sequences are written at the indentation of their parent key.
	compactSeq bool
}

func (in *input) changed() bool {
	for _, d := range in.docs {
		if d.changed {
			return true
		}
	}
	return false
}

// write writes the documents of the input to w. Unchanged documents are written byte for byte.
func (in *input) write(w io.Writer, first bool) error {
	var buf bytes.Buffer
	if !first {
		// The preceding input may not have ended with a newline.
		buf.WriteByte('\n')
	}
	for i, d := range in.docs {
		if i == 0 && !first && d.sep == "" {
			buf.WriteString("---\n")
		}
		if d.changed && d.sep != "" && !standaloneSeparator.MatchString(strings.TrimRight(d.sep, "\r\n")) {
			// Inline content is emitted by the encoder, not as part of the separator.
			buf.WriteString("---\n")
		} else {
			buf.WriteString(d.sep)
		}
		text, err := d.text()
		if err != nil {
			return err
		}
		buf.Write(text)
	}
	_, err := w.Write(buf.Bytes())
	return err
}

func (in *input) writeFile() error {
	info, err := os.Stat(in.path)
	if err != nil {
		return err
	}
	var buf bytes.Buffer
	if err := in.write(&buf, true); err != nil {
		return err
	}
	// Resolve symlinks so replacing the file preserves the link itself.
	target, err := filepath.EvalSymlinks(in.path)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".egctl-migrate-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	if _, err := tmp.Write(buf.Bytes()); err != nil {
		return err
	}
	if err := tmp.Chmod(info.Mode().Perm()); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

func (d *document) text() ([]byte, error) {
	if !d.changed {
		return d.raw, nil
	}
	d.node.Content[0] = d.obj.Root
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if d.compactSeq {
		enc.CompactSeqIndent()
	}
	if err := enc.Encode(d.node); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return padLineComments(buf.Bytes(), d.node), nil
}

// padLineComments puts two spaces before each trailing comment of the encoded node n. The encoder
// writes one, which yamllint and most YAML style guides reject.
//
// The comments are matched to output lines in document order, so a block scalar line that happens
// to end with the same text isn't changed.
func padLineComments(data []byte, n *yaml.Node) []byte {
	var comments []string
	var walk func(*yaml.Node)
	walk = func(n *yaml.Node) {
		if n.LineComment != "" {
			comments = append(comments, n.LineComment)
		}
		for _, c := range n.Content {
			walk(c)
		}
	}
	walk(n)
	if len(comments) == 0 {
		return data
	}

	lines := bytes.SplitAfter(data, []byte("\n"))
	for i, line := range lines {
		if len(comments) == 0 {
			break
		}
		body := bytes.TrimRight(line, "\n")
		suffix := " " + comments[0]
		if !bytes.HasSuffix(body, []byte(suffix)) || len(body) == len(suffix) {
			continue
		}
		code := body[:len(body)-len(suffix)]
		if !bytes.HasSuffix(code, []byte(" ")) {
			lines[i] = slices.Concat(code, []byte("  "+comments[0]), line[len(body):])
		}
		comments = comments[1:]
	}
	return bytes.Join(lines, nil)
}

var (
	docSeparator        = regexp.MustCompile(`^---(\s.*)?$`)
	standaloneSeparator = regexp.MustCompile(`^---[ \t]*(#.*)?$`)
	// indentedSeq matches a sequence indented below its parent key.
	indentedSeq = regexp.MustCompile(`(?m)^( *)[^\s#-][^\n]*:[ \t]*\n( *\n)*( *)- `)
)

// parseDocuments splits data into YAML documents.
func parseDocuments(source string, data []byte) ([]*document, error) {
	var (
		docs []*document
		cur  = &document{}
	)
	lines := bytes.SplitAfter(data, []byte("\n"))
	for _, line := range lines {
		if docSeparator.Match(bytes.TrimRight(line, "\r\n")) {
			docs = append(docs, cur)
			cur = &document{sep: string(line)}
			continue
		}
		cur.raw = append(cur.raw, line...)
	}
	docs = append(docs, cur)

	// A document without a sequence of its own follows the style of the file, or kubectl's.
	fileCompact, ok := seqStyle(data)
	if !ok {
		fileCompact = true
	}
	for i, d := range docs {
		var node yaml.Node
		data := d.raw
		if d.sep != "" && !standaloneSeparator.MatchString(strings.TrimRight(d.sep, "\r\n")) {
			data = append([]byte(d.sep), d.raw...)
		}
		if err := yaml.Unmarshal(data, &node); err != nil {
			return nil, fmt.Errorf("%s: document %d: %w", source, i+1, err)
		}
		if node.Kind != yaml.DocumentNode || len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
			continue
		}
		d.node = &node
		d.obj = newObject(node.Content[0], source)
		if d.compactSeq, ok = seqStyle(d.raw); !ok {
			d.compactSeq = fileCompact
		}
	}
	return docs, nil
}

// seqStyle reports whether the first sequence in data is written at the indentation of its parent
// key, which is what kubectl writes, rather than indented below it. ok is false if data has no
// block sequence under a key.
func seqStyle(data []byte) (compact, ok bool) {
	m := indentedSeq.FindSubmatch(data)
	if m == nil {
		return false, false
	}
	return len(m[3]) == len(m[1]), true
}

func newObject(root *yaml.Node, source string) *Object {
	obj := &Object{Root: root, Source: source}
	if v := Get(root, "apiVersion"); v != nil {
		obj.APIVersion = v.Value
	}
	if v := Get(root, "kind"); v != nil {
		obj.Kind = v.Value
	}
	if md := GetMapping(root, "metadata"); md != nil {
		if v := Get(md, "name"); v != nil {
			obj.Name = v.Value
		}
		if v := Get(md, "namespace"); v != nil {
			obj.Namespace = v.Value
		}
	}
	return obj
}

func loadFiles(paths []string) ([]*input, error) {
	var inputs []*input
	for _, p := range paths {
		if p == "-" {
			data, err := io.ReadAll(os.Stdin)
			if err != nil {
				return nil, fmt.Errorf("reading stdin: %w", err)
			}
			in, err := newFileInput("-", data)
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, in)
			continue
		}

		files, err := yamlFiles(p)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			data, err := os.ReadFile(f)
			if err != nil {
				return nil, err
			}
			in, err := newFileInput(f, data)
			if err != nil {
				return nil, err
			}
			inputs = append(inputs, in)
		}
	}
	return inputs, nil
}

func newFileInput(path string, data []byte) (*input, error) {
	docs, err := parseDocuments(path, data)
	if err != nil {
		return nil, err
	}
	return &input{path: path, docs: docs}, nil
}

// yamlFiles returns p if it's a file, or the YAML files under p, recursively, if it's a directory.
func yamlFiles(p string) ([]string, error) {
	info, err := os.Stat(p)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return []string{p}, nil
	}
	// WalkDir doesn't follow a symlink at its root, so walk its target, and report the files under
	// the path as given.
	root, err := filepath.EvalSymlinks(p)
	if err != nil {
		return nil, err
	}
	var files []string
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && (strings.HasSuffix(path, ".yaml") || strings.HasSuffix(path, ".yml")) {
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			files = append(files, filepath.Join(p, rel))
		}
		return nil
	})
	return files, err
}

// loadCluster reads the objects of path from the cluster, as a single input of YAML documents.
func loadCluster(ctx context.Context, path Path, cli client.Client, namespace string) ([]*input, error) {
	if cli == nil {
		return nil, fmt.Errorf("no Kubernetes client")
	}
	objs, err := path.Fetch(ctx, cli, namespace)
	if err != nil {
		return nil, err
	}
	in := &input{}
	for i, o := range objs {
		gvk, err := apiutil.GVKForObject(o, cli.Scheme())
		if err != nil {
			return nil, err
		}
		o.GetObjectKind().SetGroupVersionKind(gvk)
		managedBy := managedBy(o)

		data, err := sigsyaml.Marshal(o)
		if err != nil {
			return nil, err
		}
		var m map[string]any
		if err := sigsyaml.Unmarshal(data, &m); err != nil {
			return nil, err
		}
		stripServerFields(m)
		if data, err = sigsyaml.Marshal(m); err != nil {
			return nil, err
		}

		sep := ""
		if i > 0 {
			sep = "---\n"
		}
		docs, err := parseDocuments(fmt.Sprintf("%s %s/%s", gvk.Kind, o.GetNamespace(), o.GetName()), data)
		if err != nil {
			return nil, err
		}
		docs[0].sep = sep
		docs[0].managedBy = managedBy
		if docs[0].obj != nil {
			docs[0].obj.Source = ""
		}
		in.docs = append(in.docs, docs...)
	}
	return []*input{in}, nil
}

// stripServerFields removes the fields the server sets, so the printed object can be applied.
func stripServerFields(m map[string]any) {
	delete(m, "status")
	md, ok := m["metadata"].(map[string]any)
	if !ok {
		return
	}
	for _, f := range []string{
		"managedFields", "resourceVersion", "uid", "generation", "creationTimestamp",
		"selfLink", "deletionTimestamp", "deletionGracePeriodSeconds",
	} {
		delete(md, f)
	}
	if ann, ok := md["annotations"].(map[string]any); ok {
		// It holds the previously applied object, including the deprecated fields.
		delete(ann, "kubectl.kubernetes.io/last-applied-configuration")
		if len(ann) == 0 {
			delete(md, "annotations")
		}
	}
}

// managedBy returns the tool that manages the object, from its labels and annotations, or "".
func managedBy(o client.Object) string {
	labels, annotations := o.GetLabels(), o.GetAnnotations()
	switch {
	case labels["app.kubernetes.io/managed-by"] == "Helm" || annotations["meta.helm.sh/release-name"] != "":
		return "Helm"
	case labels["argocd.argoproj.io/instance"] != "" || annotations["argocd.argoproj.io/tracking-id"] != "":
		return "Argo CD"
	case labels["kustomize.toolkit.fluxcd.io/name"] != "" || labels["helm.toolkit.fluxcd.io/name"] != "":
		return "Flux"
	}
	return ""
}
