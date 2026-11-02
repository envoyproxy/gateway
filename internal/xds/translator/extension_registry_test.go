// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestNewExtensionResourceIndex(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "obj1"}}}
	xdsIR := &ir.Xds{
		ExtensionResources: []*ir.UnstructuredRef{
			{Name: "foo/ns1/obj1", Object: obj},
		},
	}

	idx := newExtensionResourceIndex(xdsIR)
	require.Len(t, idx, 1)
	require.Same(t, obj, idx["foo/ns1/obj1"])
}

func TestResolveUnstructuredRef(t *testing.T) {
	obj := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "obj1"}}}
	idx := extensionResourceIndex{"foo/ns1/obj1": obj}

	t.Run("nil ref resolves to nil", func(t *testing.T) {
		require.Nil(t, resolveUnstructuredRef(nil, idx))
	})

	t.Run("embedded object is returned directly, bypassing the index", func(t *testing.T) {
		embedded := &unstructured.Unstructured{Object: map[string]any{"metadata": map[string]any{"name": "embedded"}}}
		ref := &ir.UnstructuredRef{Object: embedded}
		require.Same(t, embedded, resolveUnstructuredRef(ref, idx))
	})

	t.Run("name-only ref resolves through the index", func(t *testing.T) {
		ref := &ir.UnstructuredRef{Name: "foo/ns1/obj1"}
		require.Same(t, obj, resolveUnstructuredRef(ref, idx))
	})

	t.Run("name-only ref with no matching entry resolves to nil", func(t *testing.T) {
		ref := &ir.UnstructuredRef{Name: "missing"}
		require.Nil(t, resolveUnstructuredRef(ref, idx))
	})
}
