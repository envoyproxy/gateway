// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package message_test

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestXdsIRDeepCopyOwnership(t *testing.T) {
	setting := &ir.DestinationSetting{
		Name: "backend", Weight: new(uint32(1)),
		Endpoints: []*ir.DestinationEndpoint{{Host: "192.0.2.1", Port: 8080, Zone: new("zone-a")}},
		TLS:       &ir.TLSUpstreamConfig{SNI: new("backend.example")},
		Filters:   &ir.DestinationFilters{RemoveRequestHeaders: []string{"x-private"}},
	}
	equalSetting := setting.DeepCopy()
	route := &ir.HTTPRoute{
		Name: "route",
		Destination: &ir.RouteDestination{
			Name: "destination", Settings: []*ir.DestinationSetting{setting, nil, setting, equalSetting},
			BackendClusterRefs: []*ir.BackendClusterRef{{Name: "merged", Weight: new(uint32(2))}},
		},
		RemoveRequestHeaders: []string{"x-route-private"},
	}
	in := &message.XdsIRWithContext{
		Context: context.Background(), StoredAt: time.Now(),
		XdsIR: &ir.Xds{
			HTTP: []*ir.HTTPListener{
				{Routes: []*ir.HTTPRoute{route}},
				{Routes: []*ir.HTTPRoute{route}},
			},
			EnvoyPatchPolicies: []*ir.EnvoyPatchPolicy{{EnvoyPatchPolicyStatus: ir.EnvoyPatchPolicyStatus{
				Status: &gwapiv1.PolicyStatus{},
			}}},
			BackendClusters: []*ir.BackendCluster{{Name: "merged", Setting: setting}},
		},
	}
	want := in.XdsIR.DeepCopy()
	a, b := in.DeepCopy(), in.DeepCopy()
	require.Equal(t, want, a.XdsIR)
	require.Equal(t, in.Context, a.Context)
	require.Equal(t, in.StoredAt, a.StoredAt)
	copied := a.XdsIR.HTTP[0].Routes[0].Destination.Settings
	require.Same(t, copied[0], copied[2])
	require.Same(t, copied[0], a.XdsIR.HTTP[1].Routes[0].Destination.Settings[0])
	require.NotSame(t, copied[0], copied[3], "equal but distinct inputs must not be merged")

	// Mutating one snapshot must not affect its source or another snapshot.
	copied[0].Endpoints[0].Host = "192.0.2.2"
	*copied[0].Endpoints[0].Zone = "zone-b"
	*copied[0].Weight = 10
	*copied[0].TLS.SNI = "changed.example"
	copied[0].Filters.RemoveRequestHeaders[0] = "x-changed"
	a.XdsIR.HTTP[0].Routes[0].RemoveRequestHeaders[0] = "x-changed"
	*a.XdsIR.HTTP[0].Routes[0].Destination.BackendClusterRefs[0].Weight = 3
	a.XdsIR.EnvoyPatchPolicies[0].Status.Ancestors = []gwapiv1.PolicyAncestorStatus{{}}
	a.XdsIR.BackendClusters[0].Setting.Name = "changed"
	require.Equal(t, want, in.XdsIR)
	require.Equal(t, want, b.XdsIR)
	require.Equal(t, "x-route-private", a.XdsIR.HTTP[1].Routes[0].RemoveRequestHeaders[0])
	require.Equal(t, equalSetting, copied[3])

	setting.Endpoints[0].Host = "192.0.2.3"
	require.Equal(t, want, b.XdsIR)
}

func TestXdsIRDeepCopyEmptyValues(t *testing.T) {
	shared := &ir.DestinationSetting{}
	for name, in := range map[string]*message.XdsIRWithContext{
		"nil":      nil,
		"nil IR":   {},
		"empty IR": {XdsIR: &ir.Xds{}},
		"empty HTTP": {XdsIR: &ir.Xds{
			HTTP: []*ir.HTTPListener{},
		}},
		"nested empty values": {XdsIR: &ir.Xds{HTTP: []*ir.HTTPListener{
			nil, {}, {Routes: []*ir.HTTPRoute{}}, {Routes: []*ir.HTTPRoute{
				nil,
				{},
				{Destination: &ir.RouteDestination{}},
				{Destination: &ir.RouteDestination{Settings: []*ir.DestinationSetting{}}},
				{Destination: &ir.RouteDestination{Settings: []*ir.DestinationSetting{nil, shared, shared}}},
			}},
		}}},
	} {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, in, in.DeepCopy())
		})
	}
}

func BenchmarkXdsIRDeepCopy(b *testing.B) {
	for _, tc := range []struct {
		name                                  string
		routes, matches, listeners, endpoints int
	}{
		{"no-destinations", 1000, 1, 1, 0},
		{"small-unique", 10, 1, 1, 1},
		{"unique-small-endpoints", 1000, 1, 1, 1},
		{"unique", 1000, 1, 1, 10},
		{"listeners-only", 1000, 1, 2, 10},
		{"matches-only", 1000, 4, 1, 10},
		{"shared-small", 1000, 4, 2, 1},
		{"shared-medium", 1000, 4, 2, 10},
		{"shared-large", 1000, 4, 2, 100},
	} {
		b.Run(tc.name, func(b *testing.B) {
			settings := make([]*ir.DestinationSetting, tc.routes)
			for i := range settings {
				settings[i] = &ir.DestinationSetting{Name: fmt.Sprintf("backend-%d", i), Weight: new(uint32(1))}
				for j := range tc.endpoints {
					settings[i].Endpoints = append(settings[i].Endpoints, &ir.DestinationEndpoint{
						Host: fmt.Sprintf("192.0.2.%d", j+1), Port: 8080, Zone: new("zone-a"),
					})
				}
			}
			value := &message.XdsIRWithContext{XdsIR: &ir.Xds{}}
			for range tc.listeners {
				listener := &ir.HTTPListener{}
				for i, setting := range settings {
					for j := range tc.matches {
						route := &ir.HTTPRoute{
							Name:        fmt.Sprintf("route-%d-match-%d", i, j),
							Destination: &ir.RouteDestination{Name: setting.Name, Settings: []*ir.DestinationSetting{setting}},
						}
						if tc.endpoints == 0 {
							route.Destination = nil
						}
						listener.Routes = append(listener.Routes, route)
					}
				}
				value.XdsIR.HTTP = append(value.XdsIR.HTTP, listener)
			}
			b.ReportAllocs()
			for b.Loop() {
				runtime.KeepAlive(value.DeepCopy())
			}
		})
	}
}
