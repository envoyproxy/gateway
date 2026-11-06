// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"maps"

	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	protobuf "google.golang.org/protobuf/proto"
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
		"Endpoint updates not applied to a cluster, by path (fast or full) and reason: already_applied (fast only), empty_transition or address_type.")
	endpointSkipReason = metrics.NewLabel("reason")
	endpointSkipPath   = metrics.NewLabel("path")
)

// Values of the skip metric's path label.
const (
	pathFull = "full"
	pathFast = "fast"
)

func captureEndpointContext(table *types.ResourceVersionTable, args *xdsClusterArgs, assignmentArgs *types.LoadAssignmentArgs) {
	if table.EndpointState.Contexts == nil || args.endpointType != EndpointTypeStatic {
		return
	}
	var hasSource bool
	settings := make([]*ir.DestinationSetting, len(assignmentArgs.Settings))
	settingEndpoints := make(map[string]types.SettingEndpoints, len(assignmentArgs.Settings))
	for i, setting := range assignmentArgs.Settings {
		if _, duplicate := settingEndpoints[setting.Name]; duplicate {
			return
		}
		hasSource = hasSource || setting.EndpointSource != nil
		// Keep every field by default, so a field the builder starts reading is
		// kept too. Drop the endpoint fields (endpoints, address type, source),
		// which are saved separately and change with each update, and the filters
		// and TLS material, which the assignment doesn't use; only TLS presence does.
		kept := *setting
		kept.Endpoints, kept.AddressType, kept.EndpointSource = nil, nil, nil
		kept.Filters = nil
		if setting.TLS != nil {
			kept.TLS = &ir.TLSUpstreamConfig{}
		}
		settings[i] = &kept
		settingEndpoints[setting.Name] = types.SettingEndpoints{Endpoints: setting.Endpoints, AddressType: setting.AddressType, Source: setting.EndpointSource}
	}
	if !hasSource {
		return
	}
	saved := *assignmentArgs
	saved.Settings = settings
	table.EndpointState.Contexts[args.name] = &saved
	table.EndpointState.Endpoints[args.name] = settingEndpoints
}

// endpointsMayBePatched reports whether anything after cluster construction can
// change a cluster's endpoints: an EnvoyPatchPolicy, or an extension hook for
// endpoints or the whole translation.
func (t *Translator) endpointsMayBePatched(xdsIR *ir.Xds) bool {
	if len(xdsIR.EnvoyPatchPolicies) > 0 {
		return true
	}
	if t.ExtensionManager == nil {
		return false
	}
	for _, hook := range []egv1a1.XDSTranslatorHook{egv1a1.XDSEndpoints, egv1a1.XDSTranslation} {
		if client, err := (*t.ExtensionManager).GetPostXDSHookClient(hook); client != nil || err != nil {
			return true
		}
	}
	return false
}

// dropPatchedEndpointContexts removes the context of every cluster whose
// assignment a patch targets or an extension hook changed, since the fast path
// doesn't apply them. A patch is enough even if it changed nothing: one that
// matches no endpoint now may match a new endpoint later.
func dropPatchedEndpointContexts(table *types.ResourceVersionTable, policies []*ir.EnvoyPatchPolicy) {
	if len(table.EndpointState.Contexts) == 0 {
		return
	}
	for _, policy := range policies {
		for _, patch := range policy.JSONPatches {
			if patch.Type == resourcev3.EndpointType {
				delete(table.EndpointState.Contexts, patch.Name)
				delete(table.EndpointState.Endpoints, patch.Name)
			}
		}
	}
	assignments := make(map[string]*endpointv3.ClusterLoadAssignment)
	for _, value := range table.XdsResources[resourcev3.EndpointType] {
		if assignment, ok := value.(*endpointv3.ClusterLoadAssignment); ok {
			assignments[assignment.ClusterName] = assignment
		}
	}
	for name, context := range table.EndpointState.Contexts {
		unpatched := buildXdsClusterLoadAssignment(name, context.WithEndpoints(table.EndpointState.Endpoints[name]))
		if !protobuf.Equal(unpatched, assignments[name]) {
			delete(table.EndpointState.Contexts, name)
			delete(table.EndpointState.Endpoints, name)
		}
	}
}

// ReplayPublishedEndpoints is the full path. It keeps the full result's
// configuration and replaces each setting's endpoints with the published ones
// for the same backend port when those carry a newer revision.
// This prevents a slow full translation from overwriting newer endpoints that
// were published by a fast update.
func ReplayPublishedEndpoints(contexts types.EndpointContexts, candidate, published types.EndpointsByCluster) (types.EndpointsByCluster, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(contexts, candidate, pathFull, func(cluster, setting string, have *types.SettingEndpoints) *types.SettingEndpoints {
		input, ok := published[cluster][setting]
		if !ok || input.Source == nil || !input.Source.SamePort(have.Source) {
			return nil
		}
		return &input
	})
}

// ApplyEndpointUpdates is the fast path. It keeps the current contexts and
// replaces each setting's endpoints with the updated backend's, for the port
// the setting uses, when the update carries a newer revision.
func ApplyEndpointUpdates(contexts types.EndpointContexts, applied types.EndpointsByCluster, updates map[ir.ServiceKey]*endpoint.BackendEndpoints) (types.EndpointsByCluster, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(contexts, applied, pathFast, func(_, _ string, have *types.SettingEndpoints) *types.SettingEndpoints {
		backendEndpoints := updates[have.Source.Service]
		if backendEndpoints == nil {
			return nil
		}
		endpoints, addressType, source := backendEndpoints.ForPort(have.Source.PortName, corev1.Protocol(have.Source.Protocol))
		return &types.SettingEndpoints{Endpoints: endpoints, AddressType: addressType, Source: source}
	})
}

// applyNewerEndpoints is shared by both paths; path ("fast" or "full") labels the
// skip metric. incoming returns the candidate endpoints for one applied setting, or nil. A candidate is applied only when
// its revision is newer and the change is safe; every cluster that changed is
// rebuilt. It returns the applied endpoints with those clusters replaced, or
// the input unchanged, and the rebuilt assignments by cluster name.
func applyNewerEndpoints(contexts types.EndpointContexts, applied types.EndpointsByCluster, path string, incoming func(string, string, *types.SettingEndpoints) *types.SettingEndpoints) (types.EndpointsByCluster, map[string]*endpointv3.ClusterLoadAssignment, error) {
	next := applied
	changed := make(map[string]*endpointv3.ClusterLoadAssignment)
	for cluster, context := range contexts {
		var updated map[string]types.SettingEndpoints
		for _, setting := range context.Settings {
			have, ok := applied[cluster][setting.Name]
			if !ok || have.Source == nil || setting.Invalid {
				continue
			}
			candidate := incoming(cluster, setting.Name, &have)
			if candidate == nil {
				continue
			}
			var reason string
			switch {
			case candidate.Source.Revision <= have.Source.Revision:
				// On the full path this just means there is no newer revision, but on the fast path
				// it means the full path has already published this update, so skip it.
				if path == pathFull {
					continue
				}
				reason = "already_applied"
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
			if updated == nil {
				updated = maps.Clone(applied[cluster])
			}
			updated[setting.Name] = *candidate
		}
		if updated == nil {
			continue
		}
		assignment := buildXdsClusterLoadAssignment(cluster, context.WithEndpoints(updated))
		if err := proto.Validate(assignment); err != nil {
			return nil, nil, err
		}
		if len(changed) == 0 {
			next = maps.Clone(applied)
		}
		next[cluster], changed[cluster] = updated, assignment
	}
	return next, changed, nil
}
