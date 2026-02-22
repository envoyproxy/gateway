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

// MergeEnvoyProxyConfigs merges EnvoyProxy configurations using a 3-level hierarchy:
// 1. defaultSpec - EnvoyProxySpec from EnvoyGateway.Provider.Kubernetes.EnvoyProxyDefault (base defaults)
// 2. gatewayClassProxy - EnvoyProxy from GatewayClass parametersRef (overrides defaults)
// 3. gatewayProxy - EnvoyProxy from Gateway parametersRef (highest priority). Note that this is only present if the MergeGateways option is false.
//
// The merge behavior depends on the MergeType field:
// - nil or Replace: More specific configs completely replace less specific ones (no merging)
// - StrategicMerge: Configs are merged using Kubernetes strategic merge patch
// - JSONMerge: Configs are merged using JSON merge patch
//
// The MergeType is determined by looking at all provided configs in priority order (gateway > gatewayClass > default).
//
// Note:  If the MergeGateways option is specified then the gatewayProxy will be nil thus will
// not affect the resulting merged configuration.  Furthermore, if there are settings not
// supplied by the merged EnvoyProxy, those are applied later at infrastructure creation time
// via GetEnvoyProxyKubeProvider() which calls the existing default functions.
//
// Returns:
//   - merged: The merged EnvoyProxy configuration. On error, this contains the fallback configuration
//     using priority-based selection so the gateway can continue to function.
//   - err: Any error that occurred during merging. Even when err is non-nil, merged will contain
//     a valid fallback configuration.
func MergeEnvoyProxyConfigs(
	defaultSpec *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) (*egv1a1.EnvoyProxy, error) {

	var defaultProxy *egv1a1.EnvoyProxy
	if defaultSpec != nil {
		defaultProxy = &egv1a1.EnvoyProxy{Spec: *defaultSpec}
	}

	// Merge GatewayClass over any EnvoyGateway defaults
	base, err := mergeEnvoyProxies(defaultProxy, gatewayClassProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to merge GatewayClass EnvoyProxy with EnvoyGateway defaults: %w", err)
	}

	// Merge Gateway over the GatewayClass result
	merged, err := mergeEnvoyProxies(base, gatewayProxy)
	if err != nil {
		return nil, fmt.Errorf("failed to merge Gateway EnvoyProxy with GatewayClass config: %w", err)
	}

	return merged, nil
}

// determineMergeType finds the MergeType to use for a base and override EnvoyProxy.
func determineMergeType(
	base *egv1a1.EnvoyProxy,
	override *egv1a1.EnvoyProxy,
) egv1a1.MergeType {
	// Check the override first as that will have higher priority.
	if override != nil && override.Spec.MergeType != nil {
		return *override.Spec.MergeType
	}

	if base != nil && base.Spec.MergeType != nil {
		return *base.Spec.MergeType
	}

	// No MergeType specified anywhere, return default Replace
	return egv1a1.Replace
}

// mergeEnvoyProxies merges an override EnvoyProxy over a base EnvoyProxy.
// If base is nil, returns override. If override is nil, returns base.
// For Replace strategy, override completely replaces base (no field-level merging).
// For StrategicMerge or JSONMerge, fields are merged according to the strategy.
func mergeEnvoyProxies(
	base *egv1a1.EnvoyProxy,
	override *egv1a1.EnvoyProxy,
) (*egv1a1.EnvoyProxy, error) {
	if override == nil {
		return base, nil
	}
	if base == nil {
		return override, nil
	}

	mergeType := determineMergeType(base, override)

	if mergeType == egv1a1.Replace {
		return override, nil
	}

	merged, err := utils.Merge(*base, *override, mergeType)
	if err != nil {
		return nil, err
	}
	return &merged, nil
}
