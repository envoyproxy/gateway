package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExpectedConfigMap(t *testing.T) {
	// Setup the infra.
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	infra := ir.NewInfra()
	infra.Proxy.Name = "test"
	cm := kube.expectedConfigMap(infra)

	require.Equal(t, "envoy-test", cm.Name)
	require.Equal(t, "envoy-gateway-system", cm.Namespace)
	require.Contains(t, cm.Data, sdsCAFilename)
	assert.Equal(t, sdsCAConfigMapData, cm.Data[sdsCAFilename])
	require.Contains(t, cm.Data, sdsCertFilename)
	assert.Equal(t, sdsCertConfigMapData, cm.Data[sdsCertFilename])
}

func TestCreateOrUpdateConfigMap(t *testing.T) {
	kube := NewInfra(nil)
	infra := ir.NewInfra()
	infra.Proxy.Name = "test"

	testCases := []struct {
		name    string
		current *corev1.ConfigMap
		expect  *corev1.ConfigMap
	}{
		{
			name: "create configmap",
			expect: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.EnvoyGatewayNamespace,
					Name:      "envoy-test",
				},
				Data: map[string]string{sdsCAFilename: sdsCAConfigMapData, sdsCertFilename: sdsCertConfigMapData},
			},
		},
		{
			name: "update configmap",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.EnvoyGatewayNamespace,
					Name:      "envoy-test",
				},
				Data: map[string]string{"foo": "bar"},
			},
			expect: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.EnvoyGatewayNamespace,
					Name:      "envoy-test",
				},
				Data: map[string]string{sdsCAFilename: sdsCAConfigMapData, sdsCertFilename: sdsCertConfigMapData},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			cm, err := kube.createOrUpdateConfigMap(context.Background(), infra)
			require.NoError(t, err)
			require.Equal(t, tc.expect.Namespace, cm.Namespace)
			require.Equal(t, tc.expect.Name, cm.Name)
			require.Equal(t, tc.expect.Data, cm.Data)
		})
	}
}

func TestDeleteConfigMap(t *testing.T) {
	infra := ir.NewInfra()
	infra.Proxy.Name = "test"

	testCases := []struct {
		name    string
		current *corev1.ConfigMap
		expect  bool
	}{
		{
			name: "delete configmap",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.EnvoyGatewayNamespace,
					Name:      "envoy-test",
				},
			},
			expect: true,
		},
		{
			name: "configmap not found",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: config.EnvoyGatewayNamespace,
					Name:      "foo",
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := NewInfra(fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build())
			err := kube.deleteConfigMap(context.Background(), infra)
			require.NoError(t, err)
		})
	}
}
