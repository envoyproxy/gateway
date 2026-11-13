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

// clusterExtensionResources returns the extension resources passed to the cluster hook for the
// cluster built from settings. Route-level resources (e.g. extensionRef filters) are passed to
// every cluster of the route, but a custom backend resource is only passed to the cluster built
// from its own backendRef, so that the extension never modifies another backendRef's cluster.
//
// routeRefs holds the route-level resources followed by one entry per custom backendRef in
// allSettings, so the same object can appear several times: once per backendRef that uses it, and
// also as a route-level resource. The latter is not a common case, but it is allowed: objects of a
// backendResources kind also resolve as extensionRef filters, and a kind can be listed in both
// resources and backendResources, so a rule can use the same object as a filter and as a backendRef.
//
// An object is passed if it is used by settings, or if it appears in routeRefs more often than it
// is used by backendRefs, i.e. it is also a route-level resource. Each object is passed only once.
func clusterExtensionResources(routeRefs []*ir.UnstructuredRef, allSettings, settings []*ir.DestinationSetting,
	idx extensionResourceIndex,
) []*unstructured.Unstructured {
	if len(routeRefs) == 0 {
		return nil
	}

	owned := make(map[string]struct{}, len(settings))
	for _, s := range settings {
		if s.CustomBackendRef != nil {
			owned[unstructuredKey(resolveUnstructuredRef(s.CustomBackendRef, idx))] = struct{}{}
		}
	}
	// occurrences counts each object in routeRefs, less one for every backendRef that uses it.
	// What remains is the number of times it is a route-level resource.
	occurrences := make(map[string]int, len(routeRefs))
	for _, ref := range routeRefs {
		occurrences[unstructuredKey(resolveUnstructuredRef(ref, idx))]++
	}
	for _, s := range allSettings {
		if s.CustomBackendRef != nil {
			occurrences[unstructuredKey(resolveUnstructuredRef(s.CustomBackendRef, idx))]--
		}
	}

	seen := make(map[string]struct{}, len(routeRefs))
	var resources []*unstructured.Unstructured
	for _, ref := range routeRefs {
		res := resolveUnstructuredRef(ref, idx)
		key := unstructuredKey(res)
		if _, ok := owned[key]; !ok && occurrences[key] <= 0 {
			continue
		}
		if res != nil {
			if _, ok := seen[key]; ok {
				continue
			}
			seen[key] = struct{}{}
		}
		resources = append(resources, res)
	}
	return resources
}

func unstructuredKey(u *unstructured.Unstructured) string {
	if u == nil {
		return ""
	}
	return u.GetAPIVersion() + "/" + u.GetKind() + "/" + u.GetNamespace() + "/" + u.GetName()
}
