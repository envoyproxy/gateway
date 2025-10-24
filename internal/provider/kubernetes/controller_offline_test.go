// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

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
		cfg, err := config.New(os.Stdout, os.Stderr)
		require.NoError(t, err)

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeKubernetes,
		}
		pResources := new(message.ProviderResources)
		_, err = NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.Error(t, err)
	})

	t.Run("offline controller creation success", func(t *testing.T) {
		cfg, err := config.New(os.Stdout, os.Stderr)
		require.NoError(t, err)

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeCustom,
		}
		pResources := new(message.ProviderResources)
		_, err = NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.NoError(t, err)
	})

	t.Run("offline controller with extension server and custom resource creation success", func(t *testing.T) {
		cfg, err := config.New(os.Stdout, os.Stderr)
		require.NoError(t, err)

		gvk := egv1a1.GroupVersionKind{
			Group:   "gateway.example.io",
			Version: "v1alpha1",
			Kind:    "ExampleExtPolicy",
		}

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeCustom,
		}
		cfg.EnvoyGateway.ExtensionManager = &egv1a1.ExtensionManager{
			PolicyResources: []egv1a1.GroupVersionKind{gvk},
		}

		pResources := new(message.ProviderResources)
		reconciler, err := NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.NoError(t, err)
		require.NotNil(t, reconciler)

		// Verify version registration and that the custom resource is recognized by the scheme
		require.True(t, reconciler.client.Scheme().IsVersionRegistered(schema.GroupVersion{Group: gvk.Group, Version: gvk.Version}))
		require.True(t, reconciler.client.Scheme().IsGroupRegistered(gvk.Group))
		require.True(t, reconciler.client.Scheme().Recognizes(schema.GroupVersionKind{Group: gvk.Group, Version: gvk.Version, Kind: gvk.Kind}))

		// Verify the custom resource can be loaded from yaml
		inFile := "./testdata/custom-resource.yaml"
		data, err := os.ReadFile(inFile)
		require.NoError(t, err)
		resources, err := resource.LoadResourcesFromYAMLBytes(cfg, data, true)
		require.NoError(t, err)
		require.Equal(t, 1, len(resources.ExtensionServerPolicies))

		// Verify the custom resource gets the default namespace
		require.Equal(t, "envoy-gateway-system", resources.ExtensionServerPolicies[0].GetNamespace())
	})
}
