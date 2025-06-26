// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"context"
	"sync"

	"github.com/telepresenceio/watchable"
)

// preSubscribedWatchableMap is a watchable.Map wrapper that has a list of pre-subscribed subscriptions.
type preSubscribedWatchableMap[K comparable, V any] struct {
	watchableMap  watchable.Map[K, V]
	subscriptions []<-chan watchable.Snapshot[K, V]
	mutex         sync.Mutex
	lastUsedIndex int
}

// newPreSubscribedWatchableMap creates a new preSubscribedWatchableMap subscribing to given count number of subscriptions.
// Given context is used to subscribe to the subscriptions.
func newPreSubscribedWatchableMap[K comparable, V any](ctx context.Context, count int) *preSubscribedWatchableMap[K, V] {
	store := &preSubscribedWatchableMap[K, V]{
		watchableMap:  watchable.Map[K, V]{},
		subscriptions: make([]<-chan watchable.Snapshot[K, V], count),
		mutex:         sync.Mutex{},
		lastUsedIndex: -1,
	}
	for i := range count {
		store.subscriptions[i] = store.watchableMap.Subscribe(ctx)
	}
	return store
}

// GetSubscription returns the next available subscription.
func (s *preSubscribedWatchableMap[K, V]) GetSubscription() <-chan watchable.Snapshot[K, V] {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.lastUsedIndex++
	if s.lastUsedIndex >= len(s.subscriptions) {
		panic("no available subscriptions, please report this as a bug as it should never happen")
	}
	return s.subscriptions[s.lastUsedIndex]
}

// Transparent methods for the underlying map.
// Store stores the given key-value pair in the map.
func (s *preSubscribedWatchableMap[K, V]) Store(key K, value V) {
	s.watchableMap.Store(key, value)
}

// Load returns the value for the given key.
func (s *preSubscribedWatchableMap[K, V]) Load(key K) (V, bool) {
	return s.watchableMap.Load(key)
}

// LoadAll returns all values in the map.
func (s *preSubscribedWatchableMap[K, V]) LoadAll() map[K]V {
	return s.watchableMap.LoadAll()
}

// Delete deletes the value for the given key.
func (s *preSubscribedWatchableMap[K, V]) Delete(key K) {
	s.watchableMap.Delete(key)
}

// Close closes the map and all its subscriptions.
func (s *preSubscribedWatchableMap[K, V]) Close() {
	s.watchableMap.Close()
}

// Len returns the number of elements in the map.
func (s *preSubscribedWatchableMap[K, V]) Len() int {
	return s.watchableMap.Len()
}
