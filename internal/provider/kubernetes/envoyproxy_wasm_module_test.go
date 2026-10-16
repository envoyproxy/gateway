// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/logging"
)

// envoyProxyWithRemoteWasmModules registers one module per ref kind, a pull
// secret, a cross-namespace pull secret, and a Local module without refs.
func envoyProxyWithRemoteWasmModules() *egv1a1.EnvoyProxy {
	return &egv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "proxy"},
		Spec: egv1a1.EnvoyProxySpec{
			WasmModules: []egv1a1.WasmModuleEntry{
				{
					Name: "http-secret-ca",
					Source: egv1a1.WasmModuleSource{
						Type: new(egv1a1.HTTPWasmModuleSourceType),
						HTTP: &egv1a1.HTTPWasmCodeSource{
							URL: "https://example.com/a.wasm",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{Name: "ca-secret"},
							},
						},
					},
				},
				{
					Name: "image-configmap-ca",
					Source: egv1a1.WasmModuleSource{
						Type: new(egv1a1.ImageWasmModuleSourceType),
						Image: &egv1a1.ImageWasmCodeSource{
							URL:           "example.com/b:v1",
							PullSecretRef: &gwapiv1.SecretObjectReference{Name: "pull-secret"},
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: new(gwapiv1.Kind(resource.KindConfigMap)),
									Name: "ca-configmap",
								},
							},
						},
					},
				},
				{
					Name: "http-ctb-ca",
					Source: egv1a1.WasmModuleSource{
						Type: new(egv1a1.HTTPWasmModuleSourceType),
						HTTP: &egv1a1.HTTPWasmCodeSource{
							URL: "https://example.com/c.wasm",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: new(gwapiv1.Kind(resource.KindClusterTrustBundle)),
									Name: "ca-bundle",
								},
							},
						},
					},
				},
				{
					Name: "image-cross-namespace-secret",
					Source: egv1a1.WasmModuleSource{
						Type: new(egv1a1.ImageWasmModuleSourceType),
						Image: &egv1a1.ImageWasmCodeSource{
							URL: "example.com/d:v1",
							PullSecretRef: &gwapiv1.SecretObjectReference{
								Namespace: new(gwapiv1.Namespace("other-ns")),
								Name:      "other-secret",
							},
						},
					},
				},
				{
					Name: "local",
					Source: egv1a1.WasmModuleSource{
						Local: &egv1a1.LocalWasmModuleSource{Path: "/var/lib/envoy/e.wasm"},
					},
				},
			},
		},
	}
}

func TestEnvoyProxyWasmModuleIndexerFunctions(t *testing.T) {
	ep := envoyProxyWithRemoteWasmModules()
	require.ElementsMatch(t,
		[]string{"proxy-ns/ca-secret", "proxy-ns/pull-secret", "other-ns/other-secret"},
		secretEnvoyProxyIndexFunc(ep))
	require.Equal(t, []string{"proxy-ns/ca-configmap"}, configMapEnvoyProxyIndexFunc(ep))
	require.Equal(t, []string{"ca-bundle"}, clusterTrustBundleEnvoyProxyIndexFunc(ep))

	empty := &egv1a1.EnvoyProxy{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "empty"}}
	require.Empty(t, secretEnvoyProxyIndexFunc(empty))
	require.Empty(t, configMapEnvoyProxyIndexFunc(empty))
	require.Empty(t, clusterTrustBundleEnvoyProxyIndexFunc(empty))
}

