// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

type Config struct {
	config.Server
	InfraIR *message.InfraIR
}

type Runner struct {
	Config
	mgr infrastructure.Manager
}

func (r *Runner) Name() string {
	return string(v1alpha1.LogComponentInfrastructureRunner)
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) error {
	var err error
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	r.mgr, err = infrastructure.NewManager(&r.Config.Server)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}
	go r.subscribeToProxyInfraIR(ctx)

	// Enable global ratelimit if it has been configured.
	if r.EnvoyGateway.RateLimit != nil {
		go r.enableRateLimitInfra(ctx)
	}

	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeToProxyInfraIR(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(r.InfraIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Infra]) {
			r.Logger.Info("received an update")
			val := update.Value

			if update.Delete {
				if err := r.mgr.DeleteProxyInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to delete infra")
				}
			} else {
				// Manage the proxy infra.
				if err := r.mgr.CreateOrUpdateProxyInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to create new infra")
				}
			}
		},
	)
	r.Logger.Info("infra subscriber shutting down")
}

func (r *Runner) enableRateLimitInfra(ctx context.Context) {
	if err := r.mgr.CreateOrUpdateRateLimitInfra(ctx); err != nil {
		r.Logger.Error(err, "failed to create ratelimit infra")
	}
}
