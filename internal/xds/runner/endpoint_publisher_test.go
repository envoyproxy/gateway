// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"github.com/telepresenceio/watchable"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/wrapperspb"
	corev1 "k8s.io/api/core/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var testBackend = ir.EndpointBackend{Kind: "Service", Namespace: "default", Name: "orders", UID: "uid-1"}

func fullEndpointResult(t *testing.T, revision uint64, address string, weight uint32) *fullPublication {
	t.Helper()
	var endpoints []*ir.DestinationEndpoint
	var hosts []*endpointv3.LbEndpoint
	if address != "" {
		endpoints = []*ir.DestinationEndpoint{{Host: address, Port: 8080}}
		hosts = []*endpointv3.LbEndpoint{{HostIdentifier: &endpointv3.LbEndpoint_Endpoint{Endpoint: &endpointv3.Endpoint{
			Address: &corev3.Address{Address: &corev3.Address_SocketAddress{SocketAddress: &corev3.SocketAddress{
				Address: address, PortSpecifier: &corev3.SocketAddress_PortValue{PortValue: 8080},
			}}},
		}}, LoadBalancingWeight: wrapperspb.UInt32(1)}}
	}
	assignment := &endpointv3.ClusterLoadAssignment{ClusterName: "orders", Endpoints: []*endpointv3.LocalityLbEndpoints{{
		LbEndpoints: hosts, LoadBalancingWeight: wrapperspb.UInt32(weight),
	}}}
	snapshot, err := cache.PrepareSnapshot(types.XdsResources{
		resourcev3.EndpointType: {assignment}, resourcev3.RouteType: {&routev3.RouteConfiguration{Name: "route"}},
	})
	require.NoError(t, err)
	return &fullPublication{ctx: context.Background(), snapshot: snapshot, contexts: types.EndpointContexts{
		"orders": {Settings: []*ir.DestinationSetting{{
			Name: "backend", Endpoints: endpoints, AddressType: new(ir.IP), Weight: new(weight),
			EndpointSource: &ir.EndpointSource{Backend: testBackend, PortName: "http", Protocol: "TCP", Revision: revision},
		}}},
	}}
}

func endpointUpdate(revision uint64, address string) *endpoint.Snapshot {
	var endpoints []*ir.DestinationEndpoint
	if address != "" {
		endpoints = []*ir.DestinationEndpoint{{Host: address, Port: 8080}}
	}
	return &endpoint.Snapshot{
		Backend: testBackend, Revision: revision, AddressType: new(ir.IP),
		Ports: map[endpoint.Port][]*ir.DestinationEndpoint{{Name: "http", Protocol: corev1.ProtocolTCP}: endpoints},
	}
}

func testPublisher(t *testing.T) (*endpointPublisher, *[]*cachev3.Snapshot) {
	t.Helper()
	var source endpoint.Source
	var installed []*cachev3.Snapshot
	pool := newEndpointPublishers(context.Background(), &source, func(_ string, snapshot *cachev3.Snapshot, _ context.Context) error {
		installed = append(installed, snapshot)
		return nil
	}, logging.DefaultLogger(io.Discard, egv1a1.LogLevelError))
	p := &endpointPublisher{
		pool: pool, key: "default/gateway", full: make(chan watchable.Update[string, *fullPublication], 2),
		space: make(chan struct{}, 1), wake: make(chan struct{}, 1), pending: make(map[ir.EndpointBackend]*endpoint.Snapshot),
	}
	pool.keys[p.key] = p
	return p, &installed
}

func applyFull(p *endpointPublisher, publication *fullPublication) {
	p.processFullBatch(watchable.Update[string, *fullPublication]{Key: p.key, Value: publication, Delete: publication.snapshot == nil})
}

func appliedRevision(p *endpointPublisher) uint64 {
	return p.contexts["orders"].Settings[0].EndpointSource.Revision
}

