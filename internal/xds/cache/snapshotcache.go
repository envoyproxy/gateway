// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// This file contains code derived from Contour,
// https://github.com/projectcontour/contour
// from the source file
// https://github.com/projectcontour/contour/blob/main/internal/xds/v3/snapshotter.go
// and is provided here subject to the following:
// Copyright Project Contour Authors
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/metrics"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	Hash   = cachev3.IDHash{}
	tracer = otel.Tracer("envoy-gateway/xds/snapshotcache")
)

// SnapshotCacheWithCallbacks uses the go-control-plane SimpleCache to store snapshots of
// Envoy resources, sliced by Node ID so that we can do incremental xDS properly.
// It does this by also implementing callbacks to make sure that the cache is kept
// up to date for each new node.
//
// Having the cache also implement the callbacks is a little bit hacky, but it makes sure
// that all the required bookkeeping happens.
// TODO(youngnick): Talk to the go-control-plane maintainers and see if we can upstream
// this in a better way.
type SnapshotCacheWithCallbacks interface {
	cachev3.SnapshotCache
	serverv3.Callbacks
	GenerateNewSnapshot(string, types.XdsResources, context.Context) error
	PublishSnapshot(string, *cachev3.Snapshot, context.Context) error
	SnapshotHasIrKey(string) bool
	GetIrKeys() []string
}

type snapshotMap map[string]*cachev3.Snapshot

type nodeInfoMap map[int64]*corev3.Node

type nodeFrequencyMap map[string]int

type streamDurationMap map[int64]time.Time

type snapshotCache struct {
	cachev3.SnapshotCache
	streamIDNodeInfo    nodeInfoMap
	nodeFrequency       nodeFrequencyMap
	streamDuration      streamDurationMap
	deltaStreamDuration streamDurationMap
	lastSnapshot        snapshotMap
	log                 *zap.SugaredLogger
	mu                  sync.Mutex
}

// PublishSnapshot installs an already prepared immutable snapshot. A nil snapshot
// removes the IR key while sending an empty snapshot to its connected Envoys.
func (s *snapshotCache) PublishSnapshot(irKey string, snapshot *cachev3.Snapshot, ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if snapshot == nil {
		delete(s.lastSnapshot, irKey)
		var err error
		snapshot, err = PrepareSnapshot(ctx, nil)
		if err != nil {
			return err
		}
	} else {
		s.lastSnapshot[irKey] = snapshot
	}
	var errs error
	for _, node := range s.getNodeIDs(irKey) {
		s.log.Debugf("Generating a snapshot with Node %s", node)
		if err := s.SetSnapshot(ctx, node, snapshot); err != nil {
			errs = errors.Join(errs, err)
			xdsSnapshotUpdateTotal.WithFailure(metrics.ReasonError, nodeIDLabel.Value(node)).Increment()
		} else {
			xdsSnapshotUpdateTotal.WithSuccess(nodeIDLabel.Value(node)).Increment()
		}
	}
	return errs
}

// GenerateNewSnapshot takes a table of resources (the output from the IR->xDS
// translator), builds a snapshot and publishes it. Nil resources delete the IR key.
func (s *snapshotCache) GenerateNewSnapshot(irKey string, resources types.XdsResources, ctx context.Context) error {
	_, span := tracer.Start(ctx, "SnapshotCache.GenerateNewSnapshot")
	defer span.End()

	var snapshot *cachev3.Snapshot
	if resources != nil {
		var err error
		if snapshot, err = PrepareSnapshot(ctx, resources); err != nil {
			return err
		}
	}
	return s.PublishSnapshot(irKey, snapshot, ctx)
}

// PrepareSnapshot builds a snapshot and its delta xDS hashes, so the work is done
// before taking the cache lock. The version is the trace ID when ctx carries a
// trace, which ties a proxy's version_info to the translation, or a UUID otherwise.
func PrepareSnapshot(ctx context.Context, resources types.XdsResources) (*cachev3.Snapshot, error) {
	version := string(uuid.NewUUID())
	if sc := trace.SpanContextFromContext(ctx); sc.IsValid() {
		version = sc.TraceID().String()
	}
	snapshot, err := cachev3.NewSnapshot(version, resources)
	if err == nil {
		err = snapshot.ConstructVersionMap()
	}
	if err != nil {
		xdsSnapshotCreateTotal.WithFailure(metrics.ReasonError).Increment()
		return nil, err
	}
	xdsSnapshotCreateTotal.WithSuccess().Increment()
	return snapshot, nil
}

