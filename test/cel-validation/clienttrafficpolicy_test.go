// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestClientTrafficPolicyTarget(t *testing.T) {
	ctx := context.Background()
	baseCTP := egv1a1.ClientTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "ctp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.ClientTrafficPolicySpec{},
	}

	cases := []struct {
		desc         string
		mutate       func(ctp *egv1a1.ClientTrafficPolicy)
		mutateStatus func(ctp *egv1a1.ClientTrafficPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid targetRef",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "no targetRef",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("foo"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway",
			},
		},
		{
			desc: "targetRef unsupported group",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("foo"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
			},
		},
		{
			desc: "targetRefs unsupported group and kind",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("foo"),
									Kind:  gwapiv1.Kind("bar"),
									Name:  gwapiv1.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].group of gateway.networking.k8s.io",
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].kind of Gateway",
			},
		},
		{
			desc: "targetRef unsupported group and kind",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("foo"),
								Kind:  gwapiv1.Kind("bar"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway",
			},
		},
		{
			desc: "tls minimal version greater than tls maximal version",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					TLS: &egv1a1.ClientTLSSettings{
						TLSSettings: egv1a1.TLSSettings{
							MinVersion: ptr.To(egv1a1.TLSv12),
							MaxVersion: ptr.To(egv1a1.TLSv11),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.tls: Invalid value: \"object\": minVersion must be smaller or equal to maxVersion",
			},
		},
		{
			desc: "tls maximal version lesser than default tls minimal version",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					TLS: &egv1a1.ClientTLSSettings{
						TLSSettings: egv1a1.TLSSettings{
							MaxVersion: ptr.To(egv1a1.TLSv11),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.tls: Invalid value: \"object\": minVersion must be smaller or equal to maxVersion",
			},
		},
		{
			desc: "clientIPDetection xForwardedFor and customHeader set",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					ClientIPDetection: &egv1a1.ClientIPDetectionSettings{
						XForwardedFor: &egv1a1.XForwardedForSettings{
							NumTrustedHops: ptr.To(uint32(1)),
						},
						CustomHeader: &egv1a1.CustomHeaderExtensionSettings{
							Name: "x-client-ip-address",
						},
					},
				}
			},
			wantErrors: []string{
				"spec.clientIPDetection: Invalid value: \"object\": customHeader cannot be used in conjunction with xForwardedFor",
			},
		},
		{
			desc: "clientIPDetection numTrustedHops and trustedCIDRs",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					ClientIPDetection: &egv1a1.ClientIPDetectionSettings{
						XForwardedFor: &egv1a1.XForwardedForSettings{
							NumTrustedHops: ptr.To(uint32(1)),
							TrustedCIDRs: []egv1a1.CIDR{
								"192.168.1.0/24",
								"10.0.0.0/16",
								"172.16.0.0/12",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.clientIPDetection.xForwardedFor: Invalid value: \"object\": only one of numTrustedHops or trustedCIDRs must be set",
			},
		},
		{
			desc: "clientIPDetection invalid trustedCIDRs",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					ClientIPDetection: &egv1a1.ClientIPDetectionSettings{
						XForwardedFor: &egv1a1.XForwardedForSettings{
							TrustedCIDRs: []egv1a1.CIDR{
								"192.0124.1.0/24",
								"10.0.0.0/1645",
								"17212.16.0.0/123",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.clientIPDetection.xForwardedFor.trustedCIDRs[0]: Invalid value: \"192.0124.1.0/24\": spec.clientIPDetection.xForwardedFor.trustedCIDRs[0] in body should match '((25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\/([0-9]+))|((([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])|([0-9a-fA-F]{1,4}:){1,4}:((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9]))\\/([0-9]+))'",
			},
		},
		{
			desc: "clientIPDetection valid trustedCIDRs",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					ClientIPDetection: &egv1a1.ClientIPDetectionSettings{
						XForwardedFor: &egv1a1.XForwardedForSettings{
							TrustedCIDRs: []egv1a1.CIDR{
								"192.168.1.0/24",
								"10.0.0.0/16",
								"172.16.0.0/12",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "http3 enabled and ALPN protocols not set with other TLS parameters set",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					HTTP3: &egv1a1.HTTP3Settings{},
					TLS: &egv1a1.ClientTLSSettings{
						TLSSettings: egv1a1.TLSSettings{
							Ciphers: []string{"[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]"},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "setting ciphers with minimum TLS version set to 1.3",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					TLS: &egv1a1.ClientTLSSettings{
						TLSSettings: egv1a1.TLSSettings{
							MinVersion: ptr.To(egv1a1.TLSv13),
							Ciphers:    []string{"[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]"},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.tls: Invalid value: \"object\": setting ciphers has no effect if the minimum possible TLS version is 1.3",
			},
		},
		{
			desc: "valid timeout",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				d := gwapiv1.Duration("300s")
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Timeout: &egv1a1.ClientTimeout{
						HTTP: &egv1a1.HTTPClientTimeout{
							RequestReceivedTimeout: &d,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid bufferLimit format",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Connection: &egv1a1.ClientConnection{
						BufferLimit: ptr.To(resource.MustParse("15m")),
					},
				}
			},
			wantErrors: []string{
				"spec.connection.bufferLimit: Invalid value: \"15m\": spec.connection.bufferLimit in body should match '^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$', <nil>: Invalid value: \"\"",
			},
		},
		{
			desc: "invalid Connection Limit Empty",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Connection: &egv1a1.ClientConnection{
						ConnectionLimit: &egv1a1.ConnectionLimit{},
					},
				}
			},
			wantErrors: []string{
				"spec.connection.connectionLimit.value: Invalid value: 0: spec.connection.connectionLimit.value in body should be greater than or equal to 1",
			},
		},
		{
			desc: "invalid Connection Limit < 1",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Connection: &egv1a1.ClientConnection{
						ConnectionLimit: &egv1a1.ConnectionLimit{
							Value: -1, // Value: 0 is covered by existence test, as 0 is the nil value.
						},
					},
				}
			},
			wantErrors: []string{
				"spec.connection.connectionLimit.value: Invalid value: -1: spec.connection.connectionLimit.value in body should be greater than or equal to 1",
			},
		},
		{
			desc: "invalid InitialStreamWindowSize format",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					HTTP2: &egv1a1.HTTP2Settings{
						InitialStreamWindowSize: ptr.To(resource.MustParse("15m")),
					},
				}
			},
			wantErrors: []string{
				"spec.http2.initialStreamWindowSize: Invalid value: \"15m\": spec.http2.initialStreamWindowSize in body should match '^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$'",
			},
		},
		{
			desc: "invalid InitialConnectionWindowSize format",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					HTTP2: &egv1a1.HTTP2Settings{
						InitialConnectionWindowSize: ptr.To(resource.MustParse("15m")),
					},
				}
			},
			wantErrors: []string{
				"spec.http2.initialConnectionWindowSize: Invalid value: \"15m\": spec.http2.initialConnectionWindowSize in body should match '^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$'",
			},
		},
		{
			desc: "invalid xffc setting",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Headers: &egv1a1.HeaderSettings{
						XForwardedClientCert: &egv1a1.XForwardedClientCert{
							Mode: ptr.To(egv1a1.XFCCForwardModeSanitize),
							CertDetailsToAdd: []egv1a1.XFCCCertData{
								egv1a1.XFCCCertDataChain,
							},
						},
					},
				}
			},
			wantErrors: []string{
				" spec.headers.xForwardedClientCert: Invalid value: \"object\": certDetailsToAdd can only be set when mode is AppendForward or SanitizeSet",
			},
		},
		{
			desc: "both targetref and targetrefs specified",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				d := gwapiv1.Duration("300s")
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("eg"),
								},
							},
						},
					},
					Timeout: &egv1a1.ClientTimeout{
						HTTP: &egv1a1.HTTPClientTimeout{
							RequestReceivedTimeout: &d,
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "valid timeout using targetrefs",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				d := gwapiv1.Duration("300s")
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("eg"),
								},
							},
						},
					},
					Timeout: &egv1a1.ClientTimeout{
						HTTP: &egv1a1.HTTPClientTimeout{
							RequestReceivedTimeout: &d,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "target selectors without targetRefs or targetRef",
			mutate: func(sp *egv1a1.ClientTrafficPolicy) {
				sp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: ptr.To(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid x-request-id header setting",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Name = "ctp-headers"
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
						},
					},
					Headers: &egv1a1.HeaderSettings{
						PreserveXRequestID: ptr.To(true),
						RequestID:          ptr.To(egv1a1.RequestIDActionGenerate),
					},
				}
			},
			wantErrors: []string{
				"ClientTrafficPolicy.gateway.envoyproxy.io \"ctp-headers\" is invalid: spec.headers: Invalid value: \"object\": preserveXRequestID and requestID cannot both be set.",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			ctp := baseCTP.DeepCopy()
			ctp.Name = fmt.Sprintf("ctp-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(ctp)
			}
			err := c.Create(ctx, ctp)

			if tc.mutateStatus != nil {
				tc.mutateStatus(ctp)
				err = c.Status().Update(ctx, ctp)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating ClientTrafficPolicy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating ClientTrafficPolicy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
