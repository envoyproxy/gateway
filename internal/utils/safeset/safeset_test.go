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

	set := NewSafeSet[string]()
	data := func() {
		setcp := set.Insert(items...)
		for _, item := range items {
			if !setcp.Has(item) {
				t.Errorf("%s does not exist", item)
			}
		}
		require.Equal(t, 3, len(setcp.Values))
		setcp.Delete("A")
		require.Equal(t, false, set.Has("A"))
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
