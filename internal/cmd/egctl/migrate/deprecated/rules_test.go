// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package deprecated

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v3"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/envoyproxy/gateway/internal/cmd/egctl/migrate"
)

func TestTelemetryHostEmptyBackendRefs(t *testing.T) {
	for _, refs := range []string{"[]", "null"} {
		t.Run(refs, func(t *testing.T) {
			sink := parse(t, "host: collector.monitoring.svc\nport: 4317\nbackendRefs: "+refs)
			migrateTelemetryHost(sink, "sink", "cluster.local", nil)
			data, err := yaml.Marshal(sink)
			require.NoError(t, err)
			var got map[string]any
			// Decoding into a map rejects duplicate keys.
			require.NoError(t, yaml.Unmarshal(data, &got))
			require.Equal(t, map[string]any{
				"backendRefs": []any{map[string]any{"name": "collector", "namespace": "monitoring", "port": 4317}},
			}, got)
		})
	}
}

func TestTelemetryHostClusterDomain(t *testing.T) {
	tests := []struct {
		host, domain string
		migrated     bool
	}{
		{"collector.logging.svc", "cluster.local", true},
		{"collector.logging.svc.", "cluster.local", true},
		{"collector.logging.svc.cluster.local", "cluster.local", true},
		{"collector.logging.svc.cluster.local.", "cluster.local", true},
		{"collector.logging.svc.corp.internal", "corp.internal", true},
		// An external name under .svc must not be redirected to a Service.
		{"collector.logging.svc.example.com.", "cluster.local", false},
		{"collector.logging.svc.cluster.local.example.com", "cluster.local", false},
		{"collector.logging.svc.cluster.local", "corp.internal", false},
		{"collector.example.com", "cluster.local", false},
	}
	for _, tt := range tests {
		t.Run(tt.host+"/"+tt.domain, func(t *testing.T) {
			sink := parse(t, "host: "+tt.host+"\nport: 4317\n")
			f := migrateTelemetryHost(sink, "sink", tt.domain, nil)
			if tt.migrated {
				require.Equal(t, migrate.ResultMigrated, f.Result, f.Note)
				require.True(t, migrate.Has(sink, "backendRefs"))
				return
			}
			require.Equal(t, migrate.ResultError, f.Result, f.Note)
			require.Equal(t, tt.host, migrate.Scalar(migrate.Get(sink, "host")))
			require.False(t, migrate.Has(sink, "backendRefs"))
		})
	}
}

func TestTelemetryHostServiceType(t *testing.T) {
	services := map[string]*corev1.Service{
		"monitoring/collector": {Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeClusterIP}},
		"monitoring/external":  {Spec: corev1.ServiceSpec{Type: corev1.ServiceTypeExternalName, ExternalName: "otel.example.com"}},
	}
	get := func(ns, name string) (*corev1.Service, error) {
		if name == "broken" {
			return nil, errors.New("forbidden")
		}
		if svc, ok := services[ns+"/"+name]; ok {
			return svc, nil
		}
		return nil, kerrors.NewNotFound(corev1.Resource("services"), name)
	}
	tests := []struct {
		host, note string
		getter     serviceGetter
		migrated   bool
	}{
		{host: "collector.monitoring.svc", getter: get, migrated: true},
		{host: "external.monitoring.svc.cluster.local", getter: get, note: "ExternalName Service monitoring/external"},
		{host: "missing.monitoring.svc", getter: get, note: "doesn't exist"},
		{host: "broken.monitoring.svc", getter: get, note: "can't be read"},
		// Without a cluster, the Service type is unknown, so the note asks the user to check it.
		{host: "external.monitoring.svc", migrated: true, note: "isn't of type ExternalName"},
	}
	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			sink := parse(t, "host: "+tt.host+"\nport: 4317\n")
			f := migrateTelemetryHost(sink, "sink", "cluster.local", tt.getter)
			require.Contains(t, f.Note, tt.note)
			if tt.migrated {
				require.Equal(t, migrate.ResultMigrated, f.Result, f.Note)
				require.True(t, migrate.Has(sink, "backendRefs"))
				return
			}
			require.Equal(t, migrate.ResultError, f.Result, f.Note)
			require.Equal(t, tt.host, migrate.Scalar(migrate.Get(sink, "host")))
			require.False(t, migrate.Has(sink, "backendRefs"))
		})
	}
	// A Service read from the cluster has no caveat.
	sink := parse(t, "host: collector.monitoring.svc\n")
	require.NotContains(t, migrateTelemetryHost(sink, "sink", "cluster.local", get).Note, "ExternalName")
}

func TestDeprecatedBooleanAliases(t *testing.T) {
	tests := []struct {
		field, replacement, value string
		rewrite                   func(*yaml.Node, string) migrate.Finding
	}{
		{"enableProxyProtocol", "proxyProtocol", "{}", migrateProxyProtocol},
		{"preserveXRequestID", "requestID", "PreserveOrGenerate", migrateXRequestID},
		{"optional", "mode", "VerifyIfGiven", migrateClientValidation},
	}
	for _, tt := range tests {
		for _, enabled := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/%t", tt.field, enabled), func(t *testing.T) {
				root := parse(t, fmt.Sprintf("anchor: &enabled %t\n%s: *enabled\n", enabled, tt.field))
				tt.rewrite(root, "spec")
				require.False(t, migrate.Has(root, tt.field))
				if enabled {
					want := parse(t, "value: "+tt.value)
					require.True(t, migrate.Equal(migrate.Get(want, "value"), migrate.Get(root, tt.replacement)))
				} else {
					require.False(t, migrate.Has(root, tt.replacement))
				}
			})
		}
	}
}

func TestCompressionNullSettings(t *testing.T) {
	for typ, block := range compressorBlocks {
		for _, settings := range []string{"null", "*unset"} {
			t.Run(typ+"/"+settings, func(t *testing.T) {
				spec := parse(t, fmt.Sprintf("anchor: &unset null\ncompression:\n- type: %s\n  %s: %s\n", typ, block, settings))
				migrateCompression(spec, "spec")
				require.False(t, migrate.Has(spec, "compression"))
				entries := migrate.Get(spec, "compressor")
				require.Len(t, entries.Content, 1)
				settings := migrate.Get(entries.Content[0], block)
				require.NotNil(t, settings)
				require.Equal(t, yaml.MappingNode, settings.Kind)
				require.Empty(t, settings.Content)
			})
		}
	}
}

func TestExternalAuthBackendRefStillMigrates(t *testing.T) {
	for _, protocol := range []string{"http", "grpc"} {
		t.Run(protocol, func(t *testing.T) {
			root := parse(t, "kind: SecurityPolicy\nspec:\n  extAuth:\n    "+protocol+":\n      backendRef: {name: auth, port: 9002}\n")
			findings, err := (&Path{ClusterDomain: "cluster.local"}).migrateEnvoyGatewayResource(&migrate.Object{Kind: "SecurityPolicy", Root: root})
			require.NoError(t, err)
			require.Len(t, findings, 1)
			cluster := migrate.Lookup(root, "spec.extAuth."+protocol)
			require.False(t, migrate.Has(cluster, "backendRef"))
			require.Equal(t, "auth", migrate.Scalar(migrate.Lookup(cluster, "backendRefs[0].name")))
		})
	}
}
