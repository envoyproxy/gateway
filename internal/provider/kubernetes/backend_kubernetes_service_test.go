// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

func TestProcessBackendRefsKubernetesServiceBackend(t *testing.T) {
	const ns = "default"
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "svc-1", Namespace: ns},
	}
	backend := &egv1a1.Backend{
		ObjectMeta: metav1.ObjectMeta{Name: "backend-svc", Namespace: ns},
		Spec: egv1a1.BackendSpec{
			Type: new(egv1a1.BackendTypeKubernetesService),
			KubernetesService: &egv1a1.KubernetesServiceReference{
				Name: gwapiv1.ObjectName(service.Name),
				Port: 8080,
			},
		},
	}
	matchingEndpointSlice := test.GetEndpointSlice(types.NamespacedName{Namespace: ns, Name: "es-svc-1"}, service.Name, false)
	otherEndpointSlice := test.GetEndpointSlice(types.NamespacedName{Namespace: ns, Name: "es-other"}, "other", false)

	fakeClient := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects(service, backend, matchingEndpointSlice, otherEndpointSlice).
		WithIndex(&discoveryv1.EndpointSlice{}, serviceEndpointSliceIndex, serviceEndpointSliceIndexFunc).
		Build()

	r := &gatewayAPIReconciler{
		log:              logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		client:           fakeClient,
		backendCRDExists: true,
		envoyGateway: &egv1a1.EnvoyGateway{
			EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{
				ExtensionAPIs: &egv1a1.ExtensionAPISettings{EnableBackend: true},
			},
		},
	}

	resourceMappings := newResourceMapping()
	resourceMappings.insertBackendRef(gwapiv1.BackendObjectReference{
		Group:     new(gwapiv1.Group(egv1a1.GroupName)),
		Kind:      new(gwapiv1.Kind(egv1a1.KindBackend)),
		Name:      gwapiv1.ObjectName(backend.Name),
		Namespace: gatewayapi.NamespacePtr(ns),
	})

	gwcResource := resource.NewResources()
	require.NoError(t, r.processBackendRefs(t.Context(), gwcResource, resourceMappings))

	require.Len(t, gwcResource.Backends, 1)
	require.Len(t, gwcResource.Services, 1)
	require.Equal(t, service.Name, gwcResource.Services[0].Name)
	require.Len(t, gwcResource.EndpointSlices, 1)
	require.Equal(t, matchingEndpointSlice.Name, gwcResource.EndpointSlices[0].Name)
}

func TestServiceBackendIndexFunc(t *testing.T) {
	kubernetesServiceType := new(egv1a1.BackendTypeKubernetesService)
	endpointsType := new(egv1a1.BackendTypeEndpoints)

	tests := []struct {
		name     string
		backend  *egv1a1.Backend
		expected []string
	}{
		{
			name: "KubernetesService Backend is indexed by its Service",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "default"},
				Spec: egv1a1.BackendSpec{
					Type:              kubernetesServiceType,
					KubernetesService: &egv1a1.KubernetesServiceReference{Name: "svc-1", Port: 8080},
				},
			},
			expected: []string{"default/svc-1"},
		},
		{
			name: "Endpoints Backend is not indexed",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "default"},
				Spec:       egv1a1.BackendSpec{Type: endpointsType},
			},
		},
		{
			name: "KubernetesService Backend without kubernetesService is not indexed",
			backend: &egv1a1.Backend{
				ObjectMeta: metav1.ObjectMeta{Name: "b", Namespace: "default"},
				Spec:       egv1a1.BackendSpec{Type: kubernetesServiceType},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, serviceBackendIndexFunc(tc.backend))
		})
	}
}
