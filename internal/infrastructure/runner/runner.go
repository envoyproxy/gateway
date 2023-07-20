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
	"github.com/envoyproxy/gateway/internal/runner"
)

func Register(resources Resources, globalConfig config.Server) {
	runner.Manager().Register(New(resources, globalConfig), runner.RootParentRunner)
}

type infraRunner struct {
	*runner.GenericRunner[Resources]
	mgr infrastructure.Manager
}

func New(resources Resources, globalConfig config.Server) *infraRunner {
	return &infraRunner{GenericRunner: runner.New(string(v1alpha1.LogComponentInfrastructureRunner), resources, globalConfig)}
}

type Resources struct {
	InfraIR *message.InfraIR
}

// Start starts the infrastructure runner
func (r *infraRunner) Start(ctx context.Context) (err error) {
	r.Init(ctx)
	r.mgr, err = infrastructure.NewManager(&r.Server)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}
	go r.SubscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

// Start starts the infrastructure runner
func (r *infraRunner) ShutDown(ctx context.Context) {
	r.Resources.InfraIR.Close()
}

func (r *infraRunner) SubscribeAndTranslate(ctx context.Context) {
	// Enable global ratelimit if it has been configured.
	if r.EnvoyGateway.RateLimit != nil {
		go r.manageRateLimitInfra(ctx)
	}

	// Subscribe to resources
	message.HandleSubscription(r.Resources.InfraIR.Subscribe(ctx),
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

func (r *infraRunner) manageRateLimitInfra(ctx context.Context) {
	if err := r.mgr.CreateOrUpdateRateLimitInfra(ctx); err != nil {
		r.Logger.Error(err, "failed to create ratelimit infra")
	}

	<-ctx.Done()
	r.Logger.Info("deleting ratelimit infra")
	if err := r.mgr.DeleteRateLimitInfra(ctx); err != nil {
		r.Logger.Error(err, "failed to delete ratelimit infra")
	} else {
		r.Logger.Info("ratelimit infra deleted")
	}
}
