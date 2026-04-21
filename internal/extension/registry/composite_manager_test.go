// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	extTypes "github.com/envoyproxy/gateway/internal/extension/types"
)

// mockManager implements extTypes.Manager for testing.
type mockManager struct {
	hasExtension      func(g gwapiv1.Group, k gwapiv1.Kind) bool
	failOpen          bool
	translationConfig *egv1a1.TranslationConfig
	preHookClient     extTypes.XDSHookClient
	preHookErr        error
	postHookClient    extTypes.XDSHookClient
	postHookErr       error
}

func (m *mockManager) HasExtension(g gwapiv1.Group, k gwapiv1.Kind) bool {
	if m.hasExtension != nil {
		return m.hasExtension(g, k)
	}
	return false
}

func (m *mockManager) FailOpen() bool {
	return m.failOpen
}

func (m *mockManager) GetTranslationHookConfig() *egv1a1.TranslationConfig {
	return m.translationConfig
}

func (m *mockManager) GetPreXDSHookClient(_ egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	return m.preHookClient, m.preHookErr
}

func (m *mockManager) GetPostXDSHookClient(_ egv1a1.XDSTranslatorHook) (extTypes.XDSHookClient, error) {
	return m.postHookClient, m.postHookErr
}

func (m *mockManager) CleanupHookConns() {}

func TestCompositeManager_HasExtension(t *testing.T) {
	mgr1 := &mockManager{
		hasExtension: func(g gwapiv1.Group, k gwapiv1.Kind) bool {
			return g == "foo.io" && k == "Foo"
		},
	}
	mgr2 := &mockManager{
		hasExtension: func(g gwapiv1.Group, k gwapiv1.Kind) bool {
			return g == "bar.io" && k == "Bar"
		},
	}

	composite := NewCompositeManager([]namedManager{
		{name: "mgr1", manager: mgr1},
		{name: "mgr2", manager: mgr2},
	})

	// Union semantics: true if any child has the extension
	assert.True(t, composite.HasExtension("foo.io", "Foo"))
	assert.True(t, composite.HasExtension("bar.io", "Bar"))
	assert.False(t, composite.HasExtension("baz.io", "Baz"))
}

func TestCompositeManager_FailOpen(t *testing.T) {
	t.Run("all fail open", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{failOpen: true}},
			{name: "mgr2", manager: &mockManager{failOpen: true}},
		})
		assert.True(t, composite.FailOpen())
	})

	t.Run("one fail closed", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{failOpen: true}},
			{name: "mgr2", manager: &mockManager{failOpen: false}},
		})
		assert.False(t, composite.FailOpen())
	})

	t.Run("all fail closed", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{failOpen: false}},
			{name: "mgr2", manager: &mockManager{failOpen: false}},
		})
		assert.False(t, composite.FailOpen())
	})
}

func TestCompositeManager_GetTranslationHookConfig(t *testing.T) {
	t.Run("nil if no configs", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{}},
			{name: "mgr2", manager: &mockManager{}},
		})
		assert.Nil(t, composite.GetTranslationHookConfig())
	})

	t.Run("OR semantics merges all enabled types", func(t *testing.T) {
		mgr1 := &mockManager{
			translationConfig: &egv1a1.TranslationConfig{
				Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)},
			},
		}
		mgr2 := &mockManager{
			translationConfig: &egv1a1.TranslationConfig{
				Route: &egv1a1.RouteTranslationConfig{IncludeAll: new(true)},
			},
		}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: mgr1},
			{name: "mgr2", manager: mgr2},
		})
		tc := composite.GetTranslationHookConfig()
		require.NotNil(t, tc)
		// Listener and Route explicitly enabled by mgr1/mgr2
		assert.NotNil(t, tc.Listener)
		assert.NotNil(t, tc.Route)
		// Cluster and Secret included because they default to true
		assert.NotNil(t, tc.Cluster)
		assert.NotNil(t, tc.Secret)
	})

	t.Run("OR semantics includes if at least one enables", func(t *testing.T) {
		mgr1 := &mockManager{
			translationConfig: &egv1a1.TranslationConfig{
				Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(false)},
				Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: new(false)},
				Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: new(false)},
			},
		}
		mgr2 := &mockManager{
			translationConfig: &egv1a1.TranslationConfig{
				Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)},
				Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: new(false)},
				Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: new(false)},
			},
		}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: mgr1},
			{name: "mgr2", manager: mgr2},
		})
		tc := composite.GetTranslationHookConfig()
		require.NotNil(t, tc)
		// Listener: mgr1=false, mgr2=true → included
		assert.NotNil(t, tc.Listener)
		// Route: neither enables → excluded
		assert.Nil(t, tc.Route)
		// Cluster: both explicitly false → excluded
		assert.Nil(t, tc.Cluster)
		// Secret: both explicitly false → excluded
		assert.Nil(t, tc.Secret)
	})
}

