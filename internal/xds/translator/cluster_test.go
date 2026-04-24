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
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	cswrrv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/client_side_weighted_round_robin/v3"
	override_hostv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/override_host/v3"
	wrr_localityv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/wrr_locality/v3"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/testing/protocmp"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
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
			PanicThreshold: new(uint32(66)),
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
	dynamicXdsClusterLoadAssignment := buildXdsClusterLoadAssignment(bootstrapXdsCluster.Name, settings, nil, nil, nil)

	requireCmpNoDiff(t, bootstrapXdsCluster.LoadAssignment.Endpoints[0].LbEndpoints[0], dynamicXdsClusterLoadAssignment.Endpoints[0].LbEndpoints[0])
}

func TestBuildXdsClusterLoadAssignmentWithHealthCheckConfig(t *testing.T) {
	tests := []struct {
		name        string
		healthCheck *ir.HealthCheck
		expected    *endpointv3.Endpoint_HealthCheckConfig
	}{
		{
			name:        "nil health check config",
			healthCheck: nil,
			expected:    nil,
		},
		{
			name:        "nil active health check",
			healthCheck: &ir.HealthCheck{},
			expected:    nil,
		},
		{
			name: "nil health check overrides",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HealthyThreshold: new(uint32(3)),
				},
			},
			expected: nil,
		},
		{
			name: "health check overrides with port override",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HealthyThreshold: new(uint32(3)),
					Overrides: &ir.HealthCheckOverrides{
						Port: 9090,
					},
				},
			},
			expected: &endpointv3.Endpoint_HealthCheckConfig{
				PortValue: 9090,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			settings := []*ir.DestinationSetting{{
				Endpoints: []*ir.DestinationEndpoint{{Host: envoyGatewayXdsServerHost, Port: 8080}},
			}}

			clusterLoadAssignment := buildXdsClusterLoadAssignment("test-cluster", settings, tc.healthCheck, nil, nil)

			require.Len(t, clusterLoadAssignment.GetEndpoints(), 1)
			require.Len(t, clusterLoadAssignment.GetEndpoints()[0].GetLbEndpoints(), 1)
			endpoint := clusterLoadAssignment.GetEndpoints()[0].GetLbEndpoints()[0].GetEndpoint()
			require.NotNil(t, endpoint)

			require.Equal(t, uint32(8080), endpoint.GetAddress().GetSocketAddress().GetPortValue())
			requireCmpNoDiff(t, tc.expected, endpoint.HealthCheckConfig)
		})
	}
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
				Interval:             ir.MetaV1DurationPtr(10 * time.Second),
				BaseEjectionTime:     ir.MetaV1DurationPtr(30 * time.Second),
				MaxEjectionPercent:   new(uint32(10)),
				Consecutive5xxErrors: new(uint32(5)),
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
				Interval:                   ir.MetaV1DurationPtr(10 * time.Second),
				BaseEjectionTime:           ir.MetaV1DurationPtr(30 * time.Second),
				MaxEjectionPercent:         new(uint32(10)),
				Consecutive5xxErrors:       new(uint32(5)),
				FailurePercentageThreshold: new(uint32(90)),
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
				SplitExternalLocalOriginErrors: new(true),
				Interval:                       ir.MetaV1DurationPtr(10 * time.Second),
				ConsecutiveLocalOriginFailures: new(uint32(3)),
				ConsecutiveGatewayErrors:       new(uint32(2)),
				Consecutive5xxErrors:           new(uint32(5)),
				BaseEjectionTime:               ir.MetaV1DurationPtr(30 * time.Second),
				MaxEjectionPercent:             new(uint32(10)),
				FailurePercentageThreshold:     new(uint32(85)),
				AlwaysEjectOneEndpoint:         new(true),
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
				AlwaysEjectOneHost:                 wrapperspb.Bool(true),
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

func TestBuildClusterWithBackendUtilization(t *testing.T) {
	args := &xdsClusterArgs{
		name:         "test-cluster-bu",
		endpointType: EndpointTypeStatic,
		settings: []*ir.DestinationSetting{{
			Endpoints: []*ir.DestinationEndpoint{{Host: "127.0.0.1", Port: 8080}},
		}},
		loadBalancer: &ir.LoadBalancer{BackendUtilization: &ir.BackendUtilization{}},
	}

	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	require.NotNil(t, result)
	cluster := result.cluster
	require.NotNil(t, cluster)

	require.NotNil(t, cluster.LoadBalancingPolicy)
	require.Len(t, cluster.LoadBalancingPolicy.Policies, 1)

	policy := cluster.LoadBalancingPolicy.Policies[0]
	require.NotNil(t, policy)
	require.NotNil(t, policy.TypedExtensionConfig)
	require.Equal(t, "envoy.load_balancing_policies.client_side_weighted_round_robin", policy.TypedExtensionConfig.Name)
	require.NotNil(t, policy.TypedExtensionConfig.TypedConfig)
	require.Equal(t, "type.googleapis.com/envoy.extensions.load_balancing_policies.client_side_weighted_round_robin.v3.ClientSideWeightedRoundRobin", policy.TypedExtensionConfig.TypedConfig.TypeUrl)
}

