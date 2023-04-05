// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"sigs.k8s.io/yaml"
)

func TestBuildXdsCluster(t *testing.T) {
	bootstrapObj := getBootstrapObj(t)
	var bootstrapXdsCluster *clusterv3.Cluster
	for _, cluster := range bootstrapObj.StaticResources.Clusters {
		if cluster.Name == "xds_cluster" {
			bootstrapXdsCluster = cluster
			break
		}
	}
	dynamicXdsCluster := buildXdsCluster(bootstrapXdsCluster.Name, bootstrapXdsCluster.TransportSocket, HTTP2, DefaultEndpointType)
	require.Equal(t, bootstrapXdsCluster.Name, dynamicXdsCluster.Name)
	require.Equal(t, bootstrapXdsCluster.ClusterDiscoveryType, dynamicXdsCluster.ClusterDiscoveryType)
	require.Equal(t, bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket)
	assert.True(t, proto.Equal(bootstrapXdsCluster.ConnectTimeout, dynamicXdsCluster.ConnectTimeout))
	assert.True(t, proto.Equal(bootstrapXdsCluster.TypedExtensionProtocolOptions[extensionOptionKey], dynamicXdsCluster.TypedExtensionProtocolOptions[extensionOptionKey]))
}

func getBootstrapObj(t *testing.T) *bootstrapv3.Bootstrap {
	bootstrapObj := &bootstrapv3.Bootstrap{}
	bootstrapStr, err := bootstrap.GetRenderedBootstrapConfig()
	require.NoError(t, err)
	jsonData, err := yaml.YAMLToJSON([]byte(bootstrapStr))
	require.NoError(t, err)
	err = protojson.Unmarshal(jsonData, bootstrapObj)
	require.NoError(t, err)

	return bootstrapObj
}
