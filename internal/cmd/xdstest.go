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

	controlplane_service_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	controlplane_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	controlplane_service_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	controlplane_service_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	controlplane_service_route_v3 "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	controlplane_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	controlplane_service_secret_v3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	controlplane_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

// The xdstest command is intended just to show how updating the IR can produce different
// xDS output, including showing that Delta xDS works.
// You'll need an xDS probe like the `contour cli` command to check.
//
// It's also intended that this get removed once we have a full loop implemented in
// `gateway serve`.

// getxDSTestCommand returns the xdstest cobra command to be executed.
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

// xDSTest implements the command.
// This is deliberately verbose and unoptimized, since the purpose
// is just to illustrate how the flow will need to work.
func xDSTest() error {

	// Grab the logr.Logger.
	logger, err := log.NewLogger()
	if err != nil {
		return err
	}

	// Set the logr Logger to debug.
	// zap's logr impl requires negative levels.
	logger = logger.V(-2)

	ctx := signals.SetupSignalHandler()

	logger.Info("Starting xDS Tester service")
	defer logger.Info("Stopping xDS Tester service")

	// Create three IR versions that we'll swap between, to
	// generate xDS updates for the various methods.
	ir1 := &ir.Xds{
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

	// Now, we do the translation because everything is static.
	// Normally, we'd do this in response to updates on the
	// message bus.
	cacheVersion1, err := translator.Translate(ir1)
	if err != nil {
		return err
	}

	cacheVersion2, err := translator.Translate(ir2)
	if err != nil {
		return err
	}

	cacheVersion3, err := translator.Translate(ir3)
	if err != nil {
		return err
	}

	// Set up the gRPC server and register the xDS handler.
	g := grpc.NewServer()

	snapCache := cache.NewSnapshotCache(false, logger)
	RegisterServer(controlplane_server_v3.NewServer(ctx, snapCache, snapCache), g)

	addr := net.JoinHostPort("0.0.0.0", "8001")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	// Handle the signals and stop when the signal context does.
	go func() {
		<-ctx.Done()

		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		g.Stop()
	}()

	// Loop through the various configs, updating the SnapshotCache
	// each time. This will run until the process is killed by signal
	// (SIGINT, SIGKILL etc).
	go func() {
		// This little function sleeps 10 seconds then swaps
		// between various versions of the IR
		logger.Info("Sleeping for a bit before updating the cache")
		for {
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for first-listener with first-route")
			err := snapCache.GenerateNewSnapshot("", cacheVersion1.GetXdsResources())
			if err != nil {
				logger.Error(err, "Something went wrong with generating a snapshot")
			}
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for first-listener with second-route")
			err = snapCache.GenerateNewSnapshot("", cacheVersion2.GetXdsResources())
			if err != nil {
				logger.Error(err, "Something went wrong with generating a snapshot")
			}
			time.Sleep(10 * time.Second)
			logger.Info("Updating the cache for second-listener with second-route")
			err = snapCache.GenerateNewSnapshot("", cacheVersion3.GetXdsResources())
			if err != nil {
				logger.Error(err, "Something went wrong with generating a snapshot")
			}
		}
	}()

	return g.Serve(l)

}

// Some helper stuff that we'll need to put somewhere eventually.

// Server is a collection of handlers for streaming discovery requests.
type Server interface {
	controlplane_service_cluster_v3.ClusterDiscoveryServiceServer
	controlplane_service_endpoint_v3.EndpointDiscoveryServiceServer
	controlplane_service_listener_v3.ListenerDiscoveryServiceServer
	controlplane_service_route_v3.RouteDiscoveryServiceServer
	controlplane_service_discovery_v3.AggregatedDiscoveryServiceServer
	controlplane_service_secret_v3.SecretDiscoveryServiceServer
	controlplane_service_runtime_v3.RuntimeDiscoveryServiceServer
}

// RegisterServer registers the given xDS protocol Server with the gRPC
// runtime.
func RegisterServer(srv Server, g *grpc.Server) {
	// register services
	controlplane_service_discovery_v3.RegisterAggregatedDiscoveryServiceServer(g, srv)
	controlplane_service_secret_v3.RegisterSecretDiscoveryServiceServer(g, srv)
	controlplane_service_cluster_v3.RegisterClusterDiscoveryServiceServer(g, srv)
	controlplane_service_endpoint_v3.RegisterEndpointDiscoveryServiceServer(g, srv)
	controlplane_service_listener_v3.RegisterListenerDiscoveryServiceServer(g, srv)
	controlplane_service_route_v3.RegisterRouteDiscoveryServiceServer(g, srv)
	controlplane_service_runtime_v3.RegisterRuntimeDiscoveryServiceServer(g, srv)
}