func TestBuildClusterWithBackendUtilizationSlowStart(t *testing.T) {
	window := 5 * time.Second
	args := &xdsClusterArgs{
		name:         "test-cluster-bu-ss",
		endpointType: EndpointTypeStatic,
		settings: []*ir.DestinationSetting{{
			Endpoints: []*ir.DestinationEndpoint{{Host: "127.0.0.1", Port: 8080}},
		}},
		loadBalancer: &ir.LoadBalancer{BackendUtilization: &ir.BackendUtilization{
			SlowStart: &ir.SlowStart{Window: ir.MetaV1DurationPtr(window)},
		}},
	}

	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	require.NotNil(t, result)
	cluster := result.cluster
	require.NotNil(t, cluster)

	require.NotNil(t, cluster.LoadBalancingPolicy)
	require.Len(t, cluster.LoadBalancingPolicy.Policies, 1)
	policy := cluster.LoadBalancingPolicy.Policies[0]
	require.NotNil(t, policy)
	require.NotNil(t, policy.TypedExtensionConfig)
	require.Equal(t, "envoy.load_balancing_policies.client_side_weighted_round_robin", policy.TypedExtensionConfig.Name)

	// Unmarshal and verify SlowStartConfig is present
	cswrr := &cswrrv3.ClientSideWeightedRoundRobin{}
	err = policy.TypedExtensionConfig.TypedConfig.UnmarshalTo(cswrr)
	require.NoError(t, err)
	require.NotNil(t, cswrr.SlowStartConfig)
	require.NotNil(t, cswrr.SlowStartConfig.SlowStartWindow)
	require.Equal(t, window, cswrr.SlowStartConfig.SlowStartWindow.AsDuration())
}

func TestBuildClusterWithBackendUtilizationWeightedZones(t *testing.T) {
	args := &xdsClusterArgs{
		name:         "test-cluster-bu-wz",
		endpointType: EndpointTypeStatic,
		settings: []*ir.DestinationSetting{{
			Endpoints: []*ir.DestinationEndpoint{
				{Host: "127.0.0.1", Port: 8080, Zone: new("us-east-1a")},
				{Host: "127.0.0.2", Port: 8080, Zone: new("us-east-1b")},
			},
		}},
		loadBalancer: &ir.LoadBalancer{
			BackendUtilization: &ir.BackendUtilization{},
			WeightedZones: []ir.WeightedZoneConfig{
				{Zone: "us-east-1a", Weight: 80},
				{Zone: "us-east-1b", Weight: 20},
			},
		},
	}

	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	require.NotNil(t, result)
	cluster := result.cluster
	require.NotNil(t, cluster)

	// The top-level policy should be wrr_locality
	require.NotNil(t, cluster.LoadBalancingPolicy)
	require.Len(t, cluster.LoadBalancingPolicy.Policies, 1)
	policy := cluster.LoadBalancingPolicy.Policies[0]
	require.Equal(t, "envoy.load_balancing_policies.wrr_locality", policy.TypedExtensionConfig.Name)
	require.Equal(t,
		"type.googleapis.com/envoy.extensions.load_balancing_policies.wrr_locality.v3.WrrLocality",
		policy.TypedExtensionConfig.TypedConfig.TypeUrl,
	)

	// Unmarshal and verify the child policy is CSWRR
	wrrLocality := &wrr_localityv3.WrrLocality{}
	err = policy.TypedExtensionConfig.TypedConfig.UnmarshalTo(wrrLocality)
	require.NoError(t, err)
	require.NotNil(t, wrrLocality.EndpointPickingPolicy)
	require.Len(t, wrrLocality.EndpointPickingPolicy.Policies, 1)
	childPolicy := wrrLocality.EndpointPickingPolicy.Policies[0]
	require.Equal(t, "envoy.load_balancing_policies.client_side_weighted_round_robin", childPolicy.TypedExtensionConfig.Name)

	// Verify CSWRR can be unmarshaled
	cswrr := &cswrrv3.ClientSideWeightedRoundRobin{}
	err = childPolicy.TypedExtensionConfig.TypedConfig.UnmarshalTo(cswrr)
	require.NoError(t, err)
}

