// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateProxyServiceAccount(t *testing.T) {
	proxyInfra := &ir.ProxyInfra{
		Name: "test",
		Metadata: &ir.InfraMetadata{
			Labels: map[string]string{
				gatewayapi.OwningGatewayNamespaceLabel: "default",
				gatewayapi.OwningGatewayNameLabel:      "gateway-1",
			},
			OwnerReference: &ir.ResourceMetadata{
				Kind: resource.KindGatewayClass,
				Name: testGatewayClass,
			},
		},
	}
	testCases := []struct {
		name                 string
		ns                   string
		in                   *ir.Infra
		gatewayNamespaceMode bool
		current              *corev1.ServiceAccount
		want                 *corev1.ServiceAccount
	}{
		{
			name: "create-sa",
			ns:   "test",
			in: &ir.Infra{
				Proxy: proxyInfra,
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
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "gateway.networking.k8s.io/v1",
							Kind:       "GatewayClass",
							Name:       "envoy-gateway-class",
							UID:        "foo.bar",
						},
					},
				},
			},
		},
		{
			name: "sa-exists",
			ns:   "test",
			in: &ir.Infra{
				Proxy: proxyInfra,
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
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "gateway.networking.k8s.io/v1",
							Kind:       "GatewayClass",
							Name:       "envoy-gateway-class",
							UID:        "foo.bar",
						},
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
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
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
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "gateway.networking.k8s.io/v1",
							Kind:       "GatewayClass",
							Name:       "envoy-gateway-class",
							UID:        "foo.bar",
						},
					},
				},
			},
		},
		{
			name: "create-sa-with-gateway-namespace-mode",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name:      "gateway-1",
					Namespace: "ns1",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "ns1",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: "Gateway",
							Name: "gateway-1",
						},
					},
				},
			},
			gatewayNamespaceMode: true,
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "ns1",
					Name:      "gateway-1",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "ns1",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						gatewayapi.GatewayNameLabel:            "gateway-1",
					},
					OwnerReferences: []metav1.OwnerReference{
						{
							APIVersion: "gateway.networking.k8s.io/v1",
							Kind:       "Gateway",
							Name:       "gateway-1",
							UID:        "foo.bar",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cfg, err := config.New(os.Stdout)
			require.NoError(t, err)
			cfg.ControllerNamespace = tc.ns

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
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))
			if tc.gatewayNamespaceMode {
				kube.EnvoyGateway.Provider.Kubernetes.Deploy = &egv1a1.KubernetesDeployMode{
					Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
				}
			}

			r, err := proxy.NewResourceRender(ctx, kube, tc.in)
			require.NoError(t, err)
			err = kube.createOrUpdateServiceAccount(ctx, r)
			require.NoError(t, err)

			actual := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.GetResourceNamespace(tc.in),
					Name:      expectedName(tc.in.Proxy, tc.gatewayNamespaceMode),
				},
			}
			require.NoError(t, kube.Client.Get(ctx, client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			require.Empty(t, cmp.Diff(tc.want, actual, opts))
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
			ctx := context.Background()
			kube := newTestInfra(t)
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

			infra := ir.NewInfra()
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
				Kind: resource.KindGatewayClass,
				Name: testGatewayClass,
			}
			r, err := proxy.NewResourceRender(ctx, kube, infra)
			require.NoError(t, err)

			err = kube.createOrUpdateServiceAccount(ctx, r)
			require.NoError(t, err)

			err = kube.deleteServiceAccount(ctx, r)
			require.NoError(t, err)
		})
	}
}
