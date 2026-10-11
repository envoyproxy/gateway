// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

// fakeManager lets Start return an error immediately while its cache never
// reports synced, reproducing a startup failure that controller-runtime
// surfaces before the caches ever sync (e.g. a metrics/health listener bind
// conflict, which is started before the caches in manager.Start).
type fakeManager struct {
	manager.Manager
	startErr error
}

func (f *fakeManager) Start(context.Context) error {
	return f.startErr
}

func (f *fakeManager) GetCache() cache.Cache {
	return blockingCache{}
}

// blockingCache's WaitForCacheSync blocks until the context it is given is
// done, simulating caches that never sync.
type blockingCache struct {
	cache.Cache
}

func (blockingCache) WaitForCacheSync(ctx context.Context) bool {
	<-ctx.Done()
	return false
}

// TestProviderStartReturnsOnStartupFailureBeforeCacheSync reproduces a startup
// failure that happens before caches ever sync. Provider.Start must still
// return promptly instead of hanging in WaitForCacheSync, which would
// otherwise never observe either a completed cache sync or ctx cancellation.
func TestProviderStartReturnsOnStartupFailureBeforeCacheSync(t *testing.T) {
	startErr := errors.New("boom")
	p := &Provider{
		manager:       &fakeManager{startErr: startErr},
		providerReady: make(chan struct{}),
	}

	done := make(chan error, 1)
	go func() {
		done <- p.Start(context.Background())
	}()

	select {
	case err := <-done:
		require.Equal(t, startErr, err)
	case <-time.After(5 * time.Second):
		t.Fatal("Provider.Start did not return after manager startup failed -- deadlocked waiting for cache sync")
	}
}
