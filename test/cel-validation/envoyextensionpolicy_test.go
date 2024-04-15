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

	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestEnvoyExtensionPolicyTarget(t *testing.T) {
	ctx := context.Background()
	baseeep := egv1a1.EnvoyExtensionPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "eep",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.EnvoyExtensionPolicySpec{},
	}

	sectionName := gwapiv1a2.SectionName("foo")

	cases := []struct {
		desc         string
		mutate       func(eep *egv1a1.EnvoyExtensionPolicy)
		mutateStatus func(eep *egv1a1.EnvoyExtensionPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid gateway targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{}
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
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			desc: "sectionName disabled until supported",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
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
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy does not yet support the sectionName field",
			},
		},

		// ExtProc
		{
			desc: "ExtProc with BackendRef",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendRefs: []egv1a1.BackendRef{
								{
									BackendObjectReference: gwapiv1.BackendObjectReference{
										Name: "grpc-proc-service",
										Port: ptr.To(gwapiv1.PortNumber(80)),
									},
								},
							},
						},
					},
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "eg",
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ExtProc with invalid BackendRef Group",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendRefs: []egv1a1.BackendRef{
								{
									BackendObjectReference: gwapiv1.BackendObjectReference{
										Group: ptr.To(gwapiv1.Group("unsupported")),
										Name:  "grpc-proc-service",
										Port:  ptr.To(gwapiv1.PortNumber(80)),
									},
								},
							},
						},
					},
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "eg",
						},
					},
				}
			},
			wantErrors: []string{" spec.extProc[0].backendRefs: Invalid value: \"array\": BackendRefs only supports Core group."},
		},
		{
			desc: "ExtProc with invalid BackendRef Kind",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendRefs: []egv1a1.BackendRef{
								{
									BackendObjectReference: gwapiv1.BackendObjectReference{
										Kind: ptr.To(gwapiv1.Kind("unsupported")),
										Name: "grpc-proc-service",
										Port: ptr.To(gwapiv1.PortNumber(80)),
									},
								},
							},
						},
					},
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: "gateway.networking.k8s.io",
							Kind:  "Gateway",
							Name:  "eg",
						},
					},
				}
			},
			wantErrors: []string{"spec.extProc[0].backendRefs: Invalid value: \"array\": BackendRefs only supports Service kind."},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			eep := baseeep.DeepCopy()
			eep.Name = fmt.Sprintf("eep-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(eep)
			}
			err := c.Create(ctx, eep)

			if tc.mutateStatus != nil {
				tc.mutateStatus(eep)
				err = c.Status().Update(ctx, eep)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating EnvoyExtensionPolicy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating EnvoyExtensionPolicy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
