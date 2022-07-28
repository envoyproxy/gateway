package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/log"
)

func TestCreateIfNeeded(t *testing.T) {
	logger, err := log.NewLogger()
	require.NoError(t, err)

	kube := Infra{Log: logger}

	testCases := []struct {
		name   string
		in     *ir.Infra
		out    *Resources
		expect bool
	}{
		{
			name: "happy-path",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      "test",
					Namespace: "test",
				},
			},
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
				},
			},
			expect: true,
		},
		{
			name: "nil-infra",
			in:   nil,
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
				},
			},
			expect: false,
		},
		{
			name: "nil-infra-proxy",
			in:   &ir.Infra{},
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
				},
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.out.ServiceAccount).Build()
			err := kube.CreateInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.out, kube.Resources)
			}
		})
	}
}

func TestAddResource(t *testing.T) {
	logger, err := log.NewLogger()
	require.NoError(t, err)

	kube := Infra{Log: logger}

	testCases := []struct {
		name string
		kind Kind
		obj  client.Object
		out  *Resources
	}{
		{
			name: "happy-path-sa",
			kind: KindServiceAccount,
			obj: &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "test",
				},
			},
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "test",
						Name:      "test",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			err := kube.addResource(tc.kind, tc.obj)
			require.NoError(t, err)
			require.Equal(t, tc.out, kube.Resources)
		})
	}
}
