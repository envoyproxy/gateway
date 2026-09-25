// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/wasm"
)

func TestBuildWasmVMSharing(t *testing.T) {
	const (
		moduleSHA = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
		otherSHA  = "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"
		imageSHA  = "cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
		sharedID  = "envoyextensionpolicy/shop/wasm/" + moduleSHA
	)
	tests := []struct {
		name         string
		shareVM      *bool
		namespace    string
		policyName   string
		moduleSHA    string
		image        bool
		hostKeys     []string
		wantVMID     string
		wantHostKeys []string
	}{
		{
			name: "omitted", namespace: "shop", policyName: "a", moduleSHA: moduleSHA,
			hostKeys: []string{"REGION", "API_TOKEN"}, wantHostKeys: []string{"REGION", "API_TOKEN"},
		},
		{
			name: "disabled", shareVM: new(false), namespace: "shop", policyName: "a", moduleSHA: moduleSHA,
			hostKeys: []string{"REGION", "API_TOKEN"}, wantHostKeys: []string{"REGION", "API_TOKEN"},
		},
		{
			name: "shared with canonical host keys", shareVM: new(true), namespace: "shop", policyName: "a", moduleSHA: moduleSHA,
			hostKeys: []string{"REGION", "API_TOKEN", "REGION"}, wantHostKeys: []string{"API_TOKEN", "REGION"}, wantVMID: sharedID,
		},
		{
			name: "same module in another policy", shareVM: new(true), namespace: "shop", policyName: "b", moduleSHA: moduleSHA,
			hostKeys: []string{"API_TOKEN", "REGION"}, wantHostKeys: []string{"API_TOKEN", "REGION"}, wantVMID: sharedID,
		},
		{
			name: "another namespace", shareVM: new(true), namespace: "other", policyName: "a", moduleSHA: moduleSHA,
			wantVMID: "envoyextensionpolicy/other/wasm/" + moduleSHA,
		},
		{
			name: "changed module at the same URL", shareVM: new(true), namespace: "shop", policyName: "a", moduleSHA: otherSHA,
			wantVMID: "envoyextensionpolicy/shop/wasm/" + otherSHA,
		},
		{
			name: "OCI uses extracted module checksum", shareVM: new(true), namespace: "shop", policyName: "a", moduleSHA: moduleSHA,
			image: true, wantVMID: sharedID,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := &egv1a1.EnvoyExtensionPolicy{ObjectMeta: metav1.ObjectMeta{Namespace: tt.namespace, Name: tt.policyName}}
			config := &egv1a1.Wasm{
				ShareVM: tt.shareVM,
				Env:     &egv1a1.WasmEnv{HostKeys: tt.hostKeys},
				Code: egv1a1.WasmCodeSource{
					Type: egv1a1.HTTPWasmCodeSourceType,
					HTTP: &egv1a1.HTTPWasmCodeSource{URL: "https://example.com/plugin.wasm"},
				},
			}
			if tt.image {
				config.Code = egv1a1.WasmCodeSource{
					Type:  egv1a1.ImageWasmCodeSourceType,
					Image: &egv1a1.ImageWasmCodeSource{URL: "example.com/plugin:v1", SHA256: new(imageSHA)},
				}
			}
			original := config.DeepCopy()
			cache := &caCapturingMockWasmCache{GetFunc: func(_ string, _ *wasm.GetOptions) (string, string, error) {
				return "https://envoy-gateway/plugin.wasm", tt.moduleSHA, nil
			}}
			translator := &Translator{TranslatorContext: &TranslatorContext{}, WasmCache: cache}
			got, err := translator.buildWasm(irConfigNameForWasm(policy, 0), config, policy, 0, &resource.Resources{})
			require.NoError(t, err)
			assert.Equal(t, tt.wantVMID, got.VMID)
			assert.Equal(t, tt.wantHostKeys, got.HostKeys)
			assert.Equal(t, original, config, "translation must not mutate the policy")
		})
	}
}

func Test_hasTag(t *testing.T) {
	tests := []struct {
		name     string
		imageURL string
		want     bool
	}{
		{
			name:     "image with scheme and tag",
			imageURL: "oci://www.example.com/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image with scheme, host port and tag",
			imageURL: "oci://www.example.com:8080/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image with scheme without tag",
			imageURL: "oci://www.example.com/wasm",
			want:     false,
		},
		{
			name:     "image with scheme, host port without tag",
			imageURL: "oci://www.example.com:8080/wasm",
			want:     false,
		},
		{
			name:     "image without scheme with tag",
			imageURL: "www.example.com/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image without scheme with host port and tag",
			imageURL: "www.example.com:8080/wasm:v1.0.0",
			want:     true,
		},
		{
			name:     "image without scheme without tag",
			imageURL: "www.example.com/wasm",
			want:     false,
		},
		{
			name:     "image without scheme with host port without tag",
			imageURL: "www.example.com:8080/wasm",
			want:     false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, hasTag(tt.imageURL), "hasTag(%v)", tt.imageURL)
		})
	}
}

func TestValidateDynamicModuleRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantErr string
	}{
		{
			name:   "valid https URL",
			rawURL: "https://modules.example.com/libremote_auth.so",
		},
		{
			name:   "valid http URL with port",
			rawURL: "http://modules.example.com:8443/libremote_auth.so",
		},
		{
			name:    "missing hostname",
			rawURL:  "https:///libremote_auth.so",
			wantErr: "hostname",
		},
		{
			name:    "unsupported scheme",
			rawURL:  "ftp://modules.example.com/libremote_auth.so",
			wantErr: "unsupported URL scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateDynamicModuleRemoteURL(tt.rawURL)
			if tt.wantErr == "" {
				assert.NoError(t, err)
				return
			}

			if assert.Error(t, err) {
				assert.ErrorContains(t, err, tt.wantErr)
			}
		})
	}
}

func Test_envoyExtensionPolicyOwnerChoose(t *testing.T) {
	t.Run("route policy overrides parent for the same owner fields", func(t *testing.T) {
		parentPolicy := &egv1a1.EnvoyExtensionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "parent-ns"},
			Spec: egv1a1.EnvoyExtensionPolicySpec{
				Wasm: []egv1a1.Wasm{{
					Name: new("parent-wasm"),
				}},
				ExtProc: []egv1a1.ExtProc{{
					BackendCluster: egv1a1.BackendCluster{
						BackendRefs: []egv1a1.BackendRef{{
							BackendObjectReference: gwapiv1.BackendObjectReference{Name: "parent-extproc"},
						}},
					},
				}},
				Lua: []egv1a1.Lua{{
					Type:     egv1a1.LuaValueTypeValueRef,
					ValueRef: &gwapiv1.LocalObjectReference{Name: "parent-lua-cm"},
				}},
				DynamicModule: []egv1a1.DynamicModule{{
					Name: "parent-dm",
				}},
			},
		}

		routePolicy := &egv1a1.EnvoyExtensionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "route", Namespace: "route-ns"},
			Spec: egv1a1.EnvoyExtensionPolicySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Wasm: []egv1a1.Wasm{{
					Name: new("route-wasm"),
				}},
				ExtProc: []egv1a1.ExtProc{{
					BackendCluster: egv1a1.BackendCluster{
						BackendRefs: []egv1a1.BackendRef{{
							BackendObjectReference: gwapiv1.BackendObjectReference{Name: "route-extproc"},
						}},
					},
				}},
				Lua: []egv1a1.Lua{{
					Type:     egv1a1.LuaValueTypeValueRef,
					ValueRef: &gwapiv1.LocalObjectReference{Name: "route-lua-cm"},
				}},
				DynamicModule: []egv1a1.DynamicModule{{
					Name: "route-dm",
				}},
			},
		}

		_, owners, err := mergeEnvoyExtensionPolicy(routePolicy, parentPolicy)
		require.NoError(t, err)
		require.NotNil(t, owners)

		assert.Same(t, routePolicy, owners.wasm)
		assert.Same(t, routePolicy, owners.extProc)
		assert.Same(t, routePolicy, owners.lua)
		assert.Same(t, routePolicy, owners.dynamicModule)
	})

	t.Run("uses parent owner when route does not set the field", func(t *testing.T) {
		parentPolicy := &egv1a1.EnvoyExtensionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "parent", Namespace: "parent-ns"},
			Spec: egv1a1.EnvoyExtensionPolicySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Wasm: []egv1a1.Wasm{{
					Name: new("parent-wasm"),
				}},
				ExtProc: []egv1a1.ExtProc{{
					BackendCluster: egv1a1.BackendCluster{
						BackendRefs: []egv1a1.BackendRef{{
							BackendObjectReference: gwapiv1.BackendObjectReference{Name: "parent-extproc"},
						}},
					},
				}},
				Lua: []egv1a1.Lua{{
					Type:     egv1a1.LuaValueTypeValueRef,
					ValueRef: &gwapiv1.LocalObjectReference{Name: "parent-lua-cm"},
				}},
				DynamicModule: []egv1a1.DynamicModule{{
					Name: "parent-dm",
				}},
			},
		}

		routePolicy := &egv1a1.EnvoyExtensionPolicy{
			ObjectMeta: metav1.ObjectMeta{Name: "route", Namespace: "route-ns"},
			Spec: egv1a1.EnvoyExtensionPolicySpec{
				MergeType: new(egv1a1.StrategicMerge),
				Wasm: []egv1a1.Wasm{{
					Name: new("route-wasm"),
				}},
			},
		}

		_, owners, err := mergeEnvoyExtensionPolicy(routePolicy, parentPolicy)
		require.NoError(t, err)
		require.NotNil(t, owners)

		assert.Same(t, routePolicy, owners.wasm)
		assert.Same(t, parentPolicy, owners.extProc)
		assert.Same(t, parentPolicy, owners.lua)
		assert.Same(t, parentPolicy, owners.dynamicModule)
	})
}
