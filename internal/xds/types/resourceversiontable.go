// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package types

import (
	"errors"
	"fmt"
	"sort"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/cache/types"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	rlsconfv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

// XdsResources represents all the xds resources, indexed by type then by resource name.
// Indexing by name keeps lookups O(1) and enforces the xDS-level invariant that names
// are unique within a type.
type XdsResources = map[resourcev3.Type]map[string]types.Resource

type EnvoyPatchPolicyStatuses []*ir.EnvoyPatchPolicyStatus

// ResourceVersionTable holds all the translated xds resources
type ResourceVersionTable struct {
	XdsResources
	EnvoyPatchPolicyStatuses
}

// GetXdsResources retrieves the translated xds resources saved in the translator context.
func (t *ResourceVersionTable) GetXdsResources() XdsResources {
	return t.XdsResources
}

// xdsResourceName returns the unique name for an xDS resource based on its type.
// Returns the empty string if the resource type is unknown — callers should treat
// that as an error.
func xdsResourceName(rType resourcev3.Type, r types.Resource) string {
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
	case *rlsconfv3.RateLimitConfig:
		return v.GetName()
	default:
		// Fall back to the type so we still have a deterministic key for resource
		// types we don't explicitly know about.
		return rType
	}
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
		t.XdsResources[rType] = make(map[string]types.Resource)
	}

	name := xdsResourceName(rType, xdsResource)
	t.XdsResources[rType][name] = xdsResource
	return nil
}

// ValidateAll validates all the xds resources in the ResourceVersionTable
func (t *ResourceVersionTable) ValidateAll() error {
	var errs error

	for _, byName := range t.XdsResources {
		for _, resource := range byName {
			if err := proto.Validate(resource); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	return errs
}

// AddOrReplaceXdsResource adds or replaces a resource in the table. Now that resources
// are stored by name, the matchFunc parameter is unused — it is retained for API
// compatibility with existing callers.
func (t *ResourceVersionTable) AddOrReplaceXdsResource(rType resourcev3.Type, resource types.Resource, _ func(existing, new types.Resource) bool) error {
	return t.AddXdsResource(rType, resource)
}

// SetResources replaces all resources of a given type with the provided slice.
// The slice is converted to the internal name-indexed map.
func (t *ResourceVersionTable) SetResources(rType resourcev3.Type, xdsResources []types.Resource) {
	if t.XdsResources == nil {
		t.XdsResources = make(XdsResources)
	}

	byName := make(map[string]types.Resource, len(xdsResources))
	for _, r := range xdsResources {
		if r == nil {
			continue
		}
		byName[xdsResourceName(rType, r)] = r
	}
	t.XdsResources[rType] = byName
}

// FlattenToTypeWiseSlices converts the name-indexed XdsResources to the slice-shaped
// representation that go-control-plane's snapshot cache and other external
// boundaries expect. Resources within each type are returned in name-sorted
// order so the output is stable.
func FlattenToTypeWiseSlices(resources XdsResources) map[resourcev3.Type][]types.Resource {
	if resources == nil {
		return nil
	}
	out := make(map[resourcev3.Type][]types.Resource, len(resources))
	for rType, byName := range resources {
		out[rType] = FlattenToSlice(byName)
	}
	return out
}

// FlattenToSlice converts a name-indexed resource map to a name-sorted slice.
//
// Sorting is O(N log N) and only required for callers that need stable output
// (golden tests, egctl dumps, debug logs, the extension hook contract). It is
// not required for correctness of the xDS protocol — go-control-plane keys
// resources by name internally. If a profile shows the sort is hot on a
// per-reconcile path, introduce an unsorted variant for that path rather than
// changing this function.
func FlattenToSlice(byName map[string]types.Resource) []types.Resource {
	if len(byName) == 0 {
		return nil
	}
	names := make([]string, 0, len(byName))
	for n := range byName {
		names = append(names, n)
	}
	sort.Strings(names)
	out := make([]types.Resource, 0, len(byName))
	for _, n := range names {
		out = append(out, byName[n])
	}
	return out
}
