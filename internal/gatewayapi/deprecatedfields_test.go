// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func TestDeprecatedFieldsScheduledForRemoval(t *testing.T) {
	backendRef := &gwapiv1.BackendObjectReference{Name: "backend"}
	removed := func(replacement string) status.DeprecatedField {
		return status.DeprecatedField{Replacement: replacement, RemovedIn: "v1.11"}
	}

	t.Run("BackendTrafficPolicy consistent hash header", func(t *testing.T) {
		policy := &egv1a1.BackendTrafficPolicy{}
		policy.Spec.LoadBalancer = &egv1a1.LoadBalancer{
			ConsistentHash: &egv1a1.ConsistentHash{
				Type:   egv1a1.HeaderConsistentHashType,
				Header: &egv1a1.Header{Name: "x-user"},
			},
		}
		require.Equal(t, map[string]status.DeprecatedField{
			"spec.loadBalancer.consistentHash.type: Header": removed("spec.loadBalancer.consistentHash.type: Headers"),
			"spec.loadBalancer.consistentHash.header":       removed("spec.loadBalancer.consistentHash.headers"),
		}, deprecatedFieldsUsedInBackendTrafficPolicy(policy))
	})

	t.Run("EnvoyExtensionPolicy extProc backendRef", func(t *testing.T) {
		policy := &egv1a1.EnvoyExtensionPolicy{}
		policy.Spec.ExtProc = []egv1a1.ExtProc{
			{BackendCluster: egv1a1.BackendCluster{BackendRefs: []egv1a1.BackendRef{{}}}},
			{BackendCluster: egv1a1.BackendCluster{BackendRef: backendRef}},
		}
		require.Equal(t, map[string]status.DeprecatedField{
			"spec.extProc[1].backendRef": removed("spec.extProc[1].backendRefs"),
		}, deprecatedFieldsUsedInEnvoyExtensionPolicy(policy))
	})

	t.Run("SecurityPolicy JWT and OIDC backendRef", func(t *testing.T) {
		policy := &egv1a1.SecurityPolicy{}
		policy.Spec.JWT = &egv1a1.JWT{Providers: []egv1a1.JWTProvider{
			{RemoteJWKS: &egv1a1.RemoteJWKS{BackendCluster: egv1a1.BackendCluster{BackendRef: backendRef}}},
		}}
		policy.Spec.OIDC = &egv1a1.OIDC{Provider: egv1a1.OIDCProvider{BackendCluster: egv1a1.BackendCluster{BackendRef: backendRef}}}
		require.Equal(t, map[string]status.DeprecatedField{
			"spec.jwt.providers[0].remoteJWKS.backendRef": removed("spec.jwt.providers[0].remoteJWKS.backendRefs"),
			"spec.oidc.provider.backendRef":               removed("spec.oidc.provider.backendRefs"),
		}, deprecatedFieldsUsedInSecurityPolicy(policy))
	})

	t.Run("consistent hash header in nested backendSettings", func(t *testing.T) {
		policy := &egv1a1.SecurityPolicy{}
		policy.Spec.ExtAuth = &egv1a1.ExtAuth{GRPC: &egv1a1.GRPCExtAuthService{BackendCluster: egv1a1.BackendCluster{
			BackendRefs: []egv1a1.BackendRef{{}},
			BackendSettings: &egv1a1.BackendSettings{ClusterSettings: egv1a1.ClusterSettings{
				LoadBalancer: &egv1a1.LoadBalancer{ConsistentHash: &egv1a1.ConsistentHash{
					Type:   egv1a1.HeaderConsistentHashType,
					Header: &egv1a1.Header{Name: "x-user"},
				}},
			}},
		}}}
		require.Equal(t, map[string]status.DeprecatedField{
			"spec.extAuth.grpc.backendSettings.loadBalancer.consistentHash.type: Header": removed(
				"spec.extAuth.grpc.backendSettings.loadBalancer.consistentHash.type: Headers"),
			"spec.extAuth.grpc.backendSettings.loadBalancer.consistentHash.header": removed(
				"spec.extAuth.grpc.backendSettings.loadBalancer.consistentHash.headers"),
		}, deprecatedFieldsUsedInSecurityPolicy(policy))
	})

	t.Run("EnvoyProxy telemetry sinks", func(t *testing.T) {
		host := "otel-collector.monitoring.svc"
		ep := &egv1a1.EnvoyProxy{}
		ep.Spec.Telemetry = &egv1a1.ProxyTelemetry{
			Tracing: &egv1a1.ProxyTracing{Provider: egv1a1.TracingProvider{BackendCluster: egv1a1.BackendCluster{
				BackendSettings: &egv1a1.BackendSettings{ClusterSettings: egv1a1.ClusterSettings{
					LoadBalancer: &egv1a1.LoadBalancer{ConsistentHash: &egv1a1.ConsistentHash{Type: egv1a1.HeaderConsistentHashType}},
				}},
			}}},
			Metrics: &egv1a1.ProxyMetrics{Sinks: []egv1a1.ProxyMetricSink{
				{OpenTelemetry: &egv1a1.ProxyOpenTelemetrySink{Host: &host, Port: 4317}},
			}},
			AccessLog: &egv1a1.ProxyAccessLog{Settings: []egv1a1.ProxyAccessLogSetting{
				{Sinks: []egv1a1.ProxyAccessLogSink{
					{ALS: &egv1a1.ALSEnvoyProxyAccessLog{BackendCluster: egv1a1.BackendCluster{BackendRef: backendRef}}},
					{OpenTelemetry: &egv1a1.OpenTelemetryEnvoyProxyAccessLog{Host: &host, Port: 4317}},
				}},
			}},
		}
		require.Equal(t, map[string]status.DeprecatedField{
			"spec.telemetry.tracing.provider.backendSettings.loadBalancer.consistentHash.type: Header": removed(
				"spec.telemetry.tracing.provider.backendSettings.loadBalancer.consistentHash.type: Headers"),
			"spec.telemetry.metrics.sinks[0].openTelemetry.host":           removed("spec.telemetry.metrics.sinks[0].openTelemetry.backendRefs"),
			"spec.telemetry.accessLog.settings[0].sinks[0].als.backendRef": removed("spec.telemetry.accessLog.settings[0].sinks[0].als.backendRefs"),
			"spec.telemetry.accessLog.settings[0].sinks[1].openTelemetry.host": removed(
				"spec.telemetry.accessLog.settings[0].sinks[1].openTelemetry.backendRefs"),
		}, deprecatedFieldsUsedInEnvoyProxy(ep))
	})
}
