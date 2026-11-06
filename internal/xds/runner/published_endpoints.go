// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// publishedEndpointStore holds, per IR key, the endpoint state of the
// published snapshot, with its endpoints indexed by backend.
type publishedEndpointStore struct {
	byKey map[string]types.EndpointState

	// dependentsByBackend lists, per backend, the IR keys and EDS cluster names
	// whose endpoints come from it, so an endpoint update finds the affected
	// contexts without a scan. set and forget keep it in step with byKey.
	dependentsByBackend map[ir.ServiceKey]map[string]map[string]struct{}
}

func newPublishedEndpointStore() *publishedEndpointStore {
	return &publishedEndpointStore{
		byKey:               make(map[string]types.EndpointState),
		dependentsByBackend: make(map[ir.ServiceKey]map[string]map[string]struct{}),
	}
}

func (a *publishedEndpointStore) get(key string) types.EndpointState {
	return a.byKey[key]
}

// set replaces a key's published endpoint state and reindexes it.
func (a *publishedEndpointStore) set(key string, next types.EndpointState) {
	a.reindex(key, a.byKey[key].Endpoints, next.Endpoints)
	a.byKey[key] = next
}

// forget drops a key's published endpoints and index entries.
func (a *publishedEndpointStore) forget(key string) {
	a.reindex(key, a.byKey[key].Endpoints, nil)
	delete(a.byKey, key)
}

// snapshotsUsing returns the IR keys whose published snapshot takes endpoints
// from the backend, each with the EDS cluster names that use it. A backend no
// snapshot uses yields nothing.
func (a *publishedEndpointStore) snapshotsUsing(backend ir.ServiceKey) map[string]map[string]struct{} {
	return a.dependentsByBackend[backend]
}

// reindex swaps a key's index entries from the backends its old published
// endpoints used to the backends its new ones use.
func (a *publishedEndpointStore) reindex(key string, old, next types.EndpointsByCluster) {
	for backend := range backendsOf(old) {
		delete(a.dependentsByBackend[backend], key)
		if len(a.dependentsByBackend[backend]) == 0 {
			delete(a.dependentsByBackend, backend)
		}
	}
	for cluster, settings := range next {
		for _, input := range settings {
			if input.Source == nil {
				continue
			}
			backend := input.Source.Service
			if a.dependentsByBackend[backend] == nil {
				a.dependentsByBackend[backend] = make(map[string]map[string]struct{})
			}
			if a.dependentsByBackend[backend][key] == nil {
				a.dependentsByBackend[backend][key] = make(map[string]struct{})
			}
			a.dependentsByBackend[backend][key][cluster] = struct{}{}
		}
	}
}

func backendsOf(endpoints types.EndpointsByCluster) map[ir.ServiceKey]struct{} {
	backends := make(map[ir.ServiceKey]struct{})
	for _, settings := range endpoints {
		for _, entry := range settings {
			if entry.Source != nil {
				backends[entry.Source.Service] = struct{}{}
			}
		}
	}
	return backends
}
