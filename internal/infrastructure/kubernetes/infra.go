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
	"github.com/envoyproxy/gateway/internal/utils/env"
)

// Infra manages the creation and deletion of Kubernetes infrastructure
// based on Infra IR resources.
type Infra struct {
	Client client.Client

	// Namespace is the Namespace used for managed infra.
	Namespace string
}

// NewInfra returns a new Infra.
func NewInfra(cli client.Client) *Infra {
	return &Infra{
		Client:    cli,
		Namespace: env.Lookup("ENVOY_GATEWAY_NAMESPACE", config.EnvoyGatewayNamespace),
	}
}

// CreateOrUpdateInfra creates the managed kube infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	if err := i.createOrUpdateServiceAccount(ctx, infra); err != nil {
		return err
	}

	if _, err := i.createOrUpdateConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.createOrUpdateService(ctx, infra); err != nil {
		return err
	}

	return nil
}

// DeleteInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if err := i.deleteService(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteDeployment(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteConfigMap(ctx, infra); err != nil {
		return err
	}

	if err := i.deleteServiceAccount(ctx, infra); err != nil {
		return err
	}

	return nil
}
