// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

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
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateProxyServiceAccount(t *testing.T) {
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
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
				},
			},
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
			},
		},
		{
			name: "sa-exists",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
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
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
			},
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
			},
		},
		{
			name: "hashed-name",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
				},
			},
			current: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
			},
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-very-long-name-that-will-be-hashed-and-cut-off-b-5bacc75e",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.New()
			require.NoError(t, err)
			cfg.Namespace = tc.ns

			var cli client.Client
			if tc.current != nil {
				cli = fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithObjects(tc.current).
					WithInterceptorFuncs(interceptorFunc).
					Build()
			} else {
				cli = fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithInterceptorFuncs(interceptorFunc).
					Build()
			}

			kube := NewInfra(cli, cfg)

			r := proxy.NewResourceRender(kube.Namespace, tc.in.GetProxyInfra())
			err = kube.createOrUpdateServiceAccount(context.Background(), r)
			require.NoError(t, err)

			actual := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      proxy.ExpectedResourceHashedName(tc.in.Proxy.Name),
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			assert.True(t, cmp.Equal(tc.want, actual, opts))
		})
	}
}

func TestDeleteProxyServiceAccount(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "delete service account",
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kube := newTestInfra(t)

			infra := ir.NewInfra()
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			r := proxy.NewResourceRender(kube.Namespace, infra.GetProxyInfra())

			err := kube.createOrUpdateServiceAccount(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteServiceAccount(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
