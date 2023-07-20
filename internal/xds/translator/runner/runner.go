// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"

	"github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extensionregistry "github.com/envoyproxy/gateway/internal/extension/registry"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/runner"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

func Register(resources Resources, globalConfig config.Server) {
	runner.Manager().Register(New(resources, globalConfig), runner.RootParentRunner)
}

func New(resources Resources, globalConfig config.Server) *xdsTranslatorRunner {
	return &xdsTranslatorRunner{runner.New(string(v1alpha1.LogComponentXdsTranslatorRunner), resources, globalConfig)}
}

type Resources struct {
	XdsIR            *message.XdsIR
	Xds              *message.Xds
	ExtensionManager extension.Manager
}

type xdsTranslatorRunner struct {
	*runner.GenericRunner[Resources]
}

// Start starts the xds-translator runner
func (r *xdsTranslatorRunner) Start(ctx context.Context) error {
	r.Init(ctx)
	go r.SubscribeAndTranslate(ctx)

	r.Logger.Info("started")
	return nil
}

func (r *xdsTranslatorRunner) ShutDown(ctx context.Context) {
	r.Resources.Xds.Close()

	// Close connections to extension services
	if mgr, ok := r.Resources.ExtensionManager.(*extensionregistry.Manager); ok {
		mgr.CleanupHookConns()
	}
}

func (r *xdsTranslatorRunner) SubscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(r.Resources.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds]) {
			r.Logger.Info("received an update")
			key := update.Key
			val := update.Value

			if update.Delete {
				r.Resources.Xds.Delete(key)
			} else {
				// Translate to xds resources
				t := &translator.Translator{}

				// Set the extension manager if an extension is loaded
				if r.Resources.ExtensionManager != nil {
					t.ExtensionManager = &r.Resources.ExtensionManager
				}

				// Set the rate limit service URL if global rate limiting is enabled.
				if r.EnvoyGateway.RateLimit != nil {
					t.GlobalRateLimit = &translator.GlobalRateLimitSettings{
						ServiceURL: ratelimit.GetServiceURL(r.Namespace, r.DNSDomain),
					}
				}

				result, err := t.Translate(val)
				if err != nil {
					r.Logger.Error(err, "failed to translate xds ir")
				} else {
					// Publish
					r.Resources.Xds.Store(key, result)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}
