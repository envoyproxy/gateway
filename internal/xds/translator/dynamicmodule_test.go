// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"testing"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
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

// TestDynamicModuleSourceRetryAndTimeout verifies that retry and timeout
// settings on a remote dynamic module flow into the resulting AsyncDataSource.
func TestDynamicModuleSourceRetryAndTimeout(t *testing.T) {
	numRetries := uint32(3)
	dm := ir.DynamicModule{
		Remote: &ir.RemoteDynamicModuleSource{
			URL:    "https://modules.example.com/libremote_auth.so",
			SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Traffic: &ir.TrafficFeatures{
				Retry: &ir.Retry{
					NumRetries: &numRetries,
					RetryOn: &ir.RetryOn{
						Triggers: []ir.TriggerEnum{ir.Error5XX, ir.ConnectFailure},
					},
					PerRetry: &ir.PerRetryPolicy{
						BackOff: &ir.BackOffPolicy{
							BaseInterval: &metav1.Duration{Duration: 1 * time.Second},
							MaxInterval:  &metav1.Duration{Duration: 5 * time.Second},
						},
					},
				},
				Timeout: &ir.Timeout{
					HTTP: &ir.HTTPTimeout{
						RequestTimeout: &metav1.Duration{Duration: 30 * time.Second},
					},
				},
			},
		},
	}

	got, err := dynamicModuleSource(&dm)
	require.NoError(t, err)
	require.NotNil(t, got)

	remote, ok := got.Specifier.(*corev3.AsyncDataSource_Remote)
	require.True(t, ok)

	require.Equal(t, 30*time.Second, remote.Remote.HttpUri.Timeout.AsDuration())

	require.NotNil(t, remote.Remote.RetryPolicy)
	require.Equal(t, uint32(3), remote.Remote.RetryPolicy.NumRetries.GetValue())
	require.Equal(t, 1*time.Second, remote.Remote.RetryPolicy.RetryBackOff.BaseInterval.AsDuration())
	require.Equal(t, 5*time.Second, remote.Remote.RetryPolicy.RetryBackOff.MaxInterval.AsDuration())
	require.NotEmpty(t, remote.Remote.RetryPolicy.RetryOn)
}

// TestDynamicModuleSourceTimeoutFallsBackToDefault verifies that when no
// timeout is configured, the AsyncDataSource still receives the package
// default timeout rather than zero.
func TestDynamicModuleSourceTimeoutFallsBackToDefault(t *testing.T) {
	dm := ir.DynamicModule{
		Remote: &ir.RemoteDynamicModuleSource{
			URL:    "https://modules.example.com/libremote_auth.so",
			SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
		},
	}
	got, err := dynamicModuleSource(&dm)
	require.NoError(t, err)
	remote, ok := got.Specifier.(*corev3.AsyncDataSource_Remote)
	require.True(t, ok)
	require.Equal(t, defaultExtServiceRequestTimeout, remote.Remote.HttpUri.Timeout.AsDuration())
}

// TestDynamicModuleSourceWithDestination verifies that when a remote dynamic
// module has a resolved Destination (from BackendRefs), the AsyncDataSource
// references the destination's cluster name rather than a URL-synthesized one.
func TestDynamicModuleSourceWithDestination(t *testing.T) {
	dm := ir.DynamicModule{
		Remote: &ir.RemoteDynamicModuleSource{
			URL:    "https://modules.example.com/libremote_auth.so",
			SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
			Destination: &ir.RouteDestination{
				Name: "envoyextensionpolicy/default/policy/dynamic-module/0",
				Settings: []*ir.DestinationSetting{
					{
						Name:        "envoyextensionpolicy/default/policy/dynamic-module/0/backend/0",
						Weight:      ptrTo(uint32(1)),
						AddressType: ptrTo(ir.FQDN),
						Protocol:    ir.HTTPS,
						Endpoints:   []*ir.DestinationEndpoint{{Host: "modules.example.com", Port: 443}},
					},
				},
			},
		},
	}

	got, err := dynamicModuleSource(&dm)
	require.NoError(t, err)
	remote, ok := got.Specifier.(*corev3.AsyncDataSource_Remote)
	require.True(t, ok)
	clusterSpec, ok := remote.Remote.HttpUri.HttpUpstreamType.(*corev3.HttpUri_Cluster)
	require.True(t, ok)
	require.Equal(t, "envoyextensionpolicy/default/policy/dynamic-module/0", clusterSpec.Cluster)
}

// TestDynamicModulePatchResourcesUsesDestinationCluster verifies that
// patchResources creates a cluster from the resolved Destination (carrying
// BackendTLSPolicy-derived TLS) when one is set, instead of synthesizing
// a default-trust cluster from the URL. The transport socket on the resulting
// cluster must reflect the destination's TLS config rather than the
// system-trust default that addClusterFromURL would have produced.
func TestDynamicModulePatchResourcesUsesDestinationCluster(t *testing.T) {
	tCtx := &types.ResourceVersionTable{XdsResources: types.XdsResources{}}
	dmFilter := &dynamicModule{}

	sni := "modules.example.com"
	tlsConfig := &ir.TLSUpstreamConfig{
		SNI: &sni,
		CACertificate: &ir.TLSCACertificate{
			Name:        "ca",
			Certificate: []byte("dummy-ca-pem"),
		},
	}

	routes := []*ir.HTTPRoute{
		{
			EnvoyExtensions: &ir.EnvoyExtensionFeatures{
				DynamicModules: []ir.DynamicModule{
					{
						Name:       "envoyextensionpolicy/default/policy/dynamic-module/0",
						FilterName: "test_filter",
						Remote: &ir.RemoteDynamicModuleSource{
							URL:    "https://modules.example.com/libremote_auth.so",
							SHA256: "abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789",
							Destination: &ir.RouteDestination{
								Name: "envoyextensionpolicy/default/policy/dynamic-module/0",
								Settings: []*ir.DestinationSetting{
									{
										Name:        "envoyextensionpolicy/default/policy/dynamic-module/0/backend/0",
										Weight:      ptrTo(uint32(1)),
										AddressType: ptrTo(ir.FQDN),
										Protocol:    ir.HTTPS,
										Endpoints:   []*ir.DestinationEndpoint{{Host: "modules.example.com", Port: 443}},
										TLS:         tlsConfig,
									},
								},
							},
						},
					},
				},
			},
		},
	}

	require.NoError(t, dmFilter.patchResources(tCtx, routes))

	cluster := findXdsCluster(tCtx, "envoyextensionpolicy/default/policy/dynamic-module/0")
	require.NotNil(t, cluster, "expected cluster derived from Destination, not from URL")
	require.NotNil(t, cluster.TransportSocket, "expected TLS transport socket from BackendTLSPolicy-derived destination")
}

func ptrTo[T any](v T) *T { return &v }
