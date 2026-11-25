// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"sync"

	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
)

// revisionTracker gives each backend's endpoints a revision. A higher revision
// always means a later load, and an unchanged revision means the EndpointSlices
// did not change. The xDS publisher compares these revisions to keep the newer
// endpoints.
type revisionTracker struct {
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
	lastLoaded map[ir.ServiceKey]*endpoint.BackendEndpoints
}

func (s *revisionTracker) init() {
	s.once.Do(func() { s.lastLoaded = make(map[ir.ServiceKey]*endpoint.BackendEndpoints) })
}

// Track loads a backend's EndpointSlices and returns its endpoints with a
// revision: the stored one if the slices are unchanged, otherwise the next
// one. On error the stored entry is left as is.
func (s *revisionTracker) Track(key ir.ServiceKey, load func() ([]*discoveryv1.EndpointSlice, error)) (*endpoint.BackendEndpoints, error) {
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
	// conversion below.
	if sameSliceVersions(previous, slices) {
		return previous, nil
	}
	current := endpoint.NewBackendEndpoints(key, slices)
	s.lastRevision++
	current.Revision = s.lastRevision
	s.lastLoaded[key] = current
	return current, nil
}

// Tracks reports whether the backend's endpoints are stored. Full reconciliation
// stores every referenced backend, so this tells whether a backend is in use.
func (s *revisionTracker) Tracks(key ir.ServiceKey) bool {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	_, found := s.lastLoaded[key]
	return found
}

// PruneBackends removes stored endpoints for backends absent from full
// Kubernetes collection. It waits for an in-flight Track. A Track that
// starts afterwards can restore an entry until the next full collection prunes
// it. Revisions never reset.
func (s *revisionTracker) PruneBackends(keys []ir.ServiceKey) {
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
func sameSliceVersions(previous *endpoint.BackendEndpoints, slices []*discoveryv1.EndpointSlice) bool {
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
