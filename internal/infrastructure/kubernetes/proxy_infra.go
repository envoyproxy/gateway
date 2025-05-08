// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
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

	resourceNamespace := i.GetResourceNamespace(infra)
	r := proxy.NewResourceRender(resourceNamespace, i.Namespace, i.DNSDomain, infra.GetProxyInfra(), i.EnvoyGateway)
	return i.createOrUpdate(ctx, r)
}

// DeleteProxyInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	resourceNamespace := i.GetResourceNamespace(infra)
	r := proxy.NewResourceRender(resourceNamespace, i.Namespace, i.DNSDomain, infra.GetProxyInfra(), i.EnvoyGateway)
	return i.delete(ctx, r)
}

func (i *Infra) GetResourceNamespace(infra *ir.Infra) string {
	if i.EnvoyGateway.GatewayNamespaceMode() {
		return infra.Proxy.Namespace
	}
	return i.Namespace
}