func TestCompositeManager_GetTranslationHookConfig_MixedNilAndNonNil(t *testing.T) {
	// One manager has nil config, the other has a non-nil config.
	// The nil config's ShouldInclude* methods default to true for clusters/secrets.
	mgr1 := &mockManager{
		translationConfig: nil, // nil → ShouldInclude* returns default (clusters/secrets=true, listeners/routes=false)
	}
	mgr2 := &mockManager{
		translationConfig: &egv1a1.TranslationConfig{
			Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)},
		},
	}

	composite := NewCompositeManager([]namedManager{
		{name: "mgr1", manager: mgr1},
		{name: "mgr2", manager: mgr2},
	})

	tc := composite.GetTranslationHookConfig()
	require.NotNil(t, tc)
	// mgr2 explicitly enables listeners
	assert.NotNil(t, tc.Listener)
	// nil config defaults: clusters and secrets are true
	assert.NotNil(t, tc.Cluster)
	assert.NotNil(t, tc.Secret)
}

func TestCompositeManager_GetPreXDSHookClient(t *testing.T) {
	t.Run("nil when no children have hooks", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{}},
		})
		client, err := composite.GetPreXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		assert.Nil(t, client)
	})

	t.Run("returns composite when children have hooks", func(t *testing.T) {
		mockClient := &mockXDSHookClient{}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{preHookClient: mockClient}},
		})
		client, err := composite.GetPreXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		require.NotNil(t, client)
		_, ok := client.(*compositeXDSHookClient)
		assert.True(t, ok)
	})

	t.Run("returns error when child errors", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{preHookErr: fmt.Errorf("connection failed")}},
		})
		client, err := composite.GetPreXDSHookClient(egv1a1.XDSRoute)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "connection failed")
	})

	t.Run("skips erroring child when failOpen is true", func(t *testing.T) {
		mockClient := &mockXDSHookClient{}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{preHookErr: fmt.Errorf("connection failed"), failOpen: true}},
			{name: "mgr2", manager: &mockManager{preHookClient: mockClient}},
		})
		client, err := composite.GetPreXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		require.NotNil(t, client)
		compositeClient, ok := client.(*compositeXDSHookClient)
		assert.True(t, ok)
		require.Len(t, compositeClient.entries, 1)
		assert.Equal(t, "mgr2", compositeClient.entries[0].name)
	})

	t.Run("returns error when failOpen is false", func(t *testing.T) {
		mockClient := &mockXDSHookClient{}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{preHookErr: fmt.Errorf("connection failed"), failOpen: false}},
			{name: "mgr2", manager: &mockManager{preHookClient: mockClient}},
		})
		client, err := composite.GetPreXDSHookClient(egv1a1.XDSRoute)
		require.Error(t, err)
		assert.Nil(t, client)
	})
}

func TestCompositeManager_GetPostXDSHookClient(t *testing.T) {
	t.Run("nil when no children have hooks", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{}},
		})
		client, err := composite.GetPostXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		assert.Nil(t, client)
	})

	t.Run("returns composite when children have hooks", func(t *testing.T) {
		mockClient := &mockXDSHookClient{}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{postHookClient: mockClient}},
			{name: "mgr2", manager: &mockManager{postHookClient: mockClient}},
		})
		client, err := composite.GetPostXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		require.NotNil(t, client)
		compositeClient, ok := client.(*compositeXDSHookClient)
		assert.True(t, ok)
		assert.Len(t, compositeClient.entries, 2)
	})

	t.Run("returns error when failOpen is false", func(t *testing.T) {
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{postHookErr: fmt.Errorf("connection failed")}},
		})
		client, err := composite.GetPostXDSHookClient(egv1a1.XDSRoute)
		require.Error(t, err)
		assert.Nil(t, client)
		assert.Contains(t, err.Error(), "connection failed")
	})

	t.Run("skips erroring child when failOpen is true", func(t *testing.T) {
		mockClient := &mockXDSHookClient{}
		composite := NewCompositeManager([]namedManager{
			{name: "mgr1", manager: &mockManager{postHookErr: fmt.Errorf("connection failed"), failOpen: true}},
			{name: "mgr2", manager: &mockManager{postHookClient: mockClient}},
		})
		client, err := composite.GetPostXDSHookClient(egv1a1.XDSRoute)
		require.NoError(t, err)
		require.NotNil(t, client)
		compositeClient, ok := client.(*compositeXDSHookClient)
		assert.True(t, ok)
		require.Len(t, compositeClient.entries, 1)
		assert.Equal(t, "mgr2", compositeClient.entries[0].name)
	})
}

func TestCompositeManager_CleanupHookConns(t *testing.T) {
	called := make([]string, 0)
	composite := NewCompositeManager([]namedManager{
		{name: "mgr1", cleanupHookConn: func() { called = append(called, "mgr1") }},
		{name: "mgr2", cleanupHookConn: func() { called = append(called, "mgr2") }},
	})
	composite.CleanupHookConns()
	assert.Equal(t, []string{"mgr1", "mgr2"}, called)
}
