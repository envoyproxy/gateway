// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

func svcSetting(name, port string) *ir.DestinationSetting {
	return &ir.DestinationSetting{
		Name:      name,
		Weight:    new(uint32(1)),
		Endpoints: []*ir.DestinationEndpoint{{Host: "1.2.3.4", Port: 8080}},
		Metadata: &ir.ResourceMetadata{
			Kind:        "Service",
			Namespace:   "default",
			Name:        "service-1",
			SectionName: port,
		},
	}
}

func TestMergedBackendClusterName(t *testing.T) {
	tests := []struct {
		name string
		s    *ir.DestinationSetting
		want string
		ok   bool
	}{
		{
			name: "service with port",
			s:    svcSetting("route-scoped", "8080"),
			want: "backend/service/default/service-1/8080",
			ok:   true,
		},
		{
			name: "protocol is part of the identity",
			s: &ir.DestinationSetting{
				Protocol: ir.GRPC,
				Metadata: &ir.ResourceMetadata{Kind: "Service", Namespace: "default", Name: "service-1", SectionName: "8080"},
			},
			want: "backend/service/default/service-1/8080/grpc",
			ok:   true,
		},
		{
			name: "service routing is part of the identity",
			s: &ir.DestinationSetting{
				ServiceRouting: true,
				Metadata:       &ir.ResourceMetadata{Kind: "Service", Namespace: "default", Name: "service-1", SectionName: "8080"},
			},
			want: "backend/service/default/service-1/8080/serviceip",
			ok:   true,
		},
		{
			name: "credential injection filter not mergeable",
			s: &ir.DestinationSetting{
				Filters:  &ir.DestinationFilters{CredentialInjection: &ir.CredentialInjection{}},
				Metadata: &ir.ResourceMetadata{Kind: "Service", Namespace: "default", Name: "service-1", SectionName: "8080"},
			},
			ok: false,
		},
		{
			name: "header filters remain mergeable",
			s: &ir.DestinationSetting{
				Filters:  &ir.DestinationFilters{RemoveRequestHeaders: []string{"x-foo"}},
				Metadata: &ir.ResourceMetadata{Kind: "Service", Namespace: "default", Name: "service-1", SectionName: "8080"},
			},
			want: "backend/service/default/service-1/8080",
			ok:   true,
		},
		{
			name: "backend without section",
			s: &ir.DestinationSetting{
				Metadata: &ir.ResourceMetadata{Kind: "Backend", Namespace: "ns", Name: "be"},
			},
			want: "backend/backend/ns/be",
			ok:   true,
		},
		{
			name: "dynamic resolver not mergeable",
			s:    &ir.DestinationSetting{IsDynamicResolver: true, Metadata: &ir.ResourceMetadata{Kind: "Backend", Name: "be"}},
			ok:   false,
		},
		{
			name: "custom backend not mergeable",
			s:    &ir.DestinationSetting{IsCustomBackend: true, Metadata: &ir.ResourceMetadata{Kind: "Foo", Name: "be"}},
			ok:   false,
		},
		{
			name: "missing metadata not mergeable",
			s:    &ir.DestinationSetting{Name: "x"},
			ok:   false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := mergedBackendClusterName(tc.s)
			require.Equal(t, tc.ok, ok)
			if tc.ok {
				require.Equal(t, tc.want, got)
			}
		})
	}
}

func TestMergeRouteDestination(t *testing.T) {
	sharedName := "backend/service/default/service-1/8080"

	t.Run("single backend merges", func(t *testing.T) {
		d := &ir.RouteDestination{Name: "httproute/default/r/rule/0", Settings: []*ir.DestinationSetting{svcSetting("s0", "8080")}}
		mergeRouteDestination(d, false, true)
		require.True(t, d.Settings[0].Merged)
		require.Equal(t, sharedName, d.Settings[0].Name)
	})

	t.Run("route-level cluster settings de-merge", func(t *testing.T) {
		d := &ir.RouteDestination{
			Name:                      "httproute/default/r/rule/0",
			RouteLevelClusterSettings: true,
			Settings:                  []*ir.DestinationSetting{svcSetting("s0", "8080")},
		}
		mergeRouteDestination(d, false, true)
		require.False(t, d.Settings[0].Merged)
		require.Equal(t, "s0", d.Settings[0].Name)
	})

	t.Run("multi-backend weighted merges each backend", func(t *testing.T) {
		d := &ir.RouteDestination{Settings: []*ir.DestinationSetting{
			svcSetting("s0", "8080"),
			svcSetting("s1", "8443"),
		}}
		d.Settings[1].Metadata.Name = "service-2"
		mergeRouteDestination(d, false, true)
		require.True(t, d.Settings[0].Merged)
		require.True(t, d.Settings[1].Merged)
	})

	t.Run("multi-backend split-incompatible does not merge", func(t *testing.T) {
		d := &ir.RouteDestination{Settings: []*ir.DestinationSetting{
			svcSetting("s0", "8080"),
			svcSetting("s1", "8443"),
		}}
		mergeRouteDestination(d, true /* splitIncompatible */, true)
		require.False(t, d.Settings[0].Merged)
		require.False(t, d.Settings[1].Merged)
	})

	t.Run("multi-backend not merged for tcp/udp (no weighted)", func(t *testing.T) {
		d := &ir.RouteDestination{Settings: []*ir.DestinationSetting{
			svcSetting("s0", "8080"),
			svcSetting("s1", "8443"),
		}}
		mergeRouteDestination(d, false, false /* allowWeightedMerge */)
		require.False(t, d.Settings[0].Merged)
	})

	t.Run("multi-backend with priority failover not merged", func(t *testing.T) {
		d := &ir.RouteDestination{Settings: []*ir.DestinationSetting{
			svcSetting("s0", "8080"),
			svcSetting("s1", "8443"),
		}}
		d.Settings[1].Metadata.Name = "service-2"
		d.Settings[1].Priority = new(uint32(1)) // fallback backend
		mergeRouteDestination(d, false, true)
		require.False(t, d.Settings[0].Merged)
		require.False(t, d.Settings[1].Merged)
	})
}

func TestIsMergeBackendsEnabled(t *testing.T) {
	be := egv1a1.MergeBackendsModeBestEffort
	tests := []struct {
		name string
		res  *resource.Resources
		want bool
	}{
		{
			name: "gatewayclass envoyproxy set",
			res: &resource.Resources{
				EnvoyProxyForGatewayClass: &egv1a1.EnvoyProxy{Spec: egv1a1.EnvoyProxySpec{MergeBackends: &be}},
			},
			want: true,
		},
		{
			name: "default spec set",
			res:  &resource.Resources{EnvoyProxyDefaultSpec: &egv1a1.EnvoyProxySpec{MergeBackends: &be}},
			want: true,
		},
		{
			name: "unset",
			res:  &resource.Resources{},
			want: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, IsMergeBackendsEnabled(tc.res))
		})
	}
}
