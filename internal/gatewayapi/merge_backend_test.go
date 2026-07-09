// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// mergeBackendsAssertions maps each mergebackends- testdata scenario to a scenario-specific
// assertion, checked directly against the live, in-memory *TranslateResult right after
// Translate() returns (i.e. before any YAML marshal/unmarshal round-trip, since
// RouteDestination.BackendClusterRefs is tagged yaml:"-" and can never survive one).
//
// Each assertion compares the resolved Settings[0].Name of two routes' destinations directly,
// rather than counting "backend/..."-prefixed entries in Xds.Backends: a route that is
// correctly excluded from merging still gets its own registered BackendCluster, just under a
// route-scoped (not "backend/...") name - so a same-vs-different name comparison is the only
// reliable way to tell "these two routes share one cluster" from "these two routes each got
// their own cluster", regardless of which naming scheme either one landed on.
var mergeBackendsAssertions = map[string]func(t *testing.T, got *TranslateResult){
	"mergebackends-http-shared-cluster": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		require.Equal(t, destSettingNames(t, got, "http-route-1"), destSettingNames(t, got, "http-route-2"),
			"two routes referencing the identical backend should share one cluster")
	},
	"mergebackends-http-gateway-btp-floor": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		require.Equal(t, destSettingNames(t, got, "http-route-1"), destSettingNames(t, got, "http-route-2"),
			"a uniform gateway-level BackendTrafficPolicy should not block merging")
	},
	"mergebackends-tcp-shared-cluster": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		require.Equal(t, destSettingNames(t, got, "tcproute-1"), destSettingNames(t, got, "tcproute-2"),
			"two TCP routes referencing the identical backend should share one cluster")
	},
	"mergebackends-http-multi-backend-weighted": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		route1Names := destSettingNames(t, got, "http-route-1")
		route2Names := destSettingNames(t, got, "http-route-2")
		require.Len(t, route1Names, 2)
		require.Len(t, route2Names, 1)
		require.Equal(t, route1Names[0], route2Names[0],
			"the service-1 backend shared between the two routes' rules should merge to the same cluster")
		require.NotEqual(t, route1Names[0], route1Names[1],
			"service-1 and service-2 are different backends and must not collapse onto the same cluster")
		require.True(t, strings.HasPrefix(route1Names[1], "backend/"),
			"service-2 should still get an identity-merged name even though only one rule uses it")
	},
	// KNOWN, CONFIRMED GAP (see branch's final-review discussion): http-route-2's route-targeted
	// BackendTrafficPolicy sets a cluster-scoped CircuitBreaker, which should make it ineligible
	// for merging with http-route-1's identical backend - shouldMergeBackend only checks
	// RoutingType divergence today, not other cluster-scoped BTP settings, so this currently
	// fails (both routes incorrectly share one cluster). Kept red on purpose rather than baked
	// into a golden .out.yaml as "correct", until the underlying logic is fixed.
	"mergebackends-http-demerge-route-btp": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		require.NotEqual(t, destSettingNames(t, got, "http-route-1"), destSettingNames(t, got, "http-route-2"),
			"http-route-2's route-targeted CircuitBreaker BackendTrafficPolicy should exclude it from sharing http-route-1's cluster")
	},
	// KNOWN, CONFIRMED GAP: resolveBackendClusterName/getOrCreateBackendCluster have no
	// parameter carrying IsDynamicResolver/IsCustomBackend at all (that's only known later, in
	// processDestination), so nothing today excludes a dynamic-resolver backend from merging
	// with another route referencing the identical Backend CR. Kept red on purpose.
	"mergebackends-http-demerge-dynamic-resolver": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		require.NotEqual(t, destSettingNames(t, got, "http-route-1"), destSettingNames(t, got, "http-route-2"),
			"dynamic-resolver backends must never be merged, even across routes referencing the identical Backend CR")
	},
	// KNOWN, CONFIRMED GAP: the per-backendRef merge loop in route.go has no check for
	// ConsistentHash load balancing (or SessionPersistence) before merging a multi-backend
	// weighted rule's individual backends into shared clusters - splitting them onto separate,
	// independently-named clusters can break the hash's intended "same client -> same backend"
	// guarantee. Kept red on purpose.
	"mergebackends-http-demerge-consistent-hash": func(t *testing.T, got *TranslateResult) {
		t.Helper()
		names := destSettingNames(t, got, "http-route-1")
		require.Len(t, names, 2)
		for _, n := range names {
			require.False(t, strings.HasPrefix(n, "backend/"),
				"backends split across a ConsistentHash-load-balanced rule must not be identity-merged: got %q", n)
		}
	},
}

