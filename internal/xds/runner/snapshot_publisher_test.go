// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
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

const testSnapshotKey = "default/gateway"

var testBackend = ir.ServiceKey{Kind: "Service", Namespace: "default", Name: "orders"}

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
	snapshot, err := cache.PrepareSnapshot(context.Background(), types.XdsResources{
		resourcev3.EndpointType: {assignment}, resourcev3.RouteType: {&routev3.RouteConfiguration{Name: "route"}},
	})
	require.NoError(t, err)
	return &fullPublication{
		ctx: context.Background(), snapshot: snapshot,
		endpointState: types.EndpointState{
			Contexts: types.EndpointContexts{"orders": {Settings: []*ir.DestinationSetting{{Name: "backend", Weight: new(weight)}}}},
			Endpoints: types.EndpointsByCluster{"orders": {"backend": {
				Endpoints: endpoints, AddressType: new(ir.IP),
				Source: &ir.EndpointSource{Service: testBackend, PortName: "http", Protocol: "TCP", Revision: revision},
			}}},
		},
	}
}

func endpointUpdate(revision uint64, address string) *endpoint.BackendEndpoints {
	var endpoints []*ir.DestinationEndpoint
	if address != "" {
		endpoints = []*ir.DestinationEndpoint{{Host: address, Port: 8080}}
	}
	return &endpoint.BackendEndpoints{
		Backend: testBackend, Revision: revision, AddressType: new(ir.IP),
		EndpointsByPort: map[endpoint.Port][]*ir.DestinationEndpoint{{Name: "http", Protocol: corev1.ProtocolTCP}: endpoints},
	}
}

func testPublisher(t *testing.T) (*snapshotPublisher, *[]*cachev3.Snapshot) {
	t.Helper()
	var updates endpoint.UpdateQueue
	var installed []*cachev3.Snapshot
	p := newSnapshotPublisher(context.Background(), &updates, func(_ string, snapshot *cachev3.Snapshot, _ context.Context) error {
		installed = append(installed, snapshot)
		return nil
	}, logging.DefaultLogger(io.Discard, egv1a1.LogLevelError))
	return p, &installed
}

func applyFull(p *snapshotPublisher, publication *fullPublication) {
	p.processFullResults(watchable.Update[string, *fullPublication]{Key: testSnapshotKey, Value: publication, Delete: publication.snapshot == nil})
}

func publishedRevision(p *snapshotPublisher) uint64 {
	return p.publishedEndpoints.get(testSnapshotKey).Endpoints["orders"]["backend"].Source.Revision
}

func assignmentOf(snapshot *cachev3.Snapshot) *endpointv3.ClusterLoadAssignment {
	return snapshot.Resources[cachetypes.Endpoint].Items["orders"].Resource.(*endpointv3.ClusterLoadAssignment)
}

func TestPublisherBothEndpointArrivalOrders(t *testing.T) {
	t.Run("fast before full", func(t *testing.T) {
		p, _ := testPublisher(t)
		applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
		p.endpointUpdates.Add(endpointUpdate(2, "10.0.0.2"))
		p.processEndpointUpdates()
		require.Equal(t, uint64(2), publishedRevision(p))
		applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 3))
		require.Equal(t, uint64(2), publishedRevision(p))
		assignment := assignmentOf(p.snapshots[testSnapshotKey])
		require.Equal(t, uint32(3), assignment.Endpoints[0].LoadBalancingWeight.Value)
		require.Equal(t, "10.0.0.2", assignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
	})
	t.Run("full before fast", func(t *testing.T) {
		p, installed := testPublisher(t)
		applyFull(p, fullEndpointResult(t, 7, "10.0.0.7", 3))
		for _, revision := range []uint64{6, 7} {
			p.endpointUpdates.Add(endpointUpdate(revision, "10.0.0.1"))
			p.processEndpointUpdates()
		}
		require.Len(t, *installed, 1)
		require.Equal(t, uint64(7), publishedRevision(p))
		p.endpointUpdates.Add(endpointUpdate(8, "10.0.0.8"))
		p.processEndpointUpdates()
		require.Equal(t, uint64(8), publishedRevision(p))
		require.Equal(t, uint32(3), assignmentOf(p.snapshots[testSnapshotKey]).Endpoints[0].LoadBalancingWeight.Value)
	})
}

