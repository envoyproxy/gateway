// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cmd

import (
	"context"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

// leaseRunner stands in for the provider runner: its manager releases the lease when its
// context is canceled, recorded here as leaseReleased.
type leaseRunner struct {
	leaseReleased chan struct{}
	wg            sync.WaitGroup
}

func (r *leaseRunner) Name() string { return "lease" }

func (r *leaseRunner) Start(ctx context.Context) error {
	r.wg.Go(func() {
		<-ctx.Done()
		close(r.leaseReleased)
	})
	return nil
}

func (r *leaseRunner) Close() error {
	r.wg.Wait()
	return nil
}

// workRunner stands in for the infrastructure runner: leader-only work that is still in
// flight when its context is canceled, and finishes only when finish is closed.
type workRunner struct {
	canceled chan struct{}
	finish   chan struct{}
	finished chan struct{}
	wg       sync.WaitGroup
}

func (r *workRunner) Name() string { return "work" }

func (r *workRunner) Start(ctx context.Context) error {
	r.wg.Go(func() {
		<-ctx.Done()
		close(r.canceled)
		<-r.finish
		close(r.finished)
	})
	return nil
}

func (r *workRunner) Close() error {
	r.wg.Wait()
	return nil
}

type failingRunner struct{}

func (failingRunner) Name() string                { return "failing" }
func (failingRunner) Start(context.Context) error { return errors.New("start failed") }
func (failingRunner) Close() error                { return nil }

func testServerConfig() *config.Server {
	return &config.Server{Logger: logging.DefaultLogger(io.Discard, egv1a1.LogLevelInfo)}
}

// TestRunRunnersReleasesLeaseLast checks that the lease holder's context is canceled only after
// every other runner has closed, so that the lease is not released while leader-only work is
// still in flight.
func TestRunRunnersReleasesLeaseLast(t *testing.T) {
	lease := &leaseRunner{leaseReleased: make(chan struct{})}
	work := &workRunner{canceled: make(chan struct{}), finish: make(chan struct{}), finished: make(chan struct{})}
	cleanedUp := make(chan struct{})

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runRunners(ctx, testServerConfig(), []Runner{lease, work}, lease, func() { close(cleanedUp) })
	}()

	cancel()

	// The other runners are canceled at once: the shutdown has started.
	select {
	case <-work.canceled:
	case <-time.After(5 * time.Second):
		t.Fatal("the work runner's context was not canceled")
	}

	// While the work is still in flight, the lease is held.
	select {
	case <-lease.leaseReleased:
		t.Fatal("the lease was released while leader-only work was still in flight")
	case <-time.After(200 * time.Millisecond):
	}

	close(work.finish)

	select {
	case <-lease.leaseReleased:
	case <-time.After(5 * time.Second):
		t.Fatal("the lease was not released after the other runners closed")
	}
	select {
	case <-work.finished:
	default:
		t.Fatal("the lease was released before the work runner finished")
	}

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("runRunners did not return")
	}
	select {
	case <-cleanedUp:
	default:
		t.Fatal("cleanup did not run")
	}
}

// TestRunRunnersSharedContextReleasesLeaseEarly is the control: with the lease holder on the same
// context as the others (no lease holder named), the lease is released while the work is still in
// flight, so the test above can tell the two orders apart.
func TestRunRunnersSharedContextReleasesLeaseEarly(t *testing.T) {
	lease := &leaseRunner{leaseReleased: make(chan struct{})}
	work := &workRunner{canceled: make(chan struct{}), finish: make(chan struct{}), finished: make(chan struct{})}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- runRunners(ctx, testServerConfig(), []Runner{lease, work}, nil, nil)
	}()

	cancel()

	select {
	case <-lease.leaseReleased:
	case <-time.After(5 * time.Second):
		t.Fatal("with a shared context the lease was not released")
	}
	select {
	case <-work.finished:
		t.Fatal("the work finished before it was allowed to")
	default:
	}

	close(work.finish)
	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("runRunners did not return")
	}
}

// TestRunRunnersStartFailureClosesLeaseHolder checks that when a later runner fails to start,
// the lease holder is still stopped and closed, and the error is returned.
func TestRunRunnersStartFailureClosesLeaseHolder(t *testing.T) {
	lease := &leaseRunner{leaseReleased: make(chan struct{})}

	err := runRunners(context.Background(), testServerConfig(), []Runner{lease, failingRunner{}}, lease, nil)
	require.EqualError(t, err, "start failed")

	select {
	case <-lease.leaseReleased:
	default:
		t.Fatal("the lease holder was not stopped after a start failure")
	}
}
