// Portions of this code are based on code from Contour, available at:
// https://github.com/projectcontour/contour/blob/main/internal/xds/v3/snapshotter.go

package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"

	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"

	"github.com/envoyproxy/gateway/internal/log"
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
	GenerateNewSnapshot(types.XdsResources) error
}

type snapshotcache struct {
	envoy_cache_v3.SnapshotCache

	lastSnapshot *envoy_cache_v3.Snapshot

	log *log.LogrWrapper

	streamIDNodeID map[int64]string

	snapshotVersion int64
}

// GenerateNewSnapshot takes a table of resources (the output from the IR->xDS
// translator) and updates the snapshot version.
func (s *snapshotcache) GenerateNewSnapshot(resources types.XdsResources) error {

	version := s.newSnapshotVersion()

	// Create a snapshot with all xDS resources.
	snapshot, err := envoy_cache_v3.NewSnapshot(
		version,
		resources,
	)
	if err != nil {
		return err
	}

	s.lastSnapshot = snapshot

	for _, node := range s.getNodeIDs() {
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
func NewSnapshotCache(ads bool, logger *log.LogrWrapper) SnapshotCacheWithCallbacks {
	return &snapshotcache{
		SnapshotCache:  envoy_cache_v3.NewSnapshotCache(ads, &Hash, logger),
		log:            logger,
		streamIDNodeID: make(map[int64]string),
	}
}

func (s *snapshotcache) getNodeIDs() []string {

	var nodeIDs []string

	for _, node := range s.streamIDNodeID {

		nodeIDs = append(nodeIDs, node)

	}

	return nodeIDs

}

// OnStreamOpen and the other OnStream* functions implement the callbacks for the
// state-of-the-world stream types.
func (s *snapshotcache) OnStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnStreamClosed(streamID int64) {

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamRequest(streamID int64, req *envoy_service_discovery_v3.DiscoveryRequest) error {

	nodeID := Hash.ID(req.Node)

	s.streamIDNodeID[streamID] = nodeID

	var nodeVersion string

	var errorCode int32
	var errorMessage string

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot)
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
	nodeID := Hash.ID(req.Node)

	s.log.Debugf("Sending Response on stream %d to node %s", streamID, nodeID)
}

// OnDeltaStreamOpen and the other OnDeltaStream*/OnStreamDelta* functions implement
// the callbacks for the incremental xDS versions.
// Yes, the different ordering in the name is part of the go-control-plane interface.
func (s *snapshotcache) OnDeltaStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnDeltaStreamClosed(streamID int64) {

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamDeltaRequest(streamID int64, req *envoy_service_discovery_v3.DeltaDiscoveryRequest) error {

	var nodeVersion string

	var errorCode int32
	var errorMessage string

	nodeID := Hash.ID(req.Node)

	s.streamIDNodeID[streamID] = nodeID

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		err = s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot)
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

	nodeID := Hash.ID(req.Node)
	s.log.Debugf("Sending Incremental Response on stream %d to node %s", streamID, nodeID)

}

func (s *snapshotcache) OnFetchRequest(ctx context.Context, req *envoy_service_discovery_v3.DiscoveryRequest) error {

	return nil
}

func (s *snapshotcache) OnFetchResponse(req *envoy_service_discovery_v3.DiscoveryRequest, resp *envoy_service_discovery_v3.DiscoveryResponse) {

}
