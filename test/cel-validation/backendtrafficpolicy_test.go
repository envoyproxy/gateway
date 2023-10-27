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
	"strings"
	"testing"
	"time"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestBackendTrafficPolicyTarget(t *testing.T) {
	ctx := context.Background()
	baseBTP := egv1a1.BackendTrafficPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "btp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.BackendTrafficPolicySpec{},
	}

	sectionName := gwapiv1a2.SectionName("foo")

	cases := []struct {
		desc         string
		mutate       func(btp *egv1a1.BackendTrafficPolicy)
		mutateStatus func(btp *egv1a1.BackendTrafficPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid gateway targetRef",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "valid httproute targetRef",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("HTTPRoute"),
							Name:  gwapiv1a2.ObjectName("httpbin-route"),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "no targetRef",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{}
			},
			wantErrors: []string{
				"spec.targetRef.kind: Invalid value: \"\": spec.targetRef.kind in body should be at least 1 chars long",
				"spec.targetRef.name: Invalid value: \"\": spec.targetRef.name in body should be at least 1 chars long",
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef unsupported group",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
				"spec.targetRef: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef section name with gateway",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
						SectionName: &sectionName,
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "targetRef section name with httproute",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("HTTPRoute"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
						SectionName: &sectionName,
					},
				}
			},
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": sectionName can only be set when kind is 'Gateway'.",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			btp := baseBTP.DeepCopy()
			btp.Name = fmt.Sprintf("btp-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(btp)
			}
			err := c.Create(ctx, btp)

			if tc.mutateStatus != nil {
				tc.mutateStatus(btp)
				err = c.Status().Update(ctx, btp)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating BackendTrafficPolicy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating BackendTrafficPolicy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
