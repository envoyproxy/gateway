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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
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
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("HTTPRoute"),
								Name:  gwapiv1a2.ObjectName("httpbin-route"),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "both targetRef and targetRefs",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("foo"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1a2.Kind("foo"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "no targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "targetRef unsupported kind - targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("foo"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef unsupported kind - targetRefs",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1a2.Kind("foo"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef unsupported group - targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("foo"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
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
			desc: "targetRef unsupported group - targetRefs",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("foo"),
									Kind:  gwapiv1a2.Kind("Gateway"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].group of gateway.networking.k8s.io",
			},
		},
		{
			desc: "targetRef unsupported group and kind - targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("foo"),
								Kind:  gwapiv1a2.Kind("bar"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "targetRef unsupported group and kind - targetRefs",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("foo"),
									Kind:  gwapiv1a2.Kind("bar"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].group of gateway.networking.k8s.io",
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute/TCPRoute/UDPRoute/TLSRoute",
			},
		},
		{
			desc: "sectionName disabled until supported -targetRef",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
							SectionName: &sectionName,
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy does not yet support the sectionName field",
			},
		},
		{
			desc: "sectionName disabled until supported - targetRefs",
			mutate: func(eep *egv1a1.EnvoyExtensionPolicy) {
				eep.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1a2.Kind("Gateway"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
								SectionName: &sectionName,
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": this policy does not yet support the sectionName field",
			},
		},

		// ExtProc
		{
			desc: "ExtProc with BackendRef",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
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
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
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
							BackendCluster: egv1a1.BackendCluster{
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
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.extProc[0]: Invalid value: \"object\": BackendRefs only supports Core and gateway.envoyproxy.io group"},
		},
		{
			desc: "ExtProc with invalid BackendRef Kind",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
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
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{"spec.extProc[0]: Invalid value: \"object\": BackendRefs only supports Service and Backend kind"},
		},
		{
			desc: "ExtProc with invalid fields",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("not-a-body-mode")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("not-a-body-mode")),
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extProc[0].processingMode.response.body: Unsupported value: \"not-a-body-mode\": supported values: \"Streamed\", \"Buffered\", \"BufferedPartial\"",
				"spec.extProc[0].processingMode.request.body: Unsupported value: \"not-a-body-mode\": supported values: \"Streamed\", \"Buffered\", \"BufferedPartial\"",
			},
		},
		{
			desc: "valid ExtProc with request attributes and failOpen true",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Attributes: []string{"request.headers"},
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ExtProc with response attributes and failOpen true",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Response: &egv1a1.ProcessingModeOptions{
									Attributes: []string{"response.headers"},
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ExtProc with failOpen true",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ExtProc without FullDuplexStreamed body and failOpen true",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("Buffered")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("Buffered")),
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ExtProc with FullDuplexStreamed body and failOpen false",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
							},
							FailOpen: ptr.To(false),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ExtProc with FullDuplexStreamed body and failOpen nil",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid ExtProc with FullDuplexStreamed request body and failOpen",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("Buffered")),
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extProc[0]: Invalid value: \"object\": If FullDuplexStreamed body processing mode is used, FailOpen must be false.",
			},
		},
		{
			desc: "invalid ExtProc with FullDuplexStreamed response body and failOpen",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("Buffered")),
								},
								Response: &egv1a1.ProcessingModeOptions{
									Body: ptr.To(egv1a1.ExtProcBodyProcessingMode("FullDuplexStreamed")),
								},
							},
							FailOpen: ptr.To(true),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extProc[0]: Invalid value: \"object\": If FullDuplexStreamed body processing mode is used, FailOpen must be false.",
			},
		},
		{
			desc: "Valid Lua filter (inline)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type:   egv1a1.LuaValueTypeInline,
							Inline: ptr.To("function envoy_on_response(response_handle) -- Do something -- end"),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: nil,
		},
		{
			desc: "Valid Lua filter (source configmap)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type: egv1a1.LuaValueTypeValueRef,
							ValueRef: &gwapiv1a2.LocalObjectReference{
								Kind:  gwapiv1a2.Kind("ConfigMap"),
								Name:  gwapiv1a2.ObjectName("eg"),
								Group: gwapiv1a2.Group("v1"),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: nil,
		},
		{
			desc: "Invalid Lua filter (type inline but source configmap)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type: egv1a1.LuaValueTypeInline,
							ValueRef: &gwapiv1a2.LocalObjectReference{
								Kind:  gwapiv1a2.Kind("ConfigMap"),
								Name:  gwapiv1a2.ObjectName("eg"),
								Group: gwapiv1a2.Group("v1"),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.lua[0]: Invalid value: \"object\": Exactly one of inline or valueRef must be set with correct type.",
			},
		},
		{
			desc: "Invalid Lua filter (type configmap but source inline)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type:   egv1a1.LuaValueTypeValueRef,
							Inline: ptr.To("function envoy_on_response(response_handle) -- Do something -- end"),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.lua[0]: Invalid value: \"object\": Exactly one of inline or valueRef must be set with correct type.",
			},
		},
		{
			desc: "Invalid Lua filter (source object kind not configmap)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type: egv1a1.LuaValueTypeValueRef,
							ValueRef: &gwapiv1a2.LocalObjectReference{
								Kind:  gwapiv1a2.Kind("NotConfigMap"),
								Name:  gwapiv1a2.ObjectName("eg"),
								Group: gwapiv1a2.Group("v1"),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.lua[0].valueRef: Invalid value: \"object\": Only a reference to an object of kind ConfigMap belonging to default v1 API group is supported.",
			},
		},
		{
			desc: "Invalid Lua filter (source object group not default)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type: egv1a1.LuaValueTypeValueRef,
							ValueRef: &gwapiv1a2.LocalObjectReference{
								Kind:  gwapiv1a2.Kind("ConfigMap"),
								Name:  gwapiv1a2.ObjectName("eg"),
								Group: gwapiv1a2.Group(gwapiv1a2.GroupName),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.lua[0].valueRef: Invalid value: \"object\": Only a reference to an object of kind ConfigMap belonging to default v1 API group is supported.",
			},
		},
		{
			desc: "Invalid Lua filter (source both inline and configmap)",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					Lua: []egv1a1.Lua{
						{
							Type:   egv1a1.LuaValueTypeInline,
							Inline: ptr.To("function envoy_on_response(response_handle) -- Do something -- end"),
							ValueRef: &gwapiv1a2.LocalObjectReference{
								Kind:  gwapiv1a2.Kind("ConfigMap"),
								Name:  gwapiv1a2.ObjectName("eg"),
								Group: gwapiv1a2.Group("v1"),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.lua[0]: Invalid value: \"object\": Exactly one of inline or valueRef must be set with correct type.",
			},
		},
		{
			desc: "target selectors without targetRefs or targetRef",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: ptr.To(gwapiv1a2.Group("gateway.networking.k8s.io")),
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
			desc: "ExtProc with valid attributes",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Attributes: []string{
										"request.path",
										"request.url_path",
										"request.host",
										"request.scheme",
										"request.method",
										"request.headers",
										"request.referer",
										"request.useragent",
										"request.time",
										"request.id",
										"request.protocol",
										"request.query",
										"request.duration",
										"request.size",
										"request.total_size",
										"response.code",
										"response.code_details",
										"response.flags",
										"response.grpc_status",
										"response.headers",
										"response.trailers",
										"response.size",
										"response.total_size",
										"response.backend_latency",
										"source.address",
										"source.port",
										"destination.address",
										"destination.port",
									},
								},
								Response: &egv1a1.ProcessingModeOptions{
									Attributes: []string{
										"connection.id",
										"connection.mtls",
										"connection.requested_server_name",
										"connection.tls_version",
										"connection.subject_local_certificate",
										"connection.subject_peer_certificate",
										"connection.dns_san_local_certificate",
										"connection.dns_san_peer_certificate",
										"connection.uri_san_local_certificate",
										"connection.uri_san_peer_certificate",
										"connection.sha256_peer_certificate_digest",
										"connection.transport_failure_reason",
										"connection.termination_details",
										"upstream.address",
										"upstream.port",
										"upstream.tls_version",
										"upstream.subject_local_certificate",
										"upstream.subject_peer_certificate",
										"upstream.dns_san_local_certificate",
										"upstream.dns_san_peer_certificate",
										"upstream.uri_san_local_certificate",
										"upstream.uri_san_peer_certificate",
										"upstream.sha256_peer_certificate_digest",
										"upstream.local_address",
										"upstream.transport_failure_reason",
										"upstream.request_attempt_count",
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ExtProc with invalid attributes",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							ProcessingMode: &egv1a1.ExtProcProcessingMode{
								Request: &egv1a1.ProcessingModeOptions{
									Attributes: []string{
										"xds.node",
										"metadata",
										"filter_state",
										"upstream_filter_state",
									},
								},
								Response: &egv1a1.ProcessingModeOptions{
									Attributes: []string{
										"xds.node",
										"xds.cluster",
										"plugin_name",
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extProc[0].processingMode.request.attributes[0]: Invalid value: \"xds.node\": spec.extProc[0].processingMode.request.attributes[0] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.request.attributes[1]: Invalid value: \"metadata\": spec.extProc[0].processingMode.request.attributes[1] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.request.attributes[2]: Invalid value: \"filter_state\": spec.extProc[0].processingMode.request.attributes[2] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.request.attributes[3]: Invalid value: \"upstream_filter_state\": spec.extProc[0].processingMode.request.attributes[3] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.response.attributes[0]: Invalid value: \"xds.node\": spec.extProc[0].processingMode.response.attributes[0] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.response.attributes[1]: Invalid value: \"xds.cluster\": spec.extProc[0].processingMode.response.attributes[1] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
				"spec.extProc[0].processingMode.response.attributes[2]: Invalid value: \"plugin_name\": spec.extProc[0].processingMode.response.attributes[2] in body should match '^(connection\\.|source\\.|destination\\.|request\\.|response\\.|upstream\\.|xds\\.route_)[a-z_1-9]*$'",
			},
		},
		{
			desc: "ExtProc with invalid writableNamespaces",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Metadata: &egv1a1.ExtProcMetadata{
								WritableNamespaces: []string{
									"envoy.filters.http.rbac",
									"com.foocorp.custom",
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extProc[0].metadata.writableNamespaces: Invalid value: \"array\": writableNamespaces cannot contain well-known Envoy HTTP filter namespaces",
			},
		},
		{
			desc: "ExtProc with valid writableNamespaces",
			mutate: func(sp *egv1a1.EnvoyExtensionPolicy) {
				sp.Spec = egv1a1.EnvoyExtensionPolicySpec{
					ExtProc: []egv1a1.ExtProc{
						{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-proc-service",
											Port: ptr.To(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
							Metadata: &egv1a1.ExtProcMetadata{
								WritableNamespaces: []string{
									"envoy.foocrop.rbac",
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{},
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
