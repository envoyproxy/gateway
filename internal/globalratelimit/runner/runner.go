// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"fmt"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

type Config struct {
	config.Server
	XdsIR            *message.XdsIR
	RateLimitInfraIR *message.RateLimitInfraIR
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
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe
	xdsIRCh := r.XdsIR.Subscribe(ctx)
	for ctx.Err() == nil {
		var xdsIRs []*ir.Xds
		snapshot := <-xdsIRCh
		r.Logger.Info("received a notification")
		// Skip translation if state is empty
		if len(snapshot.State) == 0 {
			continue
		}

		for _, value := range snapshot.State {
			xdsIRs = append(xdsIRs, value)
		}

		// Translate to ratelimit infra IR
		result, err := r.translate(xdsIRs)
		if err != nil {
			r.Logger.Error(err, "failed to translate xds ir")
		} else {
			if result == nil {
				r.RateLimitInfraIR.Delete(r.Name())
			} else {
				// Publish ratelimit infra IR
				r.RateLimitInfraIR.Store(r.Name(), result)
			}
		}
	}
	r.Logger.Info("subscriber shutting down")
}

func (r *Runner) translate(xdsIRs []*ir.Xds) (*ir.RateLimitInfra, error) {
	rlInfra := new(ir.RateLimitInfra)

	for _, xdsIR := range xdsIRs {
		for _, listener := range xdsIR.HTTP {
			config := translator.BuildRateLimitServiceConfig(listener)
			if config != nil {
				str, err := translator.GetRateLimitServiceConfigStr(config)
				if err != nil {
					return nil, fmt.Errorf("failed to get rate limit config string: %w", err)
				}
				c := &ir.RateLimitServiceConfig{
					Name:   listener.Name,
					Config: str,
				}
				rlInfra.Configs = append(rlInfra.Configs, c)
			}
		}
	}

	rlInfra.Backend = &ir.RateLimitDBBackend{
		Redis: &ir.RateLimitRedis{
			URL: r.EnvoyGateway.RateLimit.Backend.Redis.URL,
		},
	}

	return rlInfra, nil
}
