// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"maps"
	"slices"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// WarnEnvoyGateway returns deprecation warnings for the provided EnvoyGateway configuration.
func WarnEnvoyGateway(eg *egv1a1.EnvoyGateway) []string {
	if eg == nil {
		return nil
	}
	var warnings []string
	if eg.ExtensionAPIs != nil && eg.ExtensionAPIs.DisableLua != nil {
		warnings = append(warnings, "disableLua is deprecated, use enableLua instead")
	}
	// The default EnvoyProxy spec isn't an EnvoyProxy resource, so it has no status to carry a
	// DeprecatedField warning; surface its deprecated fields at startup instead.
	envoyProxyFields := make(map[string]DeprecatedField)
	AddDeprecatedEnvoyProxySpecFields(envoyProxyFields, "envoyProxy", eg.EnvoyProxy)
	for _, path := range slices.Sorted(maps.Keys(envoyProxyFields)) {
		warnings = append(warnings, envoyProxyFields[path].Message(path))
	}
	return warnings
}
