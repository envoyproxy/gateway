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
