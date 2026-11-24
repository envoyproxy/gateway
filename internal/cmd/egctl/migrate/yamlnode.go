// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package migrate

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"go.yaml.in/yaml/v3"
)

// Get returns the value of key in the mapping node m, or nil if m isn't a mapping or has no key.
func Get(m *yaml.Node, key string) *yaml.Node {
	m = Resolve(m)
	if i := keyIndex(m, key); i >= 0 {
		return m.Content[i+1]
	}
	return nil
}

// GetMapping returns the value of key in m if it is a mapping node, following an alias, otherwise
// nil.
func GetMapping(m *yaml.Node, key string) *yaml.Node {
	if v := Resolve(Get(m, key)); v != nil && v.Kind == yaml.MappingNode {
		return v
	}
	return nil
}

// Resolve returns the node an alias node refers to, or n itself if it isn't an alias.
//
// Get returns alias nodes as they are, so that a value moved to another key stays an alias of its
// anchor. Resolve the value before reading its kind or content.
func Resolve(n *yaml.Node) *yaml.Node {
	for n != nil && n.Kind == yaml.AliasNode {
		n = n.Alias
	}
	return n
}

// Scalar returns the value of the scalar node n, following an alias, or "" if n isn't a scalar.
func Scalar(n *yaml.Node) string {
	if n = Resolve(n); n != nil && n.Kind == yaml.ScalarNode {
		return n.Value
	}
	return ""
}

// Has reports whether key is present, including when its value is null.
func Has(m *yaml.Node, key string) bool {
	return Get(m, key) != nil
}

// IsSet reports whether key is set to a non-null value in the mapping node m.
func IsSet(m *yaml.Node, key string) bool {
	v := Resolve(Get(m, key))
	return v != nil && (v.Kind != yaml.ScalarNode || v.Tag != "!!null")
}

// Delete removes key from the mapping node m. It reports whether the key was present.
func Delete(m *yaml.Node, key string) bool {
	m = Resolve(m)
	i := keyIndex(m, key)
	if i < 0 {
		return false
	}
	m.Content = append(m.Content[:i], m.Content[i+2:]...)
	return true
}

// Replace renames key in the mapping node m to newKey and sets its value to value, keeping its
// position and the comments of the key. It reports whether the key was present.
func Replace(m *yaml.Node, key, newKey string, value *yaml.Node) bool {
	m = Resolve(m)
	i := keyIndex(m, key)
	if i < 0 {
		return false
	}
	m.Content[i].Value = newKey
	old := m.Content[i+1]
	if value.LineComment == "" && old != value && old.Kind == yaml.ScalarNode {
		// Keep a trailing comment like `enabled: true # why`.
		value.LineComment = old.LineComment
	}
	m.Content[i+1] = value
	return true
}

// Set sets key in the mapping node m to value, appending the key if it isn't present.
func Set(m *yaml.Node, key string, value *yaml.Node) {
	m = Resolve(m)
	if i := keyIndex(m, key); i >= 0 {
		m.Content[i+1] = value
		return
	}
	m.Content = append(m.Content, String(key), value)
}

// IsEmpty reports whether n is nil, null or an empty mapping or sequence.
func IsEmpty(n *yaml.Node) bool {
	n = Resolve(n)
	if n == nil {
		return true
	}
	switch n.Kind {
	case yaml.MappingNode, yaml.SequenceNode:
		return len(n.Content) == 0
	case yaml.ScalarNode:
		return n.Tag == "!!null"
	}
	return false
}

// IsTrue reports whether n is the scalar true, following an alias.
func IsTrue(n *yaml.Node) bool {
	n = Resolve(n)
	var b bool
	return n != nil && n.Kind == yaml.ScalarNode && n.Decode(&b) == nil && b
}

// String returns a new block-style string scalar node.
func String(v string) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: v}
}

// Int returns a new integer scalar node.
func Int(v int64) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: strconv.FormatInt(v, 10)}
}

// Bool returns a new boolean scalar node.
func Bool(v bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(v)}
}

// Mapping returns a new block-style mapping node with the given key/value pairs.
func Mapping(kv ...any) *yaml.Node {
	m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for i := 0; i+1 < len(kv); i += 2 {
		m.Content = append(m.Content, String(kv[i].(string)), kv[i+1].(*yaml.Node))
	}
	return m
}

// Sequence returns a new block-style sequence node holding items.
func Sequence(items ...*yaml.Node) *yaml.Node {
	return &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: items}
}

// Equal reports whether a and b hold the same data, ignoring style and comments.
func Equal(a, b *yaml.Node) bool {
	var va, vb any
	if a.Decode(&va) != nil || b.Decode(&vb) != nil {
		return false
	}
	return reflect.DeepEqual(va, vb)
}

var pathSegment = regexp.MustCompile(`^([^\[\]]+)((?:\[\d+\])*)$`)

