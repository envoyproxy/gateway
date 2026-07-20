// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestProcessSDSClustersIncludesListenerCertificates(t *testing.T) {
	const sdsURL = "/var/run/sds.sock"

	xdsIR := &ir.Xds{
		HTTP: []*ir.HTTPListener{{
			TLS: &ir.TLSConfig{Certificates: []ir.TLSCertificate{{SDS: &ir.SDSConfig{URL: sdsURL}}}},
		}},
		TCP: []*ir.TCPListener{{
			TLS: &ir.TLSConfig{Certificates: []ir.TLSCertificate{{SDS: &ir.SDSConfig{URL: sdsURL}}}},
		}},
	}
	tCtx := &types.ResourceVersionTable{}

	require.NoError(t, processSDSClusters(tCtx, xdsIR))
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)

	sdsCluster, ok := tCtx.XdsResources[resourcev3.ClusterType][0].(*cluster.Cluster)
	require.True(t, ok)
	require.Equal(t, sdsClusterNameFromURL(sdsURL), sdsCluster.Name)
	require.Equal(t, sdsURL, sdsCluster.LoadAssignment.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetPipe().Path)
}
