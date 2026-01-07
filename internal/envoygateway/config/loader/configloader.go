// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package loader

import (
	"context"
	"io"
	"sync"
	"time"

	"github.com/envoyproxy/gateway/api/v1alpha1/validation"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/filewatcher"
	"github.com/envoyproxy/gateway/internal/logging"
)

type HookFunc func(c context.Context, cfg *config.Server) error

type Loader struct {
	cfgPath string
	cfg     *config.Server
	logger  logging.Logger
	cancel  context.CancelFunc
	hook    HookFunc
	mu      sync.RWMutex
	hookErr chan error

	w filewatcher.FileWatcher
}

func New(cfgPath string, cfg *config.Server, f HookFunc) *Loader {
	return &Loader{
		cfgPath: cfgPath,
		cfg:     cfg,
		logger:  cfg.Logger.WithName("config-loader"),
		hook:    f,
		hookErr: make(chan error, 1),
		w:       filewatcher.NewWatcher(),
	}
}

func (r *Loader) Start(ctx context.Context, logOut io.Writer) error {
	r.runHook(ctx)
	select {
	case err := <-r.hookErr:
		return err
	default:
	}

	if r.cfgPath == "" {
		r.logger.Info("no config file provided, skipping config watcher")
		return nil
	}

	r.logger.Info("watching for changes to the EnvoyGateway configuration", "path", r.cfgPath)
	if err := r.w.Add(r.cfgPath); err != nil {
		r.logger.Error(err, "failed to add config file to watcher")
		return err
	}

	go func() {
		defer func() {
			_ = r.w.Close()
		}()
		for {
			select {
			case e := <-r.w.Events(r.cfgPath):
				r.logger.Info("received fsnotify events", "name", e.Name, "op", e.Op.String())

				// Load the config file.
				eg, err := config.Decode(r.cfgPath)
				if err != nil {
					r.logger.Info("failed to decode config file", "name", r.cfgPath, "error", err)
					// TODO: add a metric for this?
					continue
				}

				if err := validation.ValidateEnvoyGateway(eg); err != nil {
					r.logger.Error(err, "failed to validate EnvoyGateway config")
					continue
				}

				// Set defaults for unset fields
				eg.SetEnvoyGatewayDefaults()
				eg.Logging.SetEnvoyGatewayLoggingDefaults()

				r.mu.Lock()
				r.cfg.EnvoyGateway = eg
				// update cfg logger
				r.cfg.Logger = logging.NewLogger(logOut, eg.Logging)
				r.mu.Unlock()

				// cancel last
				if r.cancel != nil {
					r.cancel()
				}

				// TODO: we need to make sure that all runners are stopped, before we start the new ones
				// Otherwise we might end up with error listening on:8081
				time.Sleep(3 * time.Second)

				r.runHook(ctx)
			case err := <-r.w.Errors(r.cfgPath):
				r.logger.Error(err, "watcher error")
			case <-ctx.Done():
				if r.cancel != nil {
					r.cancel()
				}
				return
			}
		}
	}()

	return nil
}

func (r *Loader) runHook(ctx context.Context) {
	if r.hook == nil {
		return
	}

	r.logger.Info("running hook")
	cfgCopy := r.snapshotConfig()
	c, cancel := context.WithCancel(ctx)
	r.cancel = cancel
	go func(ctx context.Context) {
		defer cancel()
		if err := r.hook(ctx, cfgCopy); err != nil {
			r.logger.Error(err, "hook error")
			// There is nothing we can do here, throw the error to the main process to exit
			// The EnvoyGateway pod will restart and hopefully any transient errors will be resolved
			r.sendHookError(err)
		}
	}(c)
}

// Errors returns a channel where hook errors are reported.
func (r *Loader) Errors() <-chan error {
	return r.hookErr
}

func (r *Loader) sendHookError(err error) {
	select {
	case r.hookErr <- err: // avoid blocking the sender
	default:
	}
}

func (r *Loader) snapshotConfig() *config.Server {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cfg == nil {
		return nil
	}

	cp := *r.cfg
	if r.cfg.EnvoyGateway != nil {
		cp.EnvoyGateway = r.cfg.EnvoyGateway.DeepCopy()
	}

	return &cp
}
