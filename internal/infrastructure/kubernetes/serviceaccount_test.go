package kubernetes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExpectedServiceAccount(t *testing.T) {
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	sa := kube.expectedServiceAccount(infra)

	// Check the serviceaccount name is as expected.
	assert.Equal(t, sa.Name, expectedServiceAccountName(infra.Proxy.Name))
}

func TestCreateOrUpdateServiceAccount(t *testing.T) {
	testCases := []struct {
		name    string
		ns      string
		in      *ir.Infra
		current *corev1.ServiceAccount
		want    *corev1.ServiceAccount
	}{
		{
			name: "create-sa",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
				},
			},
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test",
				},
			},
		},
		{
			name: "sa-exists",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
				},
			},
			current: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test",
				},
			},
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Namespace: tc.ns,
			}
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createOrUpdateServiceAccount(context.Background(), tc.in)
			require.NoError(t, err)

			actual := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      expectedServiceAccountName(tc.in.Proxy.Name),
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			assert.Equal(t, true, cmp.Equal(tc.want, actual, opts))
		})
	}
}

func TestDeleteServiceAccount(t *testing.T) {
	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "delete service account",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			infra := ir.NewInfra()
			err := kube.deleteServiceAccount(context.Background(), infra)
			require.NoError(t, err)
		})
	}
}
