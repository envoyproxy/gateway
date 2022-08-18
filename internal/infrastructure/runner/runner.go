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
	mgr *infrastructure.Manager
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
	log := r.Logger.WithValues("runner", r.Name())
	r.mgr, err = infrastructure.NewManager(&r.Config.Server)
	if err != nil {
		log.Error(err, "failed to create new manager")
	}
	go r.subscribeAndTranslate(ctx)

	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	for range r.InfraIR.Subscribe(ctx) {
		in := r.InfraIR.Get()
		// Provision infra
		if err := r.mgr.CreateInfra(ctx, in); err != nil {
			r.Logger.Error(err, "failed to create new infra")
		}
	}

	r.Logger.Info("subscriber shutting down")
}
