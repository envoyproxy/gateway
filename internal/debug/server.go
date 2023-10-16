// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package debug

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
	debugLogger = logging.DefaultLogger(v1alpha1.LogLevelInfo).WithName("debug")
)

func Init(cfg *config.Server) error {
	if cfg.EnvoyGateway.GetEnvoyGatewayDebug().EnableDumpConfig {
		spewConfig := spew.NewDefaultConfig()
		spewConfig.DisableMethods = true
		spewConfig.Dump(cfg)
	}

	return start(cfg)
}

func start(cfg *config.Server) error {
	handlers := http.NewServeMux()
	address := cfg.EnvoyGateway.GetEnvoyGatewayDebugAddress()
	enablePprof := cfg.EnvoyGateway.GetEnvoyGatewayDebug().EnablePprof

	debugLogger.Info("starting debug server", "address", address, "enablePprof", enablePprof)

	if enablePprof {
		// Serve pprof endpoints to aid in live debugging.
		handlers.HandleFunc("/debug/pprof/", pprof.Index)
		handlers.HandleFunc("/debug/pprof/profile", pprof.Profile)
		handlers.HandleFunc("/debug/pprof/trace", pprof.Trace)
		handlers.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		handlers.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	}

	debugServer := &http.Server{
		Handler:           handlers,
		Addr:              address,
		ReadTimeout:       5 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       15 * time.Second,
	}

	// Listen And Serve Debug Server.
	go func() {
		if err := debugServer.ListenAndServe(); err != nil {
			cfg.Logger.Error(err, "start debug server failed")
		}
	}()

	return nil
}