// destSettingNames returns, in order, the DestinationSetting names of the first route whose IR
// name contains routeNameSubstr, searched across every HTTP and TCP listener in every gateway
// in got.XdsIR. Fails the test if no matching route is found.
func destSettingNames(t *testing.T, got *TranslateResult, routeNameSubstr string) []string {
	t.Helper()
	for _, x := range got.XdsIR {
		for _, l := range x.HTTP {
			for _, r := range l.Routes {
				if strings.Contains(r.Name, routeNameSubstr) && r.Destination != nil {
					names := make([]string, len(r.Destination.Settings))
					for i, s := range r.Destination.Settings {
						names[i] = s.Name
					}
					return names
				}
			}
		}
		for _, l := range x.TCP {
			for _, r := range l.Routes {
				if strings.Contains(r.Name, routeNameSubstr) && r.Destination != nil {
					names := make([]string, len(r.Destination.Settings))
					for i, s := range r.Destination.Settings {
						names[i] = s.Name
					}
					return names
				}
			}
		}
	}
	t.Fatalf("no route matching %q found in translate result", routeNameSubstr)
	return nil
}

func assertMergedBackendClusterCount(t *testing.T, name string, got *TranslateResult) {
	t.Helper()
	assertion, ok := mergeBackendsAssertions[name]
	if !ok {
		return
	}
	assertion(t, got)
}

func TestIrBackendClusterName(t *testing.T) {
	tests := []struct {
		name      string
		kind      string
		namespace string
		bcName    string
		port      int32
		want      string
	}{
		{
			name:      "service with port",
			kind:      "Service",
			namespace: "default",
			bcName:    "service-1",
			port:      8080,
			want:      "backend/service/default/service-1/8080",
		},
		{
			name:      "backend kind",
			kind:      "Backend",
			namespace: "ns",
			bcName:    "be",
			port:      443,
			want:      "backend/backend/ns/be/443",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, irBackendClusterName(tc.kind, tc.namespace, tc.bcName, tc.port))
		})
	}
}

