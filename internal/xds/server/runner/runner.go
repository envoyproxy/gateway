// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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
	"time"

	clusterv3 "github.com/envoyproxy/go-control-plane/envoy/service/cluster/v3"
	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	endpointv3 "github.com/envoyproxy/go-control-plane/envoy/service/endpoint/v3"
	listenerv3 "github.com/envoyproxy/go-control-plane/envoy/service/listener/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/service/route/v3"
	runtimev3 "github.com/envoyproxy/go-control-plane/envoy/service/runtime/v3"
	secretv3 "github.com/envoyproxy/go-control-plane/envoy/service/secret/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/runner"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// XdsServerAddress is the listening address of the xds-server.
	XdsServerAddress = "0.0.0.0"
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

var _ runner.Runner = &Runner{}

type Runner struct {
	xds    *message.Xds
	grpc   *grpc.Server
	cache  cache.SnapshotCacheWithCallbacks
	logger logging.Logger
}

func New(logger logging.Logger, xds *message.Xds) *Runner {
	r := &Runner{
		xds:    xds,
		logger: logger.WithName("xds-server-runner"),
	}
	return r
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentXdsServerRunner)
}

// Start starts the xds-server runner
func (r *Runner) Start(ctx context.Context) (err error) {
	// Set up the gRPC server and register the xDS handler.
	// Create SnapshotCache before start subscribeAndTranslate,
	// prevent panics in case cache is nil.
	cfg := r.tlsConfig(xdsTLSCertFilename, xdsTLSKeyFilename, xdsTLSCaFilename)
	r.grpc = grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)), grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
		MinTime:             15 * time.Second,
		PermitWithoutStream: true,
	}))

	r.cache = cache.NewSnapshotCache(true, r.logger)
	registerServer(serverv3.NewServer(ctx, r.cache, r.cache), r.grpc)

	// Start and listen xDS gRPC Server.
	go r.serveXdsServer(ctx)

	// Start message Subscription.
	go r.subscribeAndTranslate(ctx)
	r.logger.Info("started")
	return
}

func (r *Runner) Reload(serverCfg *config.Server) error {
	r.logger = serverCfg.Logger.WithName("xds-server-runner")
	r.logger.Info("reloaded")

	// TODO: Implement reload logic
	return nil
}

func (r *Runner) serveXdsServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(bootstrap.DefaultXdsServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.logger.Error(err, "failed to listen on address", "address", addr)
		return
	}

	go func() {
		<-ctx.Done()
		r.logger.Info("grpc server shutting down")
		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		r.grpc.Stop()
	}()

	if err = r.grpc.Serve(l); err != nil {
		r.logger.Error(err, "failed to start grpc based xds server")
	}
}

// registerServer registers the given xDS protocol Server with the gRPC
// runtime.
func registerServer(srv serverv3.Server, g *grpc.Server) {
	// register services
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(g, srv)
	secretv3.RegisterSecretDiscoveryServiceServer(g, srv)
	clusterv3.RegisterClusterDiscoveryServiceServer(g, srv)
	endpointv3.RegisterEndpointDiscoveryServiceServer(g, srv)
	listenerv3.RegisterListenerDiscoveryServiceServer(g, srv)
	routev3.RegisterRouteDiscoveryServiceServer(g, srv)
	runtimev3.RegisterRuntimeDiscoveryServiceServer(g, srv)
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentXdsServerRunner), Message: "xds"}, r.xds.Subscribe(ctx),
		func(update message.Update[string, *xdstypes.ResourceVersionTable], errChan chan error) {
			key := update.Key
			val := update.Value

			r.logger.Info("received an update")
			var err error
			if update.Delete {
				err = r.cache.GenerateNewSnapshot(key, nil)
			} else if val != nil && val.XdsResources != nil {
				if r.cache == nil {
					r.logger.Error(err, "failed to init snapshot cache")
					errChan <- err
				} else {
					// Update snapshot cache
					err = r.cache.GenerateNewSnapshot(key, val.XdsResources)
				}
			}
			if err != nil {
				r.logger.Error(err, "failed to generate a snapshot")
				errChan <- err
			}
		},
	)

	r.logger.Info("subscriber shutting down")
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
			NextProtos:   []string{"h2"},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
			MinVersion:   tls.VersionTLS13,
		}, nil
	}

	// Attempt to load certificates and key to catch configuration errors early.
	if _, lerr := loadConfig(); lerr != nil {
		r.logger.Error(lerr, "failed to load certificate and key")
	}
	r.logger.Info("loaded TLS certificate and key")

	return &tls.Config{
		MinVersion: tls.VersionTLS13,
		ClientAuth: tls.RequireAndVerifyClientCert,
		Rand:       rand.Reader,
		GetConfigForClient: func(*tls.ClientHelloInfo) (*tls.Config, error) {
			return loadConfig()
		},
	}
}
