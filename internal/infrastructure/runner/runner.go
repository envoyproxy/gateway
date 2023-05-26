// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

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
	return "infrastructure"
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Manager returns the manager of infra.
func (r *Runner) Manager() infrastructure.Manager {
	return r.mgr
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) error {
	var err error
	r.Logger = r.Logger.WithValues("runner", r.Name())
	r.mgr, err = infrastructure.NewManager(&r.Config.Server)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}
	go r.subscribeToProxyInfraIR(ctx)

	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeToProxyInfraIR(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(r.InfraIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Infra]) {
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
