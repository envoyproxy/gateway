// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message

import (
	"github.com/telepresenceio/watchable"
)

// WatchableMap wrap watchable.Map override the LoadAll with no DeepCopy.
type WatchableMap[K comparable, V any] struct {
	watchable.Map[K, V]
}

func (w *WatchableMap[K, V]) LoadAll() map[K]V {
	return w.LoadAllMatching(func(K, V) bool {
		return false
	})
}
