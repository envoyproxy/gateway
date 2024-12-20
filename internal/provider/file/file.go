// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
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

	initDirs, initFiles := path.ListDirsAndFiles(p.paths)
	// Initially load resources from paths on host.
	if err := p.resourcesStore.LoadAndStore(initFiles.UnsortedList(), initDirs.UnsortedList()); err != nil {
		return fmt.Errorf("failed to load resources into store: %w", err)
	}

	// Add paths to the watcher, and aggregate all path channels into one.
	aggCh := make(chan fsnotify.Event)
	for _, path := range p.paths {
		if err := p.watcher.Add(path); err != nil {
			p.logger.Error(err, "failed to add watch", "path", path)
		} else {
			p.logger.Info("Watching path added", "path", path)
		}

		ch := p.watcher.Events(path)
		go func(c chan fsnotify.Event) {
			for msg := range c {
				aggCh <- msg
			}
		}(ch)
	}

	curDirs, curFiles := initDirs.Clone(), initFiles.Clone()
	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-aggCh:
			// Ignore the irrelevant event.
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// If a file change event is detected, regardless of the event type, it will be processed
			// as a Remove event if the file does not exist, and as a Write event if the file exists.
			//
			// The reason to do so is quite straightforward, for text edit tools like vi/vim etc.
			// They always create a temporary file, remove the existing one and replace it with the
			// temporary file when file is saved. So the watcher will only receive:
			// - Create event, with name "filename~".
			// - Remove event, with name "filename", but the file actually exist.
			if initFiles.Has(event.Name) {
				p.logger.Info("file changed", "op", event.Op, "name", event.Name)

				// For Write event, the file definitely exist.
				if event.Has(fsnotify.Write) {
					goto handle
				}

				_, err := os.Lstat(event.Name)
				if err != nil && os.IsNotExist(err) {
					curFiles.Delete(event.Name)
				} else {
					curFiles.Insert(event.Name)
				}
				goto handle
			}

			// Ignore the hidden or temporary file related change event under a directory.
			if _, name := filepath.Split(event.Name); strings.HasPrefix(name, ".") ||
				strings.HasSuffix(name, "~") {
				continue
			}
			p.logger.Info("file changed", "op", event.Op, "name", event.Name)

			switch event.Op {
			case fsnotify.Create, fsnotify.Write, fsnotify.Remove:
				// Since we do not watch any events in the subdirectories, any events involving files
				// modifications in current directory will trigger the event handling.
				goto handle
			default:
				// do nothing
				continue
			}

		handle:
			p.resourcesStore.HandleEvent(curFiles.UnsortedList(), curDirs.UnsortedList())
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
