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
	"k8s.io/utils/ptr"
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
		healthCheck: &ir.HealthCheck{
			PanicThreshold: ptr.To[uint32](66),
		},
	}
	dynamicXdsClusters := buildXdsClusters(args)
	require.Len(t, dynamicXdsClusters, 1)

	require.Equal(t, bootstrapXdsCluster.Name, dynamicXdsClusters[0].Name)
	require.Equal(t, bootstrapXdsCluster.ClusterDiscoveryType, dynamicXdsClusters[0].ClusterDiscoveryType)
	require.Equal(t, bootstrapXdsCluster.TransportSocket, dynamicXdsClusters[0].TransportSocket)
	assert.True(t, proto.Equal(bootstrapXdsCluster.TransportSocket, dynamicXdsClusters[0].TransportSocket))
	assert.True(t, proto.Equal(bootstrapXdsCluster.ConnectTimeout, dynamicXdsClusters[0].ConnectTimeout))
}

func TestBuildXdsClusterLoadAssignment(t *testing.T) {
	bootstrapXdsCluster := getXdsClusterObjFromBootstrap(t)
	ds := &ir.DestinationSetting{
		Name:      bootstrapXdsCluster.Name,
		Endpoints: []*ir.DestinationEndpoint{{Host: envoyGatewayXdsServerHost, Port: bootstrap.DefaultXdsServerPort}},
	}
	dynamicXdsClusterLoadAssignment := buildXdsClusterLoadAssignment(ds)

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
