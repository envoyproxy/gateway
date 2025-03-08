// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"crypto/tls"
	"fmt"
	"math"
	"net"
	"strconv"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachetype "github.com/envoyproxy/go-control-plane/pkg/cache/types"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/crypto"
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

	// Default certificates path for envoy-gateway with Kubernetes provider.
	// rateLimitTLSCertFilepath is the ratelimit tls cert file.
	rateLimitTLSCertFilepath = "/certs/tls.crt"
	// rateLimitTLSKeyFilepath is the ratelimit key file.
	rateLimitTLSKeyFilepath = "/certs/tls.key"
	// rateLimitTLSCACertFilepath is the ratelimit ca cert file.
	rateLimitTLSCACertFilepath = "/certs/ca.crt"

	// TODO: Make these path configurable.
	// Default certificates path for envoy-gateway with Host infrastructure provider.
	localTLSCertFilepath = "/tmp/envoy-gateway/certs/envoy-gateway/tls.crt"
	localTLSKeyFilepath  = "/tmp/envoy-gateway/certs/envoy-gateway/tls.key"
	localTLSCaFilepath   = "/tmp/envoy-gateway/certs/envoy-gateway/ca.crt"
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
	return string(egv1a1.LogComponentGlobalRateLimitRunner)
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
	tlsConfig, err := r.loadTLSConfig()
	if err != nil {
		return fmt.Errorf("failed to load TLS config: %w", err)
	}
	r.Logger.Info("loaded TLS certificate and key")

	r.grpc = grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))

	r.cache = cachev3.NewSnapshotCache(false, cachev3.IDHash{}, r.Logger.Sugar())

	// Register xDS Config server.
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(r.grpc, serverv3.NewServer(ctx, r.cache, serverv3.CallbackFuncs{}))

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

func buildXDSResourceFromCache(rateLimitConfigsCache map[string][]cachetype.Resource) types.XdsResources {
	xdsResourcesToUpdate := types.XdsResources{}
	for _, xdsR := range rateLimitConfigsCache {
		xdsResourcesToUpdate[resourcev3.RateLimitConfigType] = append(xdsResourcesToUpdate[resourcev3.RateLimitConfigType], xdsR...)
	}

	return xdsResourcesToUpdate
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// rateLimitConfigsCache is a cache of the rate limit config, which is keyed by the xdsIR key.
	rateLimitConfigsCache := map[string][]cachetype.Resource{}

	// Subscribe to resources.
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentGlobalRateLimitRunner), Message: "xds-ir"}, r.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds], errChan chan error) {
			r.Logger.Info("received a notification")

			if update.Delete {
				delete(rateLimitConfigsCache, update.Key)
				r.updateSnapshot(ctx, buildXDSResourceFromCache(rateLimitConfigsCache))
			} else {
				// Translate to ratelimit xDS Config.
				rvt, err := r.translate(update.Value)
				if err != nil {
					r.Logger.Error(err, "failed to translate an updated xds-ir to ratelimit xDS Config")
					errChan <- err
				}

				// Update ratelimit xDS config cache.
				if rvt != nil {
					// Build XdsResources to use for the snapshot update from the cache.
					rateLimitConfigsCache[update.Key] = rvt.XdsResources[resourcev3.RateLimitConfigType]
					r.updateSnapshot(ctx, buildXDSResourceFromCache(rateLimitConfigsCache))
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) translate(xdsIR *ir.Xds) (*types.ResourceVersionTable, error) {
	resourceVT := new(types.ResourceVersionTable)

	// Iterate over each HTTP listener
	for _, listener := range xdsIR.HTTP {
		// Build the rate limit service config for this listener
		configs := translator.BuildRateLimitServiceConfig(listener)

		// Iterate through each config
		for _, cfg := range configs {
			// If the config is not nil, add it to the xDS Config resources
			if cfg != nil {
				if err := resourceVT.AddXdsResource(resourcev3.RateLimitConfigType, cfg); err != nil {
					return nil, err
				}
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

func (r *Runner) loadTLSConfig() (tlsConfig *tls.Config, err error) {
	switch {
	case r.EnvoyGateway.Provider.IsRunningOnKubernetes():
		tlsConfig, err = crypto.LoadTLSConfig(rateLimitTLSCertFilepath, rateLimitTLSKeyFilepath, rateLimitTLSCACertFilepath)
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
