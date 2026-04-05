// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

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

	// If neither gatewayClassProxy nor gatewayProxy defined an EnvoyProxy, there is
	// nothing to apply defaults to — return nil regardless of defaultSpec.
	if gatewayMerged == nil {
		return nil, nil
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

	merged, err := utils.Merge(*base, *override, mergeType)
	if err != nil {
		return nil, err
	}
	return &merged, nil
}
