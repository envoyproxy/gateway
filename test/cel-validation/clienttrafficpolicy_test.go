// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build celvalidation
// +build celvalidation

package celvalidation

import (
	"context"
	"fmt"
	"k8s.io/apimachinery/pkg/api/resource"
	"strings"
	"testing"
	"time"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
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

	sectionName := gwapiv1a2.SectionName("foo")

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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
				"spec.targetRef.kind: Invalid value: \"\": spec.targetRef.kind in body should be at least 1 chars long",
				"spec.targetRef.name: Invalid value: \"\": spec.targetRef.name in body should be at least 1 chars long",
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("foo"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway",
			},
		},
		{
			desc: "targetRef unsupported group",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("foo"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
			},
		},
		{
			desc: "targetRef unsupported group and kind",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("foo"),
							Kind:  gwapiv1a2.Kind("bar"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway",
			},
		},
		{
			desc: "sectionName disabled until supported",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
						},
						SectionName: &sectionName,
					},
				}
			},
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy does not yet support the sectionName field",
			},
		},
		{
			desc: "tls minimal version greater than tls maximal version",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					TLS: &egv1a1.TLSSettings{
						MinVersion: ptr.To(egv1a1.TLSv12),
						MaxVersion: ptr.To(egv1a1.TLSv11),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					TLS: &egv1a1.TLSSettings{
						MaxVersion: ptr.To(egv1a1.TLSv11),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
			desc: "http3 enabled and ALPN protocols not set with other TLS parameters set",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HTTP3: &egv1a1.HTTP3Settings{},
					TLS: &egv1a1.TLSSettings{
						Ciphers: []string{"[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]"},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "setting ciphers with minimum TLS version set to 1.3",
			mutate: func(ctp *egv1a1.ClientTrafficPolicy) {
				ctp.Spec = egv1a1.ClientTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					TLS: &egv1a1.TLSSettings{
						MinVersion: ptr.To(egv1a1.TLSv13),
						Ciphers:    []string{"[ECDHE-ECDSA-AES128-GCM-SHA256|ECDHE-ECDSA-CHACHA20-POLY1305]"},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					Connection: &egv1a1.Connection{
						BufferLimit: ptr.To(resource.MustParse("15m")),
					},
				}
			},
			wantErrors: []string{
				"spec.connection.bufferLimit: Invalid value: \"\": bufferLimit must be of the format \"^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$\"",
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
