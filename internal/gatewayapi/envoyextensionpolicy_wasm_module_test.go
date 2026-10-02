// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/wasm"
)

func TestBuildWasmFromRegisteredRemoteModule(t *testing.T) {
	const (
		epNamespace     = "proxy-ns"
		policyNamespace = "default"
	)
	caData := []byte("fake-ca-cert")
	pullSecretData := []byte("fake-docker-config")

	secrets := []*corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: epNamespace, Name: "ca-secret"},
			Data:       map[string][]byte{"ca.crt": caData},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: epNamespace, Name: "pull-secret"},
			Data:       map[string][]byte{corev1.DockerConfigJsonKey: pullSecretData},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Namespace: policyNamespace, Name: "pull-secret"},
			Data:       map[string][]byte{corev1.DockerConfigJsonKey: pullSecretData},
		},
	}

	httpModule := egv1a1.WasmModuleSource{
		Type: new(egv1a1.HTTPWasmModuleSourceType),
		HTTP: &egv1a1.HTTPWasmCodeSource{
			URL: "https://example.com/filter.wasm",
			TLS: &egv1a1.WasmCodeSourceTLSConfig{
				CACertificateRef: gwapiv1.SecretObjectReference{Name: "ca-secret"},
			},
		},
	}
	imageModule := egv1a1.WasmModuleSource{
		Type: new(egv1a1.ImageWasmModuleSourceType),
		Image: &egv1a1.ImageWasmCodeSource{
			URL:           "example.com/filter:v1",
			PullSecretRef: &gwapiv1.SecretObjectReference{Name: "pull-secret"},
		},
		PullPolicy: new(egv1a1.ImagePullPolicyAlways),
	}
	crossNamespacePullSecret := imageModule.DeepCopy()
	crossNamespacePullSecret.Image.PullSecretRef.Namespace = new(gwapiv1.Namespace(policyNamespace))
	crossNamespaceCA := httpModule.DeepCopy()
	crossNamespaceCA.HTTP.TLS.CACertificateRef.Namespace = new(gwapiv1.Namespace(policyNamespace))
	noRefs := egv1a1.WasmModuleSource{
		Type: new(egv1a1.HTTPWasmModuleSourceType),
		HTTP: &egv1a1.HTTPWasmCodeSource{URL: "https://example.com/filter.wasm"},
	}

	namespacedProxy := metav1.ObjectMeta{Namespace: epNamespace, Name: "proxy", ResourceVersion: "7"}

	tests := []struct {
		name           string
		proxyMeta      metav1.ObjectMeta
		source         egv1a1.WasmModuleSource
		nilCache       bool
		wantErr        string
		wantURL        string
		wantCA         []byte
		wantPullSecret []byte
		wantPullPolicy wasm.PullPolicy
	}{
		{
			name:      "HTTP module resolves CA in the EnvoyProxy namespace",
			proxyMeta: namespacedProxy,
			source:    httpModule,
			wantURL:   "https://example.com/filter.wasm",
			wantCA:    caData,
		},
		{
			name:           "Image module resolves pull secret in the EnvoyProxy namespace",
			proxyMeta:      namespacedProxy,
			source:         imageModule,
			wantURL:        "oci://example.com/filter:v1",
			wantPullSecret: pullSecretData,
			wantPullPolicy: wasm.Always,
		},
		{
			name:      "cross-namespace pull secret is rejected",
			proxyMeta: namespacedProxy,
			source:    *crossNamespacePullSecret,
			wantErr:   "secret ref namespace must be unspecified/empty or proxy-ns",
		},
		{
			name:      "cross-namespace CA is rejected",
			proxyMeta: namespacedProxy,
			source:    *crossNamespaceCA,
			wantErr:   "caCertificateRef namespace must be unspecified/empty or proxy-ns",
		},
		{
			name:    "default EnvoyProxy spec rejects object refs",
			source:  imageModule,
			wantErr: "not supported in the EnvoyGateway default EnvoyProxy spec",
		},
		{
			name:    "default EnvoyProxy spec allows modules without refs",
			source:  noRefs,
			wantURL: "https://example.com/filter.wasm",
		},
		{
			name:      "missing Wasm cache",
			proxyMeta: namespacedProxy,
			source:    noRefs,
			nilCache:  true,
			wantErr:   "wasm cache is not initialized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				gotURL  string
				gotOpts *wasm.GetOptions
			)
			translator := &Translator{TranslatorContext: &TranslatorContext{}}
			if !tt.nilCache {
				translator.WasmCache = &caCapturingMockWasmCache{
					GetFunc: func(downloadURL string, opts *wasm.GetOptions) (string, string, error) {
						gotURL, gotOpts = downloadURL, opts
						return "http://eg.default/wasm", "1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef", nil
					},
				}
			}
			translator.SetSecrets(secrets)

			envoyProxy := &egv1a1.EnvoyProxy{
				ObjectMeta: tt.proxyMeta,
				Spec: egv1a1.EnvoyProxySpec{
					WasmModules: []egv1a1.WasmModuleEntry{{Name: "filter", Source: tt.source}},
				},
			}
			policy := &egv1a1.EnvoyExtensionPolicy{
				ObjectMeta: metav1.ObjectMeta{Namespace: policyNamespace, Name: "policy", ResourceVersion: "99"},
			}
			config := &egv1a1.Wasm{Name: new("filter")}

			wasmIR, err := translator.buildWasm("test-wasm", config, policy, 0, &resource.Resources{Secrets: secrets}, envoyProxy)
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, wasmIR.Code)
			assert.Empty(t, wasmIR.Path)
			assert.Equal(t, "http://eg.default/wasm", wasmIR.Code.ServingURL)
			assert.Equal(t, tt.wantURL, gotURL)
			assert.Equal(t, tt.wantCA, gotOpts.CACert)
			assert.Equal(t, tt.wantPullSecret, gotOpts.PullSecret)
			assert.Equal(t, tt.wantPullPolicy, gotOpts.PullPolicy)
			// The cache entry belongs to the EnvoyProxy, not the policy.
			assert.Equal(t, "envoyproxy/"+tt.proxyMeta.Namespace+"/"+tt.proxyMeta.Name+"/wasm/filter", gotOpts.ResourceName)
			assert.Equal(t, tt.proxyMeta.ResourceVersion, gotOpts.ResourceVersion)
		})
	}
}

func TestDeprecatedFieldsUsedInEnvoyExtensionPolicyWasmCode(t *testing.T) {
	inline := &egv1a1.EnvoyExtensionPolicy{Spec: egv1a1.EnvoyExtensionPolicySpec{
		Wasm: []egv1a1.Wasm{
			{Name: new("registered")},
			{Code: &egv1a1.WasmCodeSource{Type: egv1a1.HTTPWasmCodeSourceType}},
		},
	}}
	assert.Equal(t, map[string]string{"spec.wasm.code": "EnvoyProxy spec.wasmModules"}, deprecatedFieldsUsedInEnvoyExtensionPolicy(inline))

	registered := &egv1a1.EnvoyExtensionPolicy{Spec: egv1a1.EnvoyExtensionPolicySpec{
		Wasm: []egv1a1.Wasm{{Name: new("registered")}},
	}}
	assert.Empty(t, deprecatedFieldsUsedInEnvoyExtensionPolicy(registered))
}
