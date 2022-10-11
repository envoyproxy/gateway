package translator

import (
	"time"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsCluster(routeName string, destinations []*ir.RouteDestination) (*cluster.Cluster, error) {
	localities := make([]*endpoint.LocalityLbEndpoints, 0, 1)
	locality := &endpoint.LocalityLbEndpoints{
		Locality:    &core.Locality{},
		LbEndpoints: buildXdsEndpoints(destinations),
		Priority:    0,
		// Each locality gets the same weight 1. There is a single locality
		// per priority, so the weight value does not really matter, but some
		// load balancers need the value to be set.
		LoadBalancingWeight: &wrapperspb.UInt32Value{Value: 1}}
	localities = append(localities, locality)
	clusterName := getXdsClusterName(routeName)
	return &cluster.Cluster{
		Name:                 clusterName,
		ConnectTimeout:       durationpb.New(5 * time.Second),
		ClusterDiscoveryType: &cluster.Cluster_Type{Type: cluster.Cluster_STATIC},
		LbPolicy:             cluster.Cluster_ROUND_ROBIN,
		LoadAssignment:       &endpoint.ClusterLoadAssignment{ClusterName: clusterName, Endpoints: localities},
		DnsLookupFamily:      cluster.Cluster_V4_ONLY,
		CommonLbConfig: &cluster.Cluster_CommonLbConfig{
			LocalityConfigSpecifier: &cluster.Cluster_CommonLbConfig_LocalityWeightedLbConfig_{
				LocalityWeightedLbConfig: &cluster.Cluster_CommonLbConfig_LocalityWeightedLbConfig{}}},
		OutlierDetection: &cluster.OutlierDetection{},
	}, nil

}

func buildXdsEndpoints(destinations []*ir.RouteDestination) []*endpoint.LbEndpoint {
	endpoints := make([]*endpoint.LbEndpoint, 0, len(destinations))
	for _, destination := range destinations {
		lbEndpoint := &endpoint.LbEndpoint{
			HostIdentifier: &endpoint.LbEndpoint_Endpoint{
				Endpoint: &endpoint.Endpoint{
					Address: &core.Address{
						Address: &core.Address_SocketAddress{
							SocketAddress: &core.SocketAddress{
								Protocol: core.SocketAddress_TCP,
								Address:  destination.Host,
								PortSpecifier: &core.SocketAddress_PortValue{
									PortValue: destination.Port,
								},
							},
						},
					},
				},
			},
		}
		if destination.Weight != 0 {
			lbEndpoint.LoadBalancingWeight = &wrapperspb.UInt32Value{Value: destination.Weight}
		}
		endpoints = append(endpoints, lbEndpoint)
	}
	return endpoints
}
