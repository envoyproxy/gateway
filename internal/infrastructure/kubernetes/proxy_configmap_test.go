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
	"github.com/stretchr/testify/assert"
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
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateProxyConfigMap(t *testing.T) {
	testCases := []struct {
		name                 string
		ns                   string
		in                   *ir.Infra
		gatewayNamespaceMode bool
		current              *corev1.ConfigMap
		expect               *corev1.ConfigMap
	}{
		{
			name: "create configmap",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "test",
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
						},
					},
				},
			},
			expect: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
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
						gatewayapi.OwningGatewayNameLabel:      "test",
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
				Data: map[string]string{
					common.SdsCAFilename:   common.GetSdsCAConfigMapData(proxy.XdsTLSCaFilepath),
					common.SdsCertFilename: common.GetSdsCertConfigMapData(proxy.XdsTLSCertFilepath, proxy.XdsTLSKeyFilepath),
				},
			},
		},
		{
			name: "update configmap",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "test",
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
						},
					},
				},
			},
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "test",
					Name:      "envoy-test",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "test",
					},
				},
				Data: map[string]string{"foo": "bar"},
			},
			expect: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
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
						gatewayapi.OwningGatewayNameLabel:      "test",
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
				Data: map[string]string{
					common.SdsCAFilename:   common.GetSdsCAConfigMapData(proxy.XdsTLSCaFilepath),
					common.SdsCertFilename: common.GetSdsCertConfigMapData(proxy.XdsTLSCertFilepath, proxy.XdsTLSKeyFilepath),
				},
			},
		},
		{
			name: "create configmap with gateway namespace mode",
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
			expect: &corev1.ConfigMap{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ConfigMap",
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
				Data: map[string]string{
					common.SdsCAFilename:   common.GetSdsCAConfigMapData(proxy.XdsTLSCaFilepath),
					common.SdsCertFilename: common.GetSdsCertConfigMapData(proxy.XdsTLSCertFilepath, proxy.XdsTLSKeyFilepath),
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
			err = kube.createOrUpdateConfigMap(ctx, r)
			require.NoError(t, err)
			actual := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tc.expect.Namespace,
					Name:      tc.expect.Name,
				},
			}
			require.NoError(t, kube.Client.Get(ctx, client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			assert.True(t, cmp.Equal(tc.expect, actual, opts))
		})
	}
}

func TestDeleteConfigProxyMap(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

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
					Namespace: cfg.ControllerNamespace,
					Name:      "envoy-test",
				},
			},
			expect: true,
		},
		{
			name: "configmap not found",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.ControllerNamespace,
					Name:      "foo",
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			kube := NewInfra(cli, cfg)
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
				Kind: resource.KindGatewayClass,
				Name: testGatewayClass,
			}

			r, err := proxy.NewResourceRender(ctx, kube, infra)
			require.NoError(t, err)
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.ControllerNamespace,
					Name:      r.Name(),
				},
			}
			err = kube.Client.Delete(ctx, cm)
			require.NoError(t, err)
		})
	}
}
