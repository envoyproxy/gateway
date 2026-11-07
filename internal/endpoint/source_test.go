// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func endpointSlices(address string) []*discoveryv1.EndpointSlice {
	return []*discoveryv1.EndpointSlice{{
		ObjectMeta:  metav1.ObjectMeta{Name: "orders-1", Namespace: "default"},
		AddressType: discoveryv1.AddressTypeIPv4,
		Ports:       []discoveryv1.EndpointPort{{Name: new("http"), Protocol: new(corev1.ProtocolTCP), Port: new(int32(8080))}},
		Endpoints:   []discoveryv1.Endpoint{{Addresses: []string{address}}},
	}}
}

func TestSourceRevisionsAndFailures(t *testing.T) {
	var source Source
	backend := ir.EndpointBackend{Kind: "Service", Namespace: "default", Name: "orders", UID: "one"}
	key := BackendKey(backend)
	load := func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
		return backend, endpointSlices("10.0.0.1"), nil
	}
	first, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.Positive(t, first.Revision)
	repeated, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.Equal(t, first.Revision, repeated.Revision)
	_, err = source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
		return backend, nil, errors.New("collection failed")
	})
	require.Error(t, err)
	require.Same(t, repeated, source.backendEndpointsByKey[key])
	source.PruneBackends(nil)
	require.Empty(t, source.backendEndpointsByKey)
	recreated, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.Greater(t, recreated.Revision, first.Revision)
}

func TestSourceSerializesCaptureWithRevision(t *testing.T) {
	var source Source
	backend := ir.EndpointBackend{Kind: "Service", Name: "orders", UID: "one"}
	key := BackendKey(backend)
	entered, release := make(chan struct{}), make(chan struct{})
	results := make(chan *BackendEndpoints, 2)
	go func() {
		backendEndpoints, err := source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			close(entered)
			<-release
			return backend, endpointSlices("10.0.0.1"), nil
		})
		if err == nil {
			results <- backendEndpoints
		}
	}()
	<-entered
	started := make(chan struct{})
	go func() {
		close(started)
		backendEndpoints, err := source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			return backend, endpointSlices("10.0.0.2"), nil
		})
		if err == nil {
			results <- backendEndpoints
		}
	}()
	<-started
	close(release)
	a, b := <-results, <-results
	if a.Revision > b.Revision {
		a, b = b, a
	}
	require.Equal(t, "10.0.0.1", a.EndpointsByPort[Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
	require.Equal(t, "10.0.0.2", b.EndpointsByPort[Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
}

func TestSourcePrunesUnreferencedBackends(t *testing.T) {
	var source Source
	removed := ir.EndpointBackend{Kind: "Service", Name: "removed", UID: "one"}
	retained := ir.EndpointBackend{Kind: "Service", Name: "retained", UID: "two"}
	removedKey, retainedKey := BackendKey(removed), BackendKey(retained)
	var retainedEndpoints *BackendEndpoints
	for _, backend := range []ir.EndpointBackend{removed, retained} {
		backendEndpoints, err := source.Refresh(BackendKey(backend), func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			return backend, endpointSlices("10.0.0.1"), nil
		})
		require.NoError(t, err)
		retainedEndpoints = backendEndpoints
	}
	revision := source.lastRevision
	source.PruneBackends([]Key{retainedKey, retainedKey})
	require.NotContains(t, source.backendEndpointsByKey, removedKey)
	require.Same(t, retainedEndpoints, source.backendEndpointsByKey[retainedKey])
	require.Equal(t, revision, source.lastRevision)
}

func TestSourcePruneWaitsForInFlightRefresh(t *testing.T) {
	var source Source
	backend := ir.EndpointBackend{Kind: "Service", Name: "orders", UID: "one"}
	key := BackendKey(backend)
	entered, release, done, pruned := make(chan struct{}), make(chan struct{}), make(chan struct{}), make(chan struct{})
	var captured *BackendEndpoints
	var captureErr error
	go func() {
		captured, captureErr = source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			close(entered)
			<-release
			return backend, endpointSlices("10.0.0.1"), nil
		})
		close(done)
	}()
	<-entered
	// Refresh holds the lock, so this prune runs only after the refresh stores its entry.
	go func() {
		source.PruneBackends(nil)
		close(pruned)
	}()
	close(release)
	<-done
	<-pruned
	require.NoError(t, captureErr)
	require.NotNil(t, captured)
	require.Empty(t, source.backendEndpointsByKey)
}
