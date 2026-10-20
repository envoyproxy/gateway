// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package validation

import (
	"fmt"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

// DeprecatedFieldsRemovalRelease is the release that removes the deprecated fields tracked in
// https://github.com/envoyproxy/gateway/issues/10211.
const DeprecatedFieldsRemovalRelease = "v1.11"

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

// scheduledForRemoval describes a deprecated field replaced by replacement that is scheduled for
// removal in DeprecatedFieldsRemovalRelease.
func scheduledForRemoval(replacement string) DeprecatedField {
	return DeprecatedField{Replacement: replacement, RemovedIn: DeprecatedFieldsRemovalRelease}
}

// AddDeprecatedEnvoyProxySpecFields records the deprecated fields used in the EnvoyProxy spec at
// path, which is "spec" for an EnvoyProxy resource and "envoyProxy" for the EnvoyGateway default.
//
// explicitPorts reports whether a nonzero telemetry port was set by the user. It is false for an
// EnvoyProxy resource, whose ports always carry the CRD default, so only host is checked there. It
// is true for the EnvoyGateway default, which is loaded from the EnvoyGateway configuration without
// CRD defaulting, so a port there is explicit and is deprecated even without host.
func AddDeprecatedEnvoyProxySpecFields(fields map[string]DeprecatedField, path string, spec *egv1a1.EnvoyProxySpec, explicitPorts bool) {
	if spec == nil {
		return
	}
	if spec.LuaValidation != nil {
		fields[path+".luaValidation"] = DeprecatedField{Replacement: path + ".lua.validationType"}
	}

	telemetry := spec.Telemetry
	if telemetry == nil {
		return
	}
	if telemetry.Tracing != nil {
		providerPath := path + ".telemetry.tracing.provider"
		AddDeprecatedBackendClusterFields(fields, providerPath,
			&telemetry.Tracing.Provider.BackendCluster, telemetry.Tracing.Provider.Host)
		if explicitPorts {
			addDeprecatedPort(fields, providerPath, telemetry.Tracing.Provider.Host, telemetry.Tracing.Provider.Port)
		}
	}
	if telemetry.Metrics != nil {
		for i, sink := range telemetry.Metrics.Sinks {
			if sink.OpenTelemetry != nil {
				sinkPath := fmt.Sprintf("%s.telemetry.metrics.sinks[%d].openTelemetry", path, i)
				AddDeprecatedBackendClusterFields(fields, sinkPath, &sink.OpenTelemetry.BackendCluster, sink.OpenTelemetry.Host)
				if explicitPorts {
					addDeprecatedPort(fields, sinkPath, sink.OpenTelemetry.Host, sink.OpenTelemetry.Port)
				}
			}
		}
	}
	if telemetry.AccessLog != nil {
		for i, setting := range telemetry.AccessLog.Settings {
			for j, sink := range setting.Sinks {
				sinkPath := fmt.Sprintf("%s.telemetry.accessLog.settings[%d].sinks[%d]", path, i, j)
				if sink.OpenTelemetry != nil {
					AddDeprecatedBackendClusterFields(fields, sinkPath+".openTelemetry",
						&sink.OpenTelemetry.BackendCluster, sink.OpenTelemetry.Host)
					if explicitPorts {
						addDeprecatedPort(fields, sinkPath+".openTelemetry", sink.OpenTelemetry.Host, sink.OpenTelemetry.Port)
					}
				}
				if sink.ALS != nil {
					AddDeprecatedBackendClusterFields(fields, sinkPath+".als", &sink.ALS.BackendCluster, nil)
				}
			}
		}
	}
}

// AddDeprecatedBackendClusterFields records the deprecated fields used in the BackendCluster at
// path: backendRef, the legacy host of the owning sink (if any), and the consistent hash fields of
// its backendSettings. Port is not checked here, see AddDeprecatedEnvoyProxySpecFields.
func AddDeprecatedBackendClusterFields(fields map[string]DeprecatedField, path string, cluster *egv1a1.BackendCluster, host *string) {
	if cluster == nil {
		return
	}
	if host != nil {
		fields[path+".host"] = scheduledForRemoval(path + ".backendRefs")
	}
	if cluster.BackendRef != nil {
		fields[path+".backendRef"] = scheduledForRemoval(path + ".backendRefs")
	}
	if cluster.BackendSettings != nil {
		AddDeprecatedLoadBalancerFields(fields, path+".backendSettings.loadBalancer", cluster.BackendSettings.LoadBalancer)
	}
}

// addDeprecatedPort records the legacy port of the telemetry sink at path, if it is set without the
// legacy host. With host, the host warning already covers the pair.
func addDeprecatedPort(fields map[string]DeprecatedField, path string, host *string, port int32) {
	if host == nil && port != 0 {
		fields[path+".port"] = scheduledForRemoval(path + ".backendRefs")
	}
}

// AddDeprecatedLoadBalancerFields records the deprecated consistent hash fields used in the
// LoadBalancer at path.
func AddDeprecatedLoadBalancerFields(fields map[string]DeprecatedField, path string, lb *egv1a1.LoadBalancer) {
	if lb == nil || lb.ConsistentHash == nil {
		return
	}
	if lb.ConsistentHash.Type == egv1a1.HeaderConsistentHashType {
		fields[path+".consistentHash.type: Header"] = scheduledForRemoval(path + ".consistentHash.type: Headers")
	}
	if lb.ConsistentHash.Header != nil {
		fields[path+".consistentHash.header"] = scheduledForRemoval(path + ".consistentHash.headers")
	}
}
