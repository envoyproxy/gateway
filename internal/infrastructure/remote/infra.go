// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package remote

import (
	"context"

	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
)

// Infra manages the creation and deletion of Kubernetes infrastructure
// based on Infra IR resources.
type Infra struct {
	// EnvoyGateway is the configuration used to startup Envoy Gateway.
	EnvoyGateway *egv1a1.EnvoyGateway

	logger logging.Logger

	// errors is the notifier used to send async errors to the main control loop.
	errors message.RunnerErrorNotifier

	ic InfraClient
}

// NewInfra returns a new Infra.
func NewInfra(cfg *config.Server, k8sClient client.Client, errors message.RunnerErrorNotifier) (*Infra, error) {
	// We initialize the client here, that way if the remote connection is misconfigured, then the pod
	// crashes rather than silently failing when infrastructure changes happen.
	infraClient, err := newRemoteInfraClient(cfg, k8sClient)
	if err != nil {
		if infraClient != nil {
			_ = infraClient.Close()
		}
		return nil, err
	}
	return &Infra{
		EnvoyGateway: cfg.EnvoyGateway,
		logger:       cfg.Logger.WithName(string(egv1a1.LogComponentInfrastructureRunner)),
		errors:       errors,
		ic:           infraClient,
	}, nil
}

func (i *Infra) Close() error {
	return i.ic.Close()
}

func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	return i.ic.CreateOrUpdateProxyInfra(ctx, infra)
}

func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	return i.ic.DeleteProxyInfra(ctx, infra)
}

func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context) error {
	return i.ic.CreateOrUpdateRateLimitInfra(ctx)
}

func (i *Infra) DeleteRateLimitInfra(ctx context.Context) error {
	return i.ic.DeleteRateLimitInfra(ctx)
}
