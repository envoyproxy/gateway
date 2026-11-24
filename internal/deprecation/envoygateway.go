// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecation

import (
	"fmt"
	"maps"
	"slices"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// FieldWarning is a deprecation warning for the field at Path.
type FieldWarning struct {
	// Path is the path of the deprecated field.
	Path string
	// Message is the warning message.
	Message string
}

// WarnEnvoyGateway returns deprecation warnings for the provided EnvoyGateway configuration.
func WarnEnvoyGateway(eg *egv1a1.EnvoyGateway) []string {
	fields := EnvoyGatewayFields(eg)
	if len(fields) == 0 {
		return nil
	}
	warnings := make([]string, 0, len(fields))
	for _, f := range fields {
		warnings = append(warnings, f.Message)
	}
	return warnings
}

// EnvoyGatewayFields returns the deprecated fields used in the provided EnvoyGateway configuration.
func EnvoyGatewayFields(eg *egv1a1.EnvoyGateway) []FieldWarning {
	if eg == nil {
		return nil
	}
	var warnings []FieldWarning
	if eg.ExtensionAPIs != nil && eg.ExtensionAPIs.DisableLua != nil {
		warnings = append(warnings, FieldWarning{Path: "extensionApis.disableLua", Message: "disableLua is deprecated, use enableLua instead"})
	}
	if eg.ExtensionManager != nil {
		warnings = append(warnings, warnDeprecatedExtensionServiceFields("extensionManager.service", eg.ExtensionManager.Service)...)
	}
	for i := range eg.ExtensionManagers {
		warnings = append(warnings, warnDeprecatedExtensionServiceFields(fmt.Sprintf("extensionManagers[%d].service", i), eg.ExtensionManagers[i].Service)...)
	}
	if p := eg.Provider; p != nil && p.Custom != nil && p.Custom.Infrastructure != nil && p.Custom.Infrastructure.Remote != nil {
		warnings = append(warnings, warnDeprecatedExtensionServiceFields("provider.custom.infrastructure.remote.service", p.Custom.Infrastructure.Remote.Service)...)
	}
	// The default EnvoyProxy spec isn't an EnvoyProxy resource, so it has no status to carry a
	// DeprecatedField warning; surface its deprecated fields at startup instead.
	envoyProxyFields := make(map[string]DeprecatedField)
	AddDeprecatedEnvoyProxySpecFields(envoyProxyFields, "envoyProxy", eg.EnvoyProxy, true)
	for _, path := range slices.Sorted(maps.Keys(envoyProxyFields)) {
		warnings = append(warnings, FieldWarning{Path: path, Message: envoyProxyFields[path].Message(path)})
	}
	return warnings
}

// warnDeprecatedExtensionServiceFields warns about the deprecated host/port fields of the
// ExtensionService at path. The EnvoyGateway configuration is not CRD-defaulted, so a nonzero port
// is explicit, and is deprecated even when the target has been moved to fqdn, ip or unix.
func warnDeprecatedExtensionServiceFields(path string, svc *egv1a1.ExtensionService) []FieldWarning {
	switch {
	case svc == nil:
		return nil
	case svc.Host != "":
		return []FieldWarning{{Path: path + ".host", Message: fmt.Sprintf("%[1]s.host and %[1]s.port are deprecated and will be removed in %[2]s, use %[1]s.fqdn, %[1]s.ip or %[1]s.unix instead",
			path, DeprecatedFieldsRemovalRelease)}}
	case svc.Port != 0:
		return []FieldWarning{{Path: path + ".port", Message: fmt.Sprintf("%[1]s.port is deprecated and will be removed in %[2]s, use the port of %[1]s.fqdn or %[1]s.ip instead",
			path, DeprecatedFieldsRemovalRelease)}}
	}
	return nil
}
