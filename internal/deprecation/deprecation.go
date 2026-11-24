// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package deprecation reports the deprecated API fields in use, so that users are warned
// before they are removed.
package deprecation

import (
	"fmt"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// DeprecatedField describes the replacement for a deprecated field.
type DeprecatedField struct {
	// Replacement is the field to use instead.
	Replacement string
	// RemovedIn is the release the field is scheduled to be removed in, if any.
	RemovedIn string
}

// Message returns the warning message for the deprecated field at path.
func (f DeprecatedField) Message(path string) string {
	if f.RemovedIn != "" {
		return fmt.Sprintf("%s is deprecated and will be removed in %s, use %s instead", path, f.RemovedIn, f.Replacement)
	}
	return fmt.Sprintf("%s is deprecated, use %s instead", path, f.Replacement)
}

// AddDeprecatedEnvoyProxySpecFields records the deprecated fields used in the EnvoyProxy spec at
// path, which is "spec" for an EnvoyProxy resource and "envoyProxy" for the EnvoyGateway default.
func AddDeprecatedEnvoyProxySpecFields(fields map[string]DeprecatedField, path string, spec *egv1a1.EnvoyProxySpec) {
	if spec == nil {
		return
	}
	if spec.LuaValidation != nil {
		fields[path+".luaValidation"] = DeprecatedField{Replacement: path + ".lua.validationType"}
	}
}
