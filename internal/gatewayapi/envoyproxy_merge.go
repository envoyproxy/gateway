// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils"
)

// MergeEnvoyProxyConfigs merges EnvoyProxy configurations using a 3-level hierarchy:
// 1. template - EnvoyProxySpec from EnvoyGateway.Provider.Kubernetes.EnvoyProxyTemplate (base defaults)
// 2. gatewayClassProxy - EnvoyProxy from GatewayClass parametersRef (overrides template)
// 3. gatewayProxy - EnvoyProxy from Gateway parametersRef (highest priority).  Note that this only present if the MergeGateways option is false.
//
// Note: If there are settings which are not supplied by the merged EnvoyProxy, those are applied later
// at infrastructure creation time via GetEnvoyProxyKubeProvider() which calls the existing default functions.
//
// Returns a merged EnvoyProxy with settings from all levels, or the highest priority config if only one exists.
func MergeEnvoyProxyConfigs(
	template *egv1a1.EnvoyProxySpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) (*egv1a1.EnvoyProxy, error) {
	// Start with template as base (if provided)
	var result *egv1a1.EnvoyProxy

	if template != nil {
		// Convert template spec to full EnvoyProxy for merging
		result = &egv1a1.EnvoyProxy{
			Spec: *template,
		}
	}

	// Merge GatewayClass-level EnvoyProxy over template
	if gatewayClassProxy != nil {
		if result != nil {
			merged, err := mergeEnvoyProxy(result, gatewayClassProxy)
			if err != nil {
				return nil, err
			}
			result = &merged
		} else {
			result = gatewayClassProxy
		}
	}

	// Merge Gateway-level EnvoyProxy over GatewayClass and template (highest priority)
	if gatewayProxy != nil {
		if result != nil {
			merged, err := mergeEnvoyProxy(result, gatewayProxy)
			if err != nil {
				return nil, err
			}
			result = &merged
		} else {
			result = gatewayProxy
		}
	}

	return result, nil
}

// mergeEnvoyProxy merges two EnvoyProxy objects using strategic merge
func mergeEnvoyProxy(base, overlay *egv1a1.EnvoyProxy) (egv1a1.EnvoyProxy, error) {
	return utils.Merge(*base, *overlay, egv1a1.StrategicMerge)
}