// Lookup walks root along a field path like "spec.extProc[1].backendSettings", following aliases,
// and returns the node it ends on, or nil if the path doesn't exist.
func Lookup(root *yaml.Node, path string) *yaml.Node {
	n := Resolve(root)
	if n == nil || path == "" {
		return n
	}
	for seg := range strings.SplitSeq(path, ".") {
		m := pathSegment.FindStringSubmatch(seg)
		if m == nil {
			return nil
		}
		if n = Get(n, m[1]); n == nil {
			return nil
		}
		for idx := range strings.SplitSeq(strings.Trim(m[2], "[]"), "][") {
			if idx == "" {
				continue
			}
			i, _ := strconv.Atoi(idx)
			if n = Resolve(n); n.Kind != yaml.SequenceNode || i >= len(n.Content) {
				return nil
			}
			n = n.Content[i]
		}
	}
	return Resolve(n)
}

// Clone returns a deep copy of n, preserving shared nodes and cycles in aliases.
func Clone(n *yaml.Node) *yaml.Node {
	return clone(n, make(map[*yaml.Node]*yaml.Node))
}

// preserveAnchors prepares an output tree without changing the migration tree.
// Removing or replacing an anchored field can leave aliases pointing to nodes no
// longer in the document. Emit that anchor at its first surviving use instead.
// Track anchors in output order, including names reused for different nodes.
func preserveAnchors(root *yaml.Node) *yaml.Node {
	anchors := make(map[string]*yaml.Node)
	var visit func(*yaml.Node) *yaml.Node
	visit = func(n *yaml.Node) *yaml.Node {
		if n.Kind == yaml.AliasNode && anchors[n.Value] != n.Alias {
			n = n.Alias
		}
		out := *n
		if n.Anchor != "" {
			anchors[n.Anchor] = n
		}
		out.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			out.Content[i] = visit(child)
		}
		return &out
	}
	return visit(root)
}

func clone(n *yaml.Node, seen map[*yaml.Node]*yaml.Node) *yaml.Node {
	if n == nil {
		return nil
	}
	if c, ok := seen[n]; ok {
		return c
	}
	c := *n
	seen[n] = &c
	if n.Alias != nil {
		c.Alias = clone(n.Alias, seen)
	}
	if n.Content != nil {
		c.Content = make([]*yaml.Node, len(n.Content))
		for i, child := range n.Content {
			c.Content[i] = clone(child, seen)
		}
	}
	return &c
}

func keyIndex(m *yaml.Node, key string) int {
	if m == nil || m.Kind != yaml.MappingNode {
		return -1
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			return i
		}
	}
	return -1
}

// ExpandMergeKeys expands YAML merges before migration. Explicit keys override merged keys,
// and earlier mappings in a merge sequence override later mappings. Documents without merges
// retain their original nodes. When expanding, aliases are copied by value so removing a merge
// cannot leave dangling aliases or allow a rewrite to modify another use of the merged mapping.
func ExpandMergeKeys(root *yaml.Node) error {
	if !hasMerge(root, make(map[*yaml.Node]bool)) {
		return nil
	}
	expanded, err := expandMergedNode(root, make(map[*yaml.Node]bool))
	if err != nil {
		return err
	}
	*root = *expanded
	return nil
}

func hasMerge(n *yaml.Node, seen map[*yaml.Node]bool) bool {
	if n == nil || seen[n] {
		return false
	}
	seen[n] = true
	if n.Kind == yaml.MappingNode {
		for i := 0; i+1 < len(n.Content); i += 2 {
			if n.Content[i].Tag == "!!merge" {
				return true
			}
		}
	}
	if hasMerge(n.Alias, seen) {
		return true
	}
	for _, child := range n.Content {
		if hasMerge(child, seen) {
			return true
		}
	}
	return false
}

func expandMergedNode(n *yaml.Node, active map[*yaml.Node]bool) (*yaml.Node, error) {
	if n == nil {
		return nil, nil
	}
	if active[n] {
		return nil, fmt.Errorf("cyclic YAML alias or merge")
	}
	active[n] = true
	defer delete(active, n)
	if n.Kind == yaml.AliasNode {
		return expandMergedNode(n.Alias, active)
	}
	out := *n
	out.Anchor = ""
	out.Content = nil
	for _, child := range n.Content {
		c, err := expandMergedNode(child, active)
		if err != nil {
			return nil, err
		}
		out.Content = append(out.Content, c)
	}
	if n.Kind != yaml.MappingNode {
		return &out, nil
	}
	content := out.Content
	out.Content = nil
	keys := make(map[string]bool)
	for i := 0; i+1 < len(content); i += 2 {
		if content[i].Tag != "!!merge" {
			out.Content = append(out.Content, content[i], content[i+1])
			keys[content[i].Value] = true
		}
	}
	for i := 0; i+1 < len(content); i += 2 {
		if content[i].Tag != "!!merge" {
			continue
		}
		sources := []*yaml.Node{content[i+1]}
		if sources[0].Kind == yaml.SequenceNode {
			sources = sources[0].Content
		}
		for _, source := range sources {
			if source.Kind != yaml.MappingNode {
				return nil, fmt.Errorf("YAML merge requires a mapping or a sequence of mappings")
			}
			for j := 0; j+1 < len(source.Content); j += 2 {
				key := source.Content[j]
				if !keys[key.Value] {
					out.Content = append(out.Content, key, source.Content[j+1])
					keys[key.Value] = true
				}
			}
		}
	}
	return &out, nil
}
