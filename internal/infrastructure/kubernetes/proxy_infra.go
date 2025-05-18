// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

// CreateOrUpdateProxyInfra creates the managed kube infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	if infra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	envoyNamespace := i.GetResourceNamespace(infra)
	ownerReferenceUID, err := i.GetOwnerReferenceUID(ctx, infra)
	if err != nil {
		return err
	}

	r := proxy.NewResourceRender(envoyNamespace, i.ControllerNamespace, i.DNSDomain, infra.GetProxyInfra(), i.EnvoyGateway, ownerReferenceUID)
	return i.createOrUpdate(ctx, r)
}

// DeleteProxyInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(ctx context.Context, infra *ir.Infra) error {
	if infra == nil {
		return errors.New("infra ir is nil")
	}

	envoyNamespace := i.GetResourceNamespace(infra)
	r := proxy.NewResourceRender(envoyNamespace, i.ControllerNamespace, i.DNSDomain, infra.GetProxyInfra(), i.EnvoyGateway, nil)
	return i.delete(ctx, r)
}

func (i *Infra) GetResourceNamespace(infra *ir.Infra) string {
	if i.EnvoyGateway.GatewayNamespaceMode() {
		return infra.Proxy.Namespace
	}
	return i.ControllerNamespace
}

func (i *Infra) GetOwnerReferenceUID(ctx context.Context, infra *ir.Infra) (map[string]types.UID, error) {
	ownerReferenceUID := make(map[string]types.UID)

	if i.EnvoyGateway.GatewayNamespaceMode() {
		key := types.NamespacedName{
			Namespace: i.GetResourceNamespace(infra),
			Name:      utils.GetKubernetesResourceName(infra.Proxy.Name),
		}
		gatewayUID, err := i.Client.GetUID(ctx, key, &gwapiv1.Gateway{})
		if err != nil {
			return nil, err
		}
		ownerReferenceUID[proxy.ResourceKindGateway] = gatewayUID
	}
	// TODO: set GatewayClass UID when enable merged gateways

	return ownerReferenceUID, nil
}
