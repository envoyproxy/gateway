// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package registry

import (
	"fmt"
	"testing"

	cluster "github.com/envoyproxy/go-control-plane/envoy/config/cluster/v3"
	endpoint "github.com/envoyproxy/go-control-plane/envoy/config/endpoint/v3"
	listener "github.com/envoyproxy/go-control-plane/envoy/config/listener/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	tls "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/sets"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/types"
	"github.com/envoyproxy/gateway/internal/ir"
)

// mockXDSHookClient implements types.XDSHookClient for testing.
type mockXDSHookClient struct {
	postRouteModifyHook        func(r *route.Route, hostnames []string, resources []*unstructured.Unstructured) (*route.Route, error)
	postVirtualHostModifyHook  func(vh *route.VirtualHost) (*route.VirtualHost, error)
	postEndpointsModifyHook    func(loadAssignment *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error)
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

func (m *mockXDSHookClient) PostEndpointsModifyHook(loadAssignment *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
	if m.postEndpointsModifyHook != nil {
		return m.postEndpointsModifyHook(loadAssignment)
	}
	return loadAssignment, nil
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
				r.Name += "-ext1"
				return r, nil
			},
		}
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				r.Name += "-ext2"
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
		require.Equal(t, "test-ext1-ext2", result.Name)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		clientErr := &mockXDSHookClient{
			postRouteModifyHook: func(_ *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				return nil, fmt.Errorf("extension error")
			},
		}
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				r.Name += "-ext2"
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
		require.Equal(t, "test-ext2", result.Name)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		clientErr := &mockXDSHookClient{
			postRouteModifyHook: func(_ *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				return nil, fmt.Errorf("extension error")
			},
		}
		client2Called := false
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, _ []*unstructured.Unstructured) (*route.Route, error) {
				client2Called = true
				r.Name += "-ext2"
				return r, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: clientErr, failOpen: false},
				{name: "ext2", client: client2, failOpen: false},
			},
		}

		input := &route.Route{Name: "test"}
		_, err := composite.PostRouteModifyHook(input, nil, nil)
		require.Equal(t, fmt.Errorf(`extension "ext1": %w`, fmt.Errorf("extension error")), err)
		require.False(t, client2Called)
	})

	t.Run("per-extension resource filtering", func(t *testing.T) {
		var ext1Resources, ext2Resources []*unstructured.Unstructured

		client1 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, resources []*unstructured.Unstructured) (*route.Route, error) {
				ext1Resources = resources
				return r, nil
			},
		}
		client2 := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, resources []*unstructured.Unstructured) (*route.Route, error) {
				ext2Resources = resources
				return r, nil
			},
		}

		fooV1FooFilterGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooFilter"}
		barV1BarBackendGVK := schema.GroupVersionKind{Group: "bar.io", Version: "v1", Kind: "BarBackend"}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:          "ext1",
					client:        client1,
					resourceGKSet: sets.New(fooV1FooFilterGVK.GroupKind()),
				},
				{
					name:          "ext2",
					client:        client2,
					resourceGKSet: sets.New(barV1BarBackendGVK.GroupKind()),
				},
			},
		}

		allResources := []*unstructured.Unstructured{
			{Object: map[string]interface{}{"apiVersion": "foo.io/v1", "kind": "FooFilter"}},
			{Object: map[string]interface{}{"apiVersion": "bar.io/v1", "kind": "BarBackend"}},
		}

		_, err := composite.PostRouteModifyHook(&route.Route{Name: "test"}, nil, allResources)
		require.NoError(t, err)

		require.Len(t, ext1Resources, 1)
		assert.Equal(t, fooV1FooFilterGVK, ext1Resources[0].GetObjectKind().GroupVersionKind())

		require.Len(t, ext2Resources, 1)
		assert.Equal(t, barV1BarBackendGVK, ext2Resources[0].GetObjectKind().GroupVersionKind())
	})

	t.Run("resource matching ignores version (group/kind only)", func(t *testing.T) {
		// Regression test: a CRD served at v2 must still reach an extension that
		// declared v1 for the same group/kind, matching single-manager behavior.
		var received []*unstructured.Unstructured
		client := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, resources []*unstructured.Unstructured) (*route.Route, error) {
				received = resources
				return r, nil
			},
		}
		declaredGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooFilter"}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client, resourceGKSet: sets.New(declaredGVK.GroupKind())},
			},
		}
		servedAtV2 := []*unstructured.Unstructured{
			{Object: map[string]interface{}{"apiVersion": "foo.io/v2", "kind": "FooFilter"}},
		}

		_, err := composite.PostRouteModifyHook(&route.Route{Name: "test"}, nil, servedAtV2)
		require.NoError(t, err)
		require.Len(t, received, 1)
	})

	t.Run("no resourceGKSet passes all resources", func(t *testing.T) {
		var receivedResources []*unstructured.Unstructured
		client := &mockXDSHookClient{
			postRouteModifyHook: func(r *route.Route, _ []string, resources []*unstructured.Unstructured) (*route.Route, error) {
				receivedResources = resources
				return r, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client},
			},
		}

		allResources := []*unstructured.Unstructured{
			{Object: map[string]interface{}{"apiVersion": "foo.io/v1", "kind": "FooFilter"}},
			{Object: map[string]interface{}{"apiVersion": "bar.io/v1", "kind": "BarBackend"}},
		}

		_, err := composite.PostRouteModifyHook(&route.Route{Name: "test"}, nil, allResources)
		require.NoError(t, err)
		assert.Len(t, receivedResources, 2)
	})
}

