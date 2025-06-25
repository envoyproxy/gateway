// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

// CreateOrUpdateProxyInfra creates the managed kube infra, if it doesn't exist.
func (i *Infra) CreateOrUpdateProxyInfra(ctx context.Context, irInfra *ir.Infra) error {
	if irInfra == nil {
		return errors.New("infra ir is nil")
	}

	if irInfra.Proxy == nil {
		return errors.New("infra proxy ir is nil")
	}

	r, err := proxy.NewResourceRender(ctx, i, irInfra)
	if err != nil {
		return fmt.Errorf("failed to initialize proxy resource render: %w", err)
	}
	return i.createOrUpdate(ctx, r)
}

// DeleteProxyInfra removes the managed kube infra, if it doesn't exist.
func (i *Infra) DeleteProxyInfra(ctx context.Context, irInfra *ir.Infra) error {
	if irInfra == nil {
		return errors.New("infra ir is nil")
	}

	r, err := proxy.NewResourceRender(ctx, i, irInfra)
	if err != nil {
		return fmt.Errorf("failed to create proxy resource render: %w", err)
	}
	return i.delete(ctx, r)
}

func (i *Infra) GetControllerNamespace() string {
	return i.ControllerNamespace
}

func (i *Infra) GetDNSDomain() string {
	return i.DNSDomain
}

func (i *Infra) GetEnvoyGateway() *egv1a1.EnvoyGateway {
	return i.EnvoyGateway
}

func (i *Infra) GetOwnerReferenceUID(ctx context.Context, irInfra *ir.Infra) (map[string]types.UID, error) {
	ownerReferenceUID := make(map[string]types.UID)
	if irInfra.GetProxyInfra().GetProxyMetadata() == nil {
		return nil, errors.New("infra proxy metadata ir is nil")
	}
	if irInfra.GetProxyInfra().GetProxyMetadata().OwnerReference == nil {
		return nil, errors.New("infra proxy metadata owner reference ir is nil")
	}

	if i.EnvoyGateway.GatewayNamespaceMode() {
		key := types.NamespacedName{
			Namespace: i.GetResourceNamespace(irInfra),
			Name:      irInfra.GetProxyInfra().GetProxyMetadata().OwnerReference.Name,
		}
		gatewayUID, err := i.Client.GetUID(ctx, key, &gwapiv1.Gateway{})
		if err != nil {
			return nil, err
		}
		ownerReferenceUID[resource.KindGateway] = gatewayUID
	} else {
		key := types.NamespacedName{
			Name: irInfra.GetProxyInfra().GetProxyMetadata().OwnerReference.Name,
		}
		gatewayClassUID, err := i.Client.GetUID(ctx, key, &gwapiv1.GatewayClass{})
		if err != nil {
			return nil, err
		}
		ownerReferenceUID[resource.KindGatewayClass] = gatewayClassUID
	}

	return ownerReferenceUID, nil
}

func (i *Infra) GetResourceNamespace(irInfra *ir.Infra) string {
	if i.EnvoyGateway.GatewayNamespaceMode() {
		return irInfra.Proxy.Namespace
	}
	return i.ControllerNamespace
}
