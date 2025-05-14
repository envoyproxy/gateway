// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"reflect"
	"testing"

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestSetTLSClientValidationContext(t *testing.T) {
	sanMatcher := egv1a1.StringMatch{Type: ptr.To(egv1a1.StringMatchExact), Value: "san"}

	tests := []struct {
		name string
		tls  *egv1a1.ClientValidationContext
		want *ir.TLSConfig
	}{
		{
			name: "Certificate SPKI",
			tls: &egv1a1.ClientValidationContext{
				PublicKeyPins: []string{"NvqYIYSbgK2vCJpQhObf77vv+bQWtc5ek5RIOwPiC9A="},
			},
			want: &ir.TLSConfig{
				VerifyCertificateSpki: []string{"NvqYIYSbgK2vCJpQhObf77vv+bQWtc5ek5RIOwPiC9A="},
			},
		},
		{
			name: "Certificate hash",
			tls: &egv1a1.ClientValidationContext{
				CertificateHashes: []string{"df6ff72fe9116521268f6f2dd4966f51df479883fe7037b39f75916ac3049d1a"},
			},
			want: &ir.TLSConfig{
				VerifyCertificateHash: []string{"df6ff72fe9116521268f6f2dd4966f51df479883fe7037b39f75916ac3049d1a"},
			},
		},
		{
			name: "SANs",
			tls: &egv1a1.ClientValidationContext{
				SubjectAltNames: &egv1a1.SubjectAltNames{
					DNSNames:       []egv1a1.StringMatch{sanMatcher},
					EmailAddresses: []egv1a1.StringMatch{sanMatcher},
					IPAddresses:    []egv1a1.StringMatch{sanMatcher},
					URIs:           []egv1a1.StringMatch{sanMatcher},
					OtherNames:     []egv1a1.OtherNameMatch{{Oid: "1.3.6.1.4.1.311.20.2.3", Match: sanMatcher}},
				},
			},
			want: &ir.TLSConfig{
				MatchTypedSubjectAltNames: []*ir.StringMatch{
					{Name: "DNS", Exact: &sanMatcher.Value},
					{Name: "EMAIL", Exact: &sanMatcher.Value},
					{Name: "IP_ADDRESS", Exact: &sanMatcher.Value},
					{Name: "URI", Exact: &sanMatcher.Value},
					{Name: "1.3.6.1.4.1.311.20.2.3", Exact: &sanMatcher.Value},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := &ir.TLSConfig{}
			setTLSClientValidationContext(tt.tls, got)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("setTLSClientValidationContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
