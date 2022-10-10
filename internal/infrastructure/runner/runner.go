package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/infrastructure"
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

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) error {
	var err error
	r.Logger = r.Logger.WithValues("runner", r.Name())
	r.mgr, err = infrastructure.NewManager(&r.Config.Server)
	if err != nil {
		r.Logger.Error(err, "failed to create new manager")
	}
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	for snapshot := range r.InfraIR.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		for _, update := range snapshot.Updates {
			val := update.Value

			if update.Delete {
				if err := r.mgr.DeleteInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to delete infra")
				}
			} else {
				// Manage the proxy infra.
				if err := r.mgr.CreateOrUpdateInfra(ctx, val); err != nil {
					r.Logger.Error(err, "failed to create new infra")
				}
			}
		}
	}
	r.Logger.Info("subscriber shutting down")
}
