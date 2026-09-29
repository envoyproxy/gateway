// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"reflect"
	"testing"

	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/extension/registry"
	"github.com/envoyproxy/gateway/internal/ir"
)

func Test_toNetworkFilter(t *testing.T) {
	tests := []struct {
		name    string
		proto   proto.Message
		wantErr error
	}{
		{
			name: "valid filter",
			proto: &hcmv3.HttpConnectionManager{
				StatPrefix: "stats",
				RouteSpecifier: &hcmv3.HttpConnectionManager_RouteConfig{
					RouteConfig: &routev3.RouteConfiguration{
						Name: "route",
					},
				},
			},
			wantErr: nil,
		},
		{
			name:    "invalid proto msg",
			proto:   &hcmv3.HttpConnectionManager{},
			wantErr: errors.New("invalid HttpConnectionManager.StatPrefix: value length must be at least 1 runes; invalid HttpConnectionManager.RouteSpecifier: value is required"),
		},
		{
			name:    "nil proto msg",
			proto:   nil,
			wantErr: errors.New("empty message received"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toNetworkFilter("name", tt.proto)
			if tt.wantErr != nil {
				assert.Containsf(t, err.Error(), tt.wantErr.Error(), "toNetworkFilter(%v)", tt.proto)
			} else {
				assert.NoErrorf(t, err, "toNetworkFilter(%v)", tt.proto)
			}
		})
	}
}

func Test_buildTCPProxyHashPolicy(t *testing.T) {
	tests := []struct {
		name string
		lb   *ir.LoadBalancer
		want []*typev3.HashPolicy
	}{
		{
			name: "Nil LoadBalancer",
			lb:   nil,
			want: nil,
		},
		{
			name: "Nil ConsistentHash in LoadBalancer",
			lb:   &ir.LoadBalancer{},
			want: nil,
		},
		{
			name: "ConsistentHash without hash policy",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{}},
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to false",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: new(bool)}}, // *new(bool) defaults to false
			want: nil,
		},
		{
			name: "ConsistentHash with SourceIP set to true",
			lb:   &ir.LoadBalancer{ConsistentHash: &ir.ConsistentHash{SourceIP: func(b bool) *bool { return &b }(true)}},
			want: []*typev3.HashPolicy{{PolicySpecifier: &typev3.HashPolicy_SourceIp_{SourceIp: &typev3.HashPolicy_SourceIp{}}}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildTCPProxyHashPolicy(tt.lb)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("buildTCPProxyHashPolicy() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func newExtensionCertificate(name string) ir.TLSCertificate {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("cert.example.io/v1alpha1")
	obj.SetKind("ExampleCertificate")
	obj.SetNamespace("default")
	obj.SetName(name)
	return ir.TLSCertificate{Name: name, ExtensionRef: &ir.UnstructuredRef{Object: obj}}
}

func newSecretCertificate(name string) ir.TLSCertificate {
	return ir.TLSCertificate{Name: name, Certificate: []byte("cert"), PrivateKey: []byte("key")}
}

func Test_allExtensionCertificatesUnresolved(t *testing.T) {
	resolved := map[string]*tlsv3.SdsSecretConfig{"ext": {Name: "resolved"}}
	tests := []struct {
		name        string
		tlsConfig   *ir.TLSConfig
		resolutions map[string]*tlsv3.SdsSecretConfig
		want        bool
	}{
		{
			name:      "no TLS",
			tlsConfig: nil,
			want:      false,
		},
		{
			name:      "no certificates",
			tlsConfig: &ir.TLSConfig{},
			want:      false,
		},
		{
			name:      "secret certificate only",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newSecretCertificate("secret")}},
			want:      false,
		},
		{
			name: "secret certificate next to an unresolved extension certificate",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{
				newSecretCertificate("secret"), newExtensionCertificate("ext"),
			}},
			want: false,
		},
		{
			name:        "extension certificate resolved",
			tlsConfig:   &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("ext")}},
			resolutions: resolved,
			want:        false,
		},
		{
			name: "one of two extension certificates resolved",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{
				newExtensionCertificate("ext"), newExtensionCertificate("other"),
			}},
			resolutions: resolved,
			want:        false,
		},
		{
			name:        "extension certificate resolved to an empty name",
			tlsConfig:   &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("ext")}},
			resolutions: map[string]*tlsv3.SdsSecretConfig{"ext": {}},
			want:        true,
		},
		{
			name:      "extension certificate unresolved",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("ext")}},
			want:      true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, allExtensionCertificatesUnresolved(tt.tlsConfig, tt.resolutions))
		})
	}
}

func Test_buildXdsTLSCertSecret(t *testing.T) {
	tests := []struct {
		name       string
		cert       ir.TLSCertificate
		wantSecret bool
	}{
		{
			name:       "secret certificate",
			cert:       newSecretCertificate("secret"),
			wantSecret: true,
		},
		{
			name:       "extension certificate carries no key material",
			cert:       newExtensionCertificate("ext"),
			wantSecret: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantSecret, buildXdsTLSCertSecret(&tt.cert) != nil)
		})
	}
}

// The success and refusal paths are covered by the extension-xds-ir golden tests. These cover the
// failures, which fail translation and so have no golden output.
func Test_resolveExtensionCertificates(t *testing.T) {
	withTLSCertificateHook := buildExtensionManagerConfig(false)
	failOpen := buildExtensionManagerConfig(true)
	withoutTLSCertificateHook := buildExtensionManagerConfig(false)
	withoutTLSCertificateHook.Hooks.XDSTranslator.Post = []egv1a1.XDSTranslatorHook{egv1a1.XDSRoute}

	tests := []struct {
		name      string
		extension *egv1a1.ExtensionManager
		tlsConfig *ir.TLSConfig
		wantErr   string
	}{
		{
			name:      "no TLS needs no extension",
			tlsConfig: nil,
		},
		{
			name:      "secret certificate needs no extension",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newSecretCertificate("secret")}},
		},
		{
			name:      "extension certificate without an extension manager",
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("ext")}},
			wantErr:   "requires an extension server but none is configured",
		},
		{
			name:      "extension certificate without the TLSCertificate hook",
			extension: &withoutTLSCertificateHook,
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("ext")}},
			wantErr:   "requires the TLSCertificate hook but no extension registered it",
		},
		{
			name:      "extension server error",
			extension: &withTLSCertificateHook,
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("error-cert")}},
			wantErr:   "certificate resolve error",
		},
		{
			// With failOpen the error is logged and the certificate is left unresolved.
			name:      "extension server error with failOpen",
			extension: &failOpen,
			tlsConfig: &ir.TLSConfig{Certificates: []ir.TLSCertificate{newExtensionCertificate("error-cert")}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &Translator{}
			if tt.extension != nil {
				extMgr, closeFunc, err := registry.NewInMemoryManager(tt.extension, &testingExtensionServer{})
				require.NoError(t, err)
				defer closeFunc()
				tr.ExtensionManager = &extMgr
			}

			resolutions, err := tr.resolveExtensionCertificates(tt.tlsConfig, nil)
			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Empty(t, resolutions)
		})
	}
}
