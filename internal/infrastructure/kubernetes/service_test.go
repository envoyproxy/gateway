package kubernetes

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
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

func TestDesiredService(t *testing.T) {
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	infra := ir.NewInfra()
	infra.Proxy.Listeners[0].Ports = []ir.ListenerPort{
		{
			Name:     "gateway-system-gateway-1",
			Protocol: ir.HTTPProtocolType,
			Port:     80,
		},
		{
			Name:     "gateway-system-gateway-1",
			Protocol: ir.HTTPSProtocolType,
			Port:     443,
		},
	}
	svc := kube.expectedService(infra)

	checkServiceHasPort(t, svc, envoyServiceHTTPPort)
	checkServiceHasPort(t, svc, envoyServiceHTTPSPort)
	checkServiceHasTargetPort(t, svc, envoyHTTPPort)
	checkServiceHasTargetPort(t, svc, envoyHTTPSPort)

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
			err := kube.deleteService(context.Background())
			require.NoError(t, err)
		})
	}
}
