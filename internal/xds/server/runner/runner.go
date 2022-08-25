package runner

import (
	"context"
	"net"
	"strconv"

	"google.golang.org/grpc"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	controlplane_service_cluster_v3 "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	controlplane_service_discovery_v3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	controlplane_service_endpoint_v3 "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	controlplane_service_listener_v3 "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	controlplane_service_route_v3 "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	controlplane_service_runtime_v3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	controlplane_service_secret_v3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	controlplane_server_v3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
)

const (
	// XdsServerAddress is the listening address of the xds-server.
	XdsServerAddress = "0.0.0.0"
	// XdsServerPort is the listening port of the xds-server.
	XdsServerPort = 18000
)

type Config struct {
	config.Server
	XdsResources *message.XdsResources
	grpc         *grpc.Server
	cache        cache.SnapshotCacheWithCallbacks
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return "xds-server"
}

// Start starts the xds-server runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)
	go r.setupXdsServer(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) setupXdsServer(ctx context.Context) {
	// Set up the gRPC server and register the xDS handler.
	r.grpc = grpc.NewServer()

	r.cache = cache.NewSnapshotCache(false, r.Logger)
	registerServer(controlplane_server_v3.NewServer(ctx, r.cache, r.cache), r.grpc)

	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(XdsServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", addr)
	}
	err = r.grpc.Serve(l)
	if err != nil {
		r.Logger.Error(err, "failed to start grpc based xds server")
	}

	<-ctx.Done()
	r.Logger.Info("grpc server shutting down")
	// We don't use GracefulStop here because envoy
	// has long-lived hanging xDS requests. There's no
	// mechanism to make those pending requests fail,
	// so we forcibly terminate the TCP sessions.
	r.grpc.Stop()
}

// registerServer registers the given xDS protocol Server with the gRPC
// runtime.
func registerServer(srv controlplane_server_v3.Server, g *grpc.Server) {
	// register services
	controlplane_service_discovery_v3.RegisterAggregatedDiscoveryServiceServer(g, srv)
	controlplane_service_secret_v3.RegisterSecretDiscoveryServiceServer(g, srv)
	controlplane_service_cluster_v3.RegisterClusterDiscoveryServiceServer(g, srv)
	controlplane_service_endpoint_v3.RegisterEndpointDiscoveryServiceServer(g, srv)
	controlplane_service_listener_v3.RegisterListenerDiscoveryServiceServer(g, srv)
	controlplane_service_route_v3.RegisterRouteDiscoveryServiceServer(g, srv)
	controlplane_service_runtime_v3.RegisterRuntimeDiscoveryServiceServer(g, srv)
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	for range r.XdsResources.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		// Load all resources required for translation
		xdsResources := r.XdsResources.Get()
		if xdsResources == nil {
			r.Logger.Info("xdsResources is nil, skipping")
			continue
		}
		// Update snapshot cache
		err := r.cache.GenerateNewSnapshot(*xdsResources)
		if err != nil {
			r.Logger.Error(err, "failed to generate a snapshot")
		}
	}

	r.Logger.Info("subscriber shutting down")

}
