// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"github.com/telepresenceio/watchable"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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

// Close implements Runner interface.
func (r *Runner) Close() error {
	return r.mgr.Close()
}

// Name implements Runner interface.
func (r *Runner) Name() string {
	return string(egv1a1.LogComponentInfrastructureRunner)
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.Logger.WithName(r.Name()).WithValues("runner", r.Name())
	if r.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeCustom &&
		r.EnvoyGateway.Provider.Custom.Infrastructure == nil {
		r.Logger.Info("provider is not specified, no infrastructure is available")
		return nil
	}

	r.mgr, err = infrastructure.NewManager(ctx, &r.Server, r.Logger)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}

	// This is a blocking function that subscribes to the infraIR and initializes the infrastructure.
	subscribeInitInfraAndCloseInfraIRMessage := func() {
		// Subscribe and Close in same goroutine to avoid race condition.
		sub := r.InfraIR.Subscribe(ctx)
		go r.updateProxyInfraFromSubscription(ctx, sub)

		// Enable global ratelimit if it has been configured.
		if r.EnvoyGateway.RateLimit != nil {
			go r.enableRateLimitInfra(ctx)
		} else {
			// Delete the ratelimit infra if it exists.
			go func() {
				if err := r.mgr.DeleteRateLimitInfra(ctx); err != nil {
					r.Logger.Error(err, "failed to delete ratelimit infra")
				}
			}()
		}
		r.Logger.Info("started")
		<-ctx.Done()
		r.InfraIR.Close()
		r.Logger.Info("shutting down")
	}

	// When leader election is active, infrastructure initialization occurs only upon acquiring leadership
	// to avoid multiple EG instances processing envoy proxy infra resources.
	if r.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes &&
		!ptr.Deref(r.EnvoyGateway.Provider.Kubernetes.LeaderElection.Disable, false) {
		go func() {
			select {
			case <-ctx.Done():
				// As a follower EG instance close infraIR when the context is done.
				r.InfraIR.Close()
				return
			case <-r.Elected:
				// As a leader EG instance subscribe to infraIR to initialize the infrastructure and Close when the context is done.
				subscribeInitInfraAndCloseInfraIRMessage()
			}
		}()
	} else {
		// Since leader election is disabled subscribe to infraIR to initialize the infrastructure and Close when the context is done.
		go subscribeInitInfraAndCloseInfraIRMessage()
	}
	return
}

func (r *Runner) updateProxyInfraFromSubscription(ctx context.Context, sub <-chan watchable.Snapshot[string, *ir.Infra]) {
	// Subscribe to resources
	message.HandleSubscription(message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, sub,
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
