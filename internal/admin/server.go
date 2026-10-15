// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package admin

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/pprof"
	"net/url"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/go-logr/logr"

	"github.com/envoyproxy/gateway/internal/admin/console"
	"github.com/envoyproxy/gateway/internal/admin/gateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

// Environment variables that enable and configure the gateway route-management
// APIs. When DB_HOST is empty, the management APIs are disabled. The database
// connection is assembled from the separate DB_* variables below.
const (
	envDBHost     = "DB_HOST"
	envDBUser     = "DB_USERNAME"
	envDBPassword = "DB_PASSWORD"
	envAuthURL    = "AUTH_URL"
	envRedisURL   = "REDIS_URL"

	defaultAuthURL = "http://127.0.0.1:5120"
)

// buildDBDSN assembles a PostgreSQL connection URL from the separate DB_*
// environment variables. DB_HOST carries the address portion
// (host[:port][/dbname][?options]); url.UserPassword percent-encodes the
// credentials so passwords containing special characters are handled correctly.
func buildDBDSN() string {
	cred := url.UserPassword(os.Getenv(envDBUser), os.Getenv(envDBPassword)).String()
	return fmt.Sprintf("postgres://%s@%s", cred, os.Getenv(envDBHost))
}

type Config struct {
	Server            config.Server
	ProviderResources *message.ProviderResources
	RunnerErrors      *message.RunnerErrors
	// ReloadCh, when non-nil, enables the POST /api/reload endpoint and the
	// Redis reload subscriber. It is the file provider's reload channel.
	ReloadCh chan struct{}
}

type Runner struct {
	cfg               *config.Server
	server            *http.Server
	providerResources *message.ProviderResources
	reloadCh          chan struct{}

	// Resources owned by the gateway management API, closed on shutdown.
	db    *sql.DB
	redis *gateway.Redis
}

func New(cfg *Config) *Runner {
	return &Runner{
		cfg:               &cfg.Server,
		providerResources: cfg.ProviderResources,
		reloadCh:          cfg.ReloadCh,
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
	var firstErr error
	if r.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := r.server.Shutdown(ctx); err != nil {
			firstErr = err
		}
	}
	if r.redis != nil {
		if err := r.redis.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	if r.db != nil {
		if err := r.db.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (r *Runner) start(ctx context.Context) error {
	handlers := http.NewServeMux()
	address := r.cfg.EnvoyGateway.GetEnvoyGatewayAdminAddress()
	adminConfig := r.cfg.EnvoyGateway.GetEnvoyGatewayAdmin()
	enablePprof := adminConfig.EnablePprof

	adminLogger := r.cfg.Logger.WithName("admin")
	adminLogger.Info("starting admin server", "address", address, "enablePprof", enablePprof, "enableConsole", true)

	// Register console handlers (always enabled)
	consoleHandler := console.NewHandler(r.cfg, r.providerResources)
	consoleHandler.RegisterRoutes(handlers)

	// Register the reload endpoint when a reload channel is available (file provider mode).
	if r.reloadCh != nil {
		adminLogger.Info("reload API enabled at POST /api/reload")
		handlers.HandleFunc("/api/reload", r.handleReload)
	}

	// Wire the gateway route-management APIs when a database URL is configured.
	r.setupGatewayManagement(ctx, handlers, adminLogger.Logger)

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

// handleReload handles POST /api/reload by signalling the file provider to
// re-read all route configuration from disk.
func (r *Runner) handleReload(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed. Use POST.", http.StatusMethodNotAllowed)
		return
	}
	select {
	case r.reloadCh <- struct{}{}:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"reload triggered"}`))
	default:
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"status":"reload already pending"}`))
	}
}

// triggerReload performs a non-blocking send on the reload channel. Used by the
// Redis subscriber when another task broadcasts a config change.
func (r *Runner) triggerReload() {
	if r.reloadCh == nil {
		return
	}
	select {
	case r.reloadCh <- struct{}{}:
	default:
		// A reload is already pending; the coalesced reload will pick up the change.
	}
}

// setupGatewayManagement initializes the database, optional Redis pub/sub, and
// registers the /api/gateway/* routes. It is a no-op when DB_HOST is not set,
// keeping the admin server backward compatible.
func (r *Runner) setupGatewayManagement(ctx context.Context, mux *http.ServeMux, logger logr.Logger) {
	if os.Getenv(envDBHost) == "" {
		logger.Info("gateway management API disabled (DB_HOST not set)")
		return
	}

	db, err := gateway.InitDB(buildDBDSN())
	if err != nil {
		logger.Error(err, "failed to initialize gateway database; management API disabled")
		return
	}
	r.db = db

	// Optional Redis pub/sub for multi-task reload fan-out.
	var publish func(cluster string)
	if redisURL := os.Getenv(envRedisURL); redisURL != "" {
		rd, err := gateway.InitRedis(redisURL, logger)
		if err != nil {
			logger.Error(err, "failed to initialize redis; reload pub/sub disabled")
		} else {
			r.redis = rd
			publish = rd.Publish
			if r.reloadCh != nil {
				rd.StartSubscriber(ctx, r.triggerReload)
			}
		}
	}

	authURL := os.Getenv(envAuthURL)
	if authURL == "" {
		authURL = defaultAuthURL
	}

	repo := gateway.NewRepository(db)
	svc := gateway.NewService(repo)
	gen := gateway.NewGenerator(&gateway.GeneratorConfig{
		Repo:      repo,
		RoutesDir: r.routesDir(),
		Publish:   publish,
		Logger:    logger,
	})
	handler := gateway.NewHandler(&gateway.Config{
		Service:   svc,
		Generator: gen,
		AuthURL:   authURL,
		Logger:    logger,
	})
	handler.RegisterRoutes(mux)

	logger.Info("gateway management API enabled", "authURL", authURL, "routesDir", r.routesDir(), "redis", r.redis != nil)
}

// routesDir returns the first configured file-provider path, which is the
// directory the generator writes route files into.
func (r *Runner) routesDir() string {
	provider := r.cfg.EnvoyGateway.Provider
	if provider == nil || provider.Custom == nil || provider.Custom.Resource.File == nil {
		return ""
	}
	paths := provider.Custom.Resource.File.Paths
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}
