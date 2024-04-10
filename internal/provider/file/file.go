// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package file

import (
	"context"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

type Provider struct {
	paths     []string
	notifier  *Notifier
	resources *message.ProviderResources
}

func New(svr *config.Server, resources *message.ProviderResources) (*Provider, error) {
	notifier, err := NewNotifier(svr.Logger.Logger)
	if err != nil {
		return nil, err
	}

	return &Provider{
		paths:     svr.EnvoyGateway.Provider.Custom.Resource.File.Paths,
		notifier:  notifier,
		resources: resources,
	}, nil
}

func (p *Provider) Type() v1alpha1.ProviderType {
	return v1alpha1.ProviderTypeFile
}

func (p *Provider) Start(ctx context.Context) error {
	dirs, files, err := getDirsAndFilesForWatcher(p.paths)
	if err != nil {
		return err
	}

	// TODO: initial load for resources-store

	p.notifier.Watch(ctx, dirs, files)
	defer p.notifier.Close()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-p.notifier.Events:
			// TODO: ask resources-store to update according to the recv event
		}
	}
}
