// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

func testDesiredProxyService(t *testing.T, infra *ir.Infra, expected corev1.ServiceSpec) {
	cfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, cfg)
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
	infra.Proxy.Listeners[0].Ports = []ir.ListenerPort{
		{
			Name:          "gateway-system-gateway-1",
			Protocol:      ir.HTTPProtocolType,
			ServicePort:   80,
			ContainerPort: 2080,
		},
		{
			Name:          "gateway-system-gateway-1",
			Protocol:      ir.HTTPSProtocolType,
			ServicePort:   443,
			ContainerPort: 2443,
		},
	}
	svc, err := kube.expectedProxyService(infra)
	require.NoError(t, err)

	// Check the service name is as expected.
	assert.Equal(t, svc.Name, expectedResourceHashedName(infra.Proxy.Name))

	checkServiceHasPort(t, svc, 80)
	checkServiceHasPort(t, svc, 443)
	checkServiceHasTargetPort(t, svc, 2080)
	checkServiceHasTargetPort(t, svc, 2443)

	// Ensure the Envoy service has the expected labels.
	lbls := envoyAppLabel()
	lbls[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	lbls[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
	checkServiceHasLabels(t, svc, lbls)

	for _, port := range infra.Proxy.Listeners[0].Ports {
		checkServiceHasPortName(t, svc, port.Name)
	}

	// Make sure no service annotations are set by default
	checkServiceHasAnnotations(t, svc, nil)

	// Make sure service type are set by default with ServiceTypeLoadBalancer
	checkServiceSpec(t, svc, expected)
}

func TestDesiredProxyService(t *testing.T) {
	testDesiredProxyService(t, ir.NewInfra(), expectedServiceSpec(egcfgv1a1.DefaultKubernetesServiceType()))
}

func TestDesiredProxySpecifiedServiceSpec(t *testing.T) {
	infra := ir.NewInfra()
	clusterIPServiceType := egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeClusterIP)
	infra.Proxy.Config = &egcfgv1a1.EnvoyProxy{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: egcfgv1a1.EnvoyProxySpec{Provider: &egcfgv1a1.EnvoyProxyProvider{
			Type: egcfgv1a1.ProviderTypeKubernetes,
			Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
				EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
					Type: clusterIPServiceType,
				},
			},
		}},
		Status: egcfgv1a1.EnvoyProxyStatus{},
	}
	testDesiredProxyService(t, infra, expectedServiceSpec(clusterIPServiceType))
}

func TestExpectedAnnotations(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	// Set service annotations into EnvoyProxy API and ensure the same
	// value is set in the generated service.
	annotations := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	infra.Proxy.Config = &egcfgv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: egcfgv1a1.EnvoyProxySpec{
			Provider: &egcfgv1a1.EnvoyProxyProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
				Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
					EnvoyService: &egcfgv1a1.KubernetesServiceSpec{
						Annotations: annotations,
					},
				},
			},
		},
	}

	svc, err := kube.expectedProxyService(infra)
	require.NoError(t, err)
	checkServiceHasAnnotations(t, svc, annotations)
}

func TestDeleteProxyService(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "delete service",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			infra := ir.NewInfra()

			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

			err := kube.createOrUpdateProxyService(context.Background(), infra)
			require.NoError(t, err)

			err = kube.deleteProxyService(context.Background(), infra)
			require.NoError(t, err)
		})
	}
}
