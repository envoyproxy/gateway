// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package safeset

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSafeSet(t *testing.T) {
	items := []string{"A", "B", "C"}

	var mutex sync.Mutex
	set := NewSafeSet[string]()
	data := func() {
		mutex.Lock()
		defer mutex.Unlock()

		setcp := set.Insert(items...)
		for _, item := range items {
			if !setcp.Has(item) {
				t.Errorf("%s does not exist", item)
			}
		}
		setcp.Delete("A")
		require.False(t, set.Has("A"))
	}

	var wg sync.WaitGroup
	const concurrency = 100

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			t.Logf("concurrency safe set access at %d", idx)
			data()
		}(i)
	}

	wg.Wait()
}
