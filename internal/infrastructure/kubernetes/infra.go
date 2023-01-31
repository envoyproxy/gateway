// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

// Infra manages the creation and deletion of Kubernetes infrastructure
// based on Infra IR resources.
type Infra struct {
	Client client.Client

	// Namespace is the Namespace used for managed infra.
	Namespace string
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client, cfg *config.Server) *Infra {
	return &Infra{
		Client:    cli,
		Namespace: cfg.Namespace,
	}
}

// CreateOrUpdateProxyInfra creates the managed kube infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	if err := i.createOrUpdateProxyServiceAccount(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateProxyConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateProxyDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateProxyService(ctx, infra); err != nil {
		return err
	}

	return nil
}

// DeleteProxyInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if err := i.deleteProxyService(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteProxyDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteProxyConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteProxyServiceAccount(ctx, infra); err != nil {
		return err
	}

	return nil
}

// CreateOrUpdateRateLimitInfra creates the managed kube rate limit infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}

	if err := i.deleteRateLimitService(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteRateLimitServiceAccount(ctx, infra); err != nil {
		return err
	}

	return nil
}

// DeleteRateLimitInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteRateLimitInfra(ctx context.Context, infra *ir.RateLimitInfra) error {
	if infra == nil {
		return errors.New("ratelimit infra ir is nil")
	}
	if err := i.createOrUpdateRateLimitServiceAccount(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateRateLimitService(ctx, infra); err != nil {
		return err
	}

	return nil

}