// ReplaceEndpoints shares immutable non-EDS resources and copies EDS maps once.
// Only replacement assignments are serialized and hashed.
func ReplaceEndpoints(snapshot *cachev3.Snapshot, replacements map[string]*endpointv3.ClusterLoadAssignment) (*cachev3.Snapshot, error) {
	if len(replacements) == 0 {
		return snapshot, nil
	}
	var updated *cachev3.Snapshot
	for name, assignment := range replacements {
		if _, found := snapshot.Resources[cachetypes.Endpoint].Items[name]; !found {
			return nil, fmt.Errorf("endpoint assignment %q is not in the current snapshot", name)
		}
		data, err := cachev3.MarshalResource(assignment)
		if err != nil {
			return nil, err
		}
		hash := cachev3.HashResource(data)
		// Skip an unchanged assignment, such as one rebuilt because another port
		// of the same Service changed, so a no-op update isn't published.
		if snapshot.VersionMap[resourcev3.EndpointType][name] == hash {
			continue
		}
		if updated == nil {
			copy := *snapshot
			updated = &copy
			updated.Resources[cachetypes.Endpoint].Items = maps.Clone(snapshot.Resources[cachetypes.Endpoint].Items)
			updated.Resources[cachetypes.Endpoint].Version = string(uuid.NewUUID())
			updated.VersionMap = maps.Clone(snapshot.VersionMap)
			updated.VersionMap[resourcev3.EndpointType] = maps.Clone(snapshot.VersionMap[resourcev3.EndpointType])
		}
		updated.Resources[cachetypes.Endpoint].Items[name] = cachetypes.ResourceWithTTL{Resource: assignment}
		updated.VersionMap[resourcev3.EndpointType][name] = hash
	}
	if updated == nil {
		return snapshot, nil
	}
	return updated, nil
}

// NewSnapshotCache gives you a fresh SnapshotCache.
// It needs a logger that supports the go-control-plane
// required interface (Debugf, Infof, Warnf, and Errorf).
func NewSnapshotCache(ads bool, logger logging.Logger) SnapshotCacheWithCallbacks {
	// Set up the nasty wrapper hack.
	wrappedLogger := logger.Sugar()
	return &snapshotCache{
		SnapshotCache:       cachev3.NewSnapshotCache(ads, &Hash, wrappedLogger),
		log:                 wrappedLogger,
		lastSnapshot:        make(snapshotMap),
		streamIDNodeInfo:    make(nodeInfoMap),
		nodeFrequency:       make(nodeFrequencyMap),
		streamDuration:      make(streamDurationMap),
		deltaStreamDuration: make(streamDurationMap),
	}
}

// getNodeIDs retrieves the node ids from the node info map whose
// cluster field matches the ir key
func (s *snapshotCache) getNodeIDs(irKey string) []string {
	var nodeIDs []string
	for _, node := range s.streamIDNodeInfo {
		if node != nil && node.Cluster == irKey {
			nodeIDs = append(nodeIDs, node.Id)
		}
	}

	return nodeIDs
}

// OnStreamOpen and the other OnStream* functions implement the callbacks for the
// state-of-the-world stream types.
func (s *snapshotCache) OnStreamOpen(_ context.Context, streamID int64, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.streamIDNodeInfo[streamID] = nil
	s.streamDuration[streamID] = time.Now()

	return nil
}

func (s *snapshotCache) OnStreamClosed(streamID int64, node *corev3.Node) {
	// TODO: something with the node?
	s.mu.Lock()
	defer s.mu.Unlock()

	if startTime, ok := s.streamDuration[streamID]; ok {
		streamDuration := time.Since(startTime)
		xdsStreamDurationSeconds.With(
			streamIDLabel.Value(strconv.FormatInt(streamID, 10)),
			nodeIDLabel.Value(node.Id),
			isDeltaStreamLabel.Value("false"),
		).Record(streamDuration.Seconds())
	}

	delete(s.streamIDNodeInfo, streamID)
	delete(s.streamDuration, streamID)

	s.nodeFrequency[node.Id] -= 1
	if s.nodeFrequency[node.Id] <= 0 {
		delete(s.nodeFrequency, node.Id)

		// Only snapshots for nodes with active connections are updated, we need to clear
		// the snapshot for this node so it doesn't get stale data when it reconnects.
		s.ClearSnapshot(node.Id)
	}
}

