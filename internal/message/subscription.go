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

// SubscriptionList manages a list of subscription channels with thread-safe access
type SubscriptionList[K comparable, V any] struct {
	subscriptions []<-chan watchable.Snapshot[K, V]
	mutex         sync.Mutex
	lastUsedIndex int
}

// Subscribable represents an object that can be subscribed to for snapshots
type Subscribable[K comparable, V any] interface {
	Subscribe(ctx context.Context) <-chan watchable.Snapshot[K, V]
}

// NewSubscriptionList creates a new subscription list by calling Subscribe count times on the subscribable object
func NewSubscriptionList[K comparable, V any](ctx context.Context, subscribable Subscribable[K, V], count int) *SubscriptionList[K, V] {
	subscriptions := make([]<-chan watchable.Snapshot[K, V], count)
	for i := range count {
		subscriptions[i] = subscribable.Subscribe(ctx)
	}

	return &SubscriptionList[K, V]{
		subscriptions: subscriptions,
		mutex:         sync.Mutex{},
		lastUsedIndex: -1,
	}
}

// GetNextAvailable returns the next available subscription channel that hasn't been consumed yet.
// This method is thread-safe and atomically increments the last used index.
func (c *SubscriptionList[K, V]) GetNextAvailable() <-chan watchable.Snapshot[K, V] {
	c.mutex.Lock()
	nextIndex := c.lastUsedIndex + 1
	if nextIndex >= len(c.subscriptions) {
		c.mutex.Unlock()
		panic("no available subscriptions")
	}
	c.lastUsedIndex = nextIndex
	c.mutex.Unlock()
	return c.subscriptions[nextIndex]
}
