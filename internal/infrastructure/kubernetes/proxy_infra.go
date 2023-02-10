// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"github.com/envoyproxy/gateway/internal/ir"
)

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
