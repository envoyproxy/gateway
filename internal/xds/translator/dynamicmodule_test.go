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

func TestRouteBackendPrecedesDynamicModuleBackend(t *testing.T) {
	destination := func(name, host string) *ir.RouteDestination {
		addressType := ir.IP
		weight := uint32(1)
		return &ir.RouteDestination{
			Name: name,
			Settings: []*ir.DestinationSetting{{
				Name:        name + "/backend/0",
				AddressType: &addressType,
				Protocol:    ir.HTTP,
				Weight:      &weight,
				Endpoints:   []*ir.DestinationEndpoint{ir.NewDestEndpoint(nil, host, 8080, false, nil)},
			}},
		}
	}

	// The first listener would claim "shared" first if module backends were added per listener.
	xdsIR := &ir.Xds{HTTP: []*ir.HTTPListener{
		{
			CoreListenerDetails: ir.CoreListenerDetails{Name: "gateway/module", Address: "0.0.0.0", Port: 10080},
			Hostnames:           []string{"module.example.com"},
			Routes: []*ir.HTTPRoute{{
				Name:        "module-route",
				Hostname:    "module.example.com",
				Destination: destination("module-route", "10.0.0.1"),
				EnvoyExtensions: &ir.EnvoyExtensionFeatures{DynamicModules: []ir.DynamicModule{{
					Name:       "policy/module/0",
					Path:       "/module.so",
					FilterName: "test",
					Backends:   []*ir.RouteDestination{destination("shared", "10.0.0.2")},
				}}},
			}},
		},
		{
			CoreListenerDetails: ir.CoreListenerDetails{Name: "gateway/route", Address: "0.0.0.0", Port: 10081},
			Hostnames:           []string{"route.example.com"},
			Routes: []*ir.HTTPRoute{{
				Name:        "application-route",
				Hostname:    "route.example.com",
				Destination: destination("shared", "10.0.0.3"),
			}},
		},
	}}

	tCtx, err := (&Translator{}).Translate(t.Context(), xdsIR)
	require.NoError(t, err)

	endpoint := findXdsEndpoint(tCtx, "shared")
	require.NotNil(t, endpoint)
	require.Equal(t, "10.0.0.3", endpoint.Endpoints[0].LbEndpoints[0].GetEndpoint().Address.GetSocketAddress().Address)
}
