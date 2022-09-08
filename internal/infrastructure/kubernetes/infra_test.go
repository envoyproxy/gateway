package kubernetes

import (
	"context"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateInfra(t *testing.T) {
	testCases := []struct {
		name   string
		in     *ir.Infra
		out    *Resources
		expect bool
	}{
		{
			name: "default infra",
			in:   ir.NewInfra(),
			out: &Resources{
				ServiceAccount: &corev1.ServiceAccount{
					TypeMeta: metav1.TypeMeta{
						Kind:       "ServiceAccount",
						APIVersion: "v1",
					},
					ObjectMeta: metav1.ObjectMeta{
						Namespace:       "default",
						Name:            "envoy-default",
						ResourceVersion: "1",
					},
				},
			},
			expect: true,
		},
		{
			name:   "nil-infra",
			in:     nil,
			out:    &Resources{},
			expect: false,
		},
		{
			name: "nil-infra-proxy",
			in: &ir.Infra{
				Proxy: nil,
			},
			out:    &Resources{},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				mu:     sync.Mutex{},
				Client: fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
			}
			err := kube.CreateInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, *tc.out.ServiceAccount, *kube.Resources.ServiceAccount)
			}
		})
	}
}

func TestDeleteInfra(t *testing.T) {
	testCases := []struct {
		name   string
		in     *ir.Infra
		expect bool
	}{
		{
			name:   "nil infra",
			in:     nil,
			expect: false,
		},
		{
			name:   "default infra",
			in:     ir.NewInfra(),
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				mu:     sync.Mutex{},
				Client: fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
			}
			err := kube.DeleteInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAddResource(t *testing.T) {
	testCases := []struct {
		name string
		obj  client.Object
		out  *Resources
	}{
		{
			name: "happy-path-sa",
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				mu:     sync.Mutex{},
				Client: fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
			}
			err := kube.addResource(tc.obj)
			require.NoError(t, err)
			require.Equal(t, tc.out, kube.Resources)
		})
	}
}
