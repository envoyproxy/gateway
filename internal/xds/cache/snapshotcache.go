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
	"fmt"
	"math"
	"strconv"
	"sync"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoytypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

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
	GenerateNewSnapshot(string, types.XdsResources, context.Context) (bool, error)
	UpdateEndpointResources(string, []envoytypes.Resource, context.Context) (bool, error)
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
	snapshotVersion     int64
	lastSnapshot        snapshotMap
	log                 *zap.SugaredLogger
	mu                  sync.Mutex
	// eagerVersionMap makes GenerateNewSnapshot construct the delta version map
	// while the snapshot is still private, so UpdateEndpointResources can read
	// it without racing go-control-plane's lazy construction. Enabled only with
	// the endpoint fast path, since it costs a marshal+hash of every resource
	// per snapshot.
	eagerVersionMap bool
	// noVersionMap marks IR keys whose committed snapshot has no usable delta
	// version map, so UpdateEndpointResources stands down instead of reading a
	// field go-control-plane may be filling in lazily under its own mutex.
	noVersionMap map[string]bool
}

// GenerateNewSnapshot takes a table of resources (the output from the IR->xDS
// translator) and updates the snapshot version.
//
// The returned bool reports whether the snapshot was committed. It is false only
// when building it failed, leaving the previous snapshot untouched; it is true
// even when publishing to a node then failed, so callers can keep state in sync
// with the committed snapshot while still surfacing the error.
func (s *snapshotCache) GenerateNewSnapshot(irKey string, resources types.XdsResources, ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, span := tracer.Start(ctx, "SnapshotCache.GenerateNewSnapshot")
	defer span.End()

	sc := trace.SpanContextFromContext(ctx)
	version := sc.TraceID().String()
	if !sc.IsValid() {
		version = s.newSnapshotVersion()
	}

	// Create a snapshot with all xDS resources.
	snapshot, err := cachev3.NewSnapshot(
		version,
		resources,
	)
	if err != nil {
		xdsSnapshotCreateTotal.WithFailure(metrics.ReasonError).Increment()
		return false, err
	}
	// Build the delta version map while the snapshot is still private:
	// go-control-plane would otherwise build it lazily under its own mutex,
	// racing UpdateEndpointResources reading it under s.mu. A failure must not
	// block the publish — without the fast path this map only affects delta
	// streams — so the snapshot goes out with a nil map, rebuilt lazily.
	versionMapFailed := false
	if s.eagerVersionMap {
		if err := snapshot.ConstructVersionMap(); err != nil {
			// ConstructVersionMap allocates the map before it can fail, and
			// later lazy construction short-circuits on a non-nil map, so a
			// partially built map would never be repaired. Clear it and let
			// go-control-plane rebuild it on demand.
			snapshot.VersionMap = nil
			versionMapFailed = true
			s.log.Errorf("failed to construct the delta version map for %s: %v", irKey, err)
		}
	}
	xdsSnapshotCreateTotal.WithSuccess().Increment()

	// Commit point.
	// Delete snapshot from cache if resources are nil
	if resources == nil {
		delete(s.lastSnapshot, irKey)
		delete(s.noVersionMap, irKey)
	} else {
		// Update snapshot in cache
		s.lastSnapshot[irKey] = snapshot
		if versionMapFailed {
			s.noVersionMap[irKey] = true
		} else {
			delete(s.noVersionMap, irKey)
		}
	}

	// The snapshot is committed above, so a per-node publish failure is
	// reported with committed=true: the error still propagates to the caller
	// (and its error channel), while a node whose publish failed is served the
	// committed snapshot when it reconnects.
	for _, node := range s.getNodeIDs(irKey) {
		s.log.Debugf("Generating a snapshot with Node %s", node)

		if err = s.SetSnapshot(context.TODO(), node, snapshot); err != nil {
			xdsSnapshotUpdateTotal.WithFailure(metrics.ReasonError, nodeIDLabel.Value(node)).Increment()
			return true, err
		}
		xdsSnapshotUpdateTotal.WithSuccess(nodeIDLabel.Value(node)).Increment()
	}

	return true, nil
}

