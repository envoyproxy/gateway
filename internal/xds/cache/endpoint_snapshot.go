// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package cache

import (
	"fmt"
	"maps"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	cachetypes "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/envoyproxy/gateway/internal/metrics"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

// PrepareSnapshot builds maps and delta hashes before a full result is queued.
func PrepareSnapshot(resources types.XdsResources) (*cachev3.Snapshot, error) {
	snapshot, err := cachev3.NewSnapshot(string(uuid.NewUUID()), resources)
	if err != nil {
		xdsSnapshotCreateTotal.WithFailure(metrics.ReasonError).Increment()
		return nil, err
	}
	if err := snapshot.ConstructVersionMap(); err != nil {
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
	var next *cachev3.Snapshot
	for name, assignment := range replacements {
		if _, found := snapshot.Resources[cachetypes.Endpoint].Items[name]; !found {
			return nil, fmt.Errorf("endpoint assignment %q is not in the current snapshot", name)
		}
		data, err := cachev3.MarshalResource(assignment)
		if err != nil {
			return nil, err
		}
		hash := cachev3.HashResource(data)
		if snapshot.VersionMap[resourcev3.EndpointType][name] == hash {
			continue
		}
		if next == nil {
			copy := *snapshot
			next = &copy
			next.Resources[cachetypes.Endpoint].Items = maps.Clone(snapshot.Resources[cachetypes.Endpoint].Items)
			next.Resources[cachetypes.Endpoint].Version = string(uuid.NewUUID())
			next.VersionMap = maps.Clone(snapshot.VersionMap)
			next.VersionMap[resourcev3.EndpointType] = maps.Clone(snapshot.VersionMap[resourcev3.EndpointType])
		}
		next.Resources[cachetypes.Endpoint].Items[name] = cachetypes.ResourceWithTTL{Resource: assignment}
		next.VersionMap[resourcev3.EndpointType][name] = hash
	}
	if next == nil {
		return snapshot, nil
	}
	return next, nil
}
