// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package endpoint gives Kubernetes endpoint data a revision, and queues
// fast-path endpoint updates for the xDS publisher.
package endpoint

import (
	"sync"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/ir"
)

// RevisionTracker gives each backend's endpoints a revision. A higher revision
// always means a later load, and an unchanged revision means the endpoints did
// not change. The publisher records the revision each published snapshot was
// built with. It applies a fast update only if the update is newer, and gives a
// full result the published endpoints when those are newer than its own.
type RevisionTracker struct {
	once sync.Once

	// mu protects the fields below. Track holds it across the load, so loads
	// get revisions in the order they read Kubernetes.
	mu sync.Mutex

	// lastRevision is the last revision given out. One counter is shared by all
	// backends and never resets, even when a backend is pruned, so a backend that
	// comes back always gets a larger revision than anything published for it.
	lastRevision uint64

	// lastLoaded holds the endpoints last loaded for each referenced backend.
	// Track compares against it to keep the revision when nothing changed.
	lastLoaded map[ir.ServiceKey]*BackendEndpoints
}

func (s *RevisionTracker) init() {
	s.once.Do(func() { s.lastLoaded = make(map[ir.ServiceKey]*BackendEndpoints) })
}

// Track loads a backend's EndpointSlices and returns its endpoints with a
// revision: the stored one if the endpoints are unchanged, otherwise the next
// one. On error the stored entry is left as is.
func (s *RevisionTracker) Track(key ir.ServiceKey, load func() ([]*discoveryv1.EndpointSlice, error)) (*BackendEndpoints, error) {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	slices, err := load()
	if err != nil {
		return nil, err
	}
	previous := s.lastLoaded[key]
	// Most backends are unchanged between full reconciles. If the same slices come
	// back at the same resource versions, reuse the stored entry and skip the
	// conversion and comparison below.
	if sameSliceVersions(previous, slices) {
		return previous, nil
	}
	current := NewBackendEndpoints(key, slices)
	if sameEndpointData(previous, current) {
		current.Revision = previous.Revision
	} else {
		s.lastRevision++
		current.Revision = s.lastRevision
	}
	s.lastLoaded[key] = current
	return current, nil
}

// PruneBackends removes stored endpoints for backends absent from full
// Kubernetes collection. It waits for an in-flight Track. A Track that
// starts afterwards can restore an entry until the next full collection prunes
// it. Revisions never reset.
func (s *RevisionTracker) PruneBackends(keys []ir.ServiceKey) {
	s.init()
	next := make(map[ir.ServiceKey]struct{}, len(keys))
	for _, key := range keys {
		next[key] = struct{}{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for key := range s.lastLoaded {
		if _, found := next[key]; !found {
			delete(s.lastLoaded, key)
		}
	}
}

// sameSliceVersions reports whether slices are exactly the EndpointSlices previous
// was built from, at the same resource versions. Slices without a resource
// version never match, so they always take the full comparison.
func sameSliceVersions(previous *BackendEndpoints, slices []*discoveryv1.EndpointSlice) bool {
	if previous == nil || len(previous.EndpointSlices) != len(slices) {
		return false
	}
	versions := make(map[types.NamespacedName]string, len(slices))
	for _, slice := range previous.EndpointSlices {
		versions[types.NamespacedName{Namespace: slice.Namespace, Name: slice.Name}] = slice.ResourceVersion
	}
	for _, slice := range slices {
		version, found := versions[types.NamespacedName{Namespace: slice.Namespace, Name: slice.Name}]
		if !found || slice.ResourceVersion == "" || slice.ResourceVersion != version {
			return false
		}
	}
	return true
}
