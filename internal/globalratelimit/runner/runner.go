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

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/runner"
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

func Register(resources Resources, globalConfig config.Server) {
	runner.Manager().Register(New(resources, globalConfig), runner.RootParentRunner)
}

func New(resources Resources, globalConfig config.Server) *ratelimitRunner {
	return &ratelimitRunner{runner.New(string(v1alpha1.LogComponentGlobalRateLimitRunner), resources, globalConfig)}
}

type Resources struct {
	XdsIR           *message.XdsIR
	grpc            *grpc.Server
	cache           cachev3.SnapshotCache
	snapshotVersion int64
}

type ratelimitRunner struct {
	*runner.GenericRunner[Resources]
}

func (r *ratelimitRunner) SubscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources.
	message.HandleSubscription(r.Resources.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds]) {
			r.Logger.Info("received a notification")

			if update.Delete {
				if err := r.addNewSnapshot(ctx, nil); err != nil {
					r.Logger.Error(err, "failed to update the config snapshot")
				}
			} else {
				// Translate to ratelimit xDS Config.
				rvt := r.translate(update.Value)

				// Update ratelimit xDS config cache.
				if rvt != nil {
					r.updateSnapshot(ctx, rvt.XdsResources)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}

// Start starts the infrastructure runner
func (r *ratelimitRunner) Start(ctx context.Context) error {
	r.Init(ctx)

	// Set up the gRPC server and register the xDS handler.
	// Create SnapshotCache before start subscribeAndTranslate,
	// prevent panics in case cache is nil.
	cfg := r.tlsConfig(rateLimitTLSCertFilename, rateLimitTLSKeyFilename, rateLimitTLSCACertFilename)
	r.Resources.grpc = grpc.NewServer(grpc.Creds(credentials.NewTLS(cfg)))

	r.Resources.cache = cachev3.NewSnapshotCache(false, cachev3.IDHash{}, r.Logger.Sugar())

	// Register xDS Config server.
	cb := &test.Callbacks{}
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(r.Resources.grpc, serverv3.NewServer(ctx, r.Resources.cache, cb))

	// Start and listen xDS gRPC config Server.
	go r.serverXdsConfigServer(ctx)

	// Start message Subscription.
	go r.SubscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

func (r *ratelimitRunner) ShutDown(ctx context.Context) {
	r.Resources.XdsIR.Close()
}

func (r *ratelimitRunner) serverXdsConfigServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsGrpcSotwConfigServerAddress, strconv.Itoa(ratelimit.XdsGrpcSotwConfigServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", "address", addr)
		return
	}
	if err = r.Resources.grpc.Serve(l); err != nil {
		r.Logger.Error(err, "failed to start grpc based xds config server")
	}

	<-ctx.Done()
	r.Logger.Info("grpc config server shutting down")
	r.Resources.grpc.Stop()
}

func (r *ratelimitRunner) translate(xdsIR *ir.Xds) *types.ResourceVersionTable {
	resourceVT := new(types.ResourceVersionTable)

	for _, listener := range xdsIR.HTTP {
		cfg := translator.BuildRateLimitServiceConfig(listener)
		if cfg != nil {
			// Add to xDS Config resources.
			resourceVT.AddXdsResource(resourcev3.RateLimitConfigType, cfg)
		}
	}
	return resourceVT
}

func (r *ratelimitRunner) updateSnapshot(ctx context.Context, resource types.XdsResources) {
	if r.Resources.cache == nil {
		r.Logger.Error(nil, "failed to init the snapshot cache")
		return
	}

	if err := r.addNewSnapshot(ctx, resource); err != nil {
		r.Logger.Error(err, "failed to update the snapshot cache")
	}
}

func (r *ratelimitRunner) addNewSnapshot(ctx context.Context, resource types.XdsResources) error {
	// Increment the snapshot version.
	if r.Resources.snapshotVersion == math.MaxInt64 {
		r.Resources.snapshotVersion = 0
	}
	r.Resources.snapshotVersion++

	snapshot, err := cachev3.NewSnapshot(fmt.Sprintf("%d", r.Resources.snapshotVersion), resource)
	if err != nil {
		return fmt.Errorf("failed to generate a config snapshot: %w", err)
	}
	err = r.Resources.cache.SetSnapshot(ctx, ratelimit.InfraName, snapshot)
	if err != nil {
		return fmt.Errorf("failed to set a config snapshot: %w", err)
	}
	return nil
}

func (r *ratelimitRunner) tlsConfig(cert, key, ca string) *tls.Config {
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
