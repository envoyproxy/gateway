package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

type Config struct {
	config.Server
	XdsIR *message.XdsIR
	Xds   *message.Xds
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
	r.Logger = r.Logger.WithValues("runner", r.Name())
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	for range r.XdsIR.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		ir := r.XdsIR.Get()
		if ir == nil {
			r.Logger.Info("xds ir is nil, skipping")
			continue
		}
		// Translate to xds resources
		result, err := translator.Translate(ir)
		if err != nil {
			r.Logger.Error(err, "failed to translate xds ir")
		} else {
			// Publish
			// There should always be a single element in the map
			// Use the service name as the key for now
			r.Xds.Store(r.Name(), result)
		}
	}
	r.Logger.Info("subscriber shutting down")
}
