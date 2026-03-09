// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"fmt"
	"testing"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
)

// mockXDSHookClient implements types.XDSHookClient for testing.
type mockXDSHookClient struct {
	postRouteModifyHook        func(r *route.Route, hostnames []string, resources []*unstructured.Unstructured) (*route.Route, error)
	postVirtualHostModifyHook  func(vh *route.VirtualHost) (*route.VirtualHost, error)
	postHTTPListenerModifyHook func(l *listener.Listener, resources []*unstructured.Unstructured) (*listener.Listener, error)
	postClusterModifyHook      func(c *cluster.Cluster, resources []*unstructured.Unstructured) (*cluster.Cluster, error)
	postTranslateModifyHook    func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, policies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error)
}

var _ types.XDSHookClient = (*mockXDSHookClient)(nil)

func (m *mockXDSHookClient) PostRouteModifyHook(r *route.Route, hostnames []string, resources []*unstructured.Unstructured) (*route.Route, error) {
	if m.postRouteModifyHook != nil {
		return m.postRouteModifyHook(r, hostnames, resources)
	}
	return r, nil
}

func (m *mockXDSHookClient) PostVirtualHostModifyHook(vh *route.VirtualHost) (*route.VirtualHost, error) {
	if m.postVirtualHostModifyHook != nil {
		return m.postVirtualHostModifyHook(vh)
	}
	return vh, nil
}

func (m *mockXDSHookClient) PostHTTPListenerModifyHook(l *listener.Listener, resources []*unstructured.Unstructured) (*listener.Listener, error) {
	if m.postHTTPListenerModifyHook != nil {
		return m.postHTTPListenerModifyHook(l, resources)
	}
	return l, nil
}

func (m *mockXDSHookClient) PostClusterModifyHook(c *cluster.Cluster, resources []*unstructured.Unstructured) (*cluster.Cluster, error) {
	if m.postClusterModifyHook != nil {
		return m.postClusterModifyHook(c, resources)
	}
	return c, nil
}

func (m *mockXDSHookClient) PostTranslateModifyHook(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, policies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
	if m.postTranslateModifyHook != nil {
		return m.postTranslateModifyHook(clusters, secrets, listeners, routes, policies)
	}
	return clusters, secrets, listeners, routes, nil
}

func TestCompositeHookClient_PostRouteModifyHook(t *testing.T) {
	t.Run("chains two clients", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				r.Name = r.Name + "-ext1"
				return r, nil
			},
		}
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				r.Name = r.Name + "-ext2"
				return r, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client1},
				{name: "ext2", client: client2},
			},
		}

		input := &route.Route{Name: "test"}
		result, err := composite.PostRouteModifyHook(input, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "test-ext1-ext2", result.Name)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		clientErr := &mockXDSHookClient{
			postRouteModifyHook: func(_ *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				return nil, fmt.Errorf("extension error")
			},
		}
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				r.Name = r.Name + "-ext2"
				return r, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: clientErr, failOpen: true},
				{name: "ext2", client: client2},
			},
		}

		input := &route.Route{Name: "test"}
		result, err := composite.PostRouteModifyHook(input, nil, nil)
		require.NoError(t, err)
		assert.Equal(t, "test-ext2", result.Name)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		clientErr := &mockXDSHookClient{
			postRouteModifyHook: func(_ *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				return nil, fmt.Errorf("extension error")
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: clientErr, failOpen: false},
			},
		}

		input := &route.Route{Name: "test"}
		_, err := composite.PostRouteModifyHook(input, nil, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), `extension "ext1"`)
	})
}

func TestCompositeHookClient_PostVirtualHostModifyHook(t *testing.T) {
	client1 := &mockXDSHookClient{
		postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
			vh.Name = vh.Name + "-ext1"
			return vh, nil
		},
	}
	client2 := &mockXDSHookClient{
		postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
			vh.Name = vh.Name + "-ext2"
			return vh, nil
		},
	}

	composite := &compositeXDSHookClient{
		entries: []hookClientEntry{
			{name: "ext1", client: client1},
			{name: "ext2", client: client2},
		},
	}

	input := &route.VirtualHost{Name: "test"}
	result, err := composite.PostVirtualHostModifyHook(input)
	require.NoError(t, err)
	assert.Equal(t, "test-ext1-ext2", result.Name)
}

func TestCompositeHookClient_PostHTTPListenerModifyHook(t *testing.T) {
	client1 := &mockXDSHookClient{
		postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
			l.Name = l.Name + "-ext1"
			return l, nil
		},
	}
	client2 := &mockXDSHookClient{
		postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
			l.Name = l.Name + "-ext2"
			return l, nil
		},
	}

	composite := &compositeXDSHookClient{
		entries: []hookClientEntry{
			{name: "ext1", client: client1},
			{name: "ext2", client: client2},
		},
	}

	input := &listener.Listener{Name: "test"}
	result, err := composite.PostHTTPListenerModifyHook(input, nil)
	require.NoError(t, err)
	assert.Equal(t, "test-ext1-ext2", result.Name)
}

