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
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/intstr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

func checkServiceHasPort(t *testing.T, svc *corev1.Service, port int32) {
	t.Helper()

	for _, p := range svc.Spec.Ports {
		if p.Port == port {
			return
		}
	}
	t.Errorf("service is missing port %q", port)
}

func checkServiceHasTargetPort(t *testing.T, svc *corev1.Service, port int32) {
	t.Helper()

	intStrPort := intstr.IntOrString{IntVal: port}
	for _, p := range svc.Spec.Ports {
		if p.TargetPort == intStrPort {
			return
		}
	}
	t.Errorf("service is missing targetPort %d", port)
}

func checkServiceHasPortName(t *testing.T, svc *corev1.Service, name string) {
	t.Helper()

	for _, p := range svc.Spec.Ports {
		if p.Name == name {
			return
		}
	}
	t.Errorf("service is missing port name %q", name)
}

func checkServiceHasLabels(t *testing.T, svc *corev1.Service, expected map[string]string) {
	t.Helper()

	if apiequality.Semantic.DeepEqual(svc.Labels, expected) {
		return
	}

	t.Errorf("service has unexpected %q labels", svc.Labels)
}

func TestDesiredProxyService(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, cfg)
	infra := ir.NewInfra()
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
	assert.Equal(t, svc.Name, expectedProxyDeploymentName(infra.Proxy.Name))

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
