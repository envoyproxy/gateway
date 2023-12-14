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
	"math"
	"net"
	"os"
	"strconv"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// XdsGrpcSotwConfigServerAddress is the listening address of the ratelimit xDS config server.
	XdsGrpcSotwConfigServerAddress = "0.0.0.0"
	// rateLimitTLSCertFilename is the ratelimit tls cert file.
	rateLimitTLSCertFilename = "/certs/tls.crt"
	// rateLimitTLSKeyFilename is the ratelimit key file.
	rateLimitTLSKeyFilename = "/certs/tls.key"
	// rateLimitTLSCACertFilename is the ratelimit ca cert file.
	rateLimitTLSCACertFilename = "/certs/ca.crt"
)

type Config struct {
	config.Server
	XdsIR           *message.XdsIR
	grpc            *grpc.Server
	cache           cachev3.SnapshotCache
	snapshotVersion int64
}

type Runner struct {
	Config
}

func (r *Runner) Name() string {
	return string(v1alpha1.LogComponentGlobalRateLimitRunner)
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())

	// Set up the gRPC server and register the xDS handler.
	// Create SnapshotCache before start subscribeAndTranslate,
	// prevent panics in case cache is nil.
	cfg := r.tlsConfig(rateLimitTLSCertFilename, rateLimitTLSKeyFilename, rateLimitTLSCACertFilename)
	r.grpc = grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))

	r.cache = cachev3.NewSnapshotCache(false, cachev3.IDHash{}, r.Logger.Sugar())

	// Register xDS Config server.
	cb := &test.Callbacks{}
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(r.grpc, serverv3.NewServer(ctx, r.cache, cb))

	// Start and listen xDS gRPC config Server.
	go r.serveXdsConfigServer(ctx)

	// Start message Subscription.
	go r.subscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return
}

func (r *Runner) serveXdsConfigServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsGrpcSotwConfigServerAddress, strconv.Itoa(ratelimit.XdsGrpcSotwConfigServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", "address", addr)
		return
	}

	go func() {
		<-ctx.Done()
		r.Logger.Info("grpc server shutting down")
		r.grpc.Stop()
	}()

	if err = r.grpc.Serve(l); err != nil {
		r.Logger.Error(err, "failed to start grpc based xds config server")
	}
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources.
	message.HandleSubscription(message.Metadata{Runner: string(v1alpha1.LogComponentGlobalRateLimitRunner), Message: "xds-ir"}, r.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds], errChan chan error) {
			r.Logger.Info("received a notification")

			if update.Delete {
				if err := r.addNewSnapshot(ctx, nil); err != nil {
					r.Logger.Error(err, "failed to update the config snapshot")
					errChan <- err
				}
			} else {
				// Translate to ratelimit xDS Config.
				rvt, err := r.translate(update.Value)
				if err != nil {
					r.Logger.Error(err, err.Error())
				}

				// Update ratelimit xDS config cache.
				if rvt != nil {
					r.updateSnapshot(ctx, rvt.XdsResources)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) translate(xdsIR *ir.Xds) (*types.ResourceVersionTable, error) {
	resourceVT := new(types.ResourceVersionTable)

	for _, listener := range xdsIR.HTTP {
		cfg := translator.BuildRateLimitServiceConfig(listener)
		if cfg != nil {
			// Add to xDS Config resources.
			if err := resourceVT.AddXdsResource(resourcev3.RateLimitConfigType, cfg); err != nil {
				return nil, err
			}
		}
	}
	return resourceVT, nil
}

func (r *Runner) updateSnapshot(ctx context.Context, resource types.XdsResources) {
	if r.cache == nil {
		r.Logger.Error(nil, "failed to init the snapshot cache")
		return
	}

	if err := r.addNewSnapshot(ctx, resource); err != nil {
		r.Logger.Error(err, "failed to update the snapshot cache")
	}
}

func (r *Runner) addNewSnapshot(ctx context.Context, resource types.XdsResources) error {
	// Increment the snapshot version.
	if r.snapshotVersion == math.MaxInt64 {
		r.snapshotVersion = 0
	}
	r.snapshotVersion++

	snapshot, err := cachev3.NewSnapshot(fmt.Sprintf("%d", r.snapshotVersion), resource)
	if err != nil {
		return fmt.Errorf("failed to generate a config snapshot: %w", err)
	}
	err = r.cache.SetSnapshot(ctx, ratelimit.InfraName, snapshot)
	if err != nil {
		return fmt.Errorf("failed to set a config snapshot: %w", err)
	}
	return nil
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
