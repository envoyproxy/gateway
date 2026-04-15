// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"sync"

	"github.com/telepresenceio/watchable"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
)

type Config struct {
	config.Server
	InfraIR      *message.InfraIR
	RunnerErrors *message.RunnerErrors
}

type Runner struct {
	Config
	mgr infrastructure.Manager

	rateLimitInfraOnce sync.Once
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
	errNotifier := message.RunnerErrorNotifier{RunnerName: r.Name(), RunnerErrors: r.RunnerErrors}

	var infraClient client.Client
	if r.EnvoyGateway.Provider.Type == egv1a1.ProviderTypeKubernetes {
		select {
		case infraClient = <-r.ProviderClient:
		case <-ctx.Done():
			err = ctx.Err()
			r.Logger.Error(err, "failed to create new manager")
			return err
		}
	}

	r.mgr, err = infrastructure.NewManager(ctx, &r.Server, r.Logger, errNotifier, infraClient)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
		return err
	}

	// This is a blocking function that subscribes to the infraIR and initializes the infrastructure.
	subscribeInitInfraAndCloseInfraIRMessage := func() {
		// Subscribe and Close in same goroutine to avoid race condition.
		sub := r.InfraIR.Subscribe(ctx)
		go r.updateProxyInfraFromSubscription(ctx, sub)

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
	return err
}

func (r *Runner) updateProxyInfraFromSubscription(ctx context.Context, sub <-chan watchable.Snapshot[string, *ir.Infra]) {
	// Subscribe to resources
	message.HandleSubscription(
		r.Logger,
		message.Metadata{Runner: r.Name(), Message: message.InfraIRMessageName}, sub,
		func(update message.Update[string, *ir.Infra], errChan chan error) {
			// Check if context is done before logging to avoid writing to test output after test completes
			select {
			case <-ctx.Done():
				return
			default:
			}
			r.Logger.Info("received an update", "key", update.Key, "delete", update.Delete)

			// Since the rate limit infra is shared by all proxy infra, we only need to create it once when the first IR
			// update is received.
			r.rateLimitInfraOnce.Do(func() {
				if r.EnvoyGateway.RateLimit != nil {
					if err := r.mgr.CreateOrUpdateRateLimitInfra(ctx); err != nil {
						r.Logger.Error(err, "failed to create ratelimit infra")
					}
				} else {
					if err := r.mgr.DeleteRateLimitInfra(ctx); err != nil {
						r.Logger.Error(err, "failed to delete ratelimit infra")
					}
				}
			})

			val := update.Value

			if update.Delete {
				if err := r.mgr.DeleteProxyInfra(ctx, val); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						r.Logger.Error(err, "failed to delete infra")
					}
					errChan <- err
				}
			} else {
				// Manage the proxy infra.
				// Skip creating or updating infra if the Infra IR without any listener.
				// e.g.https://github.com/envoyproxy/gateway/issues/3044 --- Invalid Listener
				//     https://github.com/envoyproxy/gateway/issues/7735 --- Invalid EnvoyProxy
				if len(val.Proxy.Listeners) == 0 {
					select {
					case <-ctx.Done():
						return
					default:
						r.Logger.Info("Infra IR was updated, but no listeners were found. Skipping infra creation.")
					}
					return
				}

				if err := r.mgr.CreateOrUpdateProxyInfra(ctx, val); err != nil {
					select {
					case <-ctx.Done():
						return
					default:
						r.Logger.Error(err, "failed to create new infra")
					}
					errChan <- err
				}
			}
		},
	)
	select {
	case <-ctx.Done():
		return
	default:
		r.Logger.Info("infra subscriber shutting down")
	}
}
