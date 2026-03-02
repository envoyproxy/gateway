// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/require"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes/test"
)

// TestWasmTLSIndexers tests that all three WASM TLS certificate sources
// (Secrets, ConfigMaps, and ClusterTrustBundles) are properly indexed and
// can trigger reconciliation when changed.
func TestWasmTLSIndexers(t *testing.T) {
	gc := test.GetGatewayClass("test-gc", egv1a1.GatewayControllerName, nil)
	gtw := test.GetGateway(types.NamespacedName{Namespace: "default", Name: "test-gateway"}, "test-gc", 8080)

	// EnvoyExtensionPolicy with HTTP WASM using Secret
	eepHTTPSecret := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-http-secret",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("http-wasm-filter"),
					RootID: ptr.To("http_root_id"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.HTTPWasmCodeSourceType,
						HTTP: &egv1a1.HTTPWasmCodeSource{
							URL: "https://example.com/wasm-http.wasm",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Name: "http-wasm-ca-secret",
								},
							},
						},
					},
				},
			},
		},
	}

	// EnvoyExtensionPolicy with HTTP WASM using ConfigMap
	eepHTTPConfigMap := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-http-configmap",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("http-wasm-filter-cm"),
					RootID: ptr.To("http_root_id_cm"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.HTTPWasmCodeSourceType,
						HTTP: &egv1a1.HTTPWasmCodeSource{
							URL: "https://example.com/wasm-http-cm.wasm",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: ptr.To[gwapiv1.Kind]("ConfigMap"),
									Name: "http-wasm-ca-configmap",
								},
							},
						},
					},
				},
			},
		},
	}

	// EnvoyExtensionPolicy with HTTP WASM using ClusterTrustBundle
	eepHTTPClusterTrustBundle := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-http-ctb",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("http-wasm-filter-ctb"),
					RootID: ptr.To("http_root_id_ctb"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.HTTPWasmCodeSourceType,
						HTTP: &egv1a1.HTTPWasmCodeSource{
							URL: "https://example.com/wasm-http-ctb.wasm",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: ptr.To[gwapiv1.Kind](resource.KindClusterTrustBundle),
									Name: "http-wasm-ca-ctb",
								},
							},
						},
					},
				},
			},
		},
	}

	// EnvoyExtensionPolicy with Image WASM using Secret
	eepImageSecret := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-image-secret",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("image-wasm-filter"),
					RootID: ptr.To("image_root_id"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.ImageWasmCodeSourceType,
						Image: &egv1a1.ImageWasmCodeSource{
							URL: "oci://example.com/wasm-image:v1.0.0",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Name: "image-wasm-ca-secret",
								},
							},
						},
					},
				},
			},
		},
	}

	// EnvoyExtensionPolicy with Image WASM using ConfigMap
	eepImageConfigMap := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-image-configmap",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("image-wasm-filter-cm"),
					RootID: ptr.To("image_root_id_cm"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.ImageWasmCodeSourceType,
						Image: &egv1a1.ImageWasmCodeSource{
							URL: "oci://example.com/wasm-image-cm:v1.0.0",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: ptr.To[gwapiv1.Kind]("ConfigMap"),
									Name: "image-wasm-ca-configmap",
								},
							},
						},
					},
				},
			},
		},
	}

	// EnvoyExtensionPolicy with Image WASM using ClusterTrustBundle
	eepImageClusterTrustBundle := &egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "wasm-image-ctb",
			Namespace: "default",
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					{
						LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
							Kind: "Gateway",
							Name: "test-gateway",
						},
					},
				},
			},
			Wasm: []egv1a1.Wasm{
				{
					Name:   ptr.To("image-wasm-filter-ctb"),
					RootID: ptr.To("image_root_id_ctb"),
					Code: egv1a1.WasmCodeSource{
						Type: egv1a1.ImageWasmCodeSourceType,
						Image: &egv1a1.ImageWasmCodeSource{
							URL: "oci://example.com/wasm-image-ctb:v1.0.0",
							TLS: &egv1a1.WasmCodeSourceTLSConfig{
								CACertificateRef: gwapiv1.SecretObjectReference{
									Kind: ptr.To[gwapiv1.Kind](resource.KindClusterTrustBundle),
									Name: "image-wasm-ca-ctb",
								},
							},
						},
					},
				},
			},
		},
	}

	testCases := []struct {
		name     string
		configs  []egv1a1.EnvoyExtensionPolicy
		resource interface{}
		expect   bool
	}{
		{
			name:     "HTTP WASM references Secret CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepHTTPSecret},
			resource: test.GetSecret(types.NamespacedName{Namespace: "default", Name: "http-wasm-ca-secret"}),
			expect:   true,
		},
		{
			name:     "HTTP WASM references ConfigMap CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepHTTPConfigMap},
			resource: test.GetConfigMap(types.NamespacedName{Namespace: "default", Name: "http-wasm-ca-configmap"}, nil, nil),
			expect:   true,
		},
		{
			name:     "HTTP WASM references ClusterTrustBundle CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepHTTPClusterTrustBundle},
			resource: test.GetClusterTrustBundle("http-wasm-ca-ctb"),
			expect:   true,
		},
		{
			name:     "Image WASM references Secret CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepImageSecret},
			resource: test.GetSecret(types.NamespacedName{Namespace: "default", Name: "image-wasm-ca-secret"}),
			expect:   true,
		},
		{
			name:     "Image WASM references ConfigMap CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepImageConfigMap},
			resource: test.GetConfigMap(types.NamespacedName{Namespace: "default", Name: "image-wasm-ca-configmap"}, nil, nil),
			expect:   true,
		},
		{
			name:     "Image WASM references ClusterTrustBundle CA",
			configs:  []egv1a1.EnvoyExtensionPolicy{*eepImageClusterTrustBundle},
			resource: test.GetClusterTrustBundle("image-wasm-ca-ctb"),
			expect:   true,
		},
		{
			name: "All three sources in different policies",
			configs: []egv1a1.EnvoyExtensionPolicy{
				*eepHTTPSecret,
				*eepHTTPConfigMap,
				*eepHTTPClusterTrustBundle,
			},
			resource: test.GetSecret(types.NamespacedName{Namespace: "default", Name: "http-wasm-ca-secret"}),
			expect:   true,
		},
		{
			name: "All three sources but checking unrelated Secret",
			configs: []egv1a1.EnvoyExtensionPolicy{
				*eepHTTPSecret,
				*eepHTTPConfigMap,
				*eepHTTPClusterTrustBundle,
			},
			resource: test.GetSecret(types.NamespacedName{Namespace: "default", Name: "unrelated-secret"}),
			expect:   false,
		},
		{
			name: "All three sources but checking unrelated ConfigMap",
			configs: []egv1a1.EnvoyExtensionPolicy{
				*eepHTTPSecret,
				*eepHTTPConfigMap,
				*eepHTTPClusterTrustBundle,
			},
			resource: test.GetConfigMap(types.NamespacedName{Namespace: "default", Name: "unrelated-configmap"}, nil, nil),
			expect:   false,
		},
		{
			name: "All three sources but checking unrelated ClusterTrustBundle",
			configs: []egv1a1.EnvoyExtensionPolicy{
				*eepHTTPSecret,
				*eepHTTPConfigMap,
				*eepHTTPClusterTrustBundle,
			},
			resource: test.GetClusterTrustBundle("unrelated-ctb"),
			expect:   false,
		},
	}

	r := gatewayAPIReconciler{
		classController: egv1a1.GatewayControllerName,
		eepCRDExists:    true,
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			objs := []client.Object{gc, gtw}
			for _, eep := range tc.configs {
				objs = append(objs, &eep)
			}

			clientBuilder := fakeclient.NewClientBuilder().
				WithScheme(envoygateway.GetScheme()).
				WithIndex(&egv1a1.EnvoyExtensionPolicy{}, secretEnvoyExtensionPolicyIndex, secretEnvoyExtensionPolicyIndexFunc).
				WithIndex(&egv1a1.EnvoyExtensionPolicy{}, configMapEepIndex, configMapEepIndexFunc).
				WithIndex(&egv1a1.EnvoyExtensionPolicy{}, clusterTrustBundleEepIndex, clusterTrustBundleEepIndexFunc)

			for _, obj := range objs {
				clientBuilder = clientBuilder.WithObjects(obj)
			}

			r.client = clientBuilder.Build()

			var res bool
			switch _resource := tc.resource.(type) {
			case *corev1.Secret:
				res = r.validateSecretForReconcile(_resource)
			case *corev1.ConfigMap:
				res = r.validateConfigMapForReconcile(_resource)
			case *certificatesv1b1.ClusterTrustBundle:
				res = r.validateClusterTrustBundleForReconcile(_resource)
			default:
				t.Fatalf("unknown resource type: %T", tc.resource)
			}

			require.Equal(t, tc.expect, res, "validation result mismatch")
		})
	}
}

