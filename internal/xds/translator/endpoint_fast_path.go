// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"maps"
	"reflect"
	"slices"

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
	for i, setting := range args.settings {
		hasSource = hasSource || setting.EndpointSource != nil
		// Only TLS presence affects endpoint metadata. Do not retain certificate
		// material, filters, or other cluster configuration through these settings.
		settings[i] = &ir.DestinationSetting{
			Name: setting.Name, Endpoints: setting.Endpoints, EndpointSource: setting.EndpointSource,
			AddressType: setting.AddressType, Weight: setting.Weight, Priority: setting.Priority,
			PreferLocal: setting.PreferLocal, Invalid: setting.Invalid, Metadata: setting.Metadata,
		}
		if setting.TLS != nil {
			settings[i].TLS = &ir.TLSUpstreamConfig{}
		}
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
			continue
		}
		context.ClusterBaseline, context.AssignmentBaseline = nil, nil
	}
}

type endpointDependency struct {
	cluster, setting string
	backend          ir.EndpointBackend
	port, protocol   string
}

func dependency(cluster string, setting *ir.DestinationSetting) endpointDependency {
	source := setting.EndpointSource
	return endpointDependency{cluster: cluster, setting: setting.Name, backend: source.Backend, port: source.PortName, protocol: source.Protocol}
}

// ReplayEndpoints uses the candidate full result's configuration and replays only
// newer inputs successfully applied to the matching dependency in the old snapshot.
func ReplayEndpoints(candidate, applied types.EndpointContexts) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	inputs := make(map[endpointDependency]*ir.DestinationSetting)
	for cluster, context := range applied {
		for _, setting := range context.Settings {
			if setting.EndpointSource != nil {
				inputs[dependency(cluster, setting)] = setting
			}
		}
	}
	return refreshEndpointContexts(candidate, func(cluster string, setting *ir.DestinationSetting) *ir.DestinationSetting {
		return inputs[dependency(cluster, setting)]
	})
}

// RefreshEndpoints always uses the context currently owned by the publisher.
func RefreshEndpoints(current types.EndpointContexts, updates map[ir.EndpointBackend]*endpoint.Snapshot) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	return refreshEndpointContexts(current, func(_ string, setting *ir.DestinationSetting) *ir.DestinationSetting {
		source := setting.EndpointSource
		snapshot := updates[source.Backend]
		if snapshot == nil {
			return nil
		}
		return &ir.DestinationSetting{
			EndpointSource: snapshot.Reference(source.PortName, corev1.Protocol(source.Protocol)),
			Endpoints:      snapshot.Ports[endpoint.Port{Name: source.PortName, Protocol: corev1.Protocol(source.Protocol)}],
			AddressType:    snapshot.AddressType,
		}
	})
}

func refreshEndpointContexts(current types.EndpointContexts, lookup func(string, *ir.DestinationSetting) *ir.DestinationSetting) (types.EndpointContexts, map[string]*endpointv3.ClusterLoadAssignment, error) {
	var next types.EndpointContexts
	changed := make(map[string]*endpointv3.ClusterLoadAssignment)
	for name, context := range current {
		var settings []*ir.DestinationSetting
		for i, setting := range context.Settings {
			if setting.EndpointSource == nil || setting.Invalid {
				continue
			}
			input := lookup(name, setting)
			if input == nil {
				continue
			}
			var reason string
			switch {
			case input.EndpointSource.Revision <= setting.EndpointSource.Revision:
				reason = "already_applied"
			case (len(input.Endpoints) == 0) != (len(setting.Endpoints) == 0):
				reason = "empty_transition"
			case !reflect.DeepEqual(input.AddressType, setting.AddressType):
				reason = "address_type"
			}
			if reason != "" {
				endpointRefreshSkipped.With(endpointSkipReason.Value(reason)).Increment()
				continue
			}
			if settings == nil {
				settings = slices.Clone(context.Settings)
			}
			updated := *setting
			updated.Endpoints, updated.AddressType, updated.EndpointSource = input.Endpoints, input.AddressType, input.EndpointSource
			settings[i] = &updated
		}
		if settings == nil {
			continue
		}
		assignment := buildXdsClusterLoadAssignment(name, settings, context.HealthCheck, context.PreferLocal, context.WeightedZones)
		if err := proto.Validate(assignment); err != nil {
			return nil, nil, err
		}
		if next == nil {
			next = maps.Clone(current)
		}
		updated := *context
		updated.Settings = settings
		next[name], changed[name] = &updated, assignment
	}
	if next == nil {
		next = current
	}
	return next, changed, nil
}
