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
	for range r.InfraIR.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		for _, in := range r.InfraIR.LoadAll() {
			switch {
			case in == nil:
				// The resource map is nil at startup.
				r.Logger.Info("infra ir is nil, skipping")
				continue
			case in.Proxy == nil:
				if err := r.mgr.DeleteInfra(ctx, in); err != nil {
					r.Logger.Error(err, "failed to delete infra")
				}
			default:
				// Manage the proxy infra.
				if err := r.mgr.CreateInfra(ctx, in); err != nil {
					r.Logger.Error(err, "failed to create new infra")
				}
			}
		}
		r.Logger.Info("subscriber shutting down")
	}
}
