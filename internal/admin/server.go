// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"embed"
	"net/http"
	"net/http/pprof"
	"time"

	"github.com/davecgh/go-spew/spew"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
)

//go:embed webui.html logo.svg cncf.svg
var webuiFS embed.FS

func Init(cfg *config.Server, k8sClient client.Client) error {
	if cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnableDumpConfig {
		spewConfig := spew.NewDefaultConfig()
		spewConfig.DisableMethods = true
		spewConfig.Dump(cfg)
	}

	// Set the k8s client for handlers
	SetK8sClient(k8sClient)

	return start(cfg)
}

func start(cfg *config.Server) error {
	handlers := http.NewServeMux()
	address := cfg.EnvoyGateway.GetEnvoyGatewayAdminAddress()
	enablePprof := cfg.EnvoyGateway.GetEnvoyGatewayAdmin().EnablePprof

	adminLogger := cfg.Logger.WithName("admin")
	adminLogger.Info("starting admin server", "address", address, "enablePprof", enablePprof)

	// Serve the web UI
	handlers.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/":
			handleWebUI(w, r)
		case "/logo.svg":
			handleSVG(w, r, "logo.svg")
		case "/cncf.svg":
			handleSVG(w, r, "cncf.svg")
		default:
			http.NotFound(w, r)
		}
	})

	// API endpoints
	handlers.HandleFunc("/admin/api/status", handleStatus(cfg))
	handlers.HandleFunc("/admin/api/resources", handleResources)
	handlers.HandleFunc("/admin/api/stats", handleStats)

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
			adminLogger.Error(err, "start admin server failed")
		}
	}()

	return nil
}
