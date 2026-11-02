// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/envoyproxy/gateway/internal/ir"
)

// extensionResourceIndex resolves an UnstructuredRef's Name against ir.Xds.ExtensionResources.
type extensionResourceIndex map[string]*unstructured.Unstructured

func newExtensionResourceIndex(xdsIR *ir.Xds) extensionResourceIndex {
	idx := make(extensionResourceIndex, len(xdsIR.ExtensionResources))
	for _, ref := range xdsIR.ExtensionResources {
		idx[ref.Name] = ref.Object
	}
	return idx
}

// resolveUnstructuredRef returns ref's resolved object: ref.Object directly when populated
// (deduplication disabled), otherwise looked up by ref.Name (deduplication enabled).
func resolveUnstructuredRef(ref *ir.UnstructuredRef, idx extensionResourceIndex) *unstructured.Unstructured {
	if ref == nil {
		return nil
	}
	if ref.Object != nil {
		return ref.Object
	}
	return idx[ref.Name]
}
