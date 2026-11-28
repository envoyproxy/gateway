// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package publisher

import (
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// publishedEndpointStore holds, per IR key, the endpoint contexts of the
// published snapshot, with its endpoints indexed by backend.
type publishedEndpointStore struct {
	byKey map[string]types.EndpointContexts

	// clustersByBackend lists, per backend, the EDS clusters of each IR key whose
	// endpoints come from it, so an endpoint update finds the affected clusters
	// without a scan. replaceContexts and forget keep it in step with byKey.
	clustersByBackend map[ir.ServiceKey]clustersByKey
}

// clustersByKey maps an IR key to a set of EDS cluster names.
type clustersByKey map[string]map[string]struct{}

func newPublishedEndpointStore() *publishedEndpointStore {
	return &publishedEndpointStore{
		byKey:             make(map[string]types.EndpointContexts),
		clustersByBackend: make(map[ir.ServiceKey]clustersByKey),
	}
}

func (a *publishedEndpointStore) get(key string) types.EndpointContexts {
	return a.byKey[key]
}

// replaceContexts replaces a key's published endpoint contexts and reindexes them.
func (a *publishedEndpointStore) replaceContexts(key string, updatedContexts types.EndpointContexts) {
	a.reindex(key, a.byKey[key], updatedContexts)
	a.byKey[key] = updatedContexts
}

// updateEndpoints stores a key's updated endpoint contexts without reindexing.
// Use it only when every cluster keeps its backends, as on the fast path, which
// only swaps endpoints.
func (a *publishedEndpointStore) updateEndpoints(key string, updatedContexts types.EndpointContexts) {
	a.byKey[key] = updatedContexts
}

// forget drops a key's published endpoints and index entries.
func (a *publishedEndpointStore) forget(key string) {
	a.reindex(key, a.byKey[key], nil)
	delete(a.byKey, key)
}

// clustersUsing returns, per IR key, the EDS clusters whose endpoints come from
// the backend. A backend no cluster uses yields nothing.
func (a *publishedEndpointStore) clustersUsing(backend ir.ServiceKey) clustersByKey {
	return a.clustersByBackend[backend]
}

// reindex moves a key's index entries from the backends its old contexts used
// to the backends its updated ones use.
func (a *publishedEndpointStore) reindex(key string, old, updated types.EndpointContexts) {
	for backend := range backendsOf(old) {
		delete(a.clustersByBackend[backend], key)
		if len(a.clustersByBackend[backend]) == 0 {
			delete(a.clustersByBackend, backend)
		}
	}
	for cluster, args := range updated {
		for _, setting := range args.Settings {
			if setting.EndpointSource == nil {
				continue
			}
			byKey := a.clustersByBackend[setting.EndpointSource.Service]
			if byKey == nil {
				byKey = make(clustersByKey)
				a.clustersByBackend[setting.EndpointSource.Service] = byKey
			}
			if byKey[key] == nil {
				byKey[key] = make(map[string]struct{})
			}
			byKey[key][cluster] = struct{}{}
		}
	}
}

func backendsOf(contexts types.EndpointContexts) map[ir.ServiceKey]struct{} {
	backends := make(map[ir.ServiceKey]struct{})
	for _, args := range contexts {
		for _, setting := range args.Settings {
			if setting.EndpointSource != nil {
				backends[setting.EndpointSource.Service] = struct{}{}
			}
		}
	}
	return backends
}
