package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
)

func TestProviderReferenceStore(t *testing.T) {
	cache := newProviderReferenceStore()

	testCases := []struct {
		name string
		test func(t *testing.T, c *providerReferenceStore)
	}{
		{
			name: "route to service mappings",
			test: testRouteToServicesMappings,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			tc.test(t, cache)
		})
	}
}

func testRouteToServicesMappings(t *testing.T, cache *providerReferenceStore) {
	httpr1 := ObjectKindNamespacedName{"HTTPRoute", "ns1", "r1"}
	tlsr1 := ObjectKindNamespacedName{"TLSRoute", "ns1", "r1"}

	ns1svc1 := types.NamespacedName{Namespace: "ns1", Name: "svc1"}
	ns1svc2 := types.NamespacedName{Namespace: "ns1", Name: "svc2"}

	// Add HTTPRoute/ns1/r1 -> ns1/svc1 mapping
	cache.updateRouteToServicesMapping(httpr1, ns1svc1)
	require.Equal(t, map[types.NamespacedName]struct{}{ns1svc1: {}}, cache.getRouteToServicesMapping(httpr1))

	// Add HTTPRoute/ns1/r1 -> ns1/svc2 mapping
	// Add TLSRoute/ns1/r1 -> ns1/svc2 mapping
	cache.updateRouteToServicesMapping(tlsr1, ns1svc2)
	cache.updateRouteToServicesMapping(httpr1, ns1svc2)
	require.Equal(t, map[types.NamespacedName]struct{}{ns1svc2: {}}, cache.getRouteToServicesMapping(tlsr1))
	require.Equal(t, map[types.NamespacedName]struct{}{ns1svc1: {}, ns1svc2: {}}, cache.getRouteToServicesMapping(httpr1))

	// Remove HTTPRoute/ns1/r1 -> ns1/svc1 mapping
	cache.removeRouteToServicesMapping(httpr1, ns1svc1)
	require.Equal(t, map[types.NamespacedName]struct{}{ns1svc2: {}}, cache.getRouteToServicesMapping(httpr1))

	// Remove TLSRoute/ns1/r1 -> ns1/svc2 mapping
	cache.removeRouteToServicesMapping(tlsr1, ns1svc2)
	require.Equal(t, map[types.NamespacedName]struct{}(nil), cache.getRouteToServicesMapping(tlsr1))

	// Verify that ns1svc2 is still referred by another route (HTTPRoute/ns1/r1)
	require.Equal(t, true, cache.isServiceReferredByRoutes(ns1svc2))

	// Verify that ns1svc1 is not referred by any other route
	require.Equal(t, false, cache.isServiceReferredByRoutes(ns1svc1))
}
