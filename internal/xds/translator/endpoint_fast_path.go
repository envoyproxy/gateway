// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"maps"
	"reflect"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	protobuf "google.golang.org/protobuf/proto"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/metrics"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

var (
	endpointRefreshSkipped = metrics.NewCounter("xds_endpoint_refresh_skipped_total",
		"Endpoint contributions skipped during fast refresh or full-result replay, by reason.")
	endpointSkipReason = metrics.NewLabel("reason")
)

func captureEndpointContext(table *types.ResourceVersionTable, args *xdsClusterArgs, cluster *clusterv3.Cluster, assignment *endpointv3.ClusterLoadAssignment) {
	if table.EndpointContexts == nil || args.endpointType != EndpointTypeStatic {
		return
	}
	var hasSource bool
	settings := make([]*ir.DestinationSetting, len(args.settings))
	inputs := make(map[string]types.EndpointInput, len(args.settings))
	for i, setting := range args.settings {
		if _, duplicate := inputs[setting.Name]; duplicate {
			return
		}
		hasSource = hasSource || setting.EndpointSource != nil
		// Only TLS presence affects endpoint metadata. Do not retain certificate
		// material, filters, or other cluster configuration through these settings.
		settings[i] = &ir.DestinationSetting{
			Name: setting.Name, Weight: setting.Weight, Priority: setting.Priority,
			PreferLocal: setting.PreferLocal, Invalid: setting.Invalid, Metadata: setting.Metadata,
		}
		if setting.TLS != nil {
			settings[i].TLS = &ir.TLSUpstreamConfig{}
		}
		inputs[setting.Name] = types.EndpointInput{Endpoints: setting.Endpoints, AddressType: setting.AddressType, Source: setting.EndpointSource}
	}
	if !hasSource {
		return
	}
	lb := ptr.Deref(args.loadBalancer, ir.LoadBalancer{})
	table.EndpointContexts[args.name] = &types.EndpointContext{
		Settings: settings, HealthCheck: args.healthCheck, PreferLocal: lb.PreferLocal, WeightedZones: lb.WeightedZones,
		ClusterBaseline:    protobuf.Clone(cluster).(*clusterv3.Cluster),
		AssignmentBaseline: protobuf.Clone(assignment).(*endpointv3.ClusterLoadAssignment),
	}
	table.EndpointInputs[args.name] = inputs
}

func finalizeEndpointContexts(table *types.ResourceVersionTable) {
	clusters := make(map[string]*clusterv3.Cluster)
	assignments := make(map[string]*endpointv3.ClusterLoadAssignment)
	if len(table.EndpointContexts) == 0 {
		return
	}
	for _, value := range table.XdsResources[resourcev3.ClusterType] {
		if cluster, ok := value.(*clusterv3.Cluster); ok {
			clusters[cluster.Name] = cluster
		}
	}
	for _, value := range table.XdsResources[resourcev3.EndpointType] {
		if assignment, ok := value.(*endpointv3.ClusterLoadAssignment); ok {
			assignments[assignment.ClusterName] = assignment
		}
	}
	for name, context := range table.EndpointContexts {
		if !protobuf.Equal(context.ClusterBaseline, clusters[name]) || !protobuf.Equal(context.AssignmentBaseline, assignments[name]) {
			delete(table.EndpointContexts, name)
			delete(table.EndpointInputs, name)
			continue
		}
		context.ClusterBaseline, context.AssignmentBaseline = nil, nil
	}
}

// sameDependency reports whether two sources name the same backend port, so
// endpoints applied to one can stand in for the other.
func sameDependency(a, b *ir.EndpointSource) bool {
	return a.Backend == b.Backend && a.PortName == b.PortName && a.Protocol == b.Protocol
}

// ReplayPublishedEndpoints is the full path. It keeps the full result's
// configuration and replaces each setting's endpoints with the published ones
// for the same dependency when those carry a newer revision.
func ReplayPublishedEndpoints(contexts types.EndpointContexts, candidate, published types.EndpointInputs) (types.EndpointInputs, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(contexts, candidate, func(cluster, setting string, have *types.EndpointInput) *types.EndpointInput {
		input, ok := published[cluster][setting]
		if !ok || input.Source == nil || !sameDependency(input.Source, have.Source) {
			return nil
		}
		return &input
	})
}

// ApplyEndpointUpdates is the fast path. It keeps the current contexts and
// replaces each setting's endpoints with the updated backend's, for the port
// the setting uses, when the update carries a newer revision.
func ApplyEndpointUpdates(contexts types.EndpointContexts, applied types.EndpointInputs, updates map[ir.EndpointBackend]*endpoint.BackendEndpoints) (types.EndpointInputs, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return applyNewerEndpoints(contexts, applied, func(_, _ string, have *types.EndpointInput) *types.EndpointInput {
		backendEndpoints := updates[have.Source.Backend]
		if backendEndpoints == nil {
			return nil
		}
		endpoints, addressType, source := backendEndpoints.ForPort(have.Source.PortName, corev1.Protocol(have.Source.Protocol))
		return &types.EndpointInput{Endpoints: endpoints, AddressType: addressType, Source: source}
	})
}

// applyNewerEndpoints is shared by both paths. incoming returns the candidate
// endpoints for one applied setting, or nil. A candidate is applied only when
// its revision is newer and the change is safe; every cluster that changed is
// rebuilt. It returns the applied endpoints with those clusters replaced, or
// the input unchanged, and the rebuilt assignments by cluster name.
func applyNewerEndpoints(contexts types.EndpointContexts, applied types.EndpointInputs, incoming func(string, string, *types.EndpointInput) *types.EndpointInput) (types.EndpointInputs, map[string]*endpointv3.ClusterLoadAssignment, error) {
	next := applied
	changed := make(map[string]*endpointv3.ClusterLoadAssignment)
	for cluster, context := range contexts {
		var updated map[string]types.EndpointInput
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
				reason = "already_applied"
			case (len(candidate.Endpoints) == 0) != (len(have.Endpoints) == 0):
				reason = "empty_transition"
			case !reflect.DeepEqual(candidate.AddressType, have.AddressType):
				reason = "address_type"
			}
			if reason != "" {
				endpointRefreshSkipped.With(endpointSkipReason.Value(reason)).Increment()
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
		assignment := buildXdsClusterLoadAssignment(cluster, context.BuildSettings(updated), context.HealthCheck, context.PreferLocal, context.WeightedZones)
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
