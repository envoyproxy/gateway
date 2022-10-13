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
	first := true
	for snapshot := range r.XdsIR.Subscribe(ctx) {
		r.Logger.Info("received a notification")
		if first {
			first = false
			for key, val := range snapshot.State {
				result, err := translator.Translate(val)
				if err != nil {
					r.Logger.Error(err, "failed to translate xds ir")
				} else {
					// Publish
					r.Xds.Store(key, result)
				}
			}
		} else {
			for _, update := range snapshot.Updates {
				r.Logger.Info("received an update")
				key := update.Key
				val := update.Value

				if update.Delete {
					r.Xds.Delete(key)
				} else {
					// Translate to xds resources
					result, err := translator.Translate(val)
					if err != nil {
						r.Logger.Error(err, "failed to translate xds ir")
					} else {
						// Publish
						r.Xds.Store(key, result)
					}
				}
			}
		}
	}
	r.Logger.Info("subscriber shutting down")
}
