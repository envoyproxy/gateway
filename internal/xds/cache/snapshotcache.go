// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/xds/v3/snapshotter.go

package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"sync"

	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

var Hash = envoy_cache_v3.IDHash{}

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
	envoy_cache_v3.SnapshotCache
	envoy_server_v3.Callbacks
	GenerateNewSnapshot(string, types.XdsResources) error
}

type snapshotMap map[string]*envoy_cache_v3.Snapshot

type snapshotcache struct {
	envoy_cache_v3.SnapshotCache
	streamIDNodeID  map[int64]string
	snapshotVersion int64
	lastSnapshot    snapshotMap
	log             *LogrWrapper
	mu              sync.Mutex
}

// GenerateNewSnapshot takes a table of resources (the output from the IR->xDS
// translator) and updates the snapshot version.
func (s *snapshotcache) GenerateNewSnapshot(irKey string, resources types.XdsResources) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	version := s.newSnapshotVersion()

	// Create a snapshot with all xDS resources.
	snapshot, err := envoy_cache_v3.NewSnapshot(
		version,
		resources,
	)
	if err != nil {
		return err
	}

	s.lastSnapshot[irKey] = snapshot

	for _, node := range s.getNodeIDs(irKey) {
		s.log.Debugf("Generating a snapshot with Node %s", node)
		err := s.SetSnapshot(context.TODO(), node, snapshot)
		if err != nil {
			return err
		}
	}

	return nil

}

// newSnapshotVersion increments the current snapshotVersion
// and returns as a string.
func (s *snapshotcache) newSnapshotVersion() string {

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
func NewSnapshotCache(ads bool, logger logr.Logger) SnapshotCacheWithCallbacks {
	// Set up the nasty wrapper hack.
	wrappedLogger := NewLogrWrapper(logger)
	return &snapshotcache{
		SnapshotCache:  envoy_cache_v3.NewSnapshotCache(ads, &Hash, wrappedLogger),
		log:            wrappedLogger,
		lastSnapshot:   make(snapshotMap),
		streamIDNodeID: make(map[int64]string),
	}
}

func (s *snapshotcache) getNodeIDs(irKey string) []string {
	var nodeIDs []string
	for _, nodeID := range s.streamIDNodeID {
		if nodeID == irKey {
			nodeIDs = append(nodeIDs, nodeID)
		}
	}

	return nodeIDs

}

// OnStreamOpen and the other OnStream* functions implement the callbacks for the
// state-of-the-world stream types.
func (s *snapshotcache) OnStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnStreamClosed(streamID int64) {

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamRequest(streamID int64, req *envoy_service_discovery_v3.DiscoveryRequest) error {

	s.mu.Lock()
	// We could do this a little earlier than the defer, since the last half of this func is only logging
	// but that seemed like a premature optimization.
	defer s.mu.Unlock()

	// It's possible that only the first discovery request will have a node ID set.
	// We also need to save the node ID to the node list anyway.
	// So check if we have a nodeID for this stream already, then set it if not.
	nodeID := s.streamIDNodeID[streamID]
	if nodeID == "" {
		nodeID = Hash.ID(req.Node)
		if nodeID == "" {
			return fmt.Errorf("couldn't hash the node ID from the first discovery request on stream %d", streamID)
		}
		s.log.Debugf("First discovery request on stream %d, got nodeID %s", streamID, nodeID)
		s.streamIDNodeID[streamID] = nodeID
	}

	var nodeVersion string

	var errorCode int32
	var errorMessage string

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot[nodeID] == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot[nodeID])
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
		// if Envoy rejected the last update log the details here.
		// TODO(youngnick): Handle NACK properly
		errorCode = status.Code
		errorMessage = status.Message
	}

	s.log.Debugf("handling v3 xDS resource request, version_info %s, response_nonce %s, nodeID %s, node_version %s, resource_names %v, type_url %s, errorCode %d, errorMessage %s",
		req.VersionInfo, req.ResponseNonce,
		nodeID, nodeVersion, req.ResourceNames, req.GetTypeUrl(),
		errorCode, errorMessage)

	return nil
}

func (s *snapshotcache) OnStreamResponse(ctx context.Context, streamID int64, req *envoy_service_discovery_v3.DiscoveryRequest, resp *envoy_service_discovery_v3.DiscoveryResponse) {

	// No mutex lock required here because no writing to the cache.

	nodeID := s.streamIDNodeID[streamID]
	if nodeID == "" {
		s.log.Errorf("Tried to send a response to a node we haven't seen yet on stream %d", streamID)
	}

	s.log.Debugf("Sending Response on stream %d to node %s", streamID, nodeID)
}

// OnDeltaStreamOpen and the other OnDeltaStream*/OnStreamDelta* functions implement
// the callbacks for the incremental xDS versions.
// Yes, the different ordering in the name is part of the go-control-plane interface.
func (s *snapshotcache) OnDeltaStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.mu.Lock()
	defer s.mu.Unlock()

	// Ensure that we're adding the streamID to the Node ID list.
	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnDeltaStreamClosed(streamID int64) {

	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamDeltaRequest(streamID int64, req *envoy_service_discovery_v3.DeltaDiscoveryRequest) error {

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
	nodeID := s.streamIDNodeID[streamID]
	if nodeID == "" {
		nodeID = Hash.ID(req.Node)
		if nodeID == "" {
			return fmt.Errorf("couldn't hash the node ID from the first incremental discovery request on stream %d", streamID)
		}
		s.log.Debugf("First incremental discovery request on stream %d, got nodeID %s", streamID, nodeID)
		s.streamIDNodeID[streamID] = nodeID
	}

	// If no snapshot has been written into the snapshotcache yet, we can't do anything, so don't mess with
	// this request. go-control-plane will respond with an empty response, then send an update when a
	// snapshot is generated.
	if s.lastSnapshot[nodeID] == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot[nodeID])
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
		// if Envoy rejected the last update log the details here.
		// TODO(youngnick): Handle NACK properly
		errorCode = status.Code
		errorMessage = status.Message
	}
	s.log.Debugf("handling v3 xDS resource request, response_nonce %s, nodeID %s, node_version %s, resource_names_subscribe %v, resource_names_unsubscribe %v, type_url %s, errorCode %d, errorMessage %s",
		req.ResponseNonce,
		nodeID, nodeVersion,
		req.ResourceNamesSubscribe, req.ResourceNamesUnsubscribe,
		req.GetTypeUrl(),
		errorCode, errorMessage)

	return nil
}

func (s *snapshotcache) OnStreamDeltaResponse(streamID int64, req *envoy_service_discovery_v3.DeltaDiscoveryRequest, resp *envoy_service_discovery_v3.DeltaDiscoveryResponse) {
	// No mutex lock required here because no writing to the cache.
	nodeID := s.streamIDNodeID[streamID]
	if nodeID == "" {
		s.log.Errorf("Tried to send a response to a node we haven't seen yet on stream %d", streamID)
	}
	s.log.Debugf("Sending Incremental Response on stream %d to node %s", streamID, nodeID)
}

func (s *snapshotcache) OnFetchRequest(ctx context.Context, req *envoy_service_discovery_v3.DiscoveryRequest) error {
	return nil
}

func (s *snapshotcache) OnFetchResponse(req *envoy_service_discovery_v3.DiscoveryRequest, resp *envoy_service_discovery_v3.DiscoveryResponse) {
}