func TestCompositeHookClient_PostClusterModifyHook(t *testing.T) {
	client1 := &mockXDSHookClient{
		postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
			c.Name = c.Name + "-ext1"
			return c, nil
		},
	}
	client2 := &mockXDSHookClient{
		postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
			c.Name = c.Name + "-ext2"
			return c, nil
		},
	}

	composite := &compositeXDSHookClient{
		entries: []hookClientEntry{
			{name: "ext1", client: client1},
			{name: "ext2", client: client2},
		},
	}

	input := &cluster.Cluster{Name: "test"}
	result, err := composite.PostClusterModifyHook(input, nil)
	require.NoError(t, err)
	assert.Equal(t, "test-ext1-ext2", result.Name)
}

func TestCompositeHookClient_PostTranslateModifyHook(t *testing.T) {
	t.Run("chains resource lists", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				clusters = append(clusters, &cluster.Cluster{Name: "from-ext1"})
				return clusters, secrets, listeners, routes, nil
			},
		}
		client2 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				clusters = append(clusters, &cluster.Cluster{Name: "from-ext2"})
				return clusters, secrets, listeners, routes, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client1},
				{name: "ext2", client: client2},
			},
		}

		rc, _, _, _, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, rc, 2)
		assert.Equal(t, "from-ext1", rc[0].Name)
		assert.Equal(t, "from-ext2", rc[1].Name)
	})

	t.Run("per-extension policy filtering", func(t *testing.T) {
		var ext1Policies, ext2Policies []*ir.UnstructuredRef

		client1 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, policies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				ext1Policies = policies
				return clusters, secrets, listeners, routes, nil
			},
		}
		client2 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, policies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				ext2Policies = policies
				return clusters, secrets, listeners, routes, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:   "ext1",
					client: client1,
					policyGVKSet: map[string]struct{}{
						"foo.io/v1/FooPolicy": {},
					},
				},
				{
					name:   "ext2",
					client: client2,
					policyGVKSet: map[string]struct{}{
						"bar.io/v1/BarPolicy": {},
					},
				},
			},
		}

		fooPolicy := &ir.UnstructuredRef{
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "foo.io/v1",
					"kind":       "FooPolicy",
				},
			},
		}
		barPolicy := &ir.UnstructuredRef{
			Object: &unstructured.Unstructured{
				Object: map[string]interface{}{
					"apiVersion": "bar.io/v1",
					"kind":       "BarPolicy",
				},
			},
		}

		allPolicies := []*ir.UnstructuredRef{fooPolicy, barPolicy}
		_, _, _, _, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, allPolicies)
		require.NoError(t, err)

		// ext1 should only see FooPolicy
		require.Len(t, ext1Policies, 1)
		assert.Equal(t, "FooPolicy", ext1Policies[0].Object.GetKind())

		// ext2 should only see BarPolicy
		require.Len(t, ext2Policies, 1)
		assert.Equal(t, "BarPolicy", ext2Policies[0].Object.GetKind())
	})

	t.Run("per-extension resource-type gating", func(t *testing.T) {
		// ext1 only wants clusters (not listeners/routes/secrets)
		// ext2 only wants listeners (not clusters/routes/secrets)
		client1 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				// Should receive clusters but not listeners
				assert.NotNil(t, clusters)
				assert.Nil(t, listeners)
				// Modify clusters
				clusters = append(clusters, &cluster.Cluster{Name: "from-ext1"})
				// Return non-nil listeners (should be ignored since ext1 didn't declare interest)
				return clusters, secrets, []*listener.Listener{{Name: "bad-listener"}}, routes, nil
			},
		}
		client2 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				// Should receive listeners but not clusters
				assert.Nil(t, clusters)
				assert.NotNil(t, listeners)
				// Modify listeners
				listeners = append(listeners, &listener.Listener{Name: "from-ext2"})
				return clusters, secrets, listeners, routes, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:   "ext1",
					client: client1,
					translationConfig: &egv1a1.TranslationConfig{
						Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: ptr.To(true)},
						Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: ptr.To(false)},
						// Listener and Route nil → defaults false
					},
				},
				{
					name:   "ext2",
					client: client2,
					translationConfig: &egv1a1.TranslationConfig{
						Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: ptr.To(false)},
						Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: ptr.To(false)},
						Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: ptr.To(true)},
					},
				},
			},
		}

		initialClusters := []*cluster.Cluster{{Name: "original"}}
		initialListeners := []*listener.Listener{{Name: "original-listener"}}

		rc, _, rl, _, err := composite.PostTranslateModifyHook(initialClusters, nil, initialListeners, nil, nil)
		require.NoError(t, err)

		// ext1 modified clusters
		require.Len(t, rc, 2)
		assert.Equal(t, "original", rc[0].Name)
		assert.Equal(t, "from-ext1", rc[1].Name)

		// ext1's returned listeners are ignored (it didn't declare interest)
		// ext2 modified listeners
		require.Len(t, rl, 2)
		assert.Equal(t, "original-listener", rl[0].Name)
		assert.Equal(t, "from-ext2", rl[1].Name)
	})

	t.Run("no policyGVKSet passes all policies", func(t *testing.T) {
		var receivedPolicies []*ir.UnstructuredRef

		client := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, policies []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				receivedPolicies = policies
				return clusters, secrets, listeners, routes, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client},
			},
		}

		policies := []*ir.UnstructuredRef{
			{Object: &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "foo.io/v1",
				"kind":       "FooPolicy",
			}}},
			{Object: &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "bar.io/v1",
				"kind":       "BarPolicy",
			}}},
		}

		_, _, _, _, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, policies)
		require.NoError(t, err)
		assert.Len(t, receivedPolicies, 2)
	})
}
