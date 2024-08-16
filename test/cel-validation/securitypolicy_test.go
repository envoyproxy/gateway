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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
)

func TestSecurityPolicyTarget(t *testing.T) {
	ctx := context.Background()
	baseSP := egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.SecurityPolicySpec{},
	}

	sectionName := gwapiv1a2.SectionName("foo")

	cases := []struct {
		desc         string
		mutate       func(sp *egv1a1.SecurityPolicy)
		mutateStatus func(sp *egv1a1.SecurityPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
			desc: "no targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute",
			},
		},
		{
			desc: "targetRef unsupported group",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
			desc: "targetRef unsupported group and kind - targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value: \"object\": this policy can only have a targetRef.kind of Gateway/HTTPRoute/GRPCRoute",
			},
		},
		{
			desc: "targetRef unsupported group and kind - targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value: \"object\": this policy can only have a targetRefs[*].kind of Gateway/HTTPRoute/GRPCRoute",
			},
		},

		{
			desc: "sectionName disabled until supported - targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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

		// cors
		{
			desc: "cors alloworigin valid without port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://foo.bar.com", // valid
						},
					},
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
			desc: "cors alloworigin valid with port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://foo.bar.com:8080", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://*.foo.bar", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard and port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://*.foo.com:8080", // valid
						},
					},
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
			desc: "cors alloworigin valid with scheme and wildcard",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://*", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard only",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"*", // valid
						},
					},
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
			desc: "cors alloworigin valid with simple hostname without tld",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://localhost", // valid
						},
					},
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
			desc: "cors alloworigin with wildcard in the middle",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://foo.*.com", // invalid, wildcard must be at the beginning
						},
					},
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"https://foo.*.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|https?:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},
		{
			desc: "cors alloworigin invalid without scheme",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"foo.bar.com", // invalid, no scheme
						},
					},
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"foo.bar.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|https?:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},
		{
			desc: "cors alloworigin invalid with unsupported scheme",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"grpc://foo.bar.com", // invalid, unsupported scheme
						},
					},
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"grpc://foo.bar.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|https?:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},

		// ExtAuth
		{
			desc: "GRPC external auth service",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Port: ptr.To(gwapiv1.PortNumber(15001)),
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
			desc: "GRPC external auth service with backendRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Kind: ptr.To(gwapiv1a2.Kind("Service")),
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
			desc: "empty GRPC external auth service",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{},
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
			wantErrors: []string{" backendRef or backendRefs needs to be set"},
		},
		{
			desc: "HTTP external auth service",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "http-auth-service",
											Port: ptr.To(gwapiv1.PortNumber(15001)),
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
			desc: "HTTP external auth service with backendRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Kind: ptr.To(gwapiv1a2.Kind("Service")),
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
			desc: "no extAuth",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{},
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
				"spec.extAuth: Invalid value: \"object\": one of grpc or http must be specified",
			},
		},
		{
			desc: "with both extAuth services",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Name: "grpc-auth-service",
									Port: ptr.To(gwapiv1.PortNumber(80)),
								},
							},
						},
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Name: "http-auth-service",
									Port: ptr.To(gwapiv1.PortNumber(15001)),
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
				"spec.extAuth: Invalid value: \"object\": only one of grpc or http can be specified",
			},
		},
		{
			desc: "http extAuth service invalid Group",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Group: ptr.To(gwapiv1.Group("unsupported")),
											Name:  "http-auth-service",
											Port:  ptr.To(gwapiv1.PortNumber(15001)),
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
			wantErrors: []string{
				" BackendRefs only supports Core and gateway.envoyproxy.io group.",
			},
		},
		{
			desc: "http extAuth service invalid Kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Kind: ptr.To(gwapiv1.Kind("unsupported")),
									Name: "http-auth-service",
									Port: ptr.To(gwapiv1.PortNumber(15001)),
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
				"BackendRefs must be used, backendRef is not supported.",
				"Exactly one backendRef can be specified in backendRefs.",
			},
		},
		{
			desc: "http extAuth service backendRefs invalid Kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Kind: ptr.To(gwapiv1a2.Kind("unsupported")),
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
			wantErrors: []string{"BackendRefs only supports Service and Backend kind."},
		},
		{
			desc: "grpc extAuth service invalid Group",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Group: ptr.To(gwapiv1.Group("unsupported")),
											Name:  "http-auth-service",
											Port:  ptr.To(gwapiv1.PortNumber(15001)),
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
			wantErrors: []string{
				"BackendRefs only supports Core and gateway.envoyproxy.io group.",
			},
		},
		{
			desc: "grpc extAuth service invalid Kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Kind: ptr.To(gwapiv1.Kind("unsupported")),
									Name: "http-auth-service",
									Port: ptr.To(gwapiv1.PortNumber(15001)),
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
				"BackendRefs must be used, backendRef is not supported.",
				"Exactly one backendRef can be specified in backendRefs.",
			},
		},
		{
			desc: "grpc extAuth service backendRefs invalid Kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Kind: ptr.To(gwapiv1a2.Kind("unsupported")),
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
			wantErrors: []string{
				"spec.extAuth.grpc: Invalid value: \"object\": BackendRefs only supports Service and Backend kind.",
			},
		},

		// JWT
		{
			desc: "valid jwt",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
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
			desc: "jwt with claim to headers",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								ClaimToHeaders: []egv1a1.ClaimToHeader{
									{
										Claim:  "name",
										Header: "x-claim-name",
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
			desc: "jwt with recomputeRoute",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								RecomputeRoute: ptr.To(true),
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
			wantErrors: []string{"Invalid value: \"object\": no such key: claimToHeaders evaluating rule: claimToHeaders must be specified if recomputeRoute is enabled"},
		},
		{
			desc: "jwt with claim to headers and recomputeRoute",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								ClaimToHeaders: []egv1a1.ClaimToHeader{
									{
										Claim:  "name",
										Header: "x-claim-name",
									},
								},
								RecomputeRoute: ptr.To(true),
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
			desc: "target selectors without targetRefs or targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
			desc: "ext-auth-grpc-backend",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name:  "grpc-auth-backend",
											Kind:  ptr.To(gwapiv1a2.Kind("Backend")),
											Port:  ptr.To(gwapiv1.PortNumber(8080)),
											Group: ptr.To(gwapiv1a2.Group("gateway.envoyproxy.io")),
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "ext-auth-http-backend",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name:  "http-auth-backend",
											Kind:  ptr.To(gwapiv1a2.Kind("Backend")),
											Port:  ptr.To(gwapiv1.PortNumber(80)),
											Group: ptr.To(gwapiv1a2.Group("gateway.envoyproxy.io")),
										},
									},
								},
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
			sp := baseSP.DeepCopy()
			sp.Name = fmt.Sprintf("sp-%v", time.Now().UnixNano())

			if tc.mutate != nil {
				tc.mutate(sp)
			}
			err := c.Create(ctx, sp)

			if tc.mutateStatus != nil {
				tc.mutateStatus(sp)
				err = c.Status().Update(ctx, sp)
			}

			if (len(tc.wantErrors) != 0) != (err != nil) {
				t.Fatalf("Unexpected response while creating SecurityPolicy; got err=\n%v\n;want error=%v", err, tc.wantErrors)
			}

			var missingErrorStrings []string
			for _, wantError := range tc.wantErrors {
				if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(wantError)) {
					missingErrorStrings = append(missingErrorStrings, wantError)
				}
			}
			if len(missingErrorStrings) != 0 {
				t.Errorf("Unexpected response while creating SecurityPolicy; got err=\n%v\n;missing strings within error=%q", err, missingErrorStrings)
			}
		})
	}
}
