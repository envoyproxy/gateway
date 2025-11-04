// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
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

		extGVK := egv1a1.GroupVersionKind{
			Group:   "gateway.example.io",
			Version: "v1alpha1",
			Kind:    "CustomRouteFilterResource",
		}
		extServerPolicyGVK := egv1a1.GroupVersionKind{
			Group:   "extensions.example.io",
			Version: "v1alpha1",
			Kind:    "CustomExtensionServerPolicy",
		}
		extBackendGVK := egv1a1.GroupVersionKind{
			Group:   "backend.example.io",
			Version: "v1alpha1",
			Kind:    "CustomBackendResource",
		}

		cfg.EnvoyGateway.Provider = &egv1a1.EnvoyGatewayProvider{
			Type: egv1a1.ProviderTypeCustom,
		}
		cfg.EnvoyGateway.ExtensionManager = &egv1a1.ExtensionManager{
			Resources:        []egv1a1.GroupVersionKind{extGVK},
			PolicyResources:  []egv1a1.GroupVersionKind{extServerPolicyGVK},
			BackendResources: []egv1a1.GroupVersionKind{extBackendGVK},
		}

		pResources := new(message.ProviderResources)
		reconciler, err := NewOfflineGatewayAPIController(context.Background(), cfg, nil, pResources)
		require.NoError(t, err)
		require.NotNil(t, reconciler)

		// Verify version registration and that the custom resource is recognized by the scheme
		assert.True(t, reconciler.client.Scheme().IsVersionRegistered(schema.GroupVersion{Group: extGVK.Group, Version: extGVK.Version}))
		assert.True(t, reconciler.client.Scheme().IsGroupRegistered(extGVK.Group))
		assert.True(t, reconciler.client.Scheme().Recognizes(schema.GroupVersionKind{Group: extGVK.Group, Version: extGVK.Version, Kind: extGVK.Kind}))
		assert.True(t, reconciler.client.Scheme().IsVersionRegistered(schema.GroupVersion{Group: extServerPolicyGVK.Group, Version: extServerPolicyGVK.Version}))
		assert.True(t, reconciler.client.Scheme().IsGroupRegistered(extServerPolicyGVK.Group))
		assert.True(t, reconciler.client.Scheme().Recognizes(schema.GroupVersionKind{Group: extServerPolicyGVK.Group, Version: extServerPolicyGVK.Version, Kind: extServerPolicyGVK.Kind}))
		assert.True(t, reconciler.client.Scheme().IsVersionRegistered(schema.GroupVersion{Group: extBackendGVK.Group, Version: extBackendGVK.Version}))
		assert.True(t, reconciler.client.Scheme().IsGroupRegistered(extBackendGVK.Group))
		assert.True(t, reconciler.client.Scheme().Recognizes(schema.GroupVersionKind{Group: extBackendGVK.Group, Version: extBackendGVK.Version, Kind: extBackendGVK.Kind}))

		// Verify the custom resource can be loaded from YAML
		inFile := "./testdata/custom-resource.yaml"
		data, err := os.ReadFile(inFile)
		require.NoError(t, err)
		resources, err := resource.LoadResourcesFromYAMLBytes(cfg, data, true)
		require.NoError(t, err)
		// Expect 1 extension server policy and 2 extension-managed resources (route filter and backend)
		require.Len(t, resources.ExtensionServerPolicies, 1)
		require.Len(t, resources.ExtensionRefFilters, 2)

		// Verify the custom resources get the default namespace
		require.Equal(t, config.DefaultNamespace, resources.ExtensionServerPolicies[0].GetNamespace())
		require.Equal(t, config.DefaultNamespace, resources.ExtensionRefFilters[0].GetNamespace())
		require.Equal(t, config.DefaultNamespace, resources.ExtensionRefFilters[1].GetNamespace())
	})
}
