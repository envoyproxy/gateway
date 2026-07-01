// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"
	"sync"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// Infra manages the creation and deletion of remotely managed proxy and rate
// limit infrastructure by delegating to a remote provider over gRPC.
//
// The underlying InfraClient is constructed lazily on the first method call
// that requires it. This avoids dialing the remote service or reading
// Kubernetes secrets during process startup, where failures would crash the
// pod before validation that the remote provider is actually being used.
type Infra struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	logger logging.Logger

	// errors is the notifier used to send async errors to the main control loop.
	errors message.RunnerErrorNotifier

	// factory builds the InfraClient on demand. It must not be nil.
	factory InfraClientFactory

	mu sync.Mutex
	ic InfraClient
}

// NewInfra returns a new Infra that will lazily build its InfraClient via the
// provided factory. The factory is invoked at most once for a successful
// construction; if it returns an error, the next call will retry.
func NewInfra(cfg *config.Server, factory InfraClientFactory, errors message.RunnerErrorNotifier) *Infra {
	return new(Infra{
		EnvoyGateway: cfg.EnvoyGateway,
		logger:       cfg.Logger.WithName(string(egv1a1.LogComponentInfrastructureRunner)),
		errors:       errors,
		factory:      factory,
	})
}

// Close releases any resources held by the underlying InfraClient. It is a
// no-op if the client was never constructed.
func (i *Infra) Close() error {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.ic == nil {
		return nil
	}
	err := i.ic.Close()
	i.ic = nil
	return err
}

// CreateOrUpdateProxyInfra delegates to the underlying InfraClient.
func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	ic, err := i.client(ctx)
	if err != nil {
		return err
	}
	return ic.CreateOrUpdateProxyInfra(ctx, infra)
}

// DeleteProxyInfra delegates to the underlying InfraClient.
func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	ic, err := i.client(ctx)
	if err != nil {
		return err
	}
	return ic.DeleteProxyInfra(ctx, infra)
}

// CreateOrUpdateRateLimitInfra delegates to the underlying InfraClient.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	ic, err := i.client(ctx)
	if err != nil {
		return err
	}
	return ic.CreateOrUpdateRateLimitInfra(ctx)
}

// DeleteRateLimitInfra delegates to the underlying InfraClient.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context) error {
	ic, err := i.client(ctx)
	if err != nil {
		return err
	}
	return ic.DeleteRateLimitInfra(ctx)
}

// client returns the cached InfraClient, building it via the factory on the
// first successful call. Failed factory invocations are not cached so that
// transient errors during startup of the remote service are retried on the
// next request.
func (i *Infra) client(ctx context.Context) (InfraClient, error) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if i.ic != nil {
		return i.ic, nil
	}
	ic, err := i.factory(ctx)
	if err != nil {
		return nil, err
	}
	i.ic = ic
	return ic, nil
}
