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
	watcher *watcher
}

func New(svr *config.Server, resources *message.ProviderResources) *Provider {
	return &Provider{
		watcher: newWatcher(svr.EnvoyGateway.Provider.Custom.Resource.File.Paths),
	}
}

func (p *Provider) Type() v1alpha1.ProviderType {
	return v1alpha1.ProviderTypeFile
}

func (p *Provider) Start(ctx context.Context) error {
	errChan := make(chan error)
	go func() {
		errChan <- p.watcher.Watch(ctx)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		return err
	}
}
