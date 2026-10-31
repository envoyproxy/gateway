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
	source.SetReferences("full", []Key{key})
	first, changed, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.True(t, changed)
	require.Positive(t, first.Revision)
	repeated, changed, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.False(t, changed)
	require.Equal(t, first.Revision, repeated.Revision)
	_, _, err = source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
		return backend, nil, errors.New("collection failed")
	})
	require.Error(t, err)
	require.Same(t, repeated, source.entries[key])
	source.SetReferences("full", nil)
	require.Empty(t, source.entries)
	source.SetReferences("full", []Key{key})
	recreated, _, err := source.Refresh(key, load)
	require.NoError(t, err)
	require.Greater(t, recreated.Revision, first.Revision)
	source.Notify(recreated)
	source.Notify(first)
	require.Same(t, recreated, source.Drain()[key])
}

func TestSourceSerializesCaptureWithRevision(t *testing.T) {
	var source Source
	backend := ir.EndpointBackend{Kind: "Service", Name: "orders", UID: "one"}
	key := BackendKey(backend)
	entered, release := make(chan struct{}), make(chan struct{})
	results := make(chan *Snapshot, 2)
	go func() {
		snapshot, _, err := source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			close(entered)
			<-release
			return backend, endpointSlices("10.0.0.1"), nil
		})
		if err == nil {
			results <- snapshot
		}
	}()
	<-entered
	started := make(chan struct{})
	go func() {
		close(started)
		snapshot, _, err := source.Refresh(key, func() (ir.EndpointBackend, []*discoveryv1.EndpointSlice, error) {
			return backend, endpointSlices("10.0.0.2"), nil
		})
		if err == nil {
			results <- snapshot
		}
	}()
	<-started
	close(release)
	a, b := <-results, <-results
	if a.Revision > b.Revision {
		a, b = b, a
	}
	require.Equal(t, "10.0.0.1", a.Ports[Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
	require.Equal(t, "10.0.0.2", b.Ports[Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
}

func TestSourceKeepsPublishedReferences(t *testing.T) {
	var source Source
	key := Key{Kind: "Service", Name: "old-backend"}
	source.SetReferences("provider", []Key{key})
	source.SetReferences("publisher/gateway", []Key{key})
	source.SetReferences("provider", nil)
	require.True(t, source.Tracked(key))
	source.SetReferences("publisher/gateway", nil)
	require.False(t, source.Tracked(key))
}