func TestCompositeHookClient_PostVirtualHostModifyHook(t *testing.T) {
	t.Run("chains two clients", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
				vh.Name += "-ext1"
				return vh, nil
			},
		}
		client2 := &mockXDSHookClient{
			postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
				vh.Name += "-ext2"
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
		require.Equal(t, "test-ext1-ext2", result.Name)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postVirtualHostModifyHook: func(_ *route.VirtualHost) (*route.VirtualHost, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: true},
				{name: "ext2", client: &mockXDSHookClient{
					postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
						vh.Name += "-ext2"
						return vh, nil
					},
				}},
			},
		}

		result, err := composite.PostVirtualHostModifyHook(&route.VirtualHost{Name: "test"})
		require.NoError(t, err)
		require.Equal(t, "test-ext2", result.Name)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		client2Called := false
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postVirtualHostModifyHook: func(_ *route.VirtualHost) (*route.VirtualHost, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: false},
				{name: "ext2", client: &mockXDSHookClient{
					postVirtualHostModifyHook: func(vh *route.VirtualHost) (*route.VirtualHost, error) {
						client2Called = true
						vh.Name += "-ext2"
						return vh, nil
					},
				}},
			},
		}

		_, err := composite.PostVirtualHostModifyHook(&route.VirtualHost{Name: "test"})
		require.Equal(t, fmt.Errorf(`extension "ext1": %w`, fmt.Errorf("extension error")), err)
		require.False(t, client2Called)
	})
}

func TestCompositeHookClient_PostEndpointsModifyHook(t *testing.T) {
	t.Run("chains two clients", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postEndpointsModifyHook: func(la *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
				la.ClusterName += "-ext1"
				return la, nil
			},
		}
		client2 := &mockXDSHookClient{
			postEndpointsModifyHook: func(la *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
				la.ClusterName += "-ext2"
				return la, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client1},
				{name: "ext2", client: client2},
			},
		}

		input := &endpoint.ClusterLoadAssignment{ClusterName: "test"}
		result, err := composite.PostEndpointsModifyHook(input)
		require.NoError(t, err)
		require.Equal(t, "test-ext1-ext2", result.ClusterName)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postEndpointsModifyHook: func(_ *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: true},
				{name: "ext2", client: &mockXDSHookClient{
					postEndpointsModifyHook: func(la *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
						la.ClusterName += "-ext2"
						return la, nil
					},
				}},
			},
		}

		result, err := composite.PostEndpointsModifyHook(&endpoint.ClusterLoadAssignment{ClusterName: "test"})
		require.NoError(t, err)
		require.Equal(t, "test-ext2", result.ClusterName)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		client2Called := false
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postEndpointsModifyHook: func(_ *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: false},
				{name: "ext2", client: &mockXDSHookClient{
					postEndpointsModifyHook: func(la *endpoint.ClusterLoadAssignment) (*endpoint.ClusterLoadAssignment, error) {
						client2Called = true
						la.ClusterName += "-ext2"
						return la, nil
					},
				}},
			},
		}

		_, err := composite.PostEndpointsModifyHook(&endpoint.ClusterLoadAssignment{ClusterName: "test"})
		require.Equal(t, fmt.Errorf(`extension "ext1": %w`, fmt.Errorf("extension error")), err)
		require.False(t, client2Called)
	})
}