func TestEnvoyProxyWasmModuleRefsTriggerReconcile(t *testing.T) {
	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		epCRDExists:     true,
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		client: fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(envoyProxyWithRemoteWasmModules()).
			WithIndex(&egv1a1.EnvoyProxy{}, secretEnvoyProxyIndex, secretEnvoyProxyIndexFunc).
			WithIndex(&egv1a1.EnvoyProxy{}, configMapEnvoyProxyIndex, configMapEnvoyProxyIndexFunc).
			WithIndex(&egv1a1.EnvoyProxy{}, clusterTrustBundleEnvoyProxyIndex, clusterTrustBundleEnvoyProxyIndexFunc).
			Build(),
	}

	secret := func(ns, name string) *corev1.Secret {
		return &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}
	}
	configMap := func(ns, name string) *corev1.ConfigMap {
		return &corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: ns, Name: name}}
	}
	ctb := func(name string) *certificatesv1b1.ClusterTrustBundle {
		return &certificatesv1b1.ClusterTrustBundle{ObjectMeta: metav1.ObjectMeta{Name: name}}
	}

	require.True(t, r.validateSecretForReconcile(secret("proxy-ns", "pull-secret")))
	require.True(t, r.validateSecretForReconcile(secret("proxy-ns", "ca-secret")))
	require.False(t, r.validateSecretForReconcile(secret("proxy-ns", "unrelated")))
	require.True(t, r.validateConfigMapForReconcile(configMap("proxy-ns", "ca-configmap")))
	require.False(t, r.validateConfigMapForReconcile(configMap("proxy-ns", "unrelated")))
	require.True(t, r.validateClusterTrustBundleForReconcile(ctb("ca-bundle")))
	require.False(t, r.validateClusterTrustBundleForReconcile(ctb("unrelated")))
}

func TestProcessWasmModuleRefs(t *testing.T) {
	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		client: fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "ca-secret"}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "pull-secret"}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "other-ns", Name: "other-secret"}},
				&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "ca-configmap"}},
				&certificatesv1b1.ClusterTrustBundle{ObjectMeta: metav1.ObjectMeta{Name: "ca-bundle"}},
			).
			Build(),
	}

	resourceTree := resource.NewResources()
	ep := envoyProxyWithRemoteWasmModules()
	err := r.processWasmModuleRefs(context.Background(), &ep.Spec, egv1a1.KindEnvoyProxy, ep.Namespace, ep.Name, newResourceMapping(), resourceTree)
	require.NoError(t, err)

	secrets := make([]string, 0, len(resourceTree.Secrets))
	for _, s := range resourceTree.Secrets {
		secrets = append(secrets, s.Namespace+"/"+s.Name)
	}
	// The cross-namespace pull secret is not fetched.
	require.ElementsMatch(t, []string{"proxy-ns/ca-secret", "proxy-ns/pull-secret"}, secrets)
	require.Len(t, resourceTree.ConfigMaps, 1)
	require.Equal(t, "ca-configmap", resourceTree.ConfigMaps[0].Name)
	require.Len(t, resourceTree.ClusterTrustBundles, 1)
	require.Equal(t, "ca-bundle", resourceTree.ClusterTrustBundles[0].Name)
}

// Refs in the EnvoyGateway default EnvoyProxy spec resolve in the controller namespace.
func TestDefaultEnvoyProxySpecWasmModuleRefs(t *testing.T) {
	const controllerNS = "envoy-gateway-system"
	defaultSpec := &envoyProxyWithRemoteWasmModules().Spec
	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		namespace:       controllerNS,
		envoyGateway:    &egv1a1.EnvoyGateway{EnvoyGatewaySpec: egv1a1.EnvoyGatewaySpec{EnvoyProxy: defaultSpec}},
		log:             logging.DefaultLogger(os.Stdout, egv1a1.LogLevelInfo),
		client: fakeclient.NewClientBuilder().
			WithScheme(envoygateway.GetScheme()).
			WithObjects(
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: controllerNS, Name: "pull-secret"}},
				&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "pull-secret"}},
			).
			Build(),
	}

	require.True(t, r.validateSecretForReconcile(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: controllerNS, Name: "pull-secret"}}))
	require.False(t, r.validateSecretForReconcile(&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "proxy-ns", Name: "pull-secret"}}))
	require.True(t, r.validateConfigMapForReconcile(&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Namespace: controllerNS, Name: "ca-configmap"}}))
	require.True(t, r.validateClusterTrustBundleForReconcile(&certificatesv1b1.ClusterTrustBundle{ObjectMeta: metav1.ObjectMeta{Name: "ca-bundle"}}))

	resourceTree := resource.NewResources()
	require.NoError(t, r.processWasmModuleRefs(context.Background(), defaultSpec, egv1a1.KindEnvoyGateway, controllerNS, "", newResourceMapping(), resourceTree))
	require.Len(t, resourceTree.Secrets, 1)
	require.Equal(t, controllerNS, resourceTree.Secrets[0].Namespace)
}