// UpdateEndpointResources patches only the EDS resources of irKey's snapshot: the
// given ClusterLoadAssignments replace the same-named ones, the EDS version is
// bumped, and every other type keeps its version, so non-EDS resources are not
// resent. The previous snapshot is never mutated — its pointer is shared with
// live streams. Returns false when no snapshot exists for irKey yet.
func (s *snapshotCache) UpdateEndpointResources(irKey string, assignments []envoytypes.Resource, ctx context.Context) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldSnapshot, ok := s.lastSnapshot[irKey]
	if !ok || oldSnapshot == nil {
		return false, nil
	}
	// The snapshot went out without a usable delta version map, so
	// go-control-plane may be building one on it concurrently. Leave it to the
	// next full build rather than read or rebuild that map here.
	if s.noVersionMap[irKey] {
		return false, nil
	}

	_, span := tracer.Start(ctx, "SnapshotCache.UpdateEndpointResources")
	defer span.End()

	// Merge the updated CLAs over the existing EDS resources.
	edsIndex := cachev3.GetResponseType(resourcev3.EndpointType)
	oldEDS := oldSnapshot.Resources[edsIndex]
	merged := make([]envoytypes.Resource, 0, len(oldEDS.Items)+len(assignments))
	replaced := make(map[string]struct{}, len(assignments))
	for _, cla := range assignments {
		merged = append(merged, cla)
		replaced[cachev3.GetResourceName(cla)] = struct{}{}
	}
	for name, res := range oldEDS.Items {
		if _, ok := replaced[name]; !ok {
			merged = append(merged, res.Resource)
		}
	}

	// Clone the snapshot: the Resources array is copied by value, and only the
	// EDS entry is replaced under a fresh version. The per-type Resources maps
	// of the other types are shared with the old snapshot but never mutated.
	newSnapshot := &cachev3.Snapshot{Resources: oldSnapshot.Resources}
	newSnapshot.Resources[edsIndex] = cachev3.NewResources(s.newSnapshotVersion(), merged)

	// Carry the delta version map forward for unchanged resource types and for
	// unchanged EDS entries, and re-hash only the incoming CLAs, so a patch
	// costs O(changed) rather than O(all resources). The old map was built
	// eagerly while the snapshot was private, so reading it here is safe.
	if oldSnapshot.VersionMap != nil {
		versionMap := make(map[string]map[string]string, len(oldSnapshot.VersionMap))
		for typeURL, inner := range oldSnapshot.VersionMap {
			if typeURL != resourcev3.EndpointType {
				versionMap[typeURL] = inner
			}
		}
		oldEDSVersions := oldSnapshot.VersionMap[resourcev3.EndpointType]
		edsVersions := make(map[string]string, len(merged))
		for _, res := range merged {
			name := cachev3.GetResourceName(res)
			if _, ok := replaced[name]; !ok {
				if v, ok := oldEDSVersions[name]; ok {
					edsVersions[name] = v
					continue
				}
			}
			marshaled, err := cachev3.MarshalResource(res)
			if err != nil {
				return false, err
			}
			edsVersions[name] = cachev3.HashResource(marshaled)
		}
		versionMap[resourcev3.EndpointType] = edsVersions
		newSnapshot.VersionMap = versionMap
	}

	// Commit point: update the stored snapshot so reconnecting nodes see the
	// patched state too. Everything fallible above happens before this, so
	// committed=false always means nothing changed; a per-node publish failure
	// below is reported with committed=true, and that node is served the
	// committed snapshot when it reconnects.
	s.lastSnapshot[irKey] = newSnapshot

	for _, node := range s.getNodeIDs(irKey) {
		s.log.Debugf("Updating endpoint resources in snapshot with Node %s", node)

		if err := s.SetSnapshot(context.TODO(), node, newSnapshot); err != nil {
			xdsSnapshotUpdateTotal.WithFailure(metrics.ReasonError, nodeIDLabel.Value(node)).Increment()
			return true, err
		}
		xdsSnapshotUpdateTotal.WithSuccess(nodeIDLabel.Value(node)).Increment()
	}

	return true, nil
}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (s *snapshotCache) newSnapshotVersion() string {
	// Reset the snapshotVersion if it ever hits max size.
	if s.snapshotVersion == math.MaxInt64 {
		s.snapshotVersion = 0
	}

	// Increment the snapshot version & return as string.
	s.snapshotVersion++
	return strconv.FormatInt(s.snapshotVersion, 10)
}

// NewSnapshotCache gives you a fresh SnapshotCache.
// It needs a logger that supports the go-control-plane
// required interface (Debugf, Infof, Warnf, and Errorf).
// eagerVersionMap must be set when the endpoint fast path is enabled, so
// UpdateEndpointResources can read the previous snapshot's delta version map
// without racing go-control-plane's lazy construction.
func NewSnapshotCache(ads, eagerVersionMap bool, logger logging.Logger) SnapshotCacheWithCallbacks {
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
		eagerVersionMap:     eagerVersionMap,
		noVersionMap:        make(map[string]bool),
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
