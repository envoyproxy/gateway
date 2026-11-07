// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package endpoint collects versioned Kubernetes endpoint data for full
// translation, and queues fast-path endpoint updates for the xDS publisher.
package endpoint

import (
	"sync"

	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

// Source holds the latest endpoints of each Kubernetes backend and their
// revisions. The provider calls Refresh to load a backend's EndpointSlices and
// give them a revision, and full collection calls PruneBackends to drop backends
// that are no longer referenced.
type Source struct {
	once sync.Once

	// mu protects the revision counter and map below. Refresh holds it across
	// the load, so an older load of a backend can't get a newer revision.
	mu                    sync.Mutex
	lastRevision          uint64
	backendEndpointsByKey map[Key]*BackendEndpoints
}

func (s *Source) init() {
	s.once.Do(func() { s.backendEndpointsByKey = make(map[Key]*BackendEndpoints) })
}

// Refresh holds mu across the loader so an older capture cannot receive a
// newer revision. Errors leave the previous source entry untouched.
func (s *Source) Refresh(key Key, load func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error)) (*BackendEndpoints, error) {
	s.init()
	s.mu.Lock()
	defer s.mu.Unlock()
	backend, slices, err := load()
	if err != nil {
		return nil, err
	}
	current := NewBackendEndpoints(backend, slices)
	previous := s.backendEndpointsByKey[key]
	if sameEndpointData(previous, current) {
		current.Revision = previous.Revision
	} else {
		s.lastRevision++
		current.Revision = s.lastRevision
	}
	s.backendEndpointsByKey[key] = current
	return current, nil
}

// PruneBackends removes stored endpoints for backends absent from full
// Kubernetes collection. It waits for an in-flight Refresh. A Refresh that
// starts afterwards can restore an entry until the next full collection prunes
// it. Revisions never reset.
func (s *Source) PruneBackends(keys []Key) {
	s.init()
	next := make(map[Key]struct{}, len(keys))
	for _, key := range keys {
		next[key] = struct{}{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for key := range s.backendEndpointsByKey {
		if _, found := next[key]; !found {
			delete(s.backendEndpointsByKey, key)
		}
	}
}
