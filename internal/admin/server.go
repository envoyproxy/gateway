// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/davecgh/go-spew/spew"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/logging"
)

var (
	adminLogger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("admin")
)

func Init(cfg *config.Server) error {
	if cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnableDumpConfig {
		spewConfig := spew.NewDefaultConfig()
		spewConfig.DisableMethods = true
		spewConfig.Dump(cfg)
	}

	return start(cfg)
}

func start(cfg *config.Server) error {
	handlers := http.NewServeMux()
	address := cfg.EnvoyGateway.GetEnvoyGatewayAdminAddress()
	enablePprof := cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnablePprof

	adminLogger.Info("starting admin server", "address", address, "enablePprof", enablePprof)

	if enablePprof {
		// Serve pprof endpoints to aid in live debugging.
		handlers.HandleFunc("/debug/pprof/", pprof.Index)
		handlers.HandleFunc("/debug/pprof/profile", pprof.Profile)
		handlers.HandleFunc("/debug/pprof/trace", pprof.Trace)
		handlers.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		handlers.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	}

	adminServer := &http.Server{
		Handler:           handlers,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Listen And Serve Admin Server.
	go func() {
		if err := adminServer.ListenAndServe(); err != nil {
			cfg.Logger.Error(err, "start admin server failed")
		}
	}()

	return nil
}
