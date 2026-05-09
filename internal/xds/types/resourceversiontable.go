// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"errors"
	"fmt"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// XdsResources represents all the xds resources
type XdsResources = map[resourcev3.Type][]types.Resource

type EnvoyPatchPolicyStatuses []*ir.EnvoyPatchPolicyStatus

// ResourceVersionTable holds all the translated xds resources
type ResourceVersionTable struct {
	XdsResources
	EnvoyPatchPolicyStatuses

	// nameIndex maps (resource type, xDS resource name) to the position of the
	// resource in XdsResources[type]. It is a sidecar to XdsResources that lets
	// findXds* helpers do an O(1) lookup instead of an O(N) linear scan on
	// every add. Maintained by AddXdsResource, AddOrReplaceXdsResource, and
	// SetResources; not exposed externally.
	//
	// Storing the slice position (rather than the Resource itself) means
	// in-place updates to XdsResources[type][i] are observed automatically by
	// the index without re-indexing.
	nameIndex map[resourcev3.Type]map[string]int
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *ResourceVersionTable) GetXdsResources() XdsResources {
	return t.XdsResources
}

// FindXdsResource returns the xDS resource of the given type with the given
// name, and a boolean indicating whether such a resource exists. Lookup is O(1).
//
// The index stores the slice position rather than the Resource itself, so
// in-place updates to t.XdsResources[type][i] are reflected without re-indexing.
func (t *ResourceVersionTable) FindXdsResource(rType resourcev3.Type, name string) (types.Resource, bool) {
	if t == nil {
		return nil, false
	}
	idx, ok := t.nameIndex[rType][name]
	if !ok {
		return nil, false
	}
	slice := t.XdsResources[rType]
	if idx < 0 || idx >= len(slice) {
		return nil, false
	}
	return slice[idx], true
}

// xdsResourceName returns the xDS-level name of a resource based on its type.
// Different proto types expose the name under different field names (e.g.
// ClusterLoadAssignment uses ClusterName); this normalizes them.
func xdsResourceName(r types.Resource) string {
	switch v := r.(type) {
	case *clusterv3.Cluster:
		return v.GetName()
	case *endpointv3.ClusterLoadAssignment:
		return v.GetClusterName()
	case *listenerv3.Listener:
		return v.GetName()
	case *routev3.RouteConfiguration:
		return v.GetName()
	case *tlsv3.Secret:
		return v.GetName()
	default:
		return ""
	}
}

// indexResource records the position of r within t.XdsResources[rType] in the
// name index. Caller must pass the position the resource occupies in the slice.
func (t *ResourceVersionTable) indexResource(rType resourcev3.Type, r types.Resource, pos int) {
	name := xdsResourceName(r)
	if name == "" {
		return
	}
	if t.nameIndex == nil {
		t.nameIndex = make(map[resourcev3.Type]map[string]int)
	}
	if t.nameIndex[rType] == nil {
		t.nameIndex[rType] = make(map[string]int)
	}
	t.nameIndex[rType][name] = pos
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
	t.indexResource(rType, xdsResource, len(t.XdsResources[rType])-1)
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
			// Re-index in case matchFunc paired resources whose names differ;
			// the slice slot stays the same so the position we record is i.
			t.indexResource(rType, resource, i)
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

	// Rebuild the name index for this type to match the replaced slice.
	if t.nameIndex != nil {
		delete(t.nameIndex, rType)
	}
	for i, r := range xdsResources {
		t.indexResource(rType, r, i)
	}
}
