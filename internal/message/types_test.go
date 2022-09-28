package message

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"
)

func TestProviderResources(t *testing.T) {
	resources := new(ProviderResources)
	ns1 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns1",
		},
	}
	gc1 := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc1",
		},
	}
	gw1 := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test1",
			Namespace: "test",
		},
	}
	r1 := &gwapiv1b1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "route1",
			Namespace: "test",
		},
	}
	s1 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service1",
			Namespace: "test",
		},
	}

	// Check init state
	assert.Nil(t, resources.GetNamespaces())
	assert.Nil(t, resources.GetGatewayClasses())
	assert.Nil(t, resources.GetGateways())
	assert.Nil(t, resources.GetHTTPRoutes())
	assert.Nil(t, resources.GetServices())

	// Add resources
	resources.Namespaces.Store("test-ns1", ns1)
	resources.GatewayClasses.Store("test-gc1", gc1)

	gw1Key := types.NamespacedName{
		Namespace: gw1.GetNamespace(),
		Name:      gw1.GetName(),
	}
	resources.Gateways.Store(gw1Key, gw1)

	r1Key := types.NamespacedName{
		Namespace: r1.GetNamespace(),
		Name:      r1.GetName(),
	}
	resources.HTTPRoutes.Store(r1Key, r1)

	s1Key := types.NamespacedName{
		Namespace: s1.GetNamespace(),
		Name:      s1.GetName(),
	}
	resources.Services.Store(s1Key, s1)

	// Test
	namespaces := resources.GetNamespaces()
	assert.Equal(t, len(namespaces), 1)

	gcs := resources.GetGatewayClasses()
	assert.Equal(t, len(gcs), 1)

	gws := resources.GetGateways()
	assert.Equal(t, len(gws), 1)

	hrs := resources.GetHTTPRoutes()
	assert.Equal(t, len(hrs), 1)

	svcs := resources.GetServices()
	assert.Equal(t, len(svcs), 1)

	// Add more resources
	ns2 := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns2",
		},
	}

	gc2 := &gwapiv1b1.GatewayClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-gc2",
		},
	}
	gw2 := &gwapiv1b1.Gateway{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test2",
			Namespace: "test",
		},
	}
	r2 := &gwapiv1b1.HTTPRoute{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "route2",
			Namespace: "test",
		},
	}
	s2 := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "service2",
			Namespace: "test",
		},
	}

	resources.Namespaces.Store("test-ns2", ns2)
	resources.GatewayClasses.Store("test-gc2", gc2)
	gw2Key := types.NamespacedName{
		Namespace: gw2.GetNamespace(),
		Name:      gw2.GetName(),
	}
	resources.Gateways.Store(gw2Key, gw2)

	r2Key := types.NamespacedName{
		Namespace: r2.GetNamespace(),
		Name:      r2.GetName(),
	}
	resources.HTTPRoutes.Store(r2Key, r2)

	s2Key := types.NamespacedName{
		Namespace: s2.GetNamespace(),
		Name:      s2.GetName(),
	}
	resources.Services.Store(s2Key, s2)

	// Test contents
	namespaces = resources.GetNamespaces()
	assert.ElementsMatch(t, namespaces, []*corev1.Namespace{ns1, ns2})

	gcs = resources.GetGatewayClasses()
	assert.ElementsMatch(t, gcs, []*gwapiv1b1.GatewayClass{gc1, gc2})

	gws = resources.GetGateways()
	assert.ElementsMatch(t, gws, []*gwapiv1b1.Gateway{gw1, gw2})

	hrs = resources.GetHTTPRoutes()
	assert.ElementsMatch(t, hrs, []*gwapiv1b1.HTTPRoute{r1, r2})

	svcs = resources.GetServices()
	assert.ElementsMatch(t, svcs, []*corev1.Service{s1, s2})

	// Delete gatewayclasses
	resources.DeleteGatewayClasses()
	assert.Nil(t, resources.GetGatewayClasses())
}
