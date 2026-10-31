// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/endpoint"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/message"
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
	queue := workqueue.NewTypedRateLimitingQueue(workqueue.DefaultTypedControllerRateLimiter[reconcile.Request]())
	t.Cleanup(queue.ShutDown)
	r := &gatewayAPIReconciler{
		client:    fake.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(service, slice).Build(),
		resources: &message.ProviderResources{}, endpointQueue: queue, classController: egv1a1.GatewayControllerName,
		envoyGateway: &egv1a1.EnvoyGateway{RuntimeFlags: &egv1a1.RuntimeFlags{
			Enabled: []egv1a1.RuntimeFlag{egv1a1.EndpointFastPath}, Disabled: []egv1a1.RuntimeFlag{egv1a1.EndpointSliceIndex},
		}},
	}
	return r, service, slice
}

func TestEndpointCollectionSchedulesFullBeforeNotification(t *testing.T) {
	r, service, slice := endpointReconciler(t)
	tree := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{tree}))
	require.Len(t, tree.EndpointSnapshots, 1)
	initial := tree.EndpointSnapshots[0]
	require.Equal(t, "service-one", initial.Backend.UID)
	require.Equal(t, "10.0.0.1", initial.Ports[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
	// Keep a full reconcile in progress while the endpoint worker accepts new data.
	request, shutdown := r.endpointQueue.Get()
	require.False(t, shutdown)
	r.resources.Endpoints.Drain()
	<-r.resources.Endpoints.Updates()
	current := &discoveryv1.EndpointSlice{}
	require.NoError(t, r.client.Get(t.Context(), client.ObjectKeyFromObject(slice), current))
	current.Endpoints[0].Addresses[0] = "10.0.0.2"
	require.NoError(t, r.client.Update(t.Context(), current))
	updated, err := r.refreshEndpointSource(t.Context(), endpoint.BackendKey(initial.Backend))
	require.NoError(t, err)
	require.Greater(t, updated.Revision, initial.Revision)
	select {
	case <-r.resources.Endpoints.Updates():
	default:
		t.Fatal("new source data did not notify the publisher")
	}
	r.endpointQueue.Done(request)
	require.Equal(t, 1, r.endpointQueue.Len(), "an in-progress full reconcile must leave a subsequent request queued")
	require.Equal(t, "10.0.0.1", tree.EndpointSlices[0].Endpoints[0].Addresses[0], "full capture must remain immutable")
	require.Equal(t, initial.Revision, tree.EndpointSnapshots[0].Revision)
	next := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{next}))
	require.Equal(t, updated.Revision, next.EndpointSnapshots[0].Revision)
	require.Equal(t, "10.0.0.2", next.EndpointSlices[0].Endpoints[0].Addresses[0])
}

func TestEndpointCollectionErrorDoesNotPublishEmptyData(t *testing.T) {
	r, service, _ := endpointReconciler(t)
	tree := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{tree}))
	initial := tree.EndpointSnapshots[0]
	r.resources.Endpoints.Drain()
	r.client = interceptor.NewClient(r.client.(client.WithWatch), interceptor.Funcs{
		List: func(context.Context, client.WithWatch, client.ObjectList, ...client.ListOption) error {
			return errors.New("temporary cache read failure")
		},
	})
	_, err := r.refreshEndpointSource(t.Context(), endpoint.BackendKey(initial.Backend))
	require.Error(t, err, "controller-runtime must retry failed collection")
	require.Empty(t, r.resources.Endpoints.Drain())
}

func TestEndpointSourceRecreationUsesNewUID(t *testing.T) {
	r, service, _ := endpointReconciler(t)
	tree := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{tree}))
	initial := tree.EndpointSnapshots[0]
	require.NoError(t, r.client.Delete(t.Context(), service))
	service.UID, service.ResourceVersion = types.UID("service-two"), ""
	require.NoError(t, r.client.Create(t.Context(), service))
	next, err := r.refreshEndpointSource(t.Context(), endpoint.BackendKey(initial.Backend))
	require.NoError(t, err)
	require.NotEqual(t, initial.Backend.UID, next.Backend.UID)
	require.Greater(t, next.Revision, initial.Revision)
}

func TestEndpointCollectionWithDebounceSchedulesAnotherReconcile(t *testing.T) {
	r, service, slice := endpointReconciler(t)
	queue := newDebouncingQueue(r.endpointQueue, time.Second, time.Second)
	now := time.Now()
	queue.now = func() time.Time { return now }
	queue.afterFunc = func(time.Duration, func()) {}
	r.endpointQueue = queue
	defer queue.ShutDown()
	tree := &resource.Resources{Services: []*corev1.Service{service}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{tree}))
	now = now.Add(time.Second)
	queue.onTimer()
	request, shutdown := queue.Get()
	require.False(t, shutdown)
	slice.Endpoints = nil
	require.NoError(t, r.client.Update(t.Context(), slice))
	_, err := r.refreshEndpointSource(t.Context(), endpoint.BackendKey(tree.EndpointSnapshots[0].Backend))
	require.NoError(t, err)
	queue.Done(request)
	now = now.Add(time.Second)
	queue.onTimer()
	require.Equal(t, 1, queue.Len(), "debouncing must preserve the follow-up full reconcile")
}

func TestEndpointCollectionServiceImportAndSliceDeduplication(t *testing.T) {
	r, service, slice := endpointReconciler(t)
	imported := &mcsapiv1a1.ServiceImport{ObjectMeta: metav1.ObjectMeta{Namespace: "default", Name: "orders", UID: "import-one"}}
	require.NoError(t, r.client.Create(t.Context(), imported))
	slice.Labels[mcsapiv1a1.LabelServiceName] = "orders"
	require.NoError(t, r.client.Update(t.Context(), slice))
	tree := &resource.Resources{Services: []*corev1.Service{service}, ServiceImports: []*mcsapiv1a1.ServiceImport{imported}}
	require.NoError(t, r.captureEndpointSnapshots(t.Context(), resource.ControllerResources{tree}))
	require.Len(t, tree.EndpointSlices, 1)
	require.Len(t, tree.EndpointSnapshots, 2)
	require.Equal(t, "ServiceImport", tree.EndpointSnapshots[1].Backend.Kind)
	require.Equal(t, "import-one", tree.EndpointSnapshots[1].Backend.UID)
	require.Equal(t, "10.0.0.1", tree.EndpointSnapshots[1].Ports[endpoint.Port{Name: "http", Protocol: corev1.ProtocolTCP}][0].Host)
}
