// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	infra "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

type Config struct {
	config.Server
	XdsIR            *message.XdsIR
	Xds              *message.Xds
	ExtensionManager extension.Manager
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
	message.HandleSubscription(r.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds]) {
			r.Logger.Info("received an update")
			key := update.Key
			val := update.Value

			if update.Delete {
				r.Xds.Delete(key)
			} else {
				// Translate to xds resources
				t := &translator.Translator{ExtensionManager: &r.ExtensionManager}

				// Set the rate limit service URL if global rate limiting is enabled.
				if r.EnvoyGateway.RateLimit != nil {
					t.GlobalRateLimit = &translator.GlobalRateLimitSettings{
						ServiceURL: infra.GetRateLimitServiceURL(r.Namespace),
					}
				}

				result, err := t.Translate(val)
				if err != nil {
					r.Logger.Error(err, "failed to translate xds ir")
				} else {
					// Publish
					r.Xds.Store(key, result)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}