func assignmentOf(snapshot *cachev3.Snapshot) *endpointv3.ClusterLoadAssignment {
	return snapshot.Resources[cachetypes.Endpoint].Items["orders"].Resource.(*endpointv3.ClusterLoadAssignment)
}

func TestPublisherBothEndpointArrivalOrders(t *testing.T) {
	t.Run("fast before full", func(t *testing.T) {
		p, _ := testPublisher(t)
		applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
		p.offer(endpointUpdate(2, "10.0.0.2"))
		p.processEndpoints()
		require.Equal(t, uint64(2), appliedRevision(p))
		applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 3))
		require.Equal(t, uint64(2), appliedRevision(p))
		assignment := assignmentOf(p.snapshot)
		require.Equal(t, uint32(3), assignment.Endpoints[0].LoadBalancingWeight.Value)
		require.Equal(t, "10.0.0.2", assignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
	})
	t.Run("full before fast", func(t *testing.T) {
		p, installed := testPublisher(t)
		applyFull(p, fullEndpointResult(t, 7, "10.0.0.7", 3))
		for _, revision := range []uint64{6, 7} {
			p.offer(endpointUpdate(revision, "10.0.0.1"))
			p.processEndpoints()
		}
		require.Len(t, *installed, 1)
		require.Equal(t, uint64(7), appliedRevision(p))
		p.offer(endpointUpdate(8, "10.0.0.8"))
		p.processEndpoints()
		require.Equal(t, uint64(8), appliedRevision(p))
		require.Equal(t, uint32(3), assignmentOf(p.snapshot).Endpoints[0].LoadBalancingWeight.Value)
	})
}

func TestPublisherUnsafeUpdatesAndFallback(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 2, "10.0.0.2", 1))
	p.offer(endpointUpdate(3, ""))
	p.processEndpoints()
	require.Len(t, *installed, 1)
	require.Equal(t, uint64(2), appliedRevision(p))
	require.Empty(t, p.pending)
	p.offer(endpointUpdate(4, "10.0.0.4"))
	p.processEndpoints()
	require.Equal(t, uint64(4), appliedRevision(p))
	applyFull(p, fullEndpointResult(t, 3, "", 1))
	require.Equal(t, uint64(3), appliedRevision(p), "unsafe replay must record the full result's actual inputs")
	require.Empty(t, p.contexts["orders"].Settings[0].Endpoints)
}

func TestPublisherTracksEachConsumerSeparately(t *testing.T) {
	p, _ := testPublisher(t)
	full := fullEndpointResult(t, 1, "10.0.0.1", 1)
	empty := fullEndpointResult(t, 1, "", 1)
	secondAssignment := proto.Clone(assignmentOf(empty.snapshot)).(*endpointv3.ClusterLoadAssignment)
	secondAssignment.ClusterName = "empty"
	var err error
	full.snapshot, err = cache.PrepareSnapshot(types.XdsResources{resourcev3.EndpointType: {assignmentOf(full.snapshot), secondAssignment}})
	require.NoError(t, err)
	full.contexts["empty"] = empty.contexts["orders"]
	applyFull(p, full)
	p.offer(endpointUpdate(2, "10.0.0.2"))
	p.processEndpoints()
	require.Equal(t, uint64(2), appliedRevision(p))
	require.Equal(t, uint64(1), p.contexts["empty"].Settings[0].EndpointSource.Revision)
	require.Empty(t, p.contexts["empty"].Settings[0].Endpoints)
	require.Empty(t, p.pending, "the skipped consumer cannot retain the notification")
}

func TestPublisherRejectsOldBackendUID(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	p.offer(endpointUpdate(3, "10.0.0.3"))
	p.processEndpoints()
	recreated := fullEndpointResult(t, 2, "10.0.0.2", 1)
	recreated.contexts["orders"].Settings[0].EndpointSource.Backend.UID = "uid-2"
	applyFull(p, recreated)
	require.Equal(t, uint64(2), appliedRevision(p), "old UID data must not replay into a recreated backend")
	p.offer(endpointUpdate(4, "10.0.0.4"))
	p.processEndpoints()
	require.Len(t, *installed, 3)
	require.Equal(t, uint64(2), appliedRevision(p))
}

