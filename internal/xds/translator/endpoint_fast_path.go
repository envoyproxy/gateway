// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"iter"
	"maps"
	"slices"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	corev1 "k8s.io/api/core/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/metrics"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	endpointRefreshSkipped = metrics.NewCounter("xds_endpoint_refresh_skipped_total",
		"Newer endpoint updates not applied to a cluster, by path (fast or full) and reason: empty_transition or address_type.")
	endpointSkipReason = metrics.NewLabel("reason")
	endpointSkipPath   = metrics.NewLabel("path")
)

// Values of the skip metric's path label.
const (
	pathFull = "full"
	pathFast = "fast"
)

func captureEndpointContext(table *types.ResourceVersionTable, args *xdsClusterArgs, assignmentArgs *types.LoadAssignmentArgs) {
	if table.EndpointContexts == nil || args.endpointType != EndpointTypeStatic {
		return
	}
	var hasSource bool
	names := make(map[string]struct{}, len(assignmentArgs.Settings))
	settings := make([]*ir.DestinationSetting, len(assignmentArgs.Settings))
	for i, setting := range assignmentArgs.Settings {
		// A full result's settings are matched to the published ones by name.
		if _, duplicate := names[setting.Name]; duplicate {
			return
		}
		names[setting.Name] = struct{}{}
		hasSource = hasSource || setting.EndpointSource != nil
		// Keep every field by default, so a field the builder starts reading is
		// kept too. Drop the filters and TLS material, which the assignment
		// doesn't use; only TLS presence does.
		kept := *setting
		kept.Filters = nil
		if setting.TLS != nil {
			kept.TLS = &ir.TLSUpstreamConfig{}
		}
		settings[i] = &kept
	}
	if !hasSource {
		return
	}
	saved := *assignmentArgs
	saved.Settings = settings
	table.EndpointContexts[args.name] = &saved
}

// hasEndpointsHook reports whether an extension server registered the Endpoints
// hook. A lookup error counts as registered, to stay on the safe side.
func (t *Translator) hasEndpointsHook() bool {
	if t.ExtensionManager == nil {
		return false
	}
	client, err := (*t.ExtensionManager).GetPostXDSHookClient(egv1a1.XDSEndpoints)
	return client != nil || err != nil
}

// dropPatchTargetedEndpointContexts removes the context of every cluster whose
// assignment a patch targets, since the fast path doesn't apply patches. A patch
// is enough even if it changed nothing: one that matches no endpoint now may
// match a new endpoint later.
func dropPatchTargetedEndpointContexts(table *types.ResourceVersionTable, policies []*ir.EnvoyPatchPolicy) {
	if len(table.EndpointContexts) == 0 {
		return
	}
	for _, policy := range policies {
		for _, patch := range policy.JSONPatches {
			if patch.Type == resourcev3.EndpointType {
				delete(table.EndpointContexts, patch.Name)
			}
		}
	}
}

// ReplayPublishedEndpoints is the full path. It keeps the full result's
// configuration and replaces each setting's endpoints with the published ones
// for the same backend port when those carry a newer revision.
func ReplayPublishedEndpoints(candidate, published types.EndpointContexts) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(pathFull, candidate, maps.Keys(candidate), func(cluster string, have *ir.DestinationSetting) *ir.DestinationSetting {
		args := published[cluster]
		if args == nil {
			return nil
		}
		for _, setting := range args.Settings {
			if setting.Name == have.Name {
				if setting.EndpointSource == nil || !setting.EndpointSource.SamePort(have.EndpointSource) {
					return nil
				}
				return setting
			}
		}
		return nil
	})
}

// ApplyEndpointUpdates is the fast path. For the given clusters, it keeps the
// published settings and replaces each setting's endpoints with the updated
// backend's, for the port the setting uses, when the update carries a newer revision.
func ApplyEndpointUpdates(published types.EndpointContexts, clusters iter.Seq[string], updates map[ir.ServiceKey]*endpoint.BackendEndpoints) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(pathFast, published, clusters, func(_ string, have *ir.DestinationSetting) *ir.DestinationSetting {
		backendEndpoints := updates[have.EndpointSource.Service]
		if backendEndpoints == nil {
			return nil
		}
		endpoints, addressType, source := backendEndpoints.ForPort(have.EndpointSource.PortName, corev1.Protocol(have.EndpointSource.Protocol))
		return &ir.DestinationSetting{Endpoints: endpoints, AddressType: addressType, EndpointSource: source}
	})
}

// applyNewerEndpoints is shared by both paths. For each setting of the given
// clusters, incoming returns a setting holding the candidate endpoints, or nil.
// A candidate is applied only when its revision is newer and the change is safe,
// and every cluster that changed is rebuilt. It returns the contexts with those
// clusters replaced, and the rebuilt assignments. path labels the skip metric.
func applyNewerEndpoints(path string, contexts types.EndpointContexts, clusters iter.Seq[string], incoming func(string, *ir.DestinationSetting) *ir.DestinationSetting) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	var updatedContexts types.EndpointContexts // copied on the first change
	changed := make(map[string]*endpointv3.ClusterLoadAssignment)
	for cluster := range clusters {
		args := contexts[cluster]
		if args == nil {
			continue
		}

		// Replace each setting's endpoints with the candidate's when it has a newer revision and the change is safe.
		var settings []*ir.DestinationSetting
		for i, have := range args.Settings {
			// Leave alone settings whose endpoints don't come from a Service's
			// EndpointSlices, and invalid ones (e.g. a missing backend), which
			// must not get endpoints until full translation fixes them.
			if have.EndpointSource == nil || have.Invalid {
				continue
			}
			// Newer wins on both paths: skip endpoints that aren't newer than the
			// ones this setting already has.
			candidate := incoming(cluster, have)
			if candidate == nil || candidate.EndpointSource.Revision <= have.EndpointSource.Revision {
				continue
			}
			var reason string
			switch {
			// Endpoints transitions from empty to non-empty or vice versa are not safe. They need a full translation to
			// add or remove the cluster.
			// TODO(robin): let the fast path handle these once a backend with no
			// endpoints keeps its cluster: https://github.com/envoyproxy/gateway/issues/10604
			case (len(candidate.Endpoints) == 0) != (len(have.Endpoints) == 0):
				reason = "empty_transition"
			// EDS only takes IP addresses. An FQDN or mixed update needs a DNS cluster
			// or makes the backend invalid; leave it to full translation.
			case candidate.AddressType != nil && *candidate.AddressType != ir.IP:
				reason = "address_type"
			}
			if reason != "" {
				endpointRefreshSkipped.With(endpointSkipPath.Value(path), endpointSkipReason.Value(reason)).Increment()
				continue
			}
			if settings == nil {
				settings = slices.Clone(args.Settings)
			}
			setting := *have
			setting.Endpoints, setting.AddressType, setting.EndpointSource = candidate.Endpoints, candidate.AddressType, candidate.EndpointSource
			settings[i] = &setting
		}
		if settings == nil {
			continue
		}

		// Rebuild the cluster's assignment with the updated settings.
		updatedArgs := *args
		updatedArgs.Settings = settings
		assignment := buildXdsClusterLoadAssignment(cluster, &updatedArgs)
		if err := proto.Validate(assignment); err != nil {
			return nil, nil, err
		}
		if updatedContexts == nil {
			updatedContexts = maps.Clone(contexts)
		}
		updatedContexts[cluster], changed[cluster] = &updatedArgs, assignment
	}
	if updatedContexts == nil {
		return contexts, changed, nil
	}
	return updatedContexts, changed, nil
}
