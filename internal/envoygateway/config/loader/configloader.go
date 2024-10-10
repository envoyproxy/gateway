// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package loader

import (
	"context"
	"sync"
	"time"

	"github.com/gohugoio/hugo/watcher"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

type Loader struct {
	cfgPath string
	cfg     *config.Server
	logger  logging.Logger

	hooks map[string]func(cfg *config.Server) error
	m     sync.Mutex
}

func New(cfgPath string, cfg *config.Server) *Loader {
	return &Loader{
		cfgPath: cfgPath,
		cfg:     cfg,
		logger:  cfg.Logger.WithName("config-loader"),
		hooks:   map[string]func(cfg *config.Server) error{},
	}
}

func (r *Loader) RegisterHooks(name string, f func(cfg *config.Server) error) {
	r.m.Lock()
	defer r.m.Unlock()
	r.hooks[name] = f
}

func (r *Loader) Start(ctx context.Context) error {
	if r.cfgPath == "" {
		r.logger.Info("no config file provided, skipping config watcher")
		return nil
	}

	w, err := watcher.New(500*time.Millisecond, 500*time.Millisecond, false)
	if err != nil {
		r.logger.Error(err, "failed to create fsnotify watcher")
		return err
	}
	r.logger.Info("watching for changes to the EnvoyGateway configuration", "path", r.cfgPath)

	if err := w.Add(r.cfgPath); err != nil {
		r.logger.Error(err, "failed to add config file to watcher")
		return err
	}

	go func() {
		defer w.Close()
		for {
			select {
			case events, ok := <-w.Events:
				if !ok {
					return
				}

				for _, e := range events {
					r.logger.Info("received fsnotify events", "name", e.Name, "op", e.Op.String())
				}

				// Load the config file.
				eg, err := config.Decode(r.cfgPath)
				if err != nil {
					r.logger.Info("failed to decode config file", "name", r.cfgPath, "error", err)
					// TODO: add a metric for this?
					continue
				}
				// Set defaults for unset fields
				eg.SetEnvoyGatewayDefaults()
				r.cfg.EnvoyGateway = eg
				// update cfg logger
				eg.Logging.SetEnvoyGatewayLoggingDefaults()
				r.cfg.Logger = logging.NewLogger(eg.Logging)

				r.m.Lock()
				for n, f := range r.hooks {
					r.logger.Info("starting to call hook", "name", n)
					if err := f(r.cfg); err != nil {
						r.logger.Error(err, "hook error")
					}
				}
				r.m.Unlock()
			case err := <-w.Errors():
				r.logger.Error(err, "watcher error")
			case <-ctx.Done():
				return
			}
		}
	}()

	err = w.Add(r.cfgPath)
	if err != nil {
		r.logger.Error(err, "watcher error")
		return err
	}
	return nil
}
