// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"errors"
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// XdsResources represents all the xds resources
type XdsResources = map[resourcev3.Type][]types.Resource

type EnvoyPatchPolicyStatuses []*ir.EnvoyPatchPolicyStatus

// GlobalResourceStatus tracks which shared global resources were emitted during translation.
type GlobalResourceStatus struct {
	// SystemTrustStore indicates the system_ca_certificates secret was emitted.
	SystemTrustStore bool
}

// ResourceVersionTable holds all the translated xds resources
type ResourceVersionTable struct {
	XdsResources
	EnvoyPatchPolicyStatuses
	GlobalResourceStatus

	clustersByName map[string]*clusterv3.Cluster
}

// EnableClusterIndex caches cluster names during resource construction. Add, replace,
// and SetResources maintain the index. Disable it before mutating clusters in place.
func (t *ResourceVersionTable) EnableClusterIndex() {
	t.clustersByName = make(map[string]*clusterv3.Cluster, len(t.XdsResources[resourcev3.ClusterType]))
	for _, resource := range t.XdsResources[resourcev3.ClusterType] {
		t.indexCluster(resource)
	}
}

// DisableClusterIndex releases the construction index before patches or extension hooks.
func (t *ResourceVersionTable) DisableClusterIndex() {
	t.clustersByName = nil
}

// FindCluster returns the first cluster with name, or nil if it does not exist.
func (t *ResourceVersionTable) FindCluster(name string) *clusterv3.Cluster {
	if t.clustersByName != nil {
		return t.clustersByName[name]
	}
	for _, resource := range t.XdsResources[resourcev3.ClusterType] {
		if cluster, ok := resource.(*clusterv3.Cluster); ok && cluster.Name == name {
			return cluster
		}
	}
	return nil
}

func (t *ResourceVersionTable) indexCluster(resource types.Resource) {
	if t.clustersByName == nil {
		return
	}
	if cluster, ok := resource.(*clusterv3.Cluster); ok {
		if _, exists := t.clustersByName[cluster.Name]; !exists {
			t.clustersByName[cluster.Name] = cluster
		}
	}
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *ResourceVersionTable) GetXdsResources() XdsResources {
	return t.XdsResources
}

func (t *ResourceVersionTable) AddXdsResource(rType resourcev3.Type, xdsResource types.Resource) error {
	// It's a sanity check to make sure the xdsResource is not nil
	if xdsResource == nil {
		return fmt.Errorf("xds resource is nil")
	}

	if err := proto.Validate(xdsResource); err != nil {
		return fmt.Errorf("validation failed for xds resource %+v, err: %w", xdsResource, err)
	}

	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}
	if t.XdsResources[rType] == nil {
		t.XdsResources[rType] = make([]types.Resource, 0, 1)
	}

	t.XdsResources[rType] = append(t.XdsResources[rType], xdsResource)
	if rType == resourcev3.ClusterType {
		t.indexCluster(xdsResource)
	}
	return nil
}

// ValidateAll validates all the xds resources in the ResourceVersionTable
func (t *ResourceVersionTable) ValidateAll() error {
	var errs error

	for _, xdsResource := range t.XdsResources {
		for _, resource := range xdsResource {
			if err := proto.Validate(resource); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

// AddOrReplaceXdsResource will update an existing resource of rType according to matchFunc or add as a new resource
// if none satisfy the match criteria. It will only update the first match it finds, regardless
// if multiple resources satisfy the match criteria.
func (t *ResourceVersionTable) AddOrReplaceXdsResource(rType resourcev3.Type, resource types.Resource, matchFunc func(existing, new types.Resource) bool) error {
	if t.XdsResources == nil || t.XdsResources[rType] == nil {
		if err := t.AddXdsResource(rType, resource); err != nil {
			return err
		} else {
			return nil
		}
	}

	var found bool
	for i, r := range t.XdsResources[rType] {
		if matchFunc(r, resource) {
			t.XdsResources[rType][i] = resource
			if rType == resourcev3.ClusterType && t.clustersByName != nil {
				t.EnableClusterIndex()
			}
			found = true
			break
		}
	}
	if !found {
		if err := t.AddXdsResource(rType, resource); err != nil {
			return err
		} else {
			return nil
		}
	}
	return nil
}

// SetResources will update an entire entry of the XdsResources for a certain type to the provided resources
func (t *ResourceVersionTable) SetResources(rType resourcev3.Type, xdsResources []types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}

	t.XdsResources[rType] = xdsResources
	if rType == resourcev3.ClusterType && t.clustersByName != nil {
		t.EnableClusterIndex()
	}
}
