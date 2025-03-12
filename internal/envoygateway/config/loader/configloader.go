// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package loader

import (
	"context"
	"io"
	"time"

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

	w filewatcher.FileWatcher
}

func New(cfgPath string, cfg *config.Server, f HookFunc) *Loader {
	return &Loader{
		cfgPath: cfgPath,
		cfg:     cfg,
		logger:  cfg.Logger.WithName("config-loader"),
		hook:    f,
		w:       filewatcher.NewWatcher(),
	}
}

func (r *Loader) Start(ctx context.Context, logOut io.Writer) error {
	r.runHook()

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
				// Set defaults for unset fields
				eg.SetEnvoyGatewayDefaults()
				r.cfg.EnvoyGateway = eg
				// update cfg logger
				eg.Logging.SetEnvoyGatewayLoggingDefaults()
				r.cfg.Logger = logging.NewLogger(logOut, eg.Logging)

				// cancel last
				if r.cancel != nil {
					r.cancel()
				}

				// TODO: we need to make sure that all runners are stopped, before we start the new ones
				// Otherwise we might end up with error listening on:8081
				time.Sleep(3 * time.Second)

				r.runHook()
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

func (r *Loader) runHook() {
	if r.hook == nil {
		return
	}

	r.logger.Info("running hook")
	c, cancel := context.WithCancel(context.TODO())
	r.cancel = cancel
	go func(ctx context.Context) {
		if err := r.hook(ctx, r.cfg); err != nil {
			r.logger.Error(err, "hook error")
		}
	}(c)
}
