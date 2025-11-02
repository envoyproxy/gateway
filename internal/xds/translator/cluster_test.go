// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	bootstrapv3 "github.com/envoyproxy/go-control-plane/envoy/config/bootstrap/v3"
	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	require.NotNil(t, bootstrapXdsCluster)

	args := &xdsClusterArgs{
		name:         bootstrapXdsCluster.Name,
		tSocket:      bootstrapXdsCluster.TransportSocket,
		endpointType: EndpointTypeDNS,
		healthCheck: &ir.HealthCheck{
			PanicThreshold: ptr.To[uint32](66),
		},
	}
	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	dynamicXdsCluster := result.cluster
	require.Equal(t, bootstrapXdsCluster.Name, dynamicXdsCluster.Name)
	require.Equal(t, bootstrapXdsCluster.ClusterDiscoveryType, dynamicXdsCluster.ClusterDiscoveryType)
	require.Equal(t, bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket)
	requireCmpNoDiff(t, bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket)
	requireCmpNoDiff(t, bootstrapXdsCluster.ConnectTimeout, dynamicXdsCluster.ConnectTimeout)
}

func TestBuildXdsClusterLoadAssignment(t *testing.T) {
	bootstrapXdsCluster := getXdsClusterObjFromBootstrap(t)
	require.NotNil(t, bootstrapXdsCluster)
	ds := &ir.DestinationSetting{
		Endpoints: []*ir.DestinationEndpoint{{Host: envoyGatewayXdsServerHost, Port: bootstrap.DefaultXdsServerPort}},
	}
	settings := []*ir.DestinationSetting{ds}
	dynamicXdsClusterLoadAssignment := buildXdsClusterLoadAssignment(bootstrapXdsCluster.Name, settings, nil)

	requireCmpNoDiff(t, bootstrapXdsCluster.LoadAssignment.Endpoints[0].LbEndpoints[0], dynamicXdsClusterLoadAssignment.Endpoints[0].LbEndpoints[0])
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

func TestBuildXdsOutlierDetection(t *testing.T) {
	tests := []struct {
		name     string
		input    *ir.OutlierDetection
		expected *clusterv3.OutlierDetection
	}{
		{
			name: "basic outlier detection",
			input: &ir.OutlierDetection{
				Interval:             ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				BaseEjectionTime:     ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:   ptr.To[int32](10),
				Consecutive5xxErrors: ptr.To[uint32](5),
			},
			expected: &clusterv3.OutlierDetection{
				Interval:           durationpb.New(10 * time.Second),
				BaseEjectionTime:   durationpb.New(30 * time.Second),
				MaxEjectionPercent: wrapperspb.UInt32(10),
				Consecutive_5Xx:    wrapperspb.UInt32(5),
			},
		},
		{
			name: "outlier detection with failure percentage threshold",
			input: &ir.OutlierDetection{
				Interval:                   ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				BaseEjectionTime:           ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:         ptr.To[int32](10),
				Consecutive5xxErrors:       ptr.To[uint32](5),
				FailurePercentageThreshold: ptr.To[uint32](90),
			},
			expected: &clusterv3.OutlierDetection{
				Interval:                   durationpb.New(10 * time.Second),
				BaseEjectionTime:           durationpb.New(30 * time.Second),
				MaxEjectionPercent:         wrapperspb.UInt32(10),
				Consecutive_5Xx:            wrapperspb.UInt32(5),
				FailurePercentageThreshold: wrapperspb.UInt32(90),
				EnforcingFailurePercentage: wrapperspb.UInt32(100),
			},
		},
		{
			name: "outlier detection with all fields",
			input: &ir.OutlierDetection{
				SplitExternalLocalOriginErrors: ptr.To(true),
				Interval:                       ptr.To(metav1.Duration{Duration: 10 * time.Second}),
				ConsecutiveLocalOriginFailures: ptr.To[uint32](3),
				ConsecutiveGatewayErrors:       ptr.To[uint32](2),
				Consecutive5xxErrors:           ptr.To[uint32](5),
				BaseEjectionTime:               ptr.To(metav1.Duration{Duration: 30 * time.Second}),
				MaxEjectionPercent:             ptr.To[int32](10),
				FailurePercentageThreshold:     ptr.To[uint32](85),
			},
			expected: &clusterv3.OutlierDetection{
				SplitExternalLocalOriginErrors:     true,
				Interval:                           durationpb.New(10 * time.Second),
				ConsecutiveLocalOriginFailure:      wrapperspb.UInt32(3),
				EnforcingConsecutiveGatewayFailure: wrapperspb.UInt32(100),
				ConsecutiveGatewayFailure:          wrapperspb.UInt32(2),
				Consecutive_5Xx:                    wrapperspb.UInt32(5),
				BaseEjectionTime:                   durationpb.New(30 * time.Second),
				MaxEjectionPercent:                 wrapperspb.UInt32(10),
				FailurePercentageThreshold:         wrapperspb.UInt32(85),
				EnforcingFailurePercentage:         wrapperspb.UInt32(100),
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := buildXdsOutlierDetection(tc.input)
			requireCmpNoDiff(t, tc.expected, result)
		})
	}
}

func requireCmpNoDiff(t *testing.T, expected, actual interface{}) {
	require.Empty(t, cmp.Diff(expected, actual, protocmp.Transform()))
}

func TestBuildCluster_WithClientSideWeightedRoundRobin(t *testing.T) {
	args := &xdsClusterArgs{
		name:         "test-cluster-cswrr",
		endpointType: EndpointTypeStatic,
		settings: []*ir.DestinationSetting{{
			Endpoints: []*ir.DestinationEndpoint{{Host: "127.0.0.1", Port: 8080}},
		}},
		loadBalancer: &ir.LoadBalancer{ClientSideWeightedRoundRobin: &ir.ClientSideWeightedRoundRobin{}},
	}

	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	require.NotNil(t, result)
	cluster := result.cluster
	require.NotNil(t, cluster)

	// LbPolicy should be CLUSTER_PROVIDED when using typed LoadBalancingPolicy
	require.Equal(t, clusterv3.Cluster_CLUSTER_PROVIDED, cluster.LbPolicy)
	require.NotNil(t, cluster.LoadBalancingPolicy)
	require.Len(t, cluster.LoadBalancingPolicy.Policies, 1)

	policy := cluster.LoadBalancingPolicy.Policies[0]
	require.NotNil(t, policy)
	require.NotNil(t, policy.TypedExtensionConfig)
	require.Equal(t, "envoy.load_balancing_policies.client_side_weighted_round_robin", policy.TypedExtensionConfig.Name)
	require.NotNil(t, policy.TypedExtensionConfig.TypedConfig)
	require.Equal(t, "type.googleapis.com/envoy.extensions.load_balancing_policies.client_side_weighted_round_robin.v3.ClientSideWeightedRoundRobin", policy.TypedExtensionConfig.TypedConfig.TypeUrl)
}