// TestWasmTLSIndexerFunctions tests the indexer functions directly to ensure
// they correctly extract references from EnvoyExtensionPolicy objects.
func TestWasmTLSIndexerFunctions(t *testing.T) {
	testCases := []struct {
		name            string
		eep             *egv1a1.EnvoyExtensionPolicy
		expectedSecrets []string
		expectedCMs     []string
		expectedCTBs    []string
	}{
		{
			name: "HTTP WASM with Secret",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "default",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									URL: "https://example.com/wasm.wasm",
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Name: "ca-secret",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{"default/ca-secret"},
			expectedCMs:     []string{},
			expectedCTBs:    []string{},
		},
		{
			name: "Image WASM with ConfigMap",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "default",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.ImageWasmCodeSourceType,
								Image: &egv1a1.ImageWasmCodeSource{
									URL: "oci://example.com/wasm:v1.0.0",
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Kind: ptr.To[gwapiv1.Kind]("ConfigMap"),
											Name: "ca-configmap",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{},
			expectedCMs:     []string{"default/ca-configmap"},
			expectedCTBs:    []string{},
		},
		{
			name: "HTTP WASM with ClusterTrustBundle",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "default",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									URL: "https://example.com/wasm.wasm",
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Kind: ptr.To[gwapiv1.Kind](resource.KindClusterTrustBundle),
											Name: "ca-ctb",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{},
			expectedCMs:     []string{},
			expectedCTBs:    []string{"ca-ctb"},
		},
		{
			name: "Multiple WASM with different sources",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "test-ns",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Name: "secret1",
										},
									},
								},
							},
						},
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.ImageWasmCodeSourceType,
								Image: &egv1a1.ImageWasmCodeSource{
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Kind: ptr.To[gwapiv1.Kind]("ConfigMap"),
											Name: "cm1",
										},
									},
								},
							},
						},
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Kind: ptr.To[gwapiv1.Kind](resource.KindClusterTrustBundle),
											Name: "ctb1",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{"test-ns/secret1"},
			expectedCMs:     []string{"test-ns/cm1"},
			expectedCTBs:    []string{"ctb1"},
		},
		{
			name: "WASM with cross-namespace references",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "default",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									TLS: &egv1a1.WasmCodeSourceTLSConfig{
										CACertificateRef: gwapiv1.SecretObjectReference{
											Namespace: gatewayapi.NamespacePtr("other-ns"),
											Name:      "ca-secret",
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{"other-ns/ca-secret"},
			expectedCMs:     []string{},
			expectedCTBs:    []string{},
		},
		{
			name: "WASM without TLS config",
			eep: &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-eep",
					Namespace: "default",
				},
				Spec: egv1a1.EnvoyExtensionPolicySpec{
					Wasm: []egv1a1.Wasm{
						{
							Code: egv1a1.WasmCodeSource{
								Type: egv1a1.HTTPWasmCodeSourceType,
								HTTP: &egv1a1.HTTPWasmCodeSource{
									URL: "https://example.com/wasm.wasm",
								},
							},
						},
					},
				},
			},
			expectedSecrets: []string{},
			expectedCMs:     []string{},
			expectedCTBs:    []string{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			secrets := secretEnvoyExtensionPolicyIndexFunc(tc.eep)
			require.ElementsMatch(t, tc.expectedSecrets, secrets, "secret references mismatch")

			cms := configMapEepIndexFunc(tc.eep)
			require.ElementsMatch(t, tc.expectedCMs, cms, "configmap references mismatch")

			ctbs := clusterTrustBundleEepIndexFunc(tc.eep)
			require.ElementsMatch(t, tc.expectedCTBs, ctbs, "clustertrustbundle references mismatch")
		})
	}
}
