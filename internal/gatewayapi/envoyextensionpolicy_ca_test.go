// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	certificatesv1b1 "k8s.io/api/certificates/v1beta1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/wasm"
)

type caCapturingMockWasmCache struct {
	GetFunc func(downloadURL string, opts *wasm.GetOptions) (string, string, error)
}

func (m *caCapturingMockWasmCache) Start(_ context.Context) {}
func (m *caCapturingMockWasmCache) Get(downloadURL string, opts *wasm.GetOptions) (string, string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(downloadURL, opts)
	}
	return "http://eg.default/wasm", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil
}

func TestBuildWasmWithTLS(t *testing.T) {
	ns := "default"
	caData := []byte("fake-ca-cert")

	tests := []struct {
		name      string
		wasm      *egv1a1.Wasm
		resources *resource.Resources
		wantCA    []byte
		wantErr   bool
	}{
		{
			name: "HTTP with TLS",
			wasm: &egv1a1.Wasm{
				Code: egv1a1.WasmCodeSource{
					Type: egv1a1.HTTPWasmCodeSourceType,
					HTTP: &egv1a1.HTTPWasmCodeSource{
						URL: "https://example.com/wasm",
						TLS: &egv1a1.WasmCodeSourceTLSConfig{
							CACertificateRef: gwapiv1.SecretObjectReference{
								Name: "ca-secret",
							},
						},
					},
				},
			},
			resources: &resource.Resources{
				Secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-secret",
						},
						Data: map[string][]byte{
							"ca.crt": caData,
						},
					},
				},
			},
			wantCA: caData,
		},
		{
			name: "Image with TLS",
			wasm: &egv1a1.Wasm{
				Code: egv1a1.WasmCodeSource{
					Type: egv1a1.ImageWasmCodeSourceType,
					Image: &egv1a1.ImageWasmCodeSource{
						URL: "example.com/wasm:v1",
						TLS: &egv1a1.WasmCodeSourceTLSConfig{
							CACertificateRef: gwapiv1.SecretObjectReference{
								Name: "ca-secret",
							},
						},
					},
				},
			},
			resources: &resource.Resources{
				Secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-secret",
						},
						Data: map[string][]byte{
							"ca.crt": caData,
						},
					},
				},
			},
			wantCA: caData,
		},
		{
			name: "HTTP with TLS error",
			wasm: &egv1a1.Wasm{
				Code: egv1a1.WasmCodeSource{
					Type: egv1a1.HTTPWasmCodeSourceType,
					HTTP: &egv1a1.HTTPWasmCodeSource{
						URL: "https://example.com/wasm",
						TLS: &egv1a1.WasmCodeSourceTLSConfig{
							CACertificateRef: gwapiv1.SecretObjectReference{
								Name: "non-existent",
							},
						},
					},
				},
			},
			resources: &resource.Resources{},
			wantErr:   true,
		},
		{
			name: "Image with TLS error",
			wasm: &egv1a1.Wasm{
				Code: egv1a1.WasmCodeSource{
					Type: egv1a1.ImageWasmCodeSourceType,
					Image: &egv1a1.ImageWasmCodeSource{
						URL: "example.com/wasm:v1",
						TLS: &egv1a1.WasmCodeSourceTLSConfig{
							CACertificateRef: gwapiv1.SecretObjectReference{
								Name: "non-existent",
							},
						},
					},
				},
			},
			resources: &resource.Resources{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var capturedCA []byte
			mockCache := &caCapturingMockWasmCache{
				GetFunc: func(downloadURL string, opts *wasm.GetOptions) (string, string, error) {
					capturedCA = opts.CACert
					return "http://eg.default/wasm", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil
				},
			}

			translator := &Translator{
				TranslatorContext: &TranslatorContext{},
				WasmCache:         mockCache,
			}
			if tt.resources != nil {
				translator.SetSecrets(tt.resources.Secrets)
				translator.SetConfigMaps(tt.resources.ConfigMaps)
				translator.SetClusterTrustBundles(tt.resources.ClusterTrustBundles)
			}

			policy := &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: ns,
					Name:      "test-policy",
				},
			}

			_, err := translator.buildWasm("test-wasm", tt.wasm, policy, 0, tt.resources)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantCA, capturedCA)
			}
		})
	}
}