func (s *snapshotCache) OnStreamRequest(streamID int64, req *discoveryv3.DiscoveryRequest) error {
	s.mu.Lock()
	// We could do this a little earlier than the defer, since the last half of this func is only logging
	// but that seemed like a premature optimization.
	defer s.mu.Unlock()

	// It's possible that only the first discovery request will have a node ID set.
	// We also need to save the node ID to the node list anyway.
	// So check if we have a nodeID for this stream already, then set it if not.
	if s.streamIDNodeInfo[streamID] == nil {
		if req.Node.Id == "" {
			return fmt.Errorf("couldn't get the node ID from the first discovery request on stream %d", streamID)
		}
		s.log.Debugf("First discovery request on stream %d, got nodeID %s", streamID, req.Node.Id)
		s.streamIDNodeInfo[streamID] = req.Node
		s.nodeFrequency[req.Node.Id] += 1
	}
	nodeID := s.streamIDNodeInfo[streamID].Id
	cluster := s.streamIDNodeInfo[streamID].Cluster

	var nodeVersion string

	var errorCode int32
	var errorMessage string

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot[cluster] == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot[cluster])
		if err != nil {
			return err
		}
	}

	if req.Node != nil {
		if bv := req.Node.GetUserAgentBuildVersion(); bv != nil && bv.Version != nil {
			nodeVersion = fmt.Sprintf("v%d.%d.%d", bv.Version.MajorNumber, bv.Version.MinorNumber, bv.Version.Patch)
		}
	}

	s.log.Debugf("Got a new request, version_info %s, response_nonce %s, nodeID %s, node_version %s", req.VersionInfo, req.ResponseNonce, nodeID, nodeVersion)

	if status := req.ErrorDetail; status != nil {
		// Envoy rejected (NACKed) the last update. It keeps serving its last known
		// good config, so this is silent from the user's point of view, but any
		// proxy that (re)starts will fail to load the rejected config. Surface it
		// as a metric so it can be alerted on, in addition to logging the details.
		errorCode = status.Code
		errorMessage = status.Message
		xdsNACKTotal.With(nodeIDLabel.Value(nodeID), typeURLLabel.Value(req.GetTypeUrl())).Increment()
	}

	s.log.Debugf("handling v3 xDS resource request, version_info %s, response_nonce %s, nodeID %s, node_version %s, resource_names %v, type_url %s, errorCode %d, errorMessage %s",
		req.VersionInfo, req.ResponseNonce,
		nodeID, nodeVersion, req.ResourceNames, req.GetTypeUrl(),
		errorCode, errorMessage)

	if errorCode != 0 {
		s.log.Errorf("Envoy rejected the last update for type %s on node %s with code %d and message %s; the proxy is still serving its last known good config and a restarting proxy will fail to load this config",
			req.GetTypeUrl(), nodeID, errorCode, errorMessage)
	}

	return nil
}

func (s *snapshotCache) OnStreamResponse(_ context.Context, streamID int64, _ *discoveryv3.DiscoveryRequest, _ *discoveryv3.DiscoveryResponse) {
	s.mu.Lock()
	node := s.streamIDNodeInfo[streamID]
	s.mu.Unlock()
	if node == nil {
		s.log.Errorf("Tried to send a response to a node we haven't seen yet on stream %d", streamID)
	} else {
		s.log.Debugf("Sending Response on stream %d to node %s", streamID, node.Id)
	}
}

// OnDeltaStreamOpen and the other OnDeltaStream*/OnStreamDelta* functions implement
// the callbacks for the incremental xDS versions.
// Yes, the different ordering in the name is part of the go-control-plane interface.
func (s *snapshotCache) OnDeltaStreamOpen(_ context.Context, streamID int64, _ string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure that we're adding the streamID to the Node ID list.
	s.streamIDNodeInfo[streamID] = nil
	s.deltaStreamDuration[streamID] = time.Now()

	return nil
}

func (s *snapshotCache) OnDeltaStreamClosed(streamID int64, node *corev3.Node) {
	// TODO: something with the node?
	s.mu.Lock()
	defer s.mu.Unlock()

	if startTime, ok := s.deltaStreamDuration[streamID]; ok {
		deltaStreamDuration := time.Since(startTime)
		xdsStreamDurationSeconds.With(
			streamIDLabel.Value(strconv.FormatInt(streamID, 10)),
			nodeIDLabel.Value(node.Id),
			isDeltaStreamLabel.Value("true"),
		).Record(deltaStreamDuration.Seconds())
	}

	delete(s.streamIDNodeInfo, streamID)
	delete(s.deltaStreamDuration, streamID)

	s.nodeFrequency[node.Id] -= 1
	if s.nodeFrequency[node.Id] <= 0 {
		delete(s.nodeFrequency, node.Id)

		// Only snapshots for nodes with active connections are updated, we need to clear
		// the snapshot for this node so it doesn't get stale data when it reconnects.
		s.ClearSnapshot(node.Id)
	}
}

