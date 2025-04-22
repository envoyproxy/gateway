// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
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
	"github.com/telepresenceio/watchable"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/keepalive"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	xdstypes "github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// XdsServerAddress is the listening address of the xds-server.
	XdsServerAddress = "0.0.0.0"

	// Default certificates path for envoy-gateway with Kubernetes provider.
	// xdsTLSCertFilepath is the fully qualified path of the file containing the
	// xDS server TLS certificate.
	xdsTLSCertFilepath = "/certs/tls.crt"
	// xdsTLSKeyFilepath is the fully qualified path of the file containing the
	// xDS server TLS key.
	xdsTLSKeyFilepath = "/certs/tls.key"
	// xdsTLSCaFilepath is the fully qualified path of the file containing the
	// xDS server trusted CA certificate.
	xdsTLSCaFilepath = "/certs/ca.crt"

	// TODO: Make these path configurable.
	// Default certificates path for envoy-gateway with Host infrastructure provider.
	localTLSCertFilepath = "/tmp/envoy-gateway/certs/envoy-gateway/tls.crt"
	localTLSKeyFilepath  = "/tmp/envoy-gateway/certs/envoy-gateway/tls.key"
	localTLSCaFilepath   = "/tmp/envoy-gateway/certs/envoy-gateway/ca.crt"
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
	return string(egv1a1.LogComponentXdsServerRunner)
}

// Close implements Runner interface.
func (r *Runner) Close() error { return nil }

// Start starts the xds-server runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())

	// Set up the gRPC server and register the xDS handler.
	// Create SnapshotCache before start subscribeAndTranslate,
	// prevent panics in case cache is nil.
	tlsConfig, err := r.loadTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %w", err)
	}
	r.Logger.Info("loaded TLS certificate and key")

	grpcOpts := []grpc.ServerOption{
		grpc.Creds(credentials.NewTLS(tlsConfig)),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             15 * time.Second,
			PermitWithoutStream: true,
		}),
	}

	// When GatewayNamespaceMode is enabled, we will use Service Account JWT tokens to authenticate envoy proxy infra and xds server.
	if r.EnvoyGateway.GatewayNamespaceMode() {
		r.Logger.Info("gatewayNamespaceMode is enabled, setting up JWTAuthInterceptor")

		publicKey, err := r.loadKubernetesPublicKey()
		if err != nil {
			return fmt.Errorf("failed to load Kubernetes public key: %w", err)
		}

		jwtInterceptor := NewJWTAuthInterceptor(
			publicKey,
			"https://kubernetes.default.svc",
		)

		ca, err := os.ReadFile(xdsTLSCaFilepath)
		if err != nil {
			return err
		}
		certPool := x509.NewCertPool()
		if !certPool.AppendCertsFromPEM(ca) {
			return err
		}

		tlsCreds := credentials.NewTLS(&tls.Config{
			RootCAs:    certPool,
			NextProtos: []string{"h2"},
		})

		grpcOpts = []grpc.ServerOption{
			grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
				MinTime:             15 * time.Second,
				PermitWithoutStream: true,
			}),
			grpc.Creds(tlsCreds),
			grpc.StreamInterceptor(jwtInterceptor.Stream()),
		}
	}

	r.grpc = grpc.NewServer(grpcOpts...)
	r.cache = cache.NewSnapshotCache(true, r.Logger)
	registerServer(serverv3.NewServer(ctx, r.cache, r.cache), r.grpc)

	// Start and listen xDS gRPC Server.
	go r.serveXdsServer(ctx)

	// Start message Subscription.
	// Do not call .Subscribe() inside Goroutine since it is supposed to be called from the same
	// Goroutine where Close() is called.
	xdsSubCh := r.Xds.Subscribe(ctx)
	go r.subscribeAndTranslate(xdsSubCh)
	r.Logger.Info("started")
	return
}

func (r *Runner) serveXdsServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsServerAddress, strconv.Itoa(bootstrap.DefaultXdsServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", "address", addr)
		return
	}

	go func() {
		<-ctx.Done()
		r.Logger.Info("grpc server shutting down")
		// We don't use GracefulStop here because envoy
		// has long-lived hanging xDS requests. There's no
		// mechanism to make those pending requests fail,
		// so we forcibly terminate the TCP sessions.
		r.grpc.Stop()
	}()

	if err = r.grpc.Serve(l); err != nil {
		r.Logger.Error(err, "failed to start grpc based xds server")
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

func (r *Runner) subscribeAndTranslate(sub <-chan watchable.Snapshot[string, *xdstypes.ResourceVersionTable]) {
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentXdsServerRunner), Message: "xds"}, sub,
		func(update message.Update[string, *xdstypes.ResourceVersionTable], errChan chan error) {
			key := update.Key
			val := update.Value

			r.Logger.Info("received an update")
			var err error
			if update.Delete {
				err = r.cache.GenerateNewSnapshot(key, nil)
			} else if val != nil && val.XdsResources != nil {
				if r.cache == nil {
					r.Logger.Error(err, "failed to init snapshot cache")
					errChan <- err
				} else {
					// Update snapshot cache
					err = r.cache.GenerateNewSnapshot(key, val.XdsResources)
				}
			}
			if err != nil {
				r.Logger.Error(err, "failed to generate a snapshot")
				errChan <- err
			}
		},
	)

	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) loadTLSConfig() (tlsConfig *tls.Config, err error) {
	switch {
	case r.EnvoyGateway.Provider.IsRunningOnKubernetes():
		tlsConfig, err = crypto.LoadTLSConfig(xdsTLSCertFilepath, xdsTLSKeyFilepath, xdsTLSCaFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create tls config: %w", err)
		}

	case r.EnvoyGateway.Provider.IsRunningOnHost():
		tlsConfig, err = crypto.LoadTLSConfig(localTLSCertFilepath, localTLSKeyFilepath, localTLSCaFilepath)
		if err != nil {
			return nil, fmt.Errorf("failed to create tls config: %w", err)
		}

	default:
		return nil, fmt.Errorf("no valid tls certificates")
	}
	return
}

// loadKubernetesPublicKey loads the Kubernetes API server's public key for validating Service Account tokens.
func (r *Runner) loadKubernetesPublicKey() (*rsa.PublicKey, error) {
	const publicKeyPath = "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"

	pemData, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %w", err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil || block.Type != "CERTIFICATE" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	rsaPubKey, ok := cert.PublicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("not an RSA public key")
	}

	return rsaPubKey, nil
}
