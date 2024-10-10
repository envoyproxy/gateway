// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	envoyGatewayXdsServerHost = "envoy-gateway"
	xdsClusterName            = "xds_cluster"
)

func TestBuildXdsCluster(t *testing.T) {
	bootstrapXdsCluster := getXdsClusterObjFromBootstrap(t)

	args := &xdsClusterArgs{
		name:         bootstrapXdsCluster.Name,
		tSocket:      bootstrapXdsCluster.TransportSocket,
		endpointType: EndpointTypeDNS,
	}
	dynamicXdsCluster := buildXdsCluster(args)

	require.Equal(t, bootstrapXdsCluster.Name, dynamicXdsCluster.Name)
	require.Equal(t, bootstrapXdsCluster.ClusterDiscoveryType, dynamicXdsCluster.ClusterDiscoveryType)
	require.Equal(t, bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket)
	assert.True(t, proto.Equal(bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket))
	assert.True(t, proto.Equal(bootstrapXdsCluster.ConnectTimeout, dynamicXdsCluster.ConnectTimeout))
}

func TestBuildXdsClusterLoadAssignment(t *testing.T) {
	bootstrapXdsCluster := getXdsClusterObjFromBootstrap(t)
	ds := &ir.DestinationSetting{
		Endpoints: []*ir.DestinationEndpoint{{Host: envoyGatewayXdsServerHost, Port: bootstrap.DefaultXdsServerPort}},
	}
	settings := []*ir.DestinationSetting{ds}
	dynamicXdsClusterLoadAssignment := buildXdsClusterLoadAssignment(bootstrapXdsCluster.Name, settings)

	assert.True(t, proto.Equal(bootstrapXdsCluster.LoadAssignment.Endpoints[0].LbEndpoints[0], dynamicXdsClusterLoadAssignment.Endpoints[0].LbEndpoints[0]))
}

func getXdsClusterObjFromBootstrap(t *testing.T) *clusterv3.Cluster {
	bootstrapObj := &bootstrapv3.Bootstrap{}
	bootstrapStr, err := bootstrap.GetRenderedBootstrapConfig(nil)
	require.NoError(t, err)
	jsonData, err := yaml.YAMLToJSON([]byte(bootstrapStr))
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonData, bootstrapObj)
	require.NoError(t, err)

	for _, cluster := range bootstrapObj.StaticResources.Clusters {
		if cluster.Name == xdsClusterName {
			return cluster
		}
	}

	return nil
}

func TestGetDNSLookupFamily(t *testing.T) {
	tests := []struct {
		name     string
		settings []*ir.DestinationSetting
		expected clusterv3.Cluster_DnsLookupFamily
	}{
		{
			name: "IPv4 only",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "192.0.2.1"}}},
			},
			expected: clusterv3.Cluster_V4_ONLY,
		},
		{
			name: "IPv6 only",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "2001:db8::1"}}},
			},
			expected: clusterv3.Cluster_V6_ONLY,
		},
		{
			name: "Dual stack",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "192.0.2.1"}}},
				{Endpoints: []*ir.DestinationEndpoint{{Host: "2001:db8::1"}}},
			},
			expected: clusterv3.Cluster_ALL,
		},
		{
			name:     "No settings",
			settings: nil,
			expected: clusterv3.Cluster_V4_ONLY,
		},
		{
			name: "FQDN only",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "example.com"}}},
			},
			expected: clusterv3.Cluster_V4_ONLY,
		},
		{
			name: "Mixed IPv4 and FQDN",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "192.0.2.1"}}},
				{Endpoints: []*ir.DestinationEndpoint{{Host: "example.com"}}},
			},
			expected: clusterv3.Cluster_V4_ONLY,
		},
		{
			name: "Mixed IPv6 and FQDN",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "2001:db8::1"}}},
				{Endpoints: []*ir.DestinationEndpoint{{Host: "example.com"}}},
			},
			expected: clusterv3.Cluster_V6_ONLY,
		},
		{
			name: "Mixed IPv4, IPv6 and FQDN",
			settings: []*ir.DestinationSetting{
				{Endpoints: []*ir.DestinationEndpoint{{Host: "192.0.2.1"}}},
				{Endpoints: []*ir.DestinationEndpoint{{Host: "2001:db8::1"}}},
				{Endpoints: []*ir.DestinationEndpoint{{Host: "example.com"}}},
			},
			expected: clusterv3.Cluster_ALL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getDNSLookupFamily(tt.settings)
			assert.Equal(t, tt.expected, result)
		})
	}
}