func TestCompositeHookClient_PostHTTPListenerModifyHook(t *testing.T) {
	t.Run("chains two clients", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
				l.Name += "-ext1"
				return l, nil
			},
		}
		client2 := &mockXDSHookClient{
			postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
				l.Name += "-ext2"
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
		require.Equal(t, "test-ext1-ext2", result.Name)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postHTTPListenerModifyHook: func(_ *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: true},
				{name: "ext2", client: &mockXDSHookClient{
					postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
						l.Name += "-ext2"
						return l, nil
					},
				}},
			},
		}

		result, err := composite.PostHTTPListenerModifyHook(&listener.Listener{Name: "test"}, nil)
		require.NoError(t, err)
		require.Equal(t, "test-ext2", result.Name)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		client2Called := false
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postHTTPListenerModifyHook: func(_ *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: false},
				{name: "ext2", client: &mockXDSHookClient{
					postHTTPListenerModifyHook: func(l *listener.Listener, _ []*unstructured.Unstructured) (*listener.Listener, error) {
						client2Called = true
						l.Name += "-ext2"
						return l, nil
					},
				}},
			},
		}

		_, err := composite.PostHTTPListenerModifyHook(&listener.Listener{Name: "test"}, nil)
		require.Equal(t, fmt.Errorf(`extension "ext1": %w`, fmt.Errorf("extension error")), err)
		require.False(t, client2Called)
	})

	t.Run("per-extension policy filtering", func(t *testing.T) {
		var ext1Resources, ext2Resources []*unstructured.Unstructured

		client1 := &mockXDSHookClient{
			postHTTPListenerModifyHook: func(l *listener.Listener, resources []*unstructured.Unstructured) (*listener.Listener, error) {
				ext1Resources = resources
				return l, nil
			},
		}
		client2 := &mockXDSHookClient{
			postHTTPListenerModifyHook: func(l *listener.Listener, resources []*unstructured.Unstructured) (*listener.Listener, error) {
				ext2Resources = resources
				return l, nil
			},
		}

		fooV1FooPolicyGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooPolicy"}
		barV1BarPolicyGVK := schema.GroupVersionKind{Group: "bar.io", Version: "v1", Kind: "BarPolicy"}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:        "ext1",
					client:      client1,
					policyGKSet: sets.New(fooV1FooPolicyGVK.GroupKind()),
				},
				{
					name:        "ext2",
					client:      client2,
					policyGKSet: sets.New(barV1BarPolicyGVK.GroupKind()),
				},
			},
		}

		allResources := []*unstructured.Unstructured{
			{Object: map[string]interface{}{"apiVersion": "foo.io/v1", "kind": "FooPolicy"}},
			{Object: map[string]interface{}{"apiVersion": "bar.io/v1", "kind": "BarPolicy"}},
		}

		_, err := composite.PostHTTPListenerModifyHook(&listener.Listener{Name: "test"}, allResources)
		require.NoError(t, err)

		require.Len(t, ext1Resources, 1)
		assert.Equal(t, fooV1FooPolicyGVK, ext1Resources[0].GetObjectKind().GroupVersionKind())

		require.Len(t, ext2Resources, 1)
		assert.Equal(t, barV1BarPolicyGVK, ext2Resources[0].GetObjectKind().GroupVersionKind())
	})
}