func (s *snapshotCache) OnStreamDeltaRequest(streamID int64, req *discoveryv3.DeltaDiscoveryRequest) error {
	s.mu.Lock()
	// We could do this a little earlier than with a defer, since the last half of this func is logging
	// but that seemed like a premature optimization.
	defer s.mu.Unlock()

	var nodeVersion string
	var errorCode int32
	var errorMessage string

	// It's possible that only the first incremental discovery request will have a node ID set.
	// We also need to save the node ID to the node list anyway.
	// So check if we have a nodeID for this stream already, then set it if not.
	node := s.streamIDNodeInfo[streamID]
	if node == nil {
		if req.Node.Id == "" {
			return fmt.Errorf("couldn't get the node ID from the first incremental discovery request on stream %d", streamID)
		}
		s.log.Debugf("First incremental discovery request on stream %d, got nodeID %s", streamID, req.Node.Id)
		s.streamIDNodeInfo[streamID] = req.Node
		s.nodeFrequency[req.Node.Id] += 1
	}
	nodeID := s.streamIDNodeInfo[streamID].Id
	cluster := s.streamIDNodeInfo[streamID].Cluster

	// If no snapshot has been written into the snapshotCache yet, we can't do anything, so don't mess with
	// this request. go-control-plane will respond with an empty response, then send an update when a
	// snapshot is generated.
	if s.lastSnapshot[cluster] == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot[cluster])
		if err != nil {
			return err
		}
	}

	if req.Node != nil {
		if bv := req.Node.GetUserAgentBuildVersion(); bv != nil && bv.Version != nil {
			nodeVersion = fmt.Sprintf("v%d.%d.%d", bv.Version.MajorNumber, bv.Version.MinorNumber, bv.Version.Patch)
		}
	}

	s.log.Debugf("Got a new request, response_nonce %s, nodeID %s, node_version %s",
		req.ResponseNonce, nodeID, nodeVersion)
	if status := req.ErrorDetail; status != nil {
		// Envoy rejected (NACKed) the last update. It keeps serving its last known
		// good config, so this is silent from the user's point of view, but any
		// proxy that (re)starts will fail to load the rejected config. Surface it
		// as a metric so it can be alerted on, in addition to logging the details.
		errorCode = status.Code
		errorMessage = status.Message
		xdsNACKTotal.With(nodeIDLabel.Value(nodeID), typeURLLabel.Value(req.GetTypeUrl())).Increment()
	}
	s.log.Debugf("handling v3 xDS resource request, response_nonce %s, nodeID %s, node_version %s, resource_names_subscribe %v, resource_names_unsubscribe %v, type_url %s, errorCode %d, errorMessage %s",
		req.ResponseNonce,
		nodeID, nodeVersion,
		req.ResourceNamesSubscribe, req.ResourceNamesUnsubscribe,
		req.GetTypeUrl(),
		errorCode, errorMessage)

	if errorCode != 0 {
		s.log.Errorf("Envoy rejected the last update for type %s on node %s with code %d and message %s; the proxy is still serving its last known good config and a restarting proxy will fail to load this config",
			req.GetTypeUrl(), nodeID, errorCode, errorMessage)
	}

	return nil
}

func (s *snapshotCache) OnStreamDeltaResponse(streamID int64, _ *discoveryv3.DeltaDiscoveryRequest, _ *discoveryv3.DeltaDiscoveryResponse) {
	s.mu.Lock()
	node := s.streamIDNodeInfo[streamID]
	s.mu.Unlock()
	if node == nil {
		s.log.Errorf("Tried to send a response to a node we haven't seen yet on stream %d", streamID)
	} else {
		s.log.Debugf("Sending Incremental Response on stream %d to node %s", streamID, node.Id)
	}
}

func (s *snapshotCache) OnFetchRequest(_ context.Context, _ *discoveryv3.DiscoveryRequest) error {
	return nil
}

func (s *snapshotCache) OnFetchResponse(_ *discoveryv3.DiscoveryRequest, _ *discoveryv3.DiscoveryResponse) {
}

func (s *snapshotCache) SnapshotHasIrKey(irKey string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key, snapshot := range s.lastSnapshot {
		if snapshot != nil && key == irKey {
			return true
		}
	}

	return false
}

func (s *snapshotCache) GetIrKeys() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	irKeys := make([]string, 0, len(s.lastSnapshot))
	for key := range s.lastSnapshot {
		irKeys = append(irKeys, key)
	}

	return irKeys
}
