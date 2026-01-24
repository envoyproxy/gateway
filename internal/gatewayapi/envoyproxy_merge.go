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
	// Determine merge strategy from the configs (highest priority wins)
	mergeType := determineMergeType(defaultSpec, gatewayClassProxy, gatewayProxy)

	// Handle Replace strategy (nil/unset MergeType): most specific config wins, no merging
	if mergeType == egv1a1.Replace {
		return replaceEnvoyProxy(defaultSpec, gatewayClassProxy, gatewayProxy), nil
	}

	// Handle merge strategies (StrategicMerge or JSONMerge)
	merged, err := mergeEnvoyProxy(defaultSpec, gatewayClassProxy, gatewayProxy, mergeType)
	if err != nil {
		return nil, err
	}

	return merged, nil
}

// determineMergeType finds the MergeType to use by checking configs in priority order.
// Gateway > GatewayClass > Default. If no MergeType is specified, returns nil (Replace behavior).
func determineMergeType(
	defaultSpec *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) egv1a1.MergeType {
	// Check gateway level first (highest priority)
	if gatewayProxy != nil && gatewayProxy.Spec.MergeType != nil {
		return *gatewayProxy.Spec.MergeType
	}

	// Check gatewayClass level next
	if gatewayClassProxy != nil && gatewayClassProxy.Spec.MergeType != nil {
		return *gatewayClassProxy.Spec.MergeType
	}

	// Check default spec last
	if defaultSpec != nil && defaultSpec.MergeType != nil {
		return *defaultSpec.MergeType
	}

	// No MergeType specified anywhere, return default Replace
	return egv1a1.Replace
}

// replaceEnvoyProxy implements the Replace strategy where more specific configs completely replace less specific ones
func replaceEnvoyProxy(
	defaultSpec *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) *egv1a1.EnvoyProxy {
	// Gateway level has highest priority
	if gatewayProxy != nil {
		return gatewayProxy
	}

	// GatewayClass level is next
	if gatewayClassProxy != nil {
		return gatewayClassProxy
	}

	// Default spec is used as fallback
	if defaultSpec != nil {
		return &egv1a1.EnvoyProxy{
			Spec: *defaultSpec,
		}
	}

	return nil
}

// mergeEnvoyProxy merges EnvoyProxy configs using the specified merge strategy
func mergeEnvoyProxy(
	defaultSpec *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
	mergeType egv1a1.MergeType,
) (*egv1a1.EnvoyProxy, error) {
	var result *egv1a1.EnvoyProxy

	// Start with default spec as base (if provided)
	if defaultSpec != nil {
		result = &egv1a1.EnvoyProxy{
			Spec: *defaultSpec,
		}
	}

	// Merge GatewayClass-level EnvoyProxy over default
	if gatewayClassProxy != nil {
		if result != nil {
			merged, err := utils.Merge(*result, *gatewayClassProxy, mergeType)
			if err != nil {
				return nil, fmt.Errorf("failed to merge GatewayClass EnvoyProxy %s/%s with default EnvoyProxySpec using %s merge: %w",
					gatewayClassProxy.Namespace, gatewayClassProxy.Name, mergeType, err)
			}
			result = &merged
		} else {
			result = gatewayClassProxy
		}
	}

	// Merge Gateway-level EnvoyProxy over GatewayClass and default (highest priority)
	if gatewayProxy != nil {
		if result != nil {
			merged, err := utils.Merge(*result, *gatewayProxy, mergeType)
			if err != nil {
				return nil, fmt.Errorf("failed to merge Gateway EnvoyProxy %s/%s with base configuration using %s merge: %w",
					gatewayProxy.Namespace, gatewayProxy.Name, mergeType, err)
			}
			result = &merged
		} else {
			result = gatewayProxy
		}
	}

	return result, nil
}
