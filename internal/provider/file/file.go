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
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/healthz"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/filewatcher"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/utils/path"
)

type Provider struct {
	paths          []string
	logger         logr.Logger
	watcher        filewatcher.FileWatcher
	resourcesStore *resourcesStore
}

func New(svr *config.Server, resources *message.ProviderResources) (*Provider, error) {
	logger := svr.Logger.Logger
	paths := sets.New[string]()
	if svr.EnvoyGateway.Provider.Custom.Resource.File != nil {
		paths.Insert(svr.EnvoyGateway.Provider.Custom.Resource.File.Paths...)
	}

	return &Provider{
		paths:          paths.UnsortedList(),
		logger:         logger,
		watcher:        filewatcher.NewWatcher(),
		resourcesStore: newResourcesStore(svr.EnvoyGateway.Gateway.ControllerName, resources, logger),
	}, nil
}

func (p *Provider) Type() egv1a1.ProviderType {
	return egv1a1.ProviderTypeCustom
}

func (p *Provider) Start(ctx context.Context) error {
	defer func() {
		_ = p.watcher.Close()
	}()

	// Start runnable servers.
	go p.startHealthProbeServer(ctx)

	dirs, files := path.ListDirsAndFiles(p.paths)
	// Initially load resources from paths on host.
	if err := p.resourcesStore.LoadAndStore(files.UnsortedList(), dirs.UnsortedList()); err != nil {
		return fmt.Errorf("failed to load resources into store: %w", err)
	}

	// aggregate all path channel into one
	aggCh := make(chan fsnotify.Event)
	for _, path := range p.paths {
		if err := p.watcher.Add(path); err != nil {
			p.logger.Error(err, "failed to add watch", "path", path)
		}
		p.logger.Info("Watching file changed", "path", path)
		ch := p.watcher.Events(path)
		go func(c chan fsnotify.Event) {
			for msg := range c {
				aggCh <- msg
			}
		}(ch)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-aggCh:
			p.logger.Info("file changed", "op", event.Op, "name", event.Name)
			switch event.Op {
			case fsnotify.Create:
				dirs.Insert(event.Name)
				files.Insert(event.Name)
			case fsnotify.Remove:
				dirs.Delete(event.Name)
				files.Delete(event.Name)
			default:
				// do nothing
				continue
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
