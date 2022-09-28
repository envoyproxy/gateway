package runner

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"strconv"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

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
	// xdsTLSCertFilename is the fully qualified path of the file containing the
	// xDS server TLS certificate.
	xdsTLSCertFilename = "/certs/tls.crt"
	// xdsTLSKeyFilename is the fully qualified path of the file containing the
	// xDS server TLS key.
	xdsTLSKeyFilename = "/certs/tls.key"
	// xdsTLSCaFilename is the fully qualified path of the file containing the
	// xDS server trusted CA certificate.
	xdsTLSCaFilename = "/certs/ca.crt"
)

type Config struct {
	config.Server
	Xds   *message.Xds
	grpc  *grpc.Server
	cache cache.SnapshotCacheWithCallbacks
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
	cfg := r.tlsConfig(xdsTLSCertFilename, xdsTLSKeyFilename, xdsTLSCaFilename)
	r.grpc = grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))

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
	for range r.Xds.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		// Load all resources required for translation
		for key, xds := range r.Xds.LoadAll() {
			if xds == nil {
				r.Logger.Info("xds is nil, skipping")
				continue
			}
			// Update snapshot cache
			err := r.cache.GenerateNewSnapshot(key, xds.XdsResources)
			if err != nil {
				r.Logger.Error(err, "failed to generate a snapshot")
			}
		}
	}

	r.Logger.Info("subscriber shutting down")

}

func (r *Runner) tlsConfig(cert, key, ca string) *tls.Config {
	loadConfig := func() (*tls.Config, error) {
		cert, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			return nil, err
		}

		// Load the CA cert.
		ca, err := os.ReadFile(ca)
		if err != nil {
			return nil, err
		}

		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(ca) {
			return nil, fmt.Errorf("failed to parse CA certificate")
		}

		return &tls.Config{
			Certificates: []tls.Certificate{cert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	// Attempt to load certificates and key to catch configuration errors early.
	if _, lerr := loadConfig(); lerr != nil {
		r.Logger.Error(lerr, "failed to load certificate and key")
	}
	r.Logger.Info("loaded TLS certificate and key")

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Rand:       rand.Reader,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return loadConfig()
		},
	}
}
