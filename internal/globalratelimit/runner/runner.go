// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"
	"math"
	"net"
	"strconv"

	discoveryv3 "github.com/envoyproxy/go-control-plane/envoy/service/discovery/v3"
	cachev3 "github.com/envoyproxy/go-control-plane/pkg/cache/v3"
	resourcev3 "github.com/envoyproxy/go-control-plane/pkg/resource/v3"
	serverv3 "github.com/envoyproxy/go-control-plane/pkg/server/v3"
	"github.com/envoyproxy/go-control-plane/pkg/test/v3"
	"google.golang.org/grpc"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/cache"
	"github.com/envoyproxy/gateway/internal/xds/translator"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// XdsGrpcSotwConfigServerAddress is the listening address of the ratelimit xDS config server.
	XdsGrpcSotwConfigServerAddress = "0.0.0.0"
)

type Config struct {
	config.Server
	XdsIR            *message.XdsIR
	RateLimitInfraIR *message.RateLimitInfraIR
	grpc             *grpc.Server
	cache            cachev3.SnapshotCache
	snapshotVersion  int64
}

type Runner struct {
	Config
}

func (r *Runner) Name() string {
	return "global-ratelimit"
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) error {
	r.Logger = r.Logger.WithValues("runner", r.Name())

	// Set up the gRPC server for ratelimit xDS Config.
	r.grpc = grpc.NewServer()

	r.cache = cachev3.NewSnapshotCache(false, cachev3.IDHash{}, cache.NewLogrWrapper(r.Logger))

	// Register xDS Config server.
	cb := &test.Callbacks{}
	discoveryv3.RegisterAggregatedDiscoveryServiceServer(r.grpc, serverv3.NewServer(ctx, r.cache, cb))

	// Start and listen xDS gRPC config Server.
	go r.serverXdsConfigServer(ctx)

	// Start message Subscription.
	go r.subscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

func (r *Runner) serverXdsConfigServer(ctx context.Context) {
	addr := net.JoinHostPort(XdsGrpcSotwConfigServerAddress, strconv.Itoa(ratelimit.XdsGrpcSotwConfigServerPort))
	l, err := net.Listen("tcp", addr)
	if err != nil {
		r.Logger.Error(err, "failed to listen on address", "address", addr)
		return
	}
	if err = r.grpc.Serve(l); err != nil {
		r.Logger.Error(err, "failed to start grpc based xds config server")
	}

	<-ctx.Done()
	r.Logger.Info("grpc config server shutting down")
	r.grpc.Stop()
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources.
	message.HandleSubscription(r.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds]) {
			r.Logger.Info("received a notification")

			if update.Delete {
				if err := r.addNewSnapshot(ctx, nil); err != nil {
					r.Logger.Error(err, "failed to update the config snapshot")
				}
			} else {
				// Translate to ratelimit infra IR and ratelimit xDS Config.
				rlIR, rvt, err := r.translate(update.Value)
				if err != nil {
					r.Logger.Error(err, "failed to translate xds ir and config")
					return
				}

				// Update ratelimit xDS config cache.
				if rvt != nil {
					r.updateSnapshot(ctx, rvt.XdsResources)
				}

				// Publish ratelimit infra IR.
				if rlIR == nil {
					r.RateLimitInfraIR.Delete(r.Name())
				} else {
					r.RateLimitInfraIR.Store(r.Name(), rlIR)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) translate(xdsIR *ir.Xds) (*ir.RateLimitInfra, *types.ResourceVersionTable, error) {
	ratelimitInfra := new(ir.RateLimitInfra)
	resourceVT := new(types.ResourceVersionTable)

	for _, listener := range xdsIR.HTTP {
		cfg := translator.BuildRateLimitServiceConfig(listener)
		if cfg != nil {
			// Add to xDS Config resources.
			resourceVT.AddXdsResource(resourcev3.RateLimitConfigType, cfg)

			ratelimitInfra.ServiceNames = append(ratelimitInfra.ServiceNames, listener.Name)
		}
	}
	return ratelimitInfra, resourceVT, nil
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
