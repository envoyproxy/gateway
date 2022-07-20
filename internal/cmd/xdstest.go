package cmd

import (
	"net"
	"time"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/log"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	envoy_service_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	envoy_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	envoy_service_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	envoy_service_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	envoy_service_route_v3 "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	envoy_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	envoy_service_secret_v3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	envoy_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

// getServerCommand returns the server cobra command to be executed.
func getxDSTestCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "xdstest",
		Aliases: []string{"xdstest"},
		Short:   "Run a test xDS server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return xDSTest()
		},
	}
	return cmd
}

// server serves Envoy Gateway.
func xDSTest() error {

	logger, err := log.NewLogger()
	if err != nil {
		return err
	}

	// Set the logr Logger to debug.
	// zap's logr impl requires negative levels.
	logger = logger.V(-2)

	cpLogger := log.NewLogrWrapper(logger)

	ctx := signals.SetupSignalHandler()

	logger.Info("Starting xDS Tester service")
	defer logger.Info("Stopping xDS Tester service")

	ir1 := &ir.Xds{
		Name: "xdstest",
		HTTP: []*ir.HTTPListener{
			{
				Name:    "first-listener",
				Address: "0.0.0.0",
				Port:    10080,
				Hostnames: []string{
					"*",
				},
				Routes: []*ir.HTTPRoute{
					{
						Name: "first-route",
						Destinations: []*ir.RouteDestination{
							{
								Host: "1.2.3.4",
								Port: 50000,
							},
						},
					},
				},
			},
		},
	}

	ir2 := &ir.Xds{
		Name: "xdstest",
		HTTP: []*ir.HTTPListener{
			{
				Name:    "first-listener",
				Address: "0.0.0.0",
				Port:    10080,
				Hostnames: []string{
					"*",
				},
				Routes: []*ir.HTTPRoute{
					{
						Name: "second-route",
						Destinations: []*ir.RouteDestination{
							{
								Host: "1.2.3.4",
								Port: 50000,
							},
						},
					},
				},
			},
		},
	}

	ir3 := &ir.Xds{
		Name: "xdstest",
		HTTP: []*ir.HTTPListener{
			{
				Name:    "second-listener",
				Address: "0.0.0.0",
				Port:    10080,
				Hostnames: []string{
					"*",
				},
				Routes: []*ir.HTTPRoute{
					{
						Name: "second-route",
						Destinations: []*ir.RouteDestination{
							{
								Host: "1.2.3.4",
								Port: 50000,
							},
						},
					},
				},
			},
		},
	}

	cacheVersion1, err := translator.TranslateXdsIR(ir1)
	if err != nil {
		return err
	}

	cacheVersion2, err := translator.TranslateXdsIR(ir2)
	if err != nil {
		return err
	}

	cacheVersion3, err := translator.TranslateXdsIR(ir3)
	if err != nil {
		return err
	}

	g := grpc.NewServer()

	snapCache := cache.NewSnapshotCache(false, cpLogger)
	// contour_xds_v3.NewRequestLoggingCallbacks(log)
	RegisterServer(envoy_server_v3.NewServer(ctx, snapCache, snapCache), g)

	// if err := provider.Start(cfg); err != nil {
	// 	return err
	// }
	addr := net.JoinHostPort("0.0.0.0", "8001")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()

		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		g.Stop()
	}()

	go func() {
		// This little function sleeps 10 seconds then swaps
		// betweeen various versions of the IR
		logger.Info("Sleeping for a bit before updating the cache")
		for {
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for first-listener with first-route")
			snapCache.GenerateNewSnapshot(cacheVersion1.GetXdsResources())
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for first-listener with second-route")
			snapCache.GenerateNewSnapshot(cacheVersion2.GetXdsResources())
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for second-listener with second-route")
			snapCache.GenerateNewSnapshot(cacheVersion3.GetXdsResources())
		}
	}()

	return g.Serve(l)

}

// Server is a collection of handlers for streaming discovery requests.
type Server interface {
	envoy_service_cluster_v3.ClusterDiscoveryServiceServer
	envoy_service_endpoint_v3.EndpointDiscoveryServiceServer
	envoy_service_listener_v3.ListenerDiscoveryServiceServer
	envoy_service_route_v3.RouteDiscoveryServiceServer
	envoy_service_discovery_v3.AggregatedDiscoveryServiceServer
	envoy_service_secret_v3.SecretDiscoveryServiceServer
	envoy_service_runtime_v3.RuntimeDiscoveryServiceServer
}

// RegisterServer registers the given xDS protocol Server with the gRPC
// runtime.
func RegisterServer(srv Server, g *grpc.Server) {
	// register services
	envoy_service_discovery_v3.RegisterAggregatedDiscoveryServiceServer(g, srv)
	envoy_service_secret_v3.RegisterSecretDiscoveryServiceServer(g, srv)
	envoy_service_cluster_v3.RegisterClusterDiscoveryServiceServer(g, srv)
	envoy_service_endpoint_v3.RegisterEndpointDiscoveryServiceServer(g, srv)
	envoy_service_listener_v3.RegisterListenerDiscoveryServiceServer(g, srv)
	envoy_service_route_v3.RegisterRouteDiscoveryServiceServer(g, srv)
	envoy_service_runtime_v3.RegisterRuntimeDiscoveryServiceServer(g, srv)
}