func TestShouldMergeBackend(t *testing.T) {
	gwNN := types.NamespacedName{Namespace: "envoy-gateway", Name: "gateway-1"}
	serviceRT := egv1a1.ServiceRoutingType
	endpointRT := egv1a1.EndpointRoutingType

	tests := []struct {
		name                         string
		mergeEnabled                 bool
		gatewayBaselineRT            *egv1a1.RoutingType
		effectiveRT                  *egv1a1.RoutingType
		hasRouteLevelClusterSettings bool
		want                         bool
	}{
		{
			name:         "disabled globally never merges",
			mergeEnabled: false,
			want:         false,
		},
		{
			name:         "enabled, no routing type anywhere: baseline == effective (both Endpoint)",
			mergeEnabled: true,
			want:         true,
		},
		{
			name:              "enabled, uniform gateway-level routing type: baseline == effective",
			mergeEnabled:      true,
			gatewayBaselineRT: &serviceRT,
			effectiveRT:       &serviceRT,
			want:              true,
		},
		{
			name:              "enabled, route-rule overrides routing type away from gateway baseline: diverges",
			mergeEnabled:      true,
			gatewayBaselineRT: &endpointRT,
			effectiveRT:       &serviceRT,
			want:              false,
		},
		{
			name:                         "enabled, uniform routing but route-level cluster settings present: excluded",
			mergeEnabled:                 true,
			hasRouteLevelClusterSettings: true,
			want:                         false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tr := &Translator{
				MergeBackends: MergeBackendsConfig{Enabled: tc.mergeEnabled},
				TranslatorContext: &TranslatorContext{
					BTPRoutingTypeIndex: &BTPRoutingTypeIndex{
						gatewayLevel: map[btpRoutingKey]*egv1a1.RoutingType{
							{Kind: "Gateway", Namespace: gwNN.Namespace, Name: gwNN.Name}: tc.gatewayBaselineRT,
						},
					},
				},
			}
			got := tr.shouldMergeBackend(gwNN, nil, tc.effectiveRT, tc.hasRouteLevelClusterSettings)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestResolveBackendClusterName(t *testing.T) {
	identity := BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}

	t.Run("nil gatewayCtx never merges", func(t *testing.T) {
		tr := &Translator{MergeBackends: MergeBackendsConfig{Enabled: true}}
		key, name, merge := tr.resolveBackendClusterName("route-scoped-name", identity, nil, nil, false)
		require.False(t, merge)
		require.Equal(t, "route-scoped-name", name)
		require.Equal(t, BackendClusterKey{Name: "route-scoped-name"}, key)
	})

	t.Run("merge disabled falls back to route-scoped name", func(t *testing.T) {
		tr := &Translator{MergeBackends: MergeBackendsConfig{Enabled: false}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		key, name, merge := tr.resolveBackendClusterName("route-scoped-name", identity, gwCtx, nil, false)
		require.False(t, merge)
		require.Equal(t, "route-scoped-name", name)
		require.Equal(t, BackendClusterKey{Name: "route-scoped-name"}, key)
	})

	t.Run("merge enabled resolves to backend-identity name", func(t *testing.T) {
		tr := &Translator{MergeBackends: MergeBackendsConfig{Enabled: true}, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		key, name, merge := tr.resolveBackendClusterName("route-scoped-name", identity, gwCtx, nil, false)
		require.True(t, merge)
		require.Equal(t, "backend/service/default/service-1/8080", name)
		require.Equal(t, identity.Kind, key.Kind)
		require.Equal(t, identity.Name, key.Name)
	})

	t.Run("route-level cluster settings excludes even when routing type matches", func(t *testing.T) {
		tr := &Translator{MergeBackends: MergeBackendsConfig{Enabled: true}, TranslatorContext: &TranslatorContext{}}
		gwCtx := &GatewayContext{Gateway: &gwapiv1.Gateway{}}
		key, name, merge := tr.resolveBackendClusterName("route-scoped-name", identity, gwCtx, nil, true)
		require.False(t, merge)
		require.Equal(t, "route-scoped-name", name)
		require.Equal(t, BackendClusterKey{Name: "route-scoped-name"}, key)
	})
}

func TestGetOrCreateBackendCluster(t *testing.T) {
	key := BackendClusterKey{Kind: "Service", Namespace: "default", Name: "service-1", Port: 8080}
	ds1 := &ir.DestinationSetting{Name: "ds-1"}
	ds2 := &ir.DestinationSetting{Name: "ds-2"}

	t.Run("cache miss creates and registers into gwIR.Backends", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		bc := tr.getOrCreateBackendCluster(gwIR, key, "backend/service/default/service-1/8080", true, ds1, nil)
		require.Len(t, gwIR.Backends, 1)
		require.Same(t, bc, gwIR.Backends[0])
		require.Equal(t, []*ir.DestinationSetting{ds1}, bc.Settings)
	})

	t.Run("cache hit while merge=true does not append the new setting", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		first := tr.getOrCreateBackendCluster(gwIR, key, "backend/service/default/service-1/8080", true, ds1, nil)
		second := tr.getOrCreateBackendCluster(gwIR, key, "backend/service/default/service-1/8080", true, ds2, nil)
		require.Same(t, first, second)
		require.Equal(t, []*ir.DestinationSetting{ds1}, second.Settings)
		require.Len(t, gwIR.Backends, 1)
	})

	t.Run("cache hit while merge=false appends the new setting", func(t *testing.T) {
		tr := &Translator{TranslatorContext: &TranslatorContext{BackendClusterMap: map[BackendClusterKey]*ir.BackendCluster{}}}
		gwIR := &ir.Xds{}
		routeScopedKey := BackendClusterKey{Name: "route-scoped-name"}
		first := tr.getOrCreateBackendCluster(gwIR, routeScopedKey, "route-scoped-name", false, ds1, nil)
		second := tr.getOrCreateBackendCluster(gwIR, routeScopedKey, "route-scoped-name", false, ds2, nil)
		require.Same(t, first, second)
		require.Equal(t, []*ir.DestinationSetting{ds1, ds2}, second.Settings)
	})
}
