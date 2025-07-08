// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"context"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

type Runner struct {
	cfg    *config.Server
	server *http.Server
}

func New(cfg *config.Server) *Runner {
	return &Runner{
		cfg: cfg,
	}
}

func (r *Runner) Start(ctx context.Context) error {
	if r.cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnableDumpConfig {
		spewConfig := spew.NewDefaultConfig()
		spewConfig.DisableMethods = true
		spewConfig.Dump(r.cfg)
	}

	return r.start(ctx)
}

func (r *Runner) Name() string {
	return "admin"
}

func (r *Runner) Close() error {
	if r.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return r.server.Shutdown(ctx)
	}
	return nil
}

func (r *Runner) start(ctx context.Context) error {
	handlers := http.NewServeMux()
	address := r.cfg.EnvoyGateway.GetEnvoyGatewayAdminAddress()
	enablePprof := r.cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnablePprof

	adminLogger := r.cfg.Logger.WithName("admin")
	adminLogger.Info("starting admin server", "address", address, "enablePprof", enablePprof)

	if enablePprof {
		// Serve pprof endpoints to aid in live debugging.
		handlers.HandleFunc("/debug/pprof/", pprof.Index)
		handlers.HandleFunc("/debug/pprof/profile", pprof.Profile)
		handlers.HandleFunc("/debug/pprof/trace", pprof.Trace)
		handlers.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		handlers.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	}

	r.server = &http.Server{
		Handler:           handlers,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Listen And Serve Admin Server.
	go func() {
		if err := r.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			adminLogger.Error(err, "start admin server failed")
		}
	}()

	return nil
}
