// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package bootstrap

import (
	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	yamlutils "github.com/envoyproxy/gateway/internal/utils/yaml"
)

// ApplyBootstrapConfig applies the bootstrap config to the default bootstrap config and return the result config.
func ApplyBootstrapConfig(boostrapConfig *egcfgv1a1.ProxyBootstrap, defaultBootstrap string) (string, error) {
	bootstrapType := boostrapConfig.Type
	if bootstrapType != nil && *bootstrapType == egcfgv1a1.BootstrapTypeMerge {
		mergedBootstrap, err := yamlutils.MergeYAML(defaultBootstrap, boostrapConfig.Value)
		if err != nil {
			return "", err
		}
		return mergedBootstrap, nil
	}
	return boostrapConfig.Value, nil
}
