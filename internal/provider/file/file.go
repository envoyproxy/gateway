// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"

	"github.com/fsnotify/fsnotify"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

type Provider struct {
	paths          []string
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
		return err
	}

	// Initially load resources from paths on host.
	if err = p.resourcesStore.LoadAndStore(files.UnsortedList(), dirs.UnsortedList()); err != nil {
		return err
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
