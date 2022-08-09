package service

import (
	"context"
	"net"

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

type Service struct {
	config.Server
	XdsResources *message.XdsResources
	grpc         *grpc.Server
	cache        cache.SnapshotCacheWithCallbacks
}

func (s *Service) Name() string {
	return "xds-server"
}

// Start starts the GatewayAPI service
func (s *Service) Start(ctx context.Context) error {
	s.Logger = s.Logger.WithValues("service", s.Name())
	go s.subscribeAndTranslate(ctx)
	go s.setupXdsServer(ctx)

	<-ctx.Done()
	s.Logger.Info("shutting down")
	// We don't use GracefulStop here because envoy
	// has long-lived hanging xDS requests. There's no
	// mechanism to make those pending requests fail,
	// so we forcibly terminate the TCP sessions.
	s.grpc.Stop()
	return nil
}

func (s *Service) setupXdsServer(ctx context.Context) {
	// Set up the gRPC server and register the xDS handler.
	s.grpc = grpc.NewServer()

	s.cache = cache.NewSnapshotCache(false, s.Logger)
	registerServer(controlplane_server_v3.NewServer(ctx, s.cache, s.cache), s.grpc)

	// TODO: Make the listening address and port configurable
	addr := net.JoinHostPort("0.0.0.0", "8001")
	l, err := net.Listen("tcp", addr)
	if err != nil {
		s.Logger.Error(err, "failed to listen on address", addr)
	}
	err = s.grpc.Serve(l)
	if err != nil {
		s.Logger.Error(err, "failed to start grpc based xds server")
	}
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

func (s *Service) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	xdsCh := s.XdsResources.Subscribe(ctx)
	for ctx.Err() == nil {
		// Receive subscribed resource notifications
		<-xdsCh
		// Load all resources required for translation
		xdsResources := s.XdsResources.Get()
		// Update snapshot cache
		err := s.cache.GenerateNewSnapshot(*xdsResources)
		if err != nil {
			s.Logger.Error(err, "failed to generate a snapshot")
		}
	}
}