func TestBuildClusterWithEndpointOverrideBackendUtilizationWeightedZones(t *testing.T) {
	args := &xdsClusterArgs{
		name:         "test-cluster-eo-bu-wz",
		endpointType: EndpointTypeStatic,
		settings: []*ir.DestinationSetting{{
			Endpoints: []*ir.DestinationEndpoint{
				{Host: "127.0.0.1", Port: 8080, Zone: new("us-east-1a")},
				{Host: "127.0.0.2", Port: 8080, Zone: new("us-east-1b")},
			},
		}},
		loadBalancer: &ir.LoadBalancer{
			BackendUtilization: &ir.BackendUtilization{
				BlackoutPeriod: ir.MetaV1DurationPtr(10 * time.Second),
			},
			WeightedZones: []ir.WeightedZoneConfig{
				{Zone: "us-east-1a", Weight: 80},
				{Zone: "us-east-1b", Weight: 20},
			},
			EndpointOverride: &ir.EndpointOverride{
				ExtractFrom: []ir.EndpointOverrideExtractFrom{{
					Header: new("x-fallback-host"),
				}},
			},
		},
	}

	result, err := buildXdsCluster(args)
	require.NoError(t, err)
	require.NotNil(t, result)
	cluster := result.cluster
	require.NotNil(t, cluster)

	require.NotNil(t, cluster.LoadBalancingPolicy)
	require.Len(t, cluster.LoadBalancingPolicy.Policies, 1)
	policy := cluster.LoadBalancingPolicy.Policies[0]
	require.Equal(t, "envoy.load_balancing_policies.override_host", policy.TypedExtensionConfig.Name)

	overrideHost := &override_hostv3.OverrideHost{}
	err = policy.TypedExtensionConfig.TypedConfig.UnmarshalTo(overrideHost)
	require.NoError(t, err)
	require.NotNil(t, overrideHost.FallbackPolicy)
	require.Len(t, overrideHost.FallbackPolicy.Policies, 1)

	fallbackPolicy := overrideHost.FallbackPolicy.Policies[0]
	require.Equal(t, "envoy.load_balancing_policies.wrr_locality", fallbackPolicy.TypedExtensionConfig.Name)

	wrrLocality := &wrr_localityv3.WrrLocality{}
	err = fallbackPolicy.TypedExtensionConfig.TypedConfig.UnmarshalTo(wrrLocality)
	require.NoError(t, err)
	require.NotNil(t, wrrLocality.EndpointPickingPolicy)
	require.Len(t, wrrLocality.EndpointPickingPolicy.Policies, 1)

	childPolicy := wrrLocality.EndpointPickingPolicy.Policies[0]
	require.Equal(t, "envoy.load_balancing_policies.client_side_weighted_round_robin", childPolicy.TypedExtensionConfig.Name)

	cswrr := &cswrrv3.ClientSideWeightedRoundRobin{}
	err = childPolicy.TypedExtensionConfig.TypedConfig.UnmarshalTo(cswrr)
	require.NoError(t, err)
	require.Equal(t, 10*time.Second, cswrr.BlackoutPeriod.AsDuration())
}

func TestGetHealthCheckOverridesHostname(t *testing.T) {
	tests := []struct {
		name        string
		healthCheck *ir.HealthCheck
		endpoint    *ir.DestinationEndpoint
		expected    string
	}{
		{
			name: "nil HTTP health checker",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HealthyThreshold: new(uint32(3)),
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "backend.example.com",
		},
		{
			name: "HTTP health checker with empty host and endpoint has hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "",
						Path: "/health",
					},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "backend.example.com",
		},
		{
			name: "HTTP health checker with wildcard host and endpoint has hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "*",
						Path: "/health",
					},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "backend.example.com",
		},
		{
			name: "HTTP health checker with explicit host",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "health.example.com",
						Path: "/health",
					},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "",
		},
		{
			name: "HTTP health checker with empty host but nil endpoint",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "",
						Path: "/health",
					},
				},
			},
			endpoint: nil,
			expected: "",
		},
		{
			name: "HTTP health checker with empty host but endpoint has nil hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "",
						Path: "/health",
					},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: nil,
			},
			expected: "",
		},
		{
			name: "HTTP health checker with wildcard host but nil endpoint",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "*",
						Path: "/health",
					},
				},
			},
			endpoint: nil,
			expected: "",
		},
		{
			name: "HTTP health checker with wildcard host but endpoint has nil hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					HTTP: &ir.HTTPHealthChecker{
						Host: "*",
						Path: "/health",
					},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: nil,
			},
			expected: "",
		},
		{
			name: "TCP health checker with endpoint hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					TCP: &ir.TCPHealthChecker{},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "backend.example.com",
		},
		{
			name: "GRPC health checker with endpoint hostname",
			healthCheck: &ir.HealthCheck{
				Active: &ir.ActiveHealthCheck{
					GRPC: &ir.GRPCHealthChecker{},
				},
			},
			endpoint: &ir.DestinationEndpoint{
				Host:     "example.com",
				Port:     8080,
				Hostname: new("backend.example.com"),
			},
			expected: "backend.example.com",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := getHealthCheckOverridesHostname(tc.healthCheck, tc.endpoint)
			require.Equal(t, tc.expected, result)
		})
	}
}
