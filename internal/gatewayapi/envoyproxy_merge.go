// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils"
)

// MergeEnvoyProxyConfigs merges EnvoyProxy configurations using a 3-level hierarchy.
// The merge is performed in two steps, with the MergeType on the more specific
// (override) resource at each step exclusively controlling the merge strategy:
//
// Step 1 - Gateway over GatewayClass:
//   - base: gatewayClassProxy
//   - override: gatewayProxy
//   - mergeType: gatewayProxy.Spec.MergeType (nil → Replace)
//
// Step 2 - Step 1 result over EnvoyGateway defaults:
//   - base: defaultSpec
//   - override: Step 1 result
//   - mergeType: gatewayClassProxy.Spec.MergeType if set, else gatewayProxy.Spec.MergeType (nil → Replace)
//
// The MergeType field in EnvoyGateway defaultSpec has no effect on merge
// behavior; it is treated as an ordinary data field.
//
// Note: If the MergeGateways option is specified then gatewayProxy will be nil
// and will not affect the resulting configuration. Settings not supplied by the
// merged EnvoyProxy are applied later at infrastructure creation time via
// GetEnvoyProxyKubeProvider().
func MergeEnvoyProxyConfigs(
	defaultSpec *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) (*egv1a1.EnvoyProxy, error) {
	var defaultProxy *egv1a1.EnvoyProxy
	if defaultSpec != nil {
		defaultProxy = &egv1a1.EnvoyProxy{Spec: *defaultSpec}
	}

	// Step 1: Merge Gateway over GatewayClass. Gateway's MergeType controls this step.
	gatewayMerged, err := mergeEnvoyProxies(gatewayClassProxy, gatewayProxy, mergeTypeOf(gatewayProxy))
	if err != nil {
		return nil, fmt.Errorf("failed to merge Gateway EnvoyProxy with GatewayClass config: %w", err)
	}

	// Step 2: Merge Step 1 result over EnvoyGateway defaults. GatewayClass's MergeType controls
	// this step, falling back to the Gateway's MergeType so a non-nil value propagates upward.
	merged, err := mergeEnvoyProxies(defaultProxy, gatewayMerged, mergeTypeOf(gatewayClassProxy, gatewayProxy))
	if err != nil {
		return nil, fmt.Errorf("failed to merge GatewayClass result with EnvoyGateway defaults: %w", err)
	}

	return merged, nil
}

// mergeTypeOf returns the first non-nil MergeType from the given proxies,
// defaulting to Replace if none is set. This allows a more specific proxy's
// MergeType to propagate upward when a less specific proxy has none set.
func mergeTypeOf(proxies ...*egv1a1.EnvoyProxy) egv1a1.MergeType {
	for _, ep := range proxies {
		if ep != nil && ep.Spec.MergeType != nil {
			return *ep.Spec.MergeType
		}
	}
	return egv1a1.Replace
}

// mergeEnvoyProxies merges an override EnvoyProxy over a base EnvoyProxy using
// the given mergeType. If base is nil, override is returned unchanged. If
// override is nil, base is returned unchanged.
func mergeEnvoyProxies(
	base *egv1a1.EnvoyProxy,
	override *egv1a1.EnvoyProxy,
	mergeType egv1a1.MergeType,
) (*egv1a1.EnvoyProxy, error) {
	if override == nil {
		return base, nil
	}
	if base == nil {
		return override, nil
	}

	if mergeType == egv1a1.Replace {
		return override, nil
	}

	// The merged object carries the override's metadata, so a backendRef
	// inherited from base without a namespace would later be resolved in the
	// override's namespace. Pin it to the namespace it was written in first.
	merged, err := utils.Merge(*qualifyTelemetryBackendRefs(base), *override, mergeType)
	if err != nil {
		return nil, err
	}
	return &merged, nil
}

// qualifyTelemetryBackendRefs returns ep with every telemetry backendRef that
// omits a namespace set to ep's own namespace, which is where the ref resolves
// when ep is used on its own. ep itself is not modified: a copy is returned
// when there is anything to set. An EnvoyProxy without a namespace, such as
// the one built from the EnvoyGateway default spec, is returned unchanged.
func qualifyTelemetryBackendRefs(ep *egv1a1.EnvoyProxy) *egv1a1.EnvoyProxy {
	if ep.Namespace == "" || !hasUnqualifiedTelemetryBackendRef(ep) {
		return ep
	}

	ep = ep.DeepCopy()
	ns := gwapiv1.Namespace(ep.Namespace)
	for _, cluster := range telemetryBackendClusters(ep) {
		for i := range cluster.BackendRefs {
			if cluster.BackendRefs[i].Namespace == nil {
				cluster.BackendRefs[i].Namespace = new(ns)
			}
		}
	}
	return ep
}

func hasUnqualifiedTelemetryBackendRef(ep *egv1a1.EnvoyProxy) bool {
	for _, cluster := range telemetryBackendClusters(ep) {
		for i := range cluster.BackendRefs {
			if cluster.BackendRefs[i].Namespace == nil {
				return true
			}
		}
	}
	return false
}

// telemetryBackendClusters returns every BackendCluster under the telemetry
// settings of ep: the tracing provider, the ALS and OpenTelemetry access log
// sinks and the OpenTelemetry metric sinks. The returned pointers alias ep.
func telemetryBackendClusters(ep *egv1a1.EnvoyProxy) []*egv1a1.BackendCluster {
	telemetry := ep.Spec.Telemetry
	if telemetry == nil {
		return nil
	}

	var clusters []*egv1a1.BackendCluster
	if telemetry.Tracing != nil {
		clusters = append(clusters, &telemetry.Tracing.Provider.BackendCluster)
	}
	if telemetry.AccessLog != nil {
		for i := range telemetry.AccessLog.Settings {
			for j := range telemetry.AccessLog.Settings[i].Sinks {
				sink := &telemetry.AccessLog.Settings[i].Sinks[j]
				if sink.ALS != nil {
					clusters = append(clusters, &sink.ALS.BackendCluster)
				}
				if sink.OpenTelemetry != nil {
					clusters = append(clusters, &sink.OpenTelemetry.BackendCluster)
				}
			}
		}
	}
	if telemetry.Metrics != nil {
		for i := range telemetry.Metrics.Sinks {
			if sink := telemetry.Metrics.Sinks[i].OpenTelemetry; sink != nil {
				clusters = append(clusters, &sink.BackendCluster)
			}
		}
	}
	return clusters
}
