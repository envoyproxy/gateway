// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func newTestCache(size int) (*ValidationCache, *atomic.Int32) {
	var calls atomic.Int32
	c := NewValidationCache(size)
	c.validate = func(*egv1a1.ProxyBootstrap) error {
		calls.Add(1)
		return nil
	}
	return c, &calls
}

func replaceBootstrap(value string) *egv1a1.ProxyBootstrap {
	return &egv1a1.ProxyBootstrap{
		Type:  new(egv1a1.BootstrapTypeReplace),
		Value: &value,
	}
}

func TestValidationCache(t *testing.T) {
	tests := []struct {
		name        string
		cfgs        []*egv1a1.ProxyBootstrap
		wantCalls   int32
		wantEntries int
	}{
		{
			name:        "same pointer is validated once",
			cfgs:        []*egv1a1.ProxyBootstrap{replaceBootstrap("a")},
			wantCalls:   1,
			wantEntries: 1,
		},
		{
			name:        "distinct pointers with equal content share an entry",
			cfgs:        []*egv1a1.ProxyBootstrap{replaceBootstrap("a"), replaceBootstrap("a"), replaceBootstrap("a")},
			wantCalls:   1,
			wantEntries: 1,
		},
		{
			name:        "different content gets different entries",
			cfgs:        []*egv1a1.ProxyBootstrap{replaceBootstrap("a"), replaceBootstrap("b"), replaceBootstrap("a")},
			wantCalls:   2,
			wantEntries: 2,
		},
		{
			name:        "nil does not touch the cache",
			cfgs:        []*egv1a1.ProxyBootstrap{nil, nil},
			wantCalls:   0,
			wantEntries: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			c, calls := newTestCache(10)
			for _, cfg := range tc.cfgs {
				require.NoError(t, c.Validate(cfg))
				require.NoError(t, c.Validate(cfg))
			}
			require.Equal(t, tc.wantCalls, calls.Load())
			require.Equal(t, tc.wantEntries, c.cache.Len())
		})
	}
}

func TestValidationCacheEviction(t *testing.T) {
	c, calls := newTestCache(2)

	a, b, d := replaceBootstrap("a"), replaceBootstrap("b"), replaceBootstrap("c")
	require.NoError(t, c.Validate(a))
	require.NoError(t, c.Validate(b))
	require.NoError(t, c.Validate(d))
	require.Equal(t, 2, c.cache.Len())
	require.Equal(t, int32(3), calls.Load())

	// "a" was the least recently used, so it is validated again.
	require.NoError(t, c.Validate(a))
	require.Equal(t, int32(4), calls.Load())
}

func TestValidationCacheConcurrent(t *testing.T) {
	c, calls := newTestCache(10)

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			require.NoError(t, c.Validate(replaceBootstrap("a")))
			require.NoError(t, c.Validate(nil))
		})
	}
	wg.Wait()

	require.Equal(t, 1, c.cache.Len())
	require.GreaterOrEqual(t, calls.Load(), int32(1))
}

func TestValidationCacheMatchesValidate(t *testing.T) {
	c := NewValidationCache(10)
	valid := newJSONPatchBootstrap(patchOp("add", "/stats_config/use_all_default_tags", `false`))
	invalid := replaceBootstrap("not: [valid")

	for _, cfg := range []*egv1a1.ProxyBootstrap{valid, invalid} {
		want := Validate(cfg)
		for range 2 {
			if want == nil {
				require.NoError(t, c.Validate(cfg))
			} else {
				require.EqualError(t, c.Validate(cfg), want.Error())
			}
		}
	}
}