func TestResolveCACertRef(t *testing.T) {
	ns := "default"
	caData := []byte("fake-ca-cert")

	tests := []struct {
		name      string
		caCertRef gwapiv1.SecretObjectReference
		resources *resource.Resources
		want      []byte
		wantErr   bool
	}{
		{
			name: "resolve from secret",
			caCertRef: gwapiv1.SecretObjectReference{
				Name: "ca-secret",
			},
			resources: &resource.Resources{
				Secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-secret",
						},
						Data: map[string][]byte{
							"ca.crt": caData,
						},
					},
				},
			},
			want: caData,
		},
		{
			name: "resolve from configmap",
			caCertRef: gwapiv1.SecretObjectReference{
				Kind: ptrTo(gwapiv1.Kind("ConfigMap")),
				Name: "ca-cm",
			},
			resources: &resource.Resources{
				ConfigMaps: []*corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-cm",
						},
						Data: map[string]string{
							"ca.crt": string(caData),
						},
					},
				},
			},
			want: caData,
		},
		{
			name: "resolve from clustertrustbundle",
			caCertRef: gwapiv1.SecretObjectReference{
				Kind: ptrTo(gwapiv1.Kind(resource.KindClusterTrustBundle)),
				Name: "ca-ctb",
			},
			resources: &resource.Resources{
				ClusterTrustBundles: []*certificatesv1b1.ClusterTrustBundle{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "ca-ctb",
						},
						Spec: certificatesv1b1.ClusterTrustBundleSpec{
							TrustBundle: string(caData),
						},
					},
				},
			},
			want: caData,
		},
		{
			name: "secret missing ca.crt",
			caCertRef: gwapiv1.SecretObjectReference{
				Name: "ca-secret",
			},
			resources: &resource.Resources{
				Secrets: []*corev1.Secret{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-secret",
						},
						Data: map[string][]byte{
							"other.crt": caData,
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "configmap missing ca.crt",
			caCertRef: gwapiv1.SecretObjectReference{
				Kind: ptrTo(gwapiv1.Kind("ConfigMap")),
				Name: "ca-cm",
			},
			resources: &resource.Resources{
				ConfigMaps: []*corev1.ConfigMap{
					{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: ns,
							Name:      "ca-cm",
						},
						Data: map[string]string{
							"other.crt": string(caData),
						},
					},
				},
			},
			wantErr: true,
		},
		{
			name: "clustertrustbundle does not exist",
			caCertRef: gwapiv1.SecretObjectReference{
				Kind: ptrTo(gwapiv1.Kind(resource.KindClusterTrustBundle)),
				Name: "non-existent",
			},
			resources: &resource.Resources{},
			wantErr:   true,
		},
		{
			name: "secret does not exist",
			caCertRef: gwapiv1.SecretObjectReference{
				Name: "non-existent",
			},
			resources: &resource.Resources{},
			wantErr:   true,
		},
		{
			name: "configmap does not exist",
			caCertRef: gwapiv1.SecretObjectReference{
				Kind: ptrTo(gwapiv1.Kind(resource.KindConfigMap)),
				Name: "non-existent",
			},
			resources: &resource.Resources{},
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			translator := &Translator{
				TranslatorContext: &TranslatorContext{},
			}
			translator.SetSecrets(tt.resources.Secrets)
			translator.SetConfigMaps(tt.resources.ConfigMaps)
			translator.SetClusterTrustBundles(tt.resources.ClusterTrustBundles)

			got, err := translator.resolveCACertRef(ns, tt.caCertRef, tt.resources)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func ptrTo[T any](v T) *T {
	return &v
}
