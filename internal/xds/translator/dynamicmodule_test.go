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

func TestDynamicModuleSource(t *testing.T) {
	tests := []struct {
		name         string
		dm           ir.DynamicModule
		wantRemote   bool
		wantCluster  string
		wantURI      string
		wantSHA256   string
		wantFilename string
		wantErr      bool
	}{
		{
			name: "local source",
			dm: ir.DynamicModule{
				Path: "/usr/lib/envoy/modules/my_auth.so",
			},
			wantRemote:   false,
			wantFilename: "/usr/lib/envoy/modules/my_auth.so",
		},
		{
			name: "remote source with https default port",
			dm: ir.DynamicModule{
				Remote: &ir.RemoteDynamicModuleSource{
					URL:    "https://modules.example.com/libremote_auth.so",
					SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantRemote:  true,
			wantCluster: "modules_example_com_443",
			wantURI:     "https://modules.example.com/libremote_auth.so",
			wantSHA256:  "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name: "remote source with http default port",
			dm: ir.DynamicModule{
				Remote: &ir.RemoteDynamicModuleSource{
					URL:    "http://modules.example.com/libremote_auth.so",
					SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantRemote:  true,
			wantCluster: "modules_example_com_80",
			wantURI:     "http://modules.example.com/libremote_auth.so",
			wantSHA256:  "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name: "remote source with explicit port",
			dm: ir.DynamicModule{
				Remote: &ir.RemoteDynamicModuleSource{
					URL:    "https://modules.example.com:8443/libremote_auth.so",
					SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantRemote:  true,
			wantCluster: "modules_example_com_8443",
			wantURI:     "https://modules.example.com:8443/libremote_auth.so",
			wantSHA256:  "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name: "remote source with http explicit port",
			dm: ir.DynamicModule{
				Remote: &ir.RemoteDynamicModuleSource{
					URL:    "http://modules.example.com:8443/libremote_auth.so",
					SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantRemote:  true,
			wantCluster: "modules_example_com_8443",
			wantURI:     "http://modules.example.com:8443/libremote_auth.so",
			wantSHA256:  "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
		{
			name: "invalid remote URL",
			dm: ir.DynamicModule{
				Remote: &ir.RemoteDynamicModuleSource{
					URL:    "://invalid",
					SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := dynamicModuleSource(&tt.dm)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, got)

			if tt.wantRemote {
				remote, ok := got.Specifier.(*corev3.AsyncDataSource_Remote)
				require.True(t, ok, "expected remote specifier")
				require.Equal(t, tt.wantURI, remote.Remote.HttpUri.Uri)
				require.Equal(t, tt.wantSHA256, remote.Remote.Sha256)
				clusterSpec, ok := remote.Remote.HttpUri.HttpUpstreamType.(*corev3.HttpUri_Cluster)
				require.True(t, ok, "expected cluster upstream type")
				require.Equal(t, tt.wantCluster, clusterSpec.Cluster)
			} else {
				local, ok := got.Specifier.(*corev3.AsyncDataSource_Local)
				require.True(t, ok, "expected local specifier")
				filename, ok := local.Local.Specifier.(*corev3.DataSource_Filename)
				require.True(t, ok, "expected filename specifier")
				require.Equal(t, tt.wantFilename, filename.Filename)
			}
		})
	}
}