func TestPublisherUnsafeUpdatesAndFallback(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 2, "10.0.0.2", 1))
	p.endpointUpdates.Add(endpointUpdate(3, ""))
	p.processEndpointUpdates()
	require.Len(t, *installed, 1)
	require.Equal(t, uint64(2), publishedRevision(p))
	p.endpointUpdates.Add(endpointUpdate(4, "10.0.0.4"))
	p.processEndpointUpdates()
	require.Equal(t, uint64(4), publishedRevision(p))
	applyFull(p, fullEndpointResult(t, 3, "", 1))
	require.Equal(t, uint64(3), publishedRevision(p), "unsafe replay must record the full result's actual endpoints")
	require.Empty(t, p.publishedEndpoints.get(testSnapshotKey).Endpoints["orders"]["backend"].Endpoints)
}

func TestPublisherTracksEachConsumerSeparately(t *testing.T) {
	p, _ := testPublisher(t)
	full := fullEndpointResult(t, 1, "10.0.0.1", 1)
	empty := fullEndpointResult(t, 1, "", 1)
	secondAssignment := proto.Clone(assignmentOf(empty.snapshot)).(*endpointv3.ClusterLoadAssignment)
	secondAssignment.ClusterName = "empty"
	var err error
	full.snapshot, err = cache.PrepareSnapshot(context.Background(), types.XdsResources{resourcev3.EndpointType: {assignmentOf(full.snapshot), secondAssignment}})
	require.NoError(t, err)
	full.endpointState.Contexts["empty"] = empty.endpointState.Contexts["orders"]
	full.endpointState.Endpoints["empty"] = empty.endpointState.Endpoints["orders"]
	applyFull(p, full)
	p.endpointUpdates.Add(endpointUpdate(2, "10.0.0.2"))
	p.processEndpointUpdates()
	require.Equal(t, uint64(2), publishedRevision(p))
	require.Equal(t, uint64(1), p.publishedEndpoints.get(testSnapshotKey).Endpoints["empty"]["backend"].Source.Revision)
	require.Empty(t, p.publishedEndpoints.get(testSnapshotKey).Endpoints["empty"]["backend"].Endpoints)
}

func TestPublisherNoClusterDropsPreviouslyPublishedInputs(t *testing.T) {
	p, _ := testPublisher(t)
	require.Empty(t, p.publishedEndpoints.dependentsByBackend)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	require.Contains(t, p.publishedEndpoints.dependentsByBackend, testBackend)
	// A full translation can replace forwarding with a direct response.
	snapshot, err := cache.PrepareSnapshot(context.Background(), types.XdsResources{resourcev3.RouteType: {&routev3.RouteConfiguration{Name: "direct-response"}}})
	require.NoError(t, err)
	applyFull(p, &fullPublication{ctx: t.Context(), snapshot: snapshot})
	require.NotContains(t, p.publishedEndpoints.dependentsByBackend, testBackend)
	p.endpointUpdates.Add(endpointUpdate(2, "10.0.0.2"))
	p.processEndpointUpdates()
	require.Same(t, snapshot, p.snapshots[testSnapshotKey])
	require.Empty(t, p.publishedEndpoints.get(testSnapshotKey).Contexts)
	require.Empty(t, p.publishedEndpoints.get(testSnapshotKey).Endpoints)
}