func TestCompositeHookClient_PostClusterModifyHook(t *testing.T) {
	t.Run("chains two clients", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
				c.Name += "-ext1"
				return c, nil
			},
		}
		client2 := &mockXDSHookClient{
			postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
				c.Name += "-ext2"
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
		require.Equal(t, "test-ext1-ext2", result.Name)
	})

	t.Run("failOpen skips erroring extension", func(t *testing.T) {
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postClusterModifyHook: func(_ *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: true},
				{name: "ext2", client: &mockXDSHookClient{
					postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
						c.Name += "-ext2"
						return c, nil
					},
				}},
			},
		}

		result, err := composite.PostClusterModifyHook(&cluster.Cluster{Name: "test"}, nil)
		require.NoError(t, err)
		require.Equal(t, "test-ext2", result.Name)
	})

	t.Run("failClosed stops chain", func(t *testing.T) {
		client2Called := false
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postClusterModifyHook: func(_ *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
						return nil, fmt.Errorf("extension error")
					},
				}, failOpen: false},
				{name: "ext2", client: &mockXDSHookClient{
					postClusterModifyHook: func(c *cluster.Cluster, _ []*unstructured.Unstructured) (*cluster.Cluster, error) {
						client2Called = true
						c.Name += "-ext2"
						return c, nil
					},
				}},
			},
		}

		_, err := composite.PostClusterModifyHook(&cluster.Cluster{Name: "test"}, nil)
		require.Error(t, err)
		require.False(t, client2Called)
	})

	t.Run("per-extension resource filtering", func(t *testing.T) {
		var ext1Resources, ext2Resources []*unstructured.Unstructured

		client1 := &mockXDSHookClient{
			postClusterModifyHook: func(c *cluster.Cluster, resources []*unstructured.Unstructured) (*cluster.Cluster, error) {
				ext1Resources = resources
				return c, nil
			},
		}
		client2 := &mockXDSHookClient{
			postClusterModifyHook: func(c *cluster.Cluster, resources []*unstructured.Unstructured) (*cluster.Cluster, error) {
				ext2Resources = resources
				return c, nil
			},
		}

		fooV1FooBackendGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooBackend"}
		barV1BarBackend := schema.GroupVersionKind{Group: "bar.io", Version: "v1", Kind: "BarBackend"}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:          "ext1",
					client:        client1,
					resourceGKSet: sets.New(fooV1FooBackendGVK.GroupKind()),
				},
				{
					name:          "ext2",
					client:        client2,
					resourceGKSet: sets.New(barV1BarBackend.GroupKind()),
				},
			},
		}

		allResources := []*unstructured.Unstructured{
			{Object: map[string]interface{}{"apiVersion": "foo.io/v1", "kind": "FooBackend"}},
			{Object: map[string]interface{}{"apiVersion": "bar.io/v1", "kind": "BarBackend"}},
		}

		_, err := composite.PostClusterModifyHook(&cluster.Cluster{Name: "test"}, allResources)
		require.NoError(t, err)

		require.Len(t, ext1Resources, 1)
		assert.Equal(t, fooV1FooBackendGVK, ext1Resources[0].GetObjectKind().GroupVersionKind())

		require.Len(t, ext2Resources, 1)
		assert.Equal(t, barV1BarBackend, ext2Resources[0].GetObjectKind().GroupVersionKind())
	})
}

