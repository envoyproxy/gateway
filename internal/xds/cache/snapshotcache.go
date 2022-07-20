// Copyright Project Contour Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cache

import (
	"context"
	"fmt"
	"math"
	"strconv"

	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_cache_v3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	envoy_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/sirupsen/logrus"

	"github.com/envoyproxy/gateway/internal/xds/types"
)

var Hash = envoy_cache_v3.IDHash{}

type SnapshotCacheWithCallbacks interface {
	envoy_cache_v3.SnapshotCache
	envoy_server_v3.Callbacks
	GenerateNewSnapshot(types.XdsResources) error
}

type snapshotcache struct {
	envoy_cache_v3.SnapshotCache

	lastSnapshot *envoy_cache_v3.Snapshot

	log logrus.FieldLogger

	streamIDNodeID map[int64]string

	snapshotVersion int64
}

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
		fmt.Printf("Generating a snapshot with Node %s\n", node)
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

func NewSnapshotCache(ads bool, logger logrus.FieldLogger) SnapshotCacheWithCallbacks {
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

func (s *snapshotcache) OnStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnStreamClosed(streamID int64) {

	// nodes := s.streamIDNodeID[streamID]

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamRequest(streamID int64, req *envoy_service_discovery_v3.DiscoveryRequest) error {

	log := s.log.WithField("version_info", req.VersionInfo).WithField("response_nonce", req.ResponseNonce)

	nodeID := Hash.ID(req.Node)

	s.streamIDNodeID[streamID] = nodeID

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot)
	}

	if req.Node != nil {
		log = log.WithField("node_id", nodeID)

		if bv := req.Node.GetUserAgentBuildVersion(); bv != nil && bv.Version != nil {
			log = log.WithField("node_version", fmt.Sprintf("v%d.%d.%d", bv.Version.MajorNumber, bv.Version.MinorNumber, bv.Version.Patch))
		}
	}

	log.Debug("Got a new request")

	if status := req.ErrorDetail; status != nil {
		// if Envoy rejected the last update log the details here.
		// TODO(youngnick): Handle NACK properly
		log.WithField("code", status.Code).Error(status.Message)
	}

	log = log.WithField("resource_names", req.ResourceNames).WithField("type_url", req.GetTypeUrl())

	log.Debug("handling v3 xDS resource request")

	return nil
}

func (s *snapshotcache) OnStreamResponse(ctx context.Context, streamID int64, req *envoy_service_discovery_v3.DiscoveryRequest, resp *envoy_service_discovery_v3.DiscoveryResponse) {
	nodeID := Hash.ID(req.Node)

	s.log.Debugf("Sending Response on stream %d to node %s", streamID, nodeID)
}

// Will need to add the other ones too, I think.

func (s *snapshotcache) OnDeltaStreamOpen(ctx context.Context, streamID int64, typeURL string) error {

	s.streamIDNodeID[streamID] = ""

	return nil
}

func (s *snapshotcache) OnDeltaStreamClosed(streamID int64) {

	delete(s.streamIDNodeID, streamID)

}

func (s *snapshotcache) OnStreamDeltaRequest(streamID int64, req *envoy_service_discovery_v3.DeltaDiscoveryRequest) error {

	log := s.log.WithField("response_nonce", req.ResponseNonce)

	nodeID := Hash.ID(req.Node)

	s.streamIDNodeID[streamID] = nodeID

	// If no snapshot has been generated yet, we can't do anything, so don't mess with this request.
	// go-control-plane will respond with an empty response, then send an update when a snapshot is generated.
	if s.lastSnapshot == nil {
		return nil
	}

	_, err := s.GetSnapshot(nodeID)
	if err != nil {
		s.SetSnapshot(context.TODO(), nodeID, s.lastSnapshot)
	}

	if req.Node != nil {
		log = log.WithField("node_id", nodeID)

		if bv := req.Node.GetUserAgentBuildVersion(); bv != nil && bv.Version != nil {
			log = log.WithField("node_version", fmt.Sprintf("v%d.%d.%d", bv.Version.MajorNumber, bv.Version.MinorNumber, bv.Version.Patch))
		}
	}

	log.Debug("Got a new incremental request")

	if status := req.ErrorDetail; status != nil {
		// if Envoy rejected the last update log the details here.
		// TODO(youngnick): Handle NACK properly
		log.WithField("code", status.Code).Error(status.Message)
	}

	log = log.WithField("resources_to_subscribe", req.ResourceNamesSubscribe).
		WithField("resources_to_unsubscribe", req.ResourceNamesUnsubscribe).
		WithField("type_url", req.GetTypeUrl())

	log.Debug("handling v3 xDS resource request")

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
