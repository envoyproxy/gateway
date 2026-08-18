// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	wasmfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/wasm/v3"
	wasmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/wasm/v3"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/types/known/wrapperspb"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

func getVMConfig(t *testing.T, cfg *wasmfilterv3.Wasm) *wasmv3.VmConfig {
	t.Helper()
	vmWrapper, ok := cfg.Config.Vm.(*wasmv3.PluginConfig_VmConfig)
	require.True(t, ok, "expected PluginConfig_VmConfig")
	require.NotNil(t, vmWrapper.VmConfig)
	return vmWrapper.VmConfig
}

func TestWasmConfigLocalSource(t *testing.T) {
	wasm := &ir.Wasm{
		Name:     "test-wasm",
		WasmName: "my-wasm",
		FailOpen: false,
		LocalCode: &ir.LocalWasmCode{
			Filename: "/etc/wasm/plugins/test.wasm",
		},
	}

	cfg, err := wasmConfig(wasm)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.Config)
	require.Equal(t, "my-wasm", cfg.Config.Name)
	require.False(t, cfg.Config.FailOpen)

	vmConfig := getVMConfig(t, cfg)
	require.Equal(t, "test-wasm", vmConfig.VmId)
	require.Equal(t, vmRuntimeV8, vmConfig.Runtime)

	local, ok := vmConfig.Code.Specifier.(*corev3.AsyncDataSource_Local)
	require.True(t, ok, "expected local specifier")
	filename, ok := local.Local.Specifier.(*corev3.DataSource_Filename)
	require.True(t, ok, "expected filename specifier")
	require.Equal(t, "/etc/wasm/plugins/test.wasm", filename.Filename)
}

func TestWasmConfigRemoteSource(t *testing.T) {
	wasm := &ir.Wasm{
		Name:     "test-wasm",
		WasmName: "my-wasm",
		FailOpen: true,
		Code: &ir.HTTPWasmCode{
			ServingURL:  "http://envoy-gateway:18000/wasm/test.wasm",
			SHA256:      "abcdef0123456789",
			OriginalURL: "https://example.com/test.wasm",
		},
	}

	cfg, err := wasmConfig(wasm)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "my-wasm", cfg.Config.Name)
	require.True(t, cfg.Config.FailOpen)

	vmConfig := getVMConfig(t, cfg)

	remote, ok := vmConfig.Code.Specifier.(*corev3.AsyncDataSource_Remote)
	require.True(t, ok, "expected remote specifier")
	require.Equal(t, "http://envoy-gateway:18000/wasm/test.wasm", remote.Remote.HttpUri.Uri)
	require.Equal(t, "abcdef0123456789", remote.Remote.Sha256)
}

func TestWasmConfigWithRootID(t *testing.T) {
	rootID := "my-root-id"
	wasm := &ir.Wasm{
		Name:     "test-wasm",
		WasmName: "my-wasm",
		RootID:   &rootID,
		LocalCode: &ir.LocalWasmCode{
			Filename: "/etc/wasm/plugins/test.wasm",
		},
	}

	cfg, err := wasmConfig(wasm)
	require.NoError(t, err)
	require.NotNil(t, cfg)
	require.Equal(t, rootID, cfg.Config.RootId)
}

func TestWasmConfigWithEnvVariables(t *testing.T) {
	wasm := &ir.Wasm{
		Name:      "test-wasm",
		WasmName:  "my-wasm",
		HostKeys:  []string{"API_KEY", "SECRET_TOKEN"},
		LocalCode: &ir.LocalWasmCode{Filename: "/etc/wasm/plugins/test.wasm"},
	}

	cfg, err := wasmConfig(wasm)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	vmConfig := getVMConfig(t, cfg)
	require.NotNil(t, vmConfig.EnvironmentVariables)
	require.Equal(t, []string{"API_KEY", "SECRET_TOKEN"}, vmConfig.EnvironmentVariables.HostEnvKeys)
}

func TestWasmConfigWithPluginConfig(t *testing.T) {
	wasm := &ir.Wasm{
		Name:     "test-wasm",
		WasmName: "my-wasm",
		Config: &apiextensionsv1.JSON{
			Raw: []byte(`{"key":"value"}`),
		},
		LocalCode: &ir.LocalWasmCode{
			Filename: "/etc/wasm/plugins/test.wasm",
		},
	}

	cfg, err := wasmConfig(wasm)
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.NotNil(t, cfg.Config.Configuration)
	strVal := &wrapperspb.StringValue{}
	err = cfg.Config.Configuration.UnmarshalTo(strVal)
	require.NoError(t, err)
	require.Contains(t, strVal.Value, `"key":"value"`)
}

func TestBuildHCMWasmFilterLocal(t *testing.T) {
	wasm := &ir.Wasm{
		Name:      "envoyextensionpolicy/default/policy/wasm/0",
		WasmName:  "my-wasm",
		FailOpen:  false,
		LocalCode: &ir.LocalWasmCode{Filename: "/etc/wasm/plugins/test.wasm"},
	}

	filter, err := buildHCMWasmFilter(wasm)
	require.NoError(t, err)
	require.NotNil(t, filter)
	require.True(t, filter.Disabled)
	require.Equal(t, "envoy.filters.http.wasm/envoyextensionpolicy/default/policy/wasm/0", filter.Name)

	wasmProto := &wasmfilterv3.Wasm{}
	err = filter.GetTypedConfig().UnmarshalTo(wasmProto)
	require.NoError(t, err)

	vmConfig := getVMConfig(t, wasmProto)

	local, ok := vmConfig.Code.Specifier.(*corev3.AsyncDataSource_Local)
	require.True(t, ok, "expected local specifier")
	filename, ok := local.Local.Specifier.(*corev3.DataSource_Filename)
	require.True(t, ok, "expected filename specifier")
	require.Equal(t, "/etc/wasm/plugins/test.wasm", filename.Filename)
}
