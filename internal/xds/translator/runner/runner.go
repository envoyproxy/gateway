package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

type Config struct {
	config.Server
	XdsIR        *message.XdsIR
	XdsResources *message.XdsResources
}

type Runner struct {
	Config
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return "xds-translator"
}

// Start starts the xds-translator runner
func (r *Runner) Start(ctx context.Context) error {
	log := r.Logger.WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)

	<-ctx.Done()
	log.Info("shutting down")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	for range r.XdsIR.Subscribe(ctx) {
		ir := r.XdsIR.Get()
		// Translate to xds resources
		result, err := translator.Translate(ir)
		if err != nil {
			r.Logger.Error(err, "failed to translate xds ir")
		} else {
			// Publish
			// There should always be a single element in the map
			// Use the service name as the key for now
			xdsResources := result.GetXdsResources()
			r.XdsResources.Store(r.Name(), &xdsResources)
		}
	}
}
