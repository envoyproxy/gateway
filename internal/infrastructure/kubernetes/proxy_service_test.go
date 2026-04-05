// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestDeleteProxyService(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "delete service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			kube := newTestInfra(t)
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

			infra := ir.NewInfra()
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
				Kind: resource.KindGatewayClass,
				Name: testGatewayClass,
			}
			r, err := proxy.NewResourceRender(ctx, kube, infra)
			require.NoError(t, err)
			err = kube.createOrUpdateService(ctx, r)
			require.NoError(t, err)

			err = kube.deleteService(ctx, r)
			require.NoError(t, err)
		})
	}
}
