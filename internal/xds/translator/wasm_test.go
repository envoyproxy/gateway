// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/require"

	"github.com/envoyproxy/gateway/internal/ir"
)

func TestWasmConfigVMSharing(t *testing.T) {
	const (
		policyVMID       = "envoyextensionpolicy/shop/a/wasm/0"
		sharedRemoteVMID = "envoyextensionpolicy/shop/wasm/module-sha"
		sharedLocalVMID  = "envoyextensionpolicy/shop/wasm/local"
		localPath        = "/var/lib/envoy/plugin.wasm"
	)
	remoteCode := &ir.HTTPWasmCode{ServingURL: "https://envoy-gateway/plugin.wasm", SHA256: "module-sha"}
	tests := []struct {
		name         string
		code         *ir.HTTPWasmCode
		path         string
		vmID         string
		wantVMID     string
		wantFilename string
		wantURI      string
		wantSHA256   string
	}{
		{
			name: "remote/Policy", code: remoteCode,
			wantVMID: policyVMID, wantURI: remoteCode.ServingURL, wantSHA256: remoteCode.SHA256,
		},
		{
			name: "remote/Namespace", code: remoteCode, vmID: sharedRemoteVMID,
			wantVMID: sharedRemoteVMID, wantURI: remoteCode.ServingURL, wantSHA256: remoteCode.SHA256,
		},
		{
			name: "local/Policy", path: localPath,
			wantVMID: policyVMID, wantFilename: localPath,
		},
		{
			name: "local/Namespace", path: localPath, vmID: sharedLocalVMID,
			wantVMID: sharedLocalVMID, wantFilename: localPath,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wasm := &ir.Wasm{
				Name:     policyVMID,
				WasmName: "my-plugin",
				RootID:   new("my-root"),
				Code:     tt.code,
				Path:     tt.path,
				VMID:     tt.vmID,
			}
			got, err := wasmConfig(wasm)
			require.NoError(t, err)
			vm := got.Config.GetVmConfig()
			require.Equal(t, tt.wantVMID, vm.GetVmId())
			require.Equal(t, wasm.WasmName, got.Config.Name)
			require.Equal(t, *wasm.RootID, got.Config.RootId)
			require.Equal(t, tt.wantFilename, vm.GetCode().GetLocal().GetFilename())
			require.Equal(t, tt.wantURI, vm.GetCode().GetRemote().GetHttpUri().GetUri())
			require.Equal(t, tt.wantSHA256, vm.GetCode().GetRemote().GetSha256())
		})
	}
}

func TestWasmCodeSource(t *testing.T) {
	tests := []struct {
		name         string
		wasm         ir.Wasm
		wantLocal    bool
		wantFilename string
		wantURI      string
		wantSHA256   string
		wantErr      bool
	}{
		{
			name: "filesystem local path",
			wasm: ir.Wasm{
				Path: "/var/lib/envoy/filter.wasm",
			},
			wantLocal:    true,
			wantFilename: "/var/lib/envoy/filter.wasm",
		},
		{
			name: "http remote code",
			wasm: ir.Wasm{
				Code: &ir.HTTPWasmCode{
					ServingURL: "https://envoy-gateway:18002/module.wasm",
					SHA256:     "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantLocal:  false,
			wantURI:    "https://envoy-gateway:18002/module.wasm",
			wantSHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name:    "missing source",
			wasm:    ir.Wasm{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := wasmCodeSource(&tt.wasm)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)

			if tt.wantLocal {
				local, ok := got.Specifier.(*corev3.AsyncDataSource_Local)
				require.True(t, ok, "expected local specifier")
				filename, ok := local.Local.Specifier.(*corev3.DataSource_Filename)
				require.True(t, ok, "expected filename specifier")
				require.Equal(t, tt.wantFilename, filename.Filename)
				return
			}

			remote, ok := got.Specifier.(*corev3.AsyncDataSource_Remote)
			require.True(t, ok, "expected remote specifier")
			require.Equal(t, tt.wantURI, remote.Remote.HttpUri.Uri)
			require.Equal(t, tt.wantSHA256, remote.Remote.Sha256)
		})
	}
}