func TestCompositeHookClient_PostTranslateModifyHook(t *testing.T) {
	t.Run("chains resource lists", func(t *testing.T) {
		client1 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				clusters = append(clusters, &cluster.Cluster{Name: "from-ext1"})
				secrets = append(secrets, &tls.Secret{Name: "from-ext1"})
				listeners = append(listeners, &listener.Listener{Name: "from-ext1"})
				routes = append(routes, &route.RouteConfiguration{Name: "from-ext1"})
				return clusters, secrets, listeners, routes, nil
			},
		}
		client2 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				clusters = append(clusters, &cluster.Cluster{Name: "from-ext2"})
				secrets = append(secrets, &tls.Secret{Name: "from-ext2"})
				listeners = append(listeners, &listener.Listener{Name: "from-ext2"})
				routes = append(routes, &route.RouteConfiguration{Name: "from-ext2"})
				return clusters, secrets, listeners, routes, nil
			},
		}

		translationConfig := &egv1a1.TranslationConfig{
			Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)},
			Route:    &egv1a1.RouteTranslationConfig{IncludeAll: new(true)},
			Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: new(true)},
			Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: new(true)},
		}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: client1, translationConfig: translationConfig},
				{name: "ext2", client: client2, translationConfig: translationConfig},
			},
		}

		rc, rs, rl, rr, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, rc, 2)
		assert.Equal(t, "from-ext1", rc[0].Name)
		assert.Equal(t, "from-ext2", rc[1].Name)
		require.Len(t, rs, 2)
		assert.Equal(t, "from-ext1", rs[0].Name)
		assert.Equal(t, "from-ext2", rs[1].Name)
		require.Len(t, rl, 2)
		assert.Equal(t, "from-ext1", rl[0].Name)
		assert.Equal(t, "from-ext2", rl[1].Name)
		require.Len(t, rr, 2)
		assert.Equal(t, "from-ext1", rr[0].Name)
		assert.Equal(t, "from-ext2", rr[1].Name)
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

		fooV1FooPolicyGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooPolicy"}
		barV1BarPolicyGVK := schema.GroupVersionKind{Group: "bar.io", Version: "v1", Kind: "BarPolicy"}
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:        "ext1",
					client:      client1,
					policyGKSet: sets.New(fooV1FooPolicyGVK.GroupKind()),
				},
				{
					name:        "ext2",
					client:      client2,
					policyGKSet: sets.New(barV1BarPolicyGVK.GroupKind()),
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
		assert.Equal(t, fooV1FooPolicyGVK, ext1Policies[0].Object.GetObjectKind().GroupVersionKind())

		// ext2 should only see BarPolicy
		require.Len(t, ext2Policies, 1)
		assert.Equal(t, barV1BarPolicyGVK, ext2Policies[0].Object.GetObjectKind().GroupVersionKind())
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
						Cluster: &egv1a1.ClusterTranslationConfig{IncludeAll: new(true)},
						Secret:  &egv1a1.SecretTranslationConfig{IncludeAll: new(false)},
						// Listener and Route nil → defaults false
					},
				},
				{
					name:   "ext2",
					client: client2,
					translationConfig: &egv1a1.TranslationConfig{
						Cluster:  &egv1a1.ClusterTranslationConfig{IncludeAll: new(false)},
						Secret:   &egv1a1.SecretTranslationConfig{IncludeAll: new(false)},
						Listener: &egv1a1.ListenerTranslationConfig{IncludeAll: new(true)},
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

	t.Run("failOpen skips erroring extension in PostTranslateModifyHook", func(t *testing.T) {
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postTranslateModifyHook: func(_ []*cluster.Cluster, _ []*tls.Secret, _ []*listener.Listener, _ []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
						return nil, nil, nil, nil, fmt.Errorf("extension error")
					},
				}, failOpen: true},
				{name: "ext2", client: &mockXDSHookClient{
					postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
						clusters = append(clusters, &cluster.Cluster{Name: "from-ext2"})
						return clusters, secrets, listeners, routes, nil
					},
				}},
			},
		}

		rc, _, _, _, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, nil)
		require.NoError(t, err)
		require.Len(t, rc, 1)
		assert.Equal(t, "from-ext2", rc[0].Name)
	})

	t.Run("failClosed stops chain in PostTranslateModifyHook", func(t *testing.T) {
		client2Called := false
		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{name: "ext1", client: &mockXDSHookClient{
					postTranslateModifyHook: func(_ []*cluster.Cluster, _ []*tls.Secret, _ []*listener.Listener, _ []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
						return nil, nil, nil, nil, fmt.Errorf("extension error")
					},
				}, failOpen: false},
				{name: "ext2", client: &mockXDSHookClient{
					postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
						client2Called = true
						clusters = append(clusters, &cluster.Cluster{Name: "from-ext2"})
						return clusters, secrets, listeners, routes, nil
					},
				}},
			},
		}

		_, _, _, _, err := composite.PostTranslateModifyHook(nil, nil, nil, nil, nil)
		require.Equal(t, fmt.Errorf(`extension "ext1": %w`, fmt.Errorf("extension error")), err)
		require.False(t, client2Called)
	})

	t.Run("per-extension secrets and routes gating", func(t *testing.T) {
		// ext1 wants secrets only, ext2 wants routes only
		client1 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				assert.NotNil(t, secrets)
				assert.Nil(t, routes)
				secrets = append(secrets, &tls.Secret{Name: "from-ext1"})
				return clusters, secrets, listeners, routes, nil
			},
		}
		client2 := &mockXDSHookClient{
			postTranslateModifyHook: func(clusters []*cluster.Cluster, secrets []*tls.Secret, listeners []*listener.Listener, routes []*route.RouteConfiguration, _ []*ir.UnstructuredRef) ([]*cluster.Cluster, []*tls.Secret, []*listener.Listener, []*route.RouteConfiguration, error) {
				assert.Nil(t, secrets)
				assert.NotNil(t, routes)
				routes = append(routes, &route.RouteConfiguration{Name: "from-ext2"})
				return clusters, secrets, listeners, routes, nil
			},
		}

		composite := &compositeXDSHookClient{
			entries: []hookClientEntry{
				{
					name:   "ext1",
					client: client1,
					translationConfig: &egv1a1.TranslationConfig{
						Secret:  &egv1a1.SecretTranslationConfig{IncludeAll: new(true)},
						Cluster: &egv1a1.ClusterTranslationConfig{IncludeAll: new(false)},
					},
				},
				{
					name:   "ext2",
					client: client2,
					translationConfig: &egv1a1.TranslationConfig{
						Route:   &egv1a1.RouteTranslationConfig{IncludeAll: new(true)},
						Cluster: &egv1a1.ClusterTranslationConfig{IncludeAll: new(false)},
						Secret:  &egv1a1.SecretTranslationConfig{IncludeAll: new(false)},
					},
				},
			},
		}

		initialSecrets := []*tls.Secret{{Name: "original-secret"}}
		initialRoutes := []*route.RouteConfiguration{{Name: "original-route"}}

		_, rs, _, rr, err := composite.PostTranslateModifyHook(nil, initialSecrets, nil, initialRoutes, nil)
		require.NoError(t, err)

		require.Len(t, rs, 2)
		assert.Equal(t, "original-secret", rs[0].Name)
		assert.Equal(t, "from-ext1", rs[1].Name)

		require.Len(t, rr, 2)
		assert.Equal(t, "original-route", rr[0].Name)
		assert.Equal(t, "from-ext2", rr[1].Name)
	})

	t.Run("filterPoliciesByGK skips nil entries", func(t *testing.T) {
		fooV1FooPolicyGVK := schema.GroupVersionKind{Group: "foo.io", Version: "v1", Kind: "FooPolicy"}
		gkSet := sets.New(fooV1FooPolicyGVK.GroupKind())
		policies := []*ir.UnstructuredRef{
			nil,
			{Object: nil},
			{Object: &unstructured.Unstructured{Object: map[string]interface{}{
				"apiVersion": "foo.io/v1",
				"kind":       "FooPolicy",
			}}},
		}

		filtered := filterPoliciesByGK(policies, gkSet)
		require.Len(t, filtered, 1)
		assert.Equal(t, fooV1FooPolicyGVK, filtered[0].Object.GetObjectKind().GroupVersionKind())
	})

	t.Run("no policyGKSet passes all policies", func(t *testing.T) {
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
