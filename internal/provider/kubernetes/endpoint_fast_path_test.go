// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/message"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

func endpointReconciler(t *testing.T) (*gatewayAPIReconciler, *corev1.Service, *discoveryv1.EndpointSlice) {
	t.Helper()
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "orders", UID: "service-one"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Name: "http", Port: 80, Protocol: corev1.ProtocolTCP}}},
	}
	slice := &discoveryv1.EndpointSlice{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "default", Name: "orders-1",
			Labels: map[string]string{discoveryv1.LabelServiceName: "orders"},
		}, AddressType: discoveryv1.AddressTypeIPv4,
		Ports:     []discoveryv1.EndpointPort{{Name: new("http"), Protocol: new(corev1.ProtocolTCP), Port: new(int32(8080))}},
		Endpoints: []discoveryv1.Endpoint{{Addresses: []string{"10.0.0.1"}}},
	}
	nn := types.NamespacedName{Namespace: service.Namespace, Name: service.Name}
	route := test.GetHTTPRoute(nn, "gateway", test.GetServiceBackendRef(nn, 80), "")
	r := &gatewayAPIReconciler{
		client: fake.NewClientBuilder().WithScheme(envoygateway.GetScheme()).
			WithObjects(service, slice, route).
			WithIndex(&gwapiv1.HTTPRoute{}, backendHTTPRouteIndex, backendHTTPRouteIndexFunc).Build(),
		log:       logging.DefaultLogger(io.Discard, egv1a1.LogLevelError),
		resources: &message.ProviderResources{},
		envoyGateway: &egv1a1.EnvoyGateway{RuntimeFlags: &egv1a1.RuntimeFlags{
			Enabled: []egv1a1.RuntimeFlag{egv1a1.EndpointFastPath}, Disabled: []egv1a1.RuntimeFlag{egv1a1.EndpointSliceIndex},
		}},
	}
	return r, service, slice
}

func TestEndpointFullCaptureRefreshesBeforeEndpointWorker(t *testing.T) {
	for _, tc := range []struct {
		name      string
		endpoints []discoveryv1.Endpoint
	}{
		{name: "changed", endpoints: []discoveryv1.Endpoint{{Addresses: []string{"10.0.0.2"}}}},
		{name: "empty"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, service, slice := endpointReconciler(t)
			key := ir.ServiceKey{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name}
			tree := &resource.Resources{Services: []*corev1.Service{service}}
			require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{tree}))
			slice.Endpoints = tc.endpoints
			require.NoError(t, r.client.Update(t.Context(), slice))

			// Full reconciliation reads the cache even before the endpoint worker runs.
			next := &resource.Resources{Services: []*corev1.Service{service}}
			require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{next}))
			require.Equal(t, tc.endpoints, next.EndpointSlices[0].Endpoints)
			require.Len(t, next.BackendEndpoints[0].EndpointsByPort[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}], len(tc.endpoints))
			require.Greater(t, next.BackendEndpoints[0].Revision, tree.BackendEndpoints[0].Revision)
			require.Equal(t, "10.0.0.1", tree.EndpointSlices[0].Endpoints[0].Addresses[0], "the earlier full capture must stay immutable")
			require.Empty(t, r.resources.EndpointUpdates.Drain(), "full collection must not queue fast-path updates")
			require.Empty(t, r.resources.EndpointUpdates.Ready())

			_, err := r.reconcileBackendEndpoints(t.Context(), key)
			require.NoError(t, err)
			require.Equal(t, &next.BackendEndpoints[0], r.resources.EndpointUpdates.Drain()[key], "the endpoint worker must queue the update even when the full path collected the same revision first")
			require.Len(t, r.resources.EndpointUpdates.Ready(), 1)
		})
	}
}

func TestEndpointCollectionServiceImportAndSliceDeduplication(t *testing.T) {
	r, service, slice := endpointReconciler(t)
	imported := &mcsapiv1a1.ServiceImport{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "orders", UID: "import-one"}}
	require.NoError(t, r.client.Create(t.Context(), imported))
	slice.Labels[mcsapiv1a1.LabelServiceName] = "orders"
	require.NoError(t, r.client.Update(t.Context(), slice))
	tree := &resource.Resources{Services: []*corev1.Service{service}, ServiceImports: []*mcsapiv1a1.ServiceImport{imported}}
	require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{tree}))
	require.Len(t, tree.EndpointSlices, 1)
	require.Len(t, tree.BackendEndpoints, 2)
	require.Equal(t, "ServiceImport", tree.BackendEndpoints[1].Backend.Kind)
	require.Equal(t, "10.0.0.1", tree.BackendEndpoints[1].EndpointsByPort[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
}

// The endpoint controller only queues backends that full reconciliation has
// loaded into the revision tracker.
func TestEndpointReconcileQueuesOnlyTrackedBackends(t *testing.T) {
	r, service, slice := endpointReconciler(t)
	key := ir.ServiceKey{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name}
	reconcileKey := func() map[ir.ServiceKey]*endpoint.BackendEndpoints {
		t.Helper()
		_, err := r.reconcileBackendEndpoints(t.Context(), key)
		require.NoError(t, err)
		return r.resources.EndpointUpdates.Drain()
	}
	require.Empty(t, reconcileKey(), "a backend full reconciliation hasn't loaded is skipped")

	require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{{Services: []*corev1.Service{service}}}))
	updated := reconcileKey()[key]
	require.NotNil(t, updated, "a loaded backend takes the fast path")
	require.Equal(t, slice.Endpoints[0].Addresses[0], updated.EndpointsByPort[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)

	require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{{}}))
	require.Empty(t, reconcileKey(), "a backend pruned by full reconciliation is skipped")
}

func TestEndpointCacheClearedWithoutManagedGatewayClasses(t *testing.T) {
	r, service, _ := endpointReconciler(t)
	tree := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.loadBackendEndpoints(t.Context(), resource.ControllerResources{tree}))
	_, err := r.Reconcile(t.Context(), reconcile.Request{})
	require.NoError(t, err)
	key := ir.ServiceKey{Kind: resource.KindService, Namespace: service.Namespace, Name: service.Name}
	recaptured, err := r.trackBackendEndpoints(t.Context(), key)
	require.NoError(t, err)
	require.Greater(t, recaptured.Revision, tree.BackendEndpoints[0].Revision)
}
