// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strings"
	"testing"
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

func TestSDSClusterNameFromURLIncludesReadableUnixPathAndHash(t *testing.T) {
	require.Equal(t, "sds_var_run_secrets_workload-spiffe-uds_socket_2465711674b623c4f815575b37404800", ir.SDSClusterNameFromURL("unix:///var/run/secrets/workload-spiffe-uds/socket"))
}

func TestProcessSDSClustersCreatesDistinctClustersForAmbiguousUnixPaths(t *testing.T) {
	tCtx := &types.ResourceVersionTable{}
	xdsIR := xdsWithHTTPSDSURLs("unix:///run/a/b/socket", "unix:///run/a_b/socket")

	err := processSDSClusters(tCtx, xdsIR)

	require.NoError(t, err)
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 2)
	require.NotEqual(t, ir.SDSClusterNameFromURL("unix:///run/a/b/socket"), ir.SDSClusterNameFromURL("unix:///run/a_b/socket"))
}

func TestProcessSDSClustersDeduplicatesSameURLAcrossHTTPAndTCP(t *testing.T) {
	sdsURL := "unix:///var/run/secrets/workload-spiffe-uds/socket"
	tCtx := &types.ResourceVersionTable{}
	xdsIR := xdsWithHTTPSDSURLs(sdsURL)
	xdsIR.TCP = []*ir.TCPListener{{
		Routes: []*ir.TCPRoute{{
			TLS: &ir.TLS{Terminate: tlsConfigWithSDSURLs(sdsURL)},
		}},
	}}

	err := processSDSClusters(tCtx, xdsIR)

	require.NoError(t, err)
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func TestProcessSDSClustersOrdersClustersByURL(t *testing.T) {
	tCtx := &types.ResourceVersionTable{}
	xdsIR := xdsWithHTTPSDSURLs(
		"unix:///var/run/sds/z.sock",
		"unix:///var/run/sds/a.sock",
		"unix:///var/run/sds/m.sock",
	)

	err := processSDSClusters(tCtx, xdsIR)

	require.NoError(t, err)
	clusters := tCtx.XdsResources[resourcev3.ClusterType]
	clusterA := clusters[0].(*cluster.Cluster)
	clusterM := clusters[1].(*cluster.Cluster)
	clusterZ := clusters[2].(*cluster.Cluster)
	require.Equal(t, []string{
		ir.SDSClusterNameFromURL("unix:///var/run/sds/a.sock"),
		ir.SDSClusterNameFromURL("unix:///var/run/sds/m.sock"),
		ir.SDSClusterNameFromURL("unix:///var/run/sds/z.sock"),
	}, []string{
		clusterA.GetName(),
		clusterM.GetName(),
		clusterZ.GetName(),
	})
}

func TestProcessSDSClustersReturnsCollisionForPreexistingNonSDSCluster(t *testing.T) {
	sdsURL := "unix:///var/run/sds/socket"
	clusterName := ir.SDSClusterNameFromURL(sdsURL)
	tCtx := &types.ResourceVersionTable{}
	tCtx.XdsResources = types.XdsResources{
		resourcev3.ClusterType: {&cluster.Cluster{Name: clusterName}},
	}

	err := processSDSClusters(tCtx, xdsWithHTTPSDSURLs(sdsURL))

	require.EqualError(t, err, `SDS cluster "`+clusterName+`" conflicts with an existing cluster`)
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func TestProcessSDSClustersAcceptsPreexistingCanonicalSDSCluster(t *testing.T) {
	sdsURL := "unix:///var/run/sds/socket"
	tCtx := &types.ResourceVersionTable{}
	require.NoError(t, createSDSCluster(tCtx, sdsURL))

	err := processSDSClusters(tCtx, xdsWithHTTPSDSURLs(sdsURL))

	require.NoError(t, err)
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func TestProcessSDSClustersReturnsCollisionWhenPreexistingClusterMissesHTTP2Options(t *testing.T) {
	sdsURL := "unix:///var/run/sds/socket"
	tCtx := &types.ResourceVersionTable{}
	require.NoError(t, createSDSCluster(tCtx, sdsURL))
	existing := findXdsCluster(tCtx, ir.SDSClusterNameFromURL(sdsURL))
	require.NotNil(t, existing)
	http2Options := existing.ProtoReflect().Descriptor().Fields().ByName("http2_protocol_options")
	existing.ProtoReflect().Clear(http2Options)

	err := processSDSClusters(tCtx, xdsWithHTTPSDSURLs(sdsURL))

	require.EqualError(t, err, sdsCollisionError(sdsURL))
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func TestProcessSDSClustersReturnsCollisionWhenPreexistingClusterHasWrongConnectTimeout(t *testing.T) {
	sdsURL := "unix:///var/run/sds/socket"
	tCtx := &types.ResourceVersionTable{}
	require.NoError(t, createSDSCluster(tCtx, sdsURL))
	findXdsCluster(tCtx, ir.SDSClusterNameFromURL(sdsURL)).ConnectTimeout = durationpb.New(5 * time.Second)

	err := processSDSClusters(tCtx, xdsWithHTTPSDSURLs(sdsURL))

	require.EqualError(t, err, sdsCollisionError(sdsURL))
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func TestProcessSDSClustersReturnsCollisionWhenPreexistingClusterHasEmptyLoadAssignment(t *testing.T) {
	sdsURL := "unix:///var/run/sds/socket"
	tCtx := &types.ResourceVersionTable{}
	require.NoError(t, createSDSCluster(tCtx, sdsURL))
	clusterName := ir.SDSClusterNameFromURL(sdsURL)
	findXdsCluster(tCtx, clusterName).LoadAssignment = &endpoint.ClusterLoadAssignment{ClusterName: clusterName}

	err := processSDSClusters(tCtx, xdsWithHTTPSDSURLs(sdsURL))

	require.EqualError(t, err, sdsCollisionError(sdsURL))
	require.Len(t, tCtx.XdsResources[resourcev3.ClusterType], 1)
}

func xdsWithHTTPSDSURLs(urls ...string) *ir.Xds {
	return &ir.Xds{
		HTTP: []*ir.HTTPListener{{
			TLS: tlsConfigWithSDSURLs(urls...),
		}},
	}
}

func tlsConfigWithSDSURLs(urls ...string) *ir.TLSConfig {
	certificates := make([]ir.TLSCertificate, 0, len(urls))
	for _, url := range urls {
		certificates = append(certificates, ir.TLSCertificate{
			Name: strings.Trim(url, "/"),
			SDS: &ir.SDSConfig{
				SecretName: "default",
				Scheme:     "unix",
				Address:    strings.TrimPrefix(url, "unix://"),
			},
		})
	}
	return &ir.TLSConfig{Certificates: certificates}
}

func sdsCollisionError(sdsURL string) string {
	return `SDS cluster "` + ir.SDSClusterNameFromURL(sdsURL) + `" conflicts with an existing cluster`
}
