package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

type Config struct {
	config.Server
	InfraIR *message.InfraIR
}

type Runner struct {
	Config
}

func (r *Runner) Name() string {
	return "infrastructure"
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

// Start starts the infrastructure runner
func (r *Runner) Start(ctx context.Context) error {
	log := r.Logger.WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)

	<-ctx.Done()
	log.Info("shutting down")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	for range r.InfraIR.Subscribe(ctx) {
		// s.InfraIR.Get()
		// TODO: Provision infra
		// infrastructure.Translate(ctx, ir)
	}
}
