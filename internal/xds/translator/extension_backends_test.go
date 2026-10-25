// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"
	"testing"
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExtensionBackendMetadata(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/extension-backends.yaml")
	tr := &Translator{}
	table, err := tr.Translate(t.Context(), x)
	require.NoError(t, err)
	require.Len(t, table.XdsResources[resourcev3.ClusterType], 3)
	require.Len(t, table.XdsResources[resourcev3.SecretType], 2)
	routes := table.XdsResources[resourcev3.RouteType][0].(*routev3.RouteConfiguration).VirtualHosts[0].Routes
	require.Len(t, routes, 2)
	for i, route := range routes {
		bindings := route.Metadata.FilterMetadata[extensionBackendsMetadataNamespace]
		require.Equal(t, x.HTTP[0].Routes[i].EnvoyExtensions.Backends["resolver"], bindings.Fields["resolver"].GetStringValue())
		for slot := range x.HTTP[0].Routes[i].EnvoyExtensions.Luas {
			require.Equal(t, bindings, route.Metadata.FilterMetadata[luaFilterName(slot)])
		}
	}
	require.NotEqual(t, routes[0].Metadata.FilterMetadata[extensionBackendsMetadataNamespace], routes[1].Metadata.FilterMetadata[extensionBackendsMetadataNamespace])
	require.Greater(t, len(routes[0].Metadata.FilterMetadata), 3, "resource metadata must be preserved")

	// No bindings must leave existing metadata untouched, including nil metadata.
	for _, metadata := range []*corev3.Metadata{nil, {FilterMetadata: map[string]*structpb.Struct{"existing": {}}}} {
		route := &routev3.Route{Metadata: metadata}
		patchExtensionBackendMetadata(route, &ir.HTTPRoute{})
		require.Same(t, metadata, route.Metadata)
	}
}

func TestExtensionBackendSDS(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/extension-backends.yaml")
	tls := x.ExtensionBackends[1].Settings[0].TLS
	tls.CACertificate = &ir.TLSCACertificate{SDS: &ir.SDSConfig{Scheme: "unix", Address: "/run/ca.sock", SecretName: "ca"}}
	tls.ClientCertificates = []ir.TLSCertificate{{SDS: &ir.SDSConfig{Scheme: "unix", Address: "/run/client.sock", SecretName: "client"}}}
	tr := &Translator{}
	table, err := tr.Translate(t.Context(), x)
	require.NoError(t, err)
	require.Len(t, table.XdsResources[resourcev3.ClusterType], 5)
	require.Empty(t, table.XdsResources[resourcev3.SecretType])
	require.NotNil(t, findXdsCluster(table, sdsClusterNameFromURL("unix:///run/ca.sock")))
	require.NotNil(t, findXdsCluster(table, sdsClusterNameFromURL("unix:///run/client.sock")))
}

func TestExtensionBackendSettingsAndUDS(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/extension-backends.yaml")
	backend := x.ExtensionBackends[0]
	backend.Traffic = &ir.TrafficFeatures{ClusterTrafficFeatures: ir.ClusterTrafficFeatures{
		LoadBalancer:   &ir.LoadBalancer{RoundRobin: &ir.RoundRobin{}},
		Timeout:        &ir.Timeout{TCP: &ir.TCPTimeout{ConnectTimeout: ir.MetaV1DurationPtr(3 * time.Second)}},
		CircuitBreaker: &ir.CircuitBreaker{MaxConnections: new(uint32(17))},
	}}
	backend.Settings[0].AddressType = new(ir.UDS)
	backend.Settings[0].Endpoints = []*ir.DestinationEndpoint{{Path: new("/run/callout.sock")}}
	table, err := (&Translator{}).Translate(t.Context(), x)
	require.NoError(t, err)
	cluster := findXdsCluster(table, backend.Name)
	require.NotNil(t, cluster)
	require.Equal(t, clusterv3.Cluster_ROUND_ROBIN, cluster.LbPolicy)
	require.Equal(t, 3*time.Second, cluster.ConnectTimeout.AsDuration())
	require.EqualValues(t, 17, cluster.CircuitBreakers.Thresholds[0].MaxConnections.Value)
	require.Equal(t, "/run/callout.sock", findXdsEndpoint(table, backend.Name).Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetPipe().Path)
}

func TestExtensionBackendSharedResources(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/extension-backends.yaml")
	// Both routes map the aliases resolver and lookup to the same backend.
	name := x.ExtensionBackends[1].Name
	x.ExtensionBackends = x.ExtensionBackends[1:]
	for _, route := range x.HTTP[0].Routes {
		route.EnvoyExtensions.Backends = map[string]string{"resolver": name, "lookup": name}
	}
	table, err := (&Translator{}).Translate(t.Context(), x.DeepCopy())
	require.NoError(t, err)
	require.Len(t, table.XdsResources[resourcev3.ClusterType], 2, "one route backend and one shared extension backend")
	require.Len(t, table.XdsResources[resourcev3.SecretType], 2)
	routes := table.XdsResources[resourcev3.RouteType][0].(*routev3.RouteConfiguration).VirtualHosts[0].Routes
	for _, route := range routes {
		bindings := route.Metadata.FilterMetadata[extensionBackendsMetadataNamespace]
		require.Equal(t, name, bindings.Fields["resolver"].GetStringValue())
		require.Equal(t, name, bindings.Fields["lookup"].GetStringValue())
	}
}

func TestExtensionBackendResourceOrder(t *testing.T) {
	x := requireXdsIRFromInputTestData(t, "testdata/in/xds-ir/extension-backends.yaml")
	first, err := (&Translator{}).Translate(t.Context(), x)
	require.NoError(t, err)
	slices.Reverse(x.ExtensionBackends)
	second, err := (&Translator{}).Translate(t.Context(), x)
	require.NoError(t, err)
	require.Equal(t, first.XdsResources, second.XdsResources)
	require.Contains(t, x.ExtensionBackends[0].Name, "/two/", "translation must not sort the input in place")
}
