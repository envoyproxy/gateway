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
// 1. template - EnvoyProxyTemplateSpec from EnvoyGateway.Provider.Kubernetes.EnvoyProxyTemplate (base defaults)
// 2. gatewayClassProxy - EnvoyProxy from GatewayClass parametersRef (overrides template)
// 3. gatewayProxy - EnvoyProxy from Gateway parametersRef (highest priority).  Note that this only present if the MergeGateways option is false.
//
// The merge behavior depends on the template's MergeType:
// - Replace (default): More specific configs completely replace less specific ones (no merging)
// - StrategicMerge: Configs are merged using Kubernetes strategic merge patch
// - JSONMerge: Configs are merged using JSON merge patch
//
// Note: If there are settings which are not supplied by the merged EnvoyProxy, those are applied later
// at infrastructure creation time via GetEnvoyProxyKubeProvider() which calls the existing default functions.
//
// Returns a merged EnvoyProxy with settings from all levels, or the highest priority config if only one exists.
func MergeEnvoyProxyConfigs(
	template *egv1a1.EnvoyProxyTemplateSpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
) (*egv1a1.EnvoyProxy, error) {
	// Determine merge strategy (default to Replace for backwards compatibility)
	mergeType := egv1a1.EnvoyProxyTemplateMergeTypeReplace
	if template != nil && template.MergeType != nil {
		mergeType = *template.MergeType
	}

	// Handle Replace strategy: most specific config wins, no merging
	if mergeType == egv1a1.EnvoyProxyTemplateMergeTypeReplace {
		return replaceEnvoyProxy(template, gatewayClassProxy, gatewayProxy), nil
	}

	// Handle merge strategies (StrategicMerge or JSONMerge)
	return mergeEnvoyProxy(template, gatewayClassProxy, gatewayProxy, mergeType)
}

// replaceEnvoyProxy implements the Replace strategy where more specific configs completely replace less specific ones
func replaceEnvoyProxy(
	template *egv1a1.EnvoyProxyTemplateSpec,
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

	// Template is used as fallback
	if template != nil && template.Spec != nil {
		return &egv1a1.EnvoyProxy{
			Spec: *template.Spec,
		}
	}

	return nil
}

// mergeEnvoyProxy merges EnvoyProxy configs using the specified merge strategy
func mergeEnvoyProxy(
	template *egv1a1.EnvoyProxyTemplateSpec,
	gatewayClassProxy *egv1a1.EnvoyProxy,
	gatewayProxy *egv1a1.EnvoyProxy,
	mergeType egv1a1.EnvoyProxyTemplateMergeType,
) (*egv1a1.EnvoyProxy, error) {
	var result *egv1a1.EnvoyProxy

	// Start with template as base (if provided)
	if template != nil && template.Spec != nil {
		result = &egv1a1.EnvoyProxy{
			Spec: *template.Spec,
		}
	}

	// Convert merge type to the type expected by utils.Merge.
	var utilsMergeType egv1a1.MergeType
	if mergeType == egv1a1.EnvoyProxyTemplateMergeTypeJSONMerge {
		utilsMergeType = egv1a1.JSONMerge
	} else {
		utilsMergeType = egv1a1.StrategicMerge
	}

	// Merge GatewayClass-level EnvoyProxy over template
	if gatewayClassProxy != nil {
		if result != nil {
			merged, err := mergeEnvoyProxyWithMergeType(result, gatewayClassProxy, utilsMergeType)
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
			merged, err := mergeEnvoyProxyWithMergeType(result, gatewayProxy, utilsMergeType)
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

// mergeEnvoyProxyWithMergeType merges two EnvoyProxy objects using the specified merge type
func mergeEnvoyProxyWithMergeType(base, overlay *egv1a1.EnvoyProxy, mergeType egv1a1.MergeType) (egv1a1.EnvoyProxy, error) {
	return utils.Merge(*base, *overlay, mergeType)
}
