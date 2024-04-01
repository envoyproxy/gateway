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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateOrUpdateProxyConfigMap(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	infra := ir.NewInfra()
	infra.Proxy.Name = "test"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	testCases := []struct {
		name    string
		current *corev1.ConfigMap
		expect  *corev1.ConfigMap
	}{
		{
			name: "create configmap",
			expect: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "test",
					},
				},
				Data: map[string]string{
					proxy.SdsCAFilename:   proxy.SdsCAConfigMapData,
					proxy.SdsCertFilename: proxy.SdsCertConfigMapData,
				},
			},
		},
		{
			name: "update configmap",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
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
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "test",
					},
				},
				Data: map[string]string{
					proxy.SdsCAFilename:   proxy.SdsCAConfigMapData,
					proxy.SdsCertFilename: proxy.SdsCertConfigMapData,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
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
			r := proxy.NewResourceRender(kube.Namespace, infra.GetProxyInfra())
			err := kube.createOrUpdateConfigMap(context.Background(), r)
			require.NoError(t, err)
			actual := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: tc.expect.Namespace,
					Name:      tc.expect.Name,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.expect.Data, actual.Data)
			assert.True(t, apiequality.Semantic.DeepEqual(tc.expect.Labels, actual.Labels))
		})
	}
}

func TestDeleteConfigProxyMap(t *testing.T) {
	cfg, err := config.New()
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
					Namespace: cfg.Namespace,
					Name:      "envoy-test",
				},
			},
			expect: true,
		},
		{
			name: "configmap not found",
			current: &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: cfg.Namespace,
					Name:      "foo",
				},
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			kube := NewInfra(cli, cfg)

			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

			r := proxy.NewResourceRender(kube.Namespace, infra.GetProxyInfra())
			cm := &corev1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      r.Name(),
				},
			}
			err = kube.Client.Delete(context.Background(), cm)
			require.NoError(t, err)
		})
	}
}
