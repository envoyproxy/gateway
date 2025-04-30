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
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	typev3 "github.com/envoyproxy/go-control-plane/envoy/type/v3"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"

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

func Test_setTLSValidationContext(t *testing.T) {
	caCertificate := &ir.TLSCACertificate{Name: "ca"}
	sdsSecretConfig := &tlsv3.SdsSecretConfig{Name: caCertificate.Name, SdsConfig: makeConfigSource()}
	san := "san"
	sanMatcher := &matcherv3.StringMatcher{MatchPattern: &matcherv3.StringMatcher_Exact{Exact: san}}

	tests := []struct {
		name string
		tls  *ir.TLSConfig
		want *tlsv3.CommonTlsContext
	}{
		{
			name: "CA only",
			tls: &ir.TLSConfig{
				CACertificate: caCertificate,
			},
			want: &tlsv3.CommonTlsContext{
				ValidationContextType: &tlsv3.CommonTlsContext_ValidationContextSdsSecretConfig{
					ValidationContextSdsSecretConfig: sdsSecretConfig,
				},
			},
		},
		{
			name: "Certificate SPKI",
			tls: &ir.TLSConfig{
				CACertificate:         caCertificate,
				VerifyCertificateSpki: []string{"NvqYIYSbgK2vCJpQhObf77vv+bQWtc5ek5RIOwPiC9A="},
			},
			want: &tlsv3.CommonTlsContext{
				ValidationContextType: &tlsv3.CommonTlsContext_CombinedValidationContext{
					CombinedValidationContext: &tlsv3.CommonTlsContext_CombinedCertificateValidationContext{
						DefaultValidationContext: &tlsv3.CertificateValidationContext{
							VerifyCertificateSpki: []string{"NvqYIYSbgK2vCJpQhObf77vv+bQWtc5ek5RIOwPiC9A="},
						},
						ValidationContextSdsSecretConfig: sdsSecretConfig,
					},
				},
			},
		},
		{
			name: "Certificate hash",
			tls: &ir.TLSConfig{
				CACertificate:         caCertificate,
				VerifyCertificateHash: []string{"df6ff72fe9116521268f6f2dd4966f51df479883fe7037b39f75916ac3049d1a"},
			},
			want: &tlsv3.CommonTlsContext{
				ValidationContextType: &tlsv3.CommonTlsContext_CombinedValidationContext{
					CombinedValidationContext: &tlsv3.CommonTlsContext_CombinedCertificateValidationContext{
						DefaultValidationContext: &tlsv3.CertificateValidationContext{
							VerifyCertificateHash: []string{"df6ff72fe9116521268f6f2dd4966f51df479883fe7037b39f75916ac3049d1a"},
						},
						ValidationContextSdsSecretConfig: sdsSecretConfig,
					},
				},
			},
		},
		{
			name: "SANs",
			tls: &ir.TLSConfig{
				CACertificate: caCertificate,
				MatchTypedSubjectAltNames: []*ir.StringMatch{
					{Name: "", Exact: &san},
					{Name: "EMAIL", Exact: &san},
					{Name: "DNS", Exact: &san},
					{Name: "URI", Exact: &san},
					{Name: "IP_ADDRESS", Exact: &san},
					{Name: "1.3.6.1.4.1.311.20.2.3", Exact: &san},
				},
			},
			want: &tlsv3.CommonTlsContext{
				ValidationContextType: &tlsv3.CommonTlsContext_CombinedValidationContext{
					CombinedValidationContext: &tlsv3.CommonTlsContext_CombinedCertificateValidationContext{
						DefaultValidationContext: &tlsv3.CertificateValidationContext{
							MatchTypedSubjectAltNames: []*tlsv3.SubjectAltNameMatcher{
								{SanType: tlsv3.SubjectAltNameMatcher_SAN_TYPE_UNSPECIFIED, Matcher: sanMatcher},
								{SanType: tlsv3.SubjectAltNameMatcher_EMAIL, Matcher: sanMatcher},
								{SanType: tlsv3.SubjectAltNameMatcher_DNS, Matcher: sanMatcher},
								{SanType: tlsv3.SubjectAltNameMatcher_URI, Matcher: sanMatcher},
								{SanType: tlsv3.SubjectAltNameMatcher_IP_ADDRESS, Matcher: sanMatcher},
								{SanType: tlsv3.SubjectAltNameMatcher_OTHER_NAME, Oid: "1.3.6.1.4.1.311.20.2.3", Matcher: sanMatcher},
							},
						},
						ValidationContextSdsSecretConfig: sdsSecretConfig,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &tlsv3.CommonTlsContext{}
			setTLSValidationContext(tt.tls, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Test_setTLSValidationContext() got = %v, want %v", got, tt.want)
			}
		})
	}
}
