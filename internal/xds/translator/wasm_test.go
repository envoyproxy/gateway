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
	for _, local := range []bool{false, true} {
		source := "remote"
		if local {
			source = "local"
		}
		for _, shared := range []bool{false, true} {
			scope := "Policy"
			if shared {
				scope = "Namespace"
			}
			t.Run(source+"/"+scope, func(t *testing.T) {
				wasm := &ir.Wasm{
					Name:     "envoyextensionpolicy/shop/a/wasm/0",
					WasmName: "my-plugin",
					RootID:   new("my-root"),
				}
				if local {
					wasm.Path = "/var/lib/envoy/plugin.wasm"
				} else {
					wasm.Code = &ir.HTTPWasmCode{ServingURL: "https://envoy-gateway/plugin.wasm", SHA256: "module-sha"}
				}
				wantVMID := wasm.Name
				if shared {
					wasm.VMID = "envoyextensionpolicy/shop/wasm/module-sha"
					if local {
						wasm.VMID = "envoyextensionpolicy/shop/wasm/local"
					}
					wantVMID = wasm.VMID
				}

				got, err := wasmConfig(wasm)
				require.NoError(t, err)
				require.Equal(t, wantVMID, got.Config.GetVmConfig().VmId)
				require.Equal(t, wasm.WasmName, got.Config.Name)
				require.Equal(t, *wasm.RootID, got.Config.RootId)
				if local {
					require.Equal(t, wasm.Path, got.Config.GetVmConfig().Code.GetLocal().GetFilename())
				} else {
					require.Equal(t, wasm.Code.ServingURL, got.Config.GetVmConfig().Code.GetRemote().HttpUri.Uri)
					require.Equal(t, wasm.Code.SHA256, got.Config.GetVmConfig().Code.GetRemote().Sha256)
				}
			})
		}
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
