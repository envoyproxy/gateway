package kubernetes

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/intstr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
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

func TestDesiredService(t *testing.T) {
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	infra := ir.NewInfra()
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayLabel] = infra.Proxy.Name
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
	svc, err := kube.expectedService(infra)
	require.NoError(t, err)

	// Check the service name is as expected.
	assert.Equal(t, svc.Name, expectedDeploymentName(infra.Proxy.Name))

	checkServiceHasPort(t, svc, 80)
	checkServiceHasPort(t, svc, 443)
	checkServiceHasTargetPort(t, svc, 2080)
	checkServiceHasTargetPort(t, svc, 2443)

	// Ensure the Envoy service has the expected labels.
	lbls := envoyAppLabel()
	lbls[gatewayapi.OwningGatewayLabel] = infra.Proxy.Name
	checkServiceHasLabels(t, svc, lbls)

	for _, port := range infra.Proxy.Listeners[0].Ports {
		checkServiceHasPortName(t, svc, port.Name)
	}
}

func TestDeleteService(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "delete service",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				mu:        sync.Mutex{},
				Namespace: "test",
			}
			infra := ir.NewInfra()
			err := kube.deleteService(context.Background(), infra)
			require.NoError(t, err)
		})
	}
}
