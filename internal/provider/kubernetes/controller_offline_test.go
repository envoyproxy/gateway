// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/message"
)

func TestNewOfflineGatewayAPIController(t *testing.T) {
	t.Run("offline controller requires config and resources", func(t *testing.T) {
		_, err := NewOfflineGatewayAPIController(context.Background(), nil, nil, nil)
		require.Error(t, err)
	})

	t.Run("offline controller does not support k8s provider type", func(t *testing.T) {
		cfg, err := config.New(os.Stdout)
		require.NoError(t, err)

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeKubernetes,
		}
		pResources := new(message.ProviderResources)
		_, err = NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.Error(t, err)
	})

	t.Run("offline controller creation succeess", func(t *testing.T) {
		cfg, err := config.New(os.Stdout)
		require.NoError(t, err)

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeCustom,
		}
		pResources := new(message.ProviderResources)
		_, err = NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.NoError(t, err)
	})
}
