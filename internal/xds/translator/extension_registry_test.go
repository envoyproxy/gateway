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

func TestClusterExtensionResources(t *testing.T) {
	newRes := func(kind, name string) *unstructured.Unstructured {
		u := &unstructured.Unstructured{}
		u.SetAPIVersion("example.io/v1")
		u.SetKind(kind)
		u.SetNamespace("default")
		u.SetName(name)
		return u
	}
	poolA, poolB, filter := newRes("Pool", "a"), newRes("Pool", "b"), newRes("Filter", "f")
	refA := &ir.UnstructuredRef{Object: poolA}
	refB := &ir.UnstructuredRef{Object: poolB}
	refFilter := &ir.UnstructuredRef{Object: filter}

	settingA1 := &ir.DestinationSetting{Name: "a1", IsCustomBackend: true, CustomBackendRef: refA}
	settingA2 := &ir.DestinationSetting{Name: "a2", IsCustomBackend: true, CustomBackendRef: refA}
	settingB := &ir.DestinationSetting{Name: "b", IsCustomBackend: true, CustomBackendRef: refB}
	service := &ir.DestinationSetting{Name: "svc"}

	t.Run("no route refs", func(t *testing.T) {
		require.Nil(t, clusterExtensionResources(nil, []*ir.DestinationSetting{settingA1}, []*ir.DestinationSetting{settingA1}, nil))
	})

	t.Run("custom backend only goes to its own cluster", func(t *testing.T) {
		routeRefs := []*ir.UnstructuredRef{refA, refB, refFilter}
		all := []*ir.DestinationSetting{settingA1, settingB, service}
		require.Equal(t, []*unstructured.Unstructured{poolA, filter},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{settingA1}, nil))
		require.Equal(t, []*unstructured.Unstructured{poolB, filter},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{settingB}, nil))
		require.Equal(t, []*unstructured.Unstructured{filter},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{service}, nil))
	})

	t.Run("resource referenced by several backendRefs is passed once", func(t *testing.T) {
		routeRefs := []*ir.UnstructuredRef{refA, refA, refFilter}
		all := []*ir.DestinationSetting{settingA1, settingA2}
		require.Equal(t, []*unstructured.Unstructured{poolA, filter},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{settingA1}, nil))
		require.Equal(t, []*unstructured.Unstructured{poolA, filter},
			clusterExtensionResources(routeRefs, all, all, nil))
	})

	t.Run("resource used as both a filter and a custom backend", func(t *testing.T) {
		// refA is an extensionRef filter, and also the custom backend of settingA1. This is not a
		// common case, but nothing rejects it: see clusterExtensionResources.
		routeRefs := []*ir.UnstructuredRef{refA, refA}
		all := []*ir.DestinationSetting{settingA1, service}
		require.Equal(t, []*unstructured.Unstructured{poolA},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{settingA1}, nil))
		require.Equal(t, []*unstructured.Unstructured{poolA},
			clusterExtensionResources(routeRefs, all, []*ir.DestinationSetting{service}, nil))
	})
}
