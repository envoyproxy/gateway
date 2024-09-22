// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package runner

import (
	"context"
	"reflect"
	"sync"

	ktypes "k8s.io/apimachinery/pkg/types"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/common"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	extension "github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/xds/translator"
)

var _ common.Runner = &Runner{}

type Config struct {
	ServerCfg         *config.Server
	XdsIR             *message.XdsIR
	Xds               *message.Xds
	ExtensionManager  extension.Manager
	ProviderResources *message.ProviderResources
	Logger            logging.Logger
}

type Runner struct {
	Config
	m sync.Mutex
}

func New(cfg *Config) *Runner {
	return &Runner{Config: *cfg}
}

func (r *Runner) Name() string {
	return string(egv1a1.LogComponentXdsTranslatorRunner)
}

// Start starts the xds-translator runner
func (r *Runner) Start(ctx context.Context) (err error) {
	r.Logger = r.ServerCfg.Logger.WithName(r.Name())
	go r.subscribeAndTranslate(ctx)
	r.Logger.Info("started")
	return
}

func (r *Runner) Reload(serverCfg *config.Server) error {
	r.m.Lock()
	r.ServerCfg = serverCfg
	r.Logger = serverCfg.Logger.WithName(r.Name())
	r.m.Unlock()

	r.Logger.Info("reloaded")
	// TODO: trigger a reload
	return nil
}

func (r *Runner) subscribeAndTranslate(ctx context.Context) {
	// Subscribe to resources
	message.HandleSubscription(message.Metadata{Runner: string(egv1a1.LogComponentXdsTranslatorRunner), Message: "xds-ir"}, r.XdsIR.Subscribe(ctx),
		func(update message.Update[string, *ir.Xds], errChan chan error) {
			r.Logger.Info("received an update")
			key := update.Key
			val := update.Value

			if update.Delete {
				r.Xds.Delete(key)
			} else {
				// Translate to xds resources
				t := &translator.Translator{
					FilterOrder: val.FilterOrder,
				}

				// Set the extension manager if an extension is loaded
				if r.ExtensionManager != nil {
					t.ExtensionManager = &r.ExtensionManager
				}

				r.m.Lock()
				// Set the rate limit service URL if global rate limiting is enabled.
				if r.ServerCfg.EnvoyGateway.RateLimit != nil {
					t.GlobalRateLimit = &translator.GlobalRateLimitSettings{
						ServiceURL: ratelimit.GetServiceURL(r.ServerCfg.Namespace, r.ServerCfg.DNSDomain),
						FailClosed: r.ServerCfg.EnvoyGateway.RateLimit.FailClosed,
					}
					if r.ServerCfg.EnvoyGateway.RateLimit.Timeout != nil {
						t.GlobalRateLimit.Timeout = r.ServerCfg.EnvoyGateway.RateLimit.Timeout.Duration
					}
				}
				r.m.Unlock()

				result, err := t.Translate(val)
				if err != nil {
					r.Logger.Error(err, "failed to translate xds ir")
					errChan <- err
				}

				// xDS translation is done in a best-effort manner, so the result
				// may contain partial resources even if there are errors.
				if result == nil {
					r.Logger.Info("no xds resources to publish")
					return
				}

				// Get all status keys from watchable and save them in the map statusesToDelete.
				// Iterating through result.EnvoyPatchPolicyStatuses, any valid keys will be removed from statusesToDelete.
				// Remaining keys will be deleted from watchable before we exit this function.
				statusesToDelete := make(map[ktypes.NamespacedName]bool)
				for key := range r.ProviderResources.EnvoyPatchPolicyStatuses.LoadAll() {
					statusesToDelete[key] = true
				}

				// Publish EnvoyPatchPolicyStatus
				for _, e := range result.EnvoyPatchPolicyStatuses {
					key := ktypes.NamespacedName{
						Name:      e.Name,
						Namespace: e.Namespace,
					}
					// Skip updating status for policies with empty status
					// They may have been skipped in this translation because
					// their target is not found (not relevant)
					if !(reflect.ValueOf(e.Status).IsZero()) {
						r.ProviderResources.EnvoyPatchPolicyStatuses.Store(key, e.Status)
					}
					delete(statusesToDelete, key)
				}
				// Discard the EnvoyPatchPolicyStatuses to reduce memory footprint
				result.EnvoyPatchPolicyStatuses = nil

				// Publish
				r.Xds.Store(key, result)

				// Delete all the deletable status keys
				for key := range statusesToDelete {
					r.ProviderResources.EnvoyPatchPolicyStatuses.Delete(key)
				}
			}
		},
	)
	r.Logger.Info("subscriber shutting down")
}