func TestPublisherNoClusterDropsPreviouslyAppliedInputs(t *testing.T) {
	p, _ := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	// A full translation can replace forwarding with a direct response.
	snapshot, err := cache.PrepareSnapshot(types.XdsResources{resourcev3.RouteType: {&routev3.RouteConfiguration{Name: "direct-response"}}})
	require.NoError(t, err)
	applyFull(p, &fullPublication{ctx: t.Context(), snapshot: snapshot})
	p.offer(endpointUpdate(2, "10.0.0.2"))
	p.processEndpoints()
	require.Same(t, snapshot, p.snapshot)
	require.Empty(t, p.contexts)
	require.Empty(t, p.backends)
	require.Empty(t, p.pending)
}

func TestPublisherDeletionBarrierSurvivesCoalescing(t *testing.T) {
	p, _ := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 9, "10.0.0.9", 1))
	p.full <- watchable.Update[string, *fullPublication]{Key: p.key, Value: fullEndpointResult(t, 1, "10.0.0.1", 1)}
	applyFull(p, &fullPublication{ctx: t.Context()})
	require.Equal(t, uint64(1), appliedRevision(p), "recreation cannot reuse pre-deletion endpoint state")
}

func TestPublisherValidationFailurePreservesAppliedState(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	p.offer(endpointUpdate(2, "bad address"))
	// An invalid port is rejected by Envoy protobuf validation.
	p.pending[testBackend].Ports[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Port = 70000
	p.processEndpoints()
	require.Len(t, *installed, 1)
	require.Equal(t, uint64(1), appliedRevision(p))
}

func TestPublisherDeletionAndRetryUseCurrentState(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 8, "10.0.0.8", 1))
	p.pool.install = func(string, *cachev3.Snapshot, context.Context) error { return errors.New("delivery failed") }
	applyFull(p, &fullPublication{ctx: context.Background()})
	require.True(t, p.retry)
	require.Nil(t, p.snapshot)
	require.Nil(t, p.contexts)
	p.offer(endpointUpdate(9, "10.0.0.9"))
	p.processEndpoints()
	require.Nil(t, p.snapshot)
	applyFull(p, fullEndpointResult(t, 10, "10.0.0.10", 1))
	p.pool.install = func(_ string, snapshot *cachev3.Snapshot, _ context.Context) error {
		*installed = append(*installed, snapshot)
		return nil
	}
	p.publish(context.Background())
	require.False(t, p.retry)
	require.Same(t, p.snapshot, (*installed)[len(*installed)-1])
	require.Equal(t, uint64(10), appliedRevision(p))
}

func TestPublisherSerializesRefreshAndFullPublication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var source endpoint.Source
	entered, release := make(chan struct{}), make(chan struct{})
	installed := make(chan *cachev3.Snapshot, 4)
	first := true
	pool := newEndpointPublishers(ctx, &source, func(_ string, snapshot *cachev3.Snapshot, _ context.Context) error {
		if first {
			first = false
			close(entered)
			<-release
		}
		installed <- snapshot
		return nil
	}, logging.DefaultLogger(io.Discard, egv1a1.LogLevelError))
	done := make(chan struct{})
	go func() { pool.run(); close(done) }()
	require.NoError(t, pool.submit("gateway", fullEndpointResult(t, 1, "10.0.0.1", 1)))
	<-entered
	pool.mu.Lock()
	pool.keys["gateway"].offer(endpointUpdate(2, "10.0.0.2"))
	pool.mu.Unlock()
	require.NoError(t, pool.submit("gateway", fullEndpointResult(t, 1, "10.0.0.1", 3)))
	close(release)
	for range 3 {
		select {
		case <-installed:
		case <-time.After(5 * time.Second):
			t.Fatal("publisher did not process full and fast publications")
		}
	}
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not stop")
	}
}
