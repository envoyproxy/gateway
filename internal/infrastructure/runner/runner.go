// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

type Config struct {
	ServerCfg *config.Server
	InfraIR   *message.InfraIR
	Logger    logging.Logger
}

type Runner struct {
	Config
	mgr infrastructure.Manager
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentInfrastructureRunner)
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.ServerCfg.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	if r.ServerCfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeCustom &&
		r.ServerCfg.EnvoyGateway.Provider.Custom.Infrastructure == nil {
		r.Logger.Info("provider is not specified, no provider is available")
		return nil
	}

	r.mgr, err = infrastructure.NewManager(r.ServerCfg)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}

	initInfra := func() {
		go r.subscribeToProxyInfraIR(ctx)

		// Enable global ratelimit if it has been configured.
		if r.ServerCfg.EnvoyGateway.RateLimit != nil {
			go r.enableRateLimitInfra(ctx)
		}
		r.Logger.Info("started")
	}

	// When leader election is active, infrastructure initialization occurs only upon acquiring leadership
	// to avoid multiple EG instances processing envoy proxy infra resources.
	if r.ServerCfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes &&
		!ptr.Deref(r.ServerCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-r.ServerCfg.Elected:
				initInfra()
			}
		}()
		return
	}
	initInfra()
	return
}

func (r *Runner) subscribeToProxyInfraIR(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentInfrastructureRunner), Message: "infra-ir"}, r.InfraIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Infra], errChan chan error) {
			r.Logger.Info("received an update")
			val := update.Value

			if update.Delete {
				if err := r.mgr.DeleteProxyInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to delete infra")
					errChan <- err
				}
			} else {
				// Manage the proxy infra.
				if len(val.Proxy.Listeners) == 0 {
					r.Logger.Info("Infra IR was updated, but no listeners were found. Skipping infra creation.")
					return
				}

				if err := r.mgr.CreateOrUpdateProxyInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to create new infra")
					errChan <- err
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
