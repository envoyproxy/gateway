// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-logr/logr"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

type Provider struct {
	paths          []string
	logger         logr.Logger
	notifier       *Notifier
	resourcesStore *resourcesStore
}

func New(svr *config.Server, resources *message.ProviderResources) (*Provider, error) {
	logger := svr.Logger.Logger

	notifier, err := NewNotifier(logger)
	if err != nil {
		return nil, err
	}

	return &Provider{
		paths:          svr.EnvoyGateway.Provider.Custom.Resource.File.Paths,
		logger:         logger,
		notifier:       notifier,
		resourcesStore: newResourcesStore(svr.EnvoyGateway.Gateway.ControllerName, resources, logger),
	}, nil
}

func (p *Provider) Type() egv1a1.ProviderType {
	return egv1a1.ProviderTypeCustom
}

func (p *Provider) Start(ctx context.Context) error {
	dirs, files, err := getDirsAndFilesForWatcher(p.paths)
	if err != nil {
		return fmt.Errorf("failed to get directories and files for the watcher: %w", err)
	}

	// Start runnable servers.
	go p.startHealthProbeServer(ctx)

	// Initially load resources from paths on host.
	if err = p.resourcesStore.LoadAndStore(files.UnsortedList(), dirs.UnsortedList()); err != nil {
		return fmt.Errorf("failed to load resources into store: %w", err)
	}

	// Start watchers in notifier.
	p.notifier.Watch(ctx, dirs, files)
	defer p.notifier.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-p.notifier.Events:
			switch event.Op {
			case fsnotify.Create:
				dirs.Insert(event.Name)
				files.Insert(event.Name)
			case fsnotify.Remove:
				dirs.Delete(event.Name)
				files.Delete(event.Name)
			}

			p.resourcesStore.HandleEvent(event, files.UnsortedList(), dirs.UnsortedList())
		}
	}
}

func (p *Provider) startHealthProbeServer(ctx context.Context) {
	const (
		readyzEndpoint  = "/readyz"
		healthzEndpoint = "/healthz"
	)

	mux := http.NewServeMux()
	srv := &http.Server{
		Addr:              ":8081",
		Handler:           mux,
		MaxHeaderBytes:    1 << 20,
		IdleTimeout:       90 * time.Second, // matches http.DefaultTransport keep-alive timeout
		ReadHeaderTimeout: 32 * time.Second,
	}

	readyzHandler := &healthz.Handler{
		Checks: map[string]healthz.Checker{
			readyzEndpoint: healthz.Ping,
		},
	}
	mux.Handle(readyzEndpoint, http.StripPrefix(readyzEndpoint, readyzHandler))
	// Append '/' suffix to handle subpaths.
	mux.Handle(readyzEndpoint+"/", http.StripPrefix(readyzEndpoint, readyzHandler))

	healthzHandler := &healthz.Handler{
		Checks: map[string]healthz.Checker{
			healthzEndpoint: healthz.Ping,
		},
	}
	mux.Handle(healthzEndpoint, http.StripPrefix(healthzEndpoint, healthzHandler))
	// Append '/' suffix to handle subpaths.
	mux.Handle(healthzEndpoint+"/", http.StripPrefix(healthzEndpoint, readyzHandler))

	go func() {
		<-ctx.Done()
		if err := srv.Close(); err != nil {
			p.logger.Error(err, "failed to close health probe server")
		}
	}()

	p.logger.Info("starting health probe server", "address", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		p.logger.Error(err, "failed to start health probe server")
	}
}
