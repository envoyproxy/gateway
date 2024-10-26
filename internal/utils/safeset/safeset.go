// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package safeset

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/sets"
)

type SafeSet[T comparable] struct {
	lock   sync.RWMutex
	Values sets.Set[T]
}

func NewSafeSet[T comparable](items ...T) *SafeSet[T] {
	return &SafeSet[T]{Values: sets.New[T](items...)}
}

func (s *SafeSet[T]) Has(item T) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.Values.Has(item)
}

func (s *SafeSet[T]) Insert(item ...T) *SafeSet[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Values.Insert(item...)
	return s
}

func (s *SafeSet[T]) Delete(item ...T) *SafeSet[T] {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Values.Delete(item...)
	return s
}