func TestPublisherSharedBackendAcrossKeys(t *testing.T) {
	p, _ := testPublisher(t)
	installed := make(map[string][]*cachev3.Snapshot)
	p.publishSnapshot = func(key string, snapshot *cachev3.Snapshot, _ context.Context) error {
		installed[key] = append(installed[key], snapshot)
		return nil
	}
	const otherKey = "default/other"
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	p.processFullResults(watchable.Update[string, *fullPublication]{Key: otherKey, Value: fullEndpointResult(t, 1, "10.0.0.1", 3)})
	p.endpointUpdates.Add(endpointUpdate(2, "10.0.0.2"))
	p.processEndpointUpdates()
	for key, weight := range map[string]uint32{testSnapshotKey: 1, otherKey: 3} {
		require.Len(t, installed[key], 2)
		require.Equal(t, uint64(2), p.publishedEndpoints.get(key).Endpoints["orders"]["backend"].Source.Revision)
		assignment := assignmentOf(p.snapshots[key])
		require.Equal(t, "10.0.0.2", assignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
		require.Equal(t, weight, assignment.Endpoints[0].LoadBalancingWeight.Value)
	}

	applyFull(p, &fullPublication{ctx: t.Context()})
	require.NotContains(t, p.snapshots, testSnapshotKey)
	require.Contains(t, p.publishedEndpoints.dependentsByBackend[testBackend], otherKey)
	p.endpointUpdates.Add(endpointUpdate(3, "10.0.0.3"))
	p.processEndpointUpdates()
	require.Len(t, installed[testSnapshotKey], 3, "the deleted key must not receive endpoint publications")
	require.Nil(t, installed[testSnapshotKey][2])
	require.Len(t, installed[otherKey], 3)
	require.Equal(t, uint64(3), p.publishedEndpoints.get(otherKey).Endpoints["orders"]["backend"].Source.Revision)

	p.processFullResults(watchable.Update[string, *fullPublication]{Key: otherKey, Value: &fullPublication{ctx: t.Context()}, Delete: true})
	require.Empty(t, p.snapshots)
	require.Empty(t, p.publishedEndpoints.byKey)
	require.Empty(t, p.publishedEndpoints.dependentsByBackend)
}

func TestPublisherValidationFailurePreservesPublishedState(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	update := endpointUpdate(2, "bad address")
	// An invalid port is rejected by Envoy protobuf validation.
	update.EndpointsByPort[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Port = 70000
	p.endpointUpdates.Add(update)
	p.processEndpointUpdates()
	require.Len(t, *installed, 1)
	require.Equal(t, uint64(1), publishedRevision(p))
}

func TestPublisherSerializesRefreshAndFullPublication(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var updates endpoint.UpdateQueue
	entered, release := make(chan struct{}), make(chan struct{})
	installed := make(chan *cachev3.Snapshot, 4)
	first := true
	publisher := newSnapshotPublisher(ctx, &updates, func(_ string, snapshot *cachev3.Snapshot, _ context.Context) error {
		if first {
			first = false
			close(entered)
			<-release
		}
		installed <- snapshot
		return nil
	}, logging.DefaultLogger(io.Discard, egv1a1.LogLevelError))
	done := make(chan struct{})
	go func() { publisher.run(); close(done) }()
	require.NoError(t, publisher.enqueueFullResult("gateway", fullEndpointResult(t, 1, "10.0.0.1", 1)))
	<-entered
	updates.Add(endpointUpdate(2, "10.0.0.2"))
	require.NoError(t, publisher.enqueueFullResult("gateway", fullEndpointResult(t, 1, "10.0.0.1", 3)))
	close(release)
	var last *cachev3.Snapshot
	for range 3 {
		select {
		case last = <-installed:
		case <-time.After(5 * time.Second):
			t.Fatal("publisher did not process full and fast publications")
		}
	}
	assignment := assignmentOf(last)
	require.Equal(t, "10.0.0.2", assignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
	require.Equal(t, uint32(3), assignment.Endpoints[0].LoadBalancingWeight.Value)
	cancel()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("publisher did not stop")
	}
	require.Empty(t, publisher.publishedEndpoints.dependentsByBackend)
}

func TestPublisherCancellationUnblocksFullQueue(t *testing.T) {
	p, _ := testPublisher(t)
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	p.ctx = ctx
	full := fullEndpointResult(t, 1, "10.0.0.1", 1)
	for range cap(p.fullResults) {
		require.NoError(t, p.enqueueFullResult(testSnapshotKey, full))
	}
	started := make(chan struct{})
	result := make(chan error, 1)
	go func() {
		close(started)
		result <- p.enqueueFullResult("other", full)
	}()
	<-started
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("cancellation did not unblock submission to the full queue")
	}
	require.ErrorIs(t, p.enqueueFullResult("other", full), context.Canceled)
}

func TestPublisherFullResultSurvivesReplayFailure(t *testing.T) {
	p, installed := testPublisher(t)
	applyFull(p, fullEndpointResult(t, 1, "10.0.0.1", 1))
	// Corrupt the published endpoints so replaying them fails validation.
	published := fullEndpointResult(t, 5, "10.0.0.5", 1).endpointState
	published.Endpoints["orders"]["backend"].Endpoints[0].Port = 70000
	p.publishedEndpoints.set(testSnapshotKey, published)

	full := fullEndpointResult(t, 2, "10.0.0.2", 3)
	applyFull(p, full)
	require.Len(t, *installed, 2)
	require.Same(t, full.snapshot, p.snapshots[testSnapshotKey])
	require.Equal(t, uint64(2), publishedRevision(p))
}
