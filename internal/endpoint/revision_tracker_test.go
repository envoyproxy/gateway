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

func TestRevisionTrackerKeepsEntryOnError(t *testing.T) {
	var tracker RevisionTracker
	key := ir.ServiceKey{Kind: "Service", Namespace: "default", Name: "orders"}
	stored, err := tracker.Track(key, func() ([]*discoveryv1.EndpointSlice, error) {
		return endpointSlices("10.0.0.1"), nil
	})
	require.NoError(t, err)
	_, err = tracker.Track(key, func() ([]*discoveryv1.EndpointSlice, error) {
		return nil, errors.New("collection failed")
	})
	require.Error(t, err)
	require.Same(t, stored, tracker.lastLoaded[key])
}

func TestRevisionTrackerSerializesCaptureWithRevision(t *testing.T) {
	var tracker RevisionTracker
	key := ir.ServiceKey{Kind: "Service", Name: "orders"}
	entered, release := make(chan struct{}), make(chan struct{})
	results := make(chan *BackendEndpoints, 2)
	go func() {
		backendEndpoints, err := tracker.Track(key, func() ([]*discoveryv1.EndpointSlice, error) {
			close(entered)
			<-release
			return endpointSlices("10.0.0.1"), nil
		})
		if err == nil {
			results <- backendEndpoints
		}
	}()
	<-entered
	started := make(chan struct{})
	go func() {
		close(started)
		backendEndpoints, err := tracker.Track(key, func() ([]*discoveryv1.EndpointSlice, error) {
			return endpointSlices("10.0.0.2"), nil
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

func TestRevisionTrackerPrunesUnreferencedBackends(t *testing.T) {
	var tracker RevisionTracker
	removed := ir.ServiceKey{Kind: "Service", Name: "removed"}
	retained := ir.ServiceKey{Kind: "Service", Name: "retained"}
	var retainedEndpoints *BackendEndpoints
	for _, backend := range []ir.ServiceKey{removed, retained} {
		backendEndpoints, err := tracker.Track(backend, func() ([]*discoveryv1.EndpointSlice, error) {
			return endpointSlices("10.0.0.1"), nil
		})
		require.NoError(t, err)
		retainedEndpoints = backendEndpoints
	}
	revision := tracker.lastRevision
	tracker.PruneBackends([]ir.ServiceKey{retained, retained})
	require.NotContains(t, tracker.lastLoaded, removed)
	require.Same(t, retainedEndpoints, tracker.lastLoaded[retained])
	require.Equal(t, revision, tracker.lastRevision)
}

func TestRevisionTrackerReusesUnchangedSlices(t *testing.T) {
	var tracker RevisionTracker
	key := ir.ServiceKey{Kind: "Service", Namespace: "default", Name: "orders"}
	slices := endpointSlices("10.0.0.1")
	slices[0].ResourceVersion = "1"
	load := func() ([]*discoveryv1.EndpointSlice, error) { return slices, nil }
	first, err := tracker.Track(key, load)
	require.NoError(t, err)
	again, err := tracker.Track(key, load)
	require.NoError(t, err)
	require.Same(t, first, again, "unchanged slices must reuse the stored entry")

	slices = endpointSlices("10.0.0.2")
	slices[0].ResourceVersion = "2"
	changed, err := tracker.Track(key, load)
	require.NoError(t, err)
	require.Greater(t, changed.Revision, first.Revision)
}
