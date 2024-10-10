// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"sync"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/runner"
)

var _ runner.Runner = &Runner{}

type Runner struct {
	serverCfg *config.Server
	infraIR   *message.InfraIR
	logger    logging.Logger

	mgr infrastructure.Manager
	m   sync.Mutex
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentInfrastructureRunner)
}

func New(cfg *config.Server, infra *message.InfraIR) *Runner {
	r := &Runner{serverCfg: cfg, infraIR: infra}
	r.logger = r.serverCfg.Logger.WithName(r.Name())

	return r
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.m.Lock()
	defer r.m.Unlock()

	if r.serverCfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeCustom &&
		r.serverCfg.EnvoyGateway.Provider.Custom.Infrastructure == nil {
		r.logger.Info("provider is not specified, no provider is available")
		return nil
	}

	r.mgr, err = infrastructure.NewManager(r.serverCfg)
	if err != nil {
		r.logger.Error(err, "failed to create new manager")
		return err
	}

	initInfra := func() {
		go r.subscribeToProxyInfraIR(ctx)

		// Enable global ratelimit if it has been configured.
		go r.reconcileRateLimitInfra(ctx)

		r.logger.Info("started")
	}

	// When leader election is active, infrastructure initialization occurs only upon acquiring leadership
	// to avoid multiple EG instances processing envoy proxy infra resources.
	if r.serverCfg.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes &&
		!ptr.Deref(r.serverCfg.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		go func() {
			select {
			case <-ctx.Done():
				return
			case <-r.serverCfg.Elected:
				initInfra()
			}
		}()
		return
	}
	initInfra()
	return
}

func (r *Runner) Reload(serverCfg *config.Server) error {
	r.m.Lock()
	defer r.m.Unlock()

	r.serverCfg = serverCfg
	r.logger = serverCfg.Logger.WithName(r.Name())

	var err error
	r.mgr, err = infrastructure.NewManager(r.serverCfg)
	if err != nil {
		r.logger.Error(err, "failed to create new manager")
		return err
	}

	r.logger.Info("reloaded")

	r.reconcileRateLimitInfra(context.TODO())

	// TODO: how to reconcile all proxy infra.
	return nil
}

func (r *Runner) subscribeToProxyInfraIR(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentInfrastructureRunner), Message: "infra-ir"}, r.infraIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Infra], errChan chan error) {
			r.logger.Info("received an update")
			val := update.Value

			r.m.Lock()
			defer r.m.Unlock()

			if update.Delete {
				if err := r.mgr.DeleteProxyInfra(ctx, val); err != nil {
					r.logger.Error(err, "failed to delete infra")
					errChan <- err
				}
			} else {
				// Manage the proxy infra.
				if len(val.Proxy.Listeners) == 0 {
					r.logger.Info("Infra IR was updated, but no listeners were found. Skipping infra creation.")
					return
				}

				if err := r.mgr.CreateOrUpdateProxyInfra(ctx, val); err != nil {
					r.logger.Error(err, "failed to create new infra")
					errChan <- err
				}
			}
		},
	)
	r.logger.Info("infra subscriber shutting down")
}

func (r *Runner) reconcileRateLimitInfra(ctx context.Context) {
	if r.serverCfg.EnvoyGateway.RateLimit != nil {
		if err := r.mgr.CreateOrUpdateRateLimitInfra(ctx); err != nil {
			r.logger.Error(err, "failed to create ratelimit infra")
		}

		return
	}

	// TODO: r.mgr.DeleteRateLimitInfra will panic with nil RateLimit
}
