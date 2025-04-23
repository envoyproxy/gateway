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
	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	commonv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/common/v3"
	least_requestv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/least_request/v3"
	maglevv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/maglev/v3"
	randomv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/random/v3"
	round_robinv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/load_balancing_policies/round_robin/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
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
	assert.True(t, proto.Equal(bootstrapXdsCluster.TransportSocket, dynamicXdsCluster.TransportSocket))
	assert.True(t, proto.Equal(bootstrapXdsCluster.ConnectTimeout, dynamicXdsCluster.ConnectTimeout))
}

func TestCheckZoneAwareRouting(t *testing.T) {
	tests := []struct {
		name               string
		zoneRoutingEnabled bool
		loadBalancerCfg    *ir.LoadBalancer
	}{
		{
			name:               "zone-routing with default lb",
			zoneRoutingEnabled: true,
			loadBalancerCfg: &ir.LoadBalancer{
				LeastRequest: &ir.LeastRequest{},
			},
		},
		{
			name:               "zone-routing with nil lb",
			zoneRoutingEnabled: true,
			loadBalancerCfg:    nil,
		},
		{
			name:               "zone-routing with least request",
			zoneRoutingEnabled: true,
			loadBalancerCfg: &ir.LoadBalancer{
				LeastRequest: &ir.LeastRequest{
					SlowStart: &ir.SlowStart{Window: &metav1.Duration{Duration: 1 * time.Second}},
				},
			},
		},
		{
			name:               "zone-routing with round robin",
			zoneRoutingEnabled: true,
			loadBalancerCfg: &ir.LoadBalancer{
				RoundRobin: &ir.RoundRobin{
					SlowStart: &ir.SlowStart{Window: &metav1.Duration{Duration: 1 * time.Second}},
				},
			},
		},
		{
			name:               "zone-routing with random",
			zoneRoutingEnabled: true,
			loadBalancerCfg:    &ir.LoadBalancer{Random: &ir.Random{}},
		},
		{
			name:               "zone-routing with maglev",
			zoneRoutingEnabled: true,
			loadBalancerCfg: &ir.LoadBalancer{
				ConsistentHash: &ir.ConsistentHash{
					TableSize: proto.Uint64(65537),
				},
			},
		},
		{
			name:               "zone-routing with round robin",
			zoneRoutingEnabled: true,
			loadBalancerCfg: &ir.LoadBalancer{
				RoundRobin: &ir.RoundRobin{
					SlowStart: &ir.SlowStart{Window: &metav1.Duration{Duration: 1 * time.Second}},
				},
			},
		},
		{
			name:               "zone-routing disabled",
			zoneRoutingEnabled: false,
			loadBalancerCfg: &ir.LoadBalancer{
				LeastRequest: &ir.LeastRequest{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bootstrapXdsCluster := getXdsClusterObjFromBootstrap(t)
			ds := &ir.DestinationSetting{
				Endpoints:               []*ir.DestinationEndpoint{{Host: envoyGatewayXdsServerHost, Port: bootstrap.DefaultXdsServerPort}},
				ZoneAwareRoutingEnabled: tt.zoneRoutingEnabled,
			}
			args := &xdsClusterArgs{
				name:         bootstrapXdsCluster.Name,
				tSocket:      bootstrapXdsCluster.TransportSocket,
				endpointType: EndpointTypeDNS,
				healthCheck: &ir.HealthCheck{
					PanicThreshold: ptr.To[uint32](66),
				},
				loadBalancer: tt.loadBalancerCfg,
				settings:     []*ir.DestinationSetting{ds},
			}
			clusterResult, err := buildXdsCluster(args)
			dynamicXdsCluster := clusterResult.cluster
			require.NoError(t, err)
			checkZoneAwareRouting(dynamicXdsCluster, args)

			if !tt.zoneRoutingEnabled {
				require.Nil(t, dynamicXdsCluster.LoadBalancingPolicy)
				require.Equal(t, &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{LocalityWeightedLbConfig: &clusterv3.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}, dynamicXdsCluster.CommonLbConfig.LocalityConfigSpecifier)
			} else {
				require.Nil(t, dynamicXdsCluster.CommonLbConfig.LocalityConfigSpecifier)
				expectedLoadBalancingPolicy := getExpectedClusterLbPolicies(dynamicXdsCluster.LbPolicy, args.loadBalancer)
				require.Equal(t, expectedLoadBalancingPolicy.Policies[0].TypedExtensionConfig.Name, dynamicXdsCluster.LoadBalancingPolicy.Policies[0].TypedExtensionConfig.Name)
				require.Equal(t, expectedLoadBalancingPolicy.Policies[0].GetTypedExtensionConfig().GetTypedConfig().String(), dynamicXdsCluster.LoadBalancingPolicy.Policies[0].GetTypedExtensionConfig().GetTypedConfig().String())
			}
		})
	}
}

func getExpectedClusterLbPolicies(policy clusterv3.Cluster_LbPolicy, lb *ir.LoadBalancer) *clusterv3.LoadBalancingPolicy {
	localityLbConfig := &commonv3.LocalityLbConfig{
		LocalityConfigSpecifier: &commonv3.LocalityLbConfig_ZoneAwareLbConfig_{
			ZoneAwareLbConfig: &commonv3.LocalityLbConfig_ZoneAwareLbConfig{
				MinClusterSize:             wrapperspb.UInt64(1),
				ForceLocalityDirectRouting: true,
			},
		},
	}
	leastRequest := &least_requestv3.LeastRequest{
		LocalityLbConfig: localityLbConfig,
	}
	typedLeastRequest, _ := anypb.New(leastRequest)
	loadBalancingPolicy := &clusterv3.LoadBalancingPolicy{
		Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
			TypedExtensionConfig: &corev3.TypedExtensionConfig{
				Name:        "envoy.load_balancing_policies.least_request",
				TypedConfig: typedLeastRequest,
			},
		}},
	}

	if lb == nil {
		return loadBalancingPolicy
	}
	switch policy {
	case clusterv3.Cluster_LEAST_REQUEST:
		if lb.LeastRequest != nil && lb.LeastRequest.SlowStart != nil && lb.LeastRequest.SlowStart.Window != nil {
			leastRequest.SlowStartConfig = &commonv3.SlowStartConfig{
				SlowStartWindow: durationpb.New(lb.LeastRequest.SlowStart.Window.Duration),
			}
		}
		loadBalancingPolicy.Policies[0].TypedExtensionConfig.TypedConfig, _ = anypb.New(leastRequest)
		return loadBalancingPolicy
	case clusterv3.Cluster_ROUND_ROBIN:
		roundRobin := &round_robinv3.RoundRobin{
			LocalityLbConfig: localityLbConfig,
		}
		if lb.RoundRobin.SlowStart != nil && lb.RoundRobin.SlowStart.Window != nil {
			roundRobin.SlowStartConfig = &commonv3.SlowStartConfig{
				SlowStartWindow: durationpb.New(lb.RoundRobin.SlowStart.Window.Duration),
			}
		}
		typedRoundRobin, _ := anypb.New(roundRobin)
		return &clusterv3.LoadBalancingPolicy{
			Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
				TypedExtensionConfig: &corev3.TypedExtensionConfig{
					Name:        "envoy.load_balancing_policies.round_robin",
					TypedConfig: typedRoundRobin,
				},
			}},
		}
	case clusterv3.Cluster_RANDOM:
		random := &randomv3.Random{
			LocalityLbConfig: localityLbConfig,
		}
		typeRandom, _ := anypb.New(random)
		return &clusterv3.LoadBalancingPolicy{
			Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
				TypedExtensionConfig: &corev3.TypedExtensionConfig{
					Name:        "envoy.load_balancing_policies.random",
					TypedConfig: typeRandom,
				},
			}},
		}
	case clusterv3.Cluster_MAGLEV:
		consistentHash := &maglevv3.Maglev{}
		if lb.ConsistentHash.TableSize != nil {
			consistentHash.TableSize = wrapperspb.UInt64(*lb.ConsistentHash.TableSize)
		}
		typedConsistentHash, _ := anypb.New(consistentHash)

		return &clusterv3.LoadBalancingPolicy{
			Policies: []*clusterv3.LoadBalancingPolicy_Policy{{
				TypedExtensionConfig: &corev3.TypedExtensionConfig{
					Name:        "envoy.load_balancing_policies.maglev",
					TypedConfig: typedConsistentHash,
				},
			}},
		}

	}
	return nil
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
