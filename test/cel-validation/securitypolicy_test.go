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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

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
			desc: "valid ListenerSet targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("ListenerSet"),
								Name:  gwapiv1.ObjectName("xls"),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid ListenerSet targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("ListenerSet"),
									Name:  gwapiv1.ObjectName("xls"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid mergeType with xRoute targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("HTTPRoute"),
								Name:  gwapiv1.ObjectName("backend"),
							},
						},
					},
					MergeType: new(egv1a1.StrategicMerge),
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid mergeType with xRoute targetSelector",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Kind:        gwapiv1.Kind("HTTPRoute"),
								MatchLabels: map[string]string{"app": "foo"},
							},
						},
					},
					MergeType: new(egv1a1.StrategicMerge),
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "mergeType rejected on ListenerSet targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("ListenerSet"),
								Name:  gwapiv1.ObjectName("xls"),
							},
						},
					},
					MergeType: new(egv1a1.StrategicMerge),
				}
			},
			wantErrors: []string{"mergeType can only be used with xRoute targets"},
		},
		{
			desc: "mergeType rejected on ListenerSet targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("ListenerSet"),
									Name:  gwapiv1.ObjectName("xls"),
								},
							},
						},
					},
					MergeType: new(egv1a1.StrategicMerge),
				}
			},
			wantErrors: []string{"mergeType can only be used with xRoute targets"},
		},
		{
			desc: "mergeType rejected on ListenerSet targetSelector",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Kind:        gwapiv1.Kind("ListenerSet"),
								MatchLabels: map[string]string{"app": "foo"},
							},
						},
					},
					MergeType: new(egv1a1.StrategicMerge),
				}
			},
			wantErrors: []string{"mergeType can only be used with xRoute targets"},
		},
		{
			desc: "no targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{}
			},
			wantErrors: []string{
				"spec: Invalid value:",
				": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value:",
				": this policy can only have a targetRef.kind of Gateway/ListenerSet/HTTPRoute/GRPCRoute/TCPRoute",
			},
		},
		{
			desc: "targetRef unsupported group",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value:",
				": this policy can only have a targetRef.group of gateway.networking.k8s.io",
			},
		},
		{
			desc: "targetRef unsupported group and kind - targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value:",
				": this policy can only have a targetRef.group of gateway.networking.k8s.io",
				": this policy can only have a targetRef.kind of Gateway/ListenerSet/HTTPRoute/GRPCRoute/TCPRoute",
			},
		},
		{
			desc: "targetRef unsupported group and kind - targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
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
				"spec: Invalid value:",
				": this policy can only have a targetRefs[*].group of gateway.networking.k8s.io",
				": this policy can only have a targetRefs[*].kind of Gateway/ListenerSet/HTTPRoute/GRPCRoute/TCPRoute",
			},
		},

		{
			desc: "sectionName supported for kind Gateway - targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("Gateway"),
								Name:  gwapiv1.ObjectName("eg"),
							},
							SectionName: &sectionName,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "sectionName supported for kind Gateway - targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("Gateway"),
									Name:  gwapiv1.ObjectName("eg"),
								},
								SectionName: &sectionName,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "sectionName supported for kind xRoute - targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: gwapiv1.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1.Kind("HTTPRoute"),
								Name:  gwapiv1.ObjectName("backend"),
							},
							SectionName: &sectionName,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "sectionName supported for kind xRoute - targetRefs",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
									Group: gwapiv1.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1.Kind("HTTPRoute"),
									Name:  gwapiv1.ObjectName("backend"),
								},
								SectionName: &sectionName,
							},
						},
					},
				}
			},
			wantErrors: []string{},
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
			desc: "cors alloworigin valid with port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://foo.bar.com:8080", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://*.foo.bar", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard and port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://*.foo.com:8080", // valid
						},
					},
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
			desc: "cors alloworigin valid with scheme and wildcard",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://*", // valid
						},
					},
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
			desc: "cors alloworigin valid with wildcard only",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"*", // valid
						},
					},
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
			desc: "cors alloworigin valid with simple hostname without tld",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"http://localhost", // valid
						},
					},
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
			desc: "cors alloworigin with wildcard in the middle",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"https://foo.*.com", // invalid, wildcard must be at the beginning
						},
					},
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"https://foo.*.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|[A-Za-z][A-Za-z0-9+.-]*:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"foo.bar.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|[A-Za-z][A-Za-z0-9+.-]*:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},
		{
			desc: "cors alloworigin valid with browser extension scheme",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"moz-extension://example.com", // valid, scheme may contain hyphens per RFC 3986
						},
					},
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
			desc: "cors alloworigin valid with custom scheme, wildcard and port",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"foo://*.example.com:8080", // valid
						},
					},
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
			desc: "cors alloworigin invalid with scheme not starting with a letter",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CORS: &egv1a1.CORS{
						AllowOrigins: []egv1a1.Origin{
							"1http://foo.bar.com", // invalid, RFC 3986 scheme must start with a letter
						},
					},
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
			wantErrors: []string{
				"spec.cors.allowOrigins[0]: Invalid value: \"1http://foo.bar.com\": spec.cors.allowOrigins[0] in body should match '^(\\*|[A-Za-z][A-Za-z0-9+.-]*:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},

		// csrf
		{
			desc: "csrf additionalOrigins valid",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CSRF: &egv1a1.CSRF{
						AdditionalOrigins: []egv1a1.Origin{
							"https://www.example.com",
							"http://www.example.com:8080",
							"https://*.trusted.com",
							"*",
						},
					},
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
			desc: "csrf additionalOrigins invalid without scheme",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CSRF: &egv1a1.CSRF{
						// invalid, an origin is scheme://host even though the CSRF filter
						// only ever matches on the host and port.
						AdditionalOrigins: []egv1a1.Origin{"www.example.com"},
					},
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
			wantErrors: []string{
				"spec.csrf.additionalOrigins[0]: Invalid value: \"www.example.com\": spec.csrf.additionalOrigins[0] in body should match '^(\\*|[A-Za-z][A-Za-z0-9+.-]*:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
			},
		},
		{
			desc: "csrf additionalOrigins invalid with path",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					CSRF: &egv1a1.CSRF{
						// invalid, the Origin header never carries a path.
						AdditionalOrigins: []egv1a1.Origin{"https://www.example.com/app"},
					},
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
			wantErrors: []string{
				"spec.csrf.additionalOrigins[0]: Invalid value: \"https://www.example.com/app\": spec.csrf.additionalOrigins[0] in body should match '^(\\*|[A-Za-z][A-Za-z0-9+.-]*:\\/\\/(\\*|(\\*\\.)?(([\\w-]+\\.?)+)?[\\w-]+)(:\\d{1,5})?)$'",
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
											Port: new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
											Kind: new(gwapiv1.Kind("Service")),
											Port: new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "GRPC external auth service with backendRefs to ServiceImport",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Group: new(gwapiv1.Group("multicluster.x-k8s.io")),
											Name:  "grpc-auth-service",
											Kind:  new(gwapiv1.Kind("ServiceImport")),
											Port:  new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
											Port: new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "HTTP external auth service with both path and pathOverride",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "http-auth-service",
											Port: new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
							Path:         new("/auth"),
							PathOverride: new("/check"),
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				" only one of path or pathOverride can be specified",
			},
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
											Kind: new(gwapiv1.Kind("Service")),
											Port: new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "HTTP external auth service with backendRefs to ServiceImport",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Group: new(gwapiv1.Group("multicluster.x-k8s.io")),
											Name:  "grpc-auth-service",
											Kind:  new(gwapiv1.Kind("ServiceImport")),
											Port:  new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "empty HTTP external auth service",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "no extAuth",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extAuth: Invalid value:",
				": one of grpc or http must be specified",
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
									Port: new(gwapiv1.PortNumber(80)),
								},
							},
						},
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRef: &gwapiv1.BackendObjectReference{
									Name: "http-auth-service",
									Port: new(gwapiv1.PortNumber(15001)),
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extAuth: Invalid value:",
				": only one of grpc or http can be specified",
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
											Group: new(gwapiv1.Group("unsupported")),
											Name:  "http-auth-service",
											Port:  new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				" BackendRefs only supports Core, multicluster.x-k8s.io, and gateway.envoyproxy.io groups.",
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
											Kind: new(gwapiv1.Kind("unsupported")),
											Port: new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{"BackendRefs only supports Service, ServiceImport, and Backend kind."},
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
											Group: new(gwapiv1.Group("unsupported")),
											Name:  "http-auth-service",
											Port:  new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"BackendRefs only supports Core, multicluster.x-k8s.io, and gateway.envoyproxy.io groups.",
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
											Kind: new(gwapiv1.Kind("unsupported")),
											Port: new(gwapiv1.PortNumber(80)),
										},
									},
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.extAuth.grpc: Invalid value:",
				": BackendRefs only supports Service, ServiceImport, and Backend kind.",
			},
		},
		{
			desc: "GRPC external auth service with timeout",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						GRPC: &egv1a1.GRPCExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "grpc-auth-service",
											Port: new(gwapiv1.PortNumber(15001)),
										},
									},
								},
							},
						},
						Timeout: new(gwapiv1.Duration("50s")),
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "HTTP external auth service with timeout",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					ExtAuth: &egv1a1.ExtAuth{
						HTTP: &egv1a1.HTTPExtAuthService{
							BackendCluster: egv1a1.BackendCluster{
								BackendRefs: []egv1a1.BackendRef{
									{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Name: "http-auth-service",
											Port: new(gwapiv1.PortNumber(8080)),
										},
									},
								},
							},
							Path: new("/auth"),
						},
						Timeout: new(gwapiv1.Duration("2s")),
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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

		// JWT
		{
			desc: "valid jwt",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: &egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
								RemoteJWKS: &egv1a1.RemoteJWKS{
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
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
								RemoteJWKS: &egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								RecomputeRoute: new(true),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"Invalid value:",
				": no such key: claimToHeaders evaluating rule: claimToHeaders must be specified if recomputeRoute is enabled",
			},
		},
		{
			desc: "jwt with claim to headers and recomputeRoute",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: &egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								ClaimToHeaders: []egv1a1.ClaimToHeader{
									{
										Claim:  "name",
										Header: "x-claim-name",
									},
								},
								RecomputeRoute: new(true),
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
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
			desc: "jwt with both remote and local JWKS",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: &egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
								LocalJWKS: &egv1a1.LocalJWKS{
									Inline: new(`{
  "keys": [
    {
      "kid": "1234567890",
      "kty": "RSA",
      "alg": "RS256",
      "n": "n",
      "e": "e"
    }
  ]
}
									`),
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"remoteJWKS and localJWKS cannot both be specified.",
			},
		},
		{
			desc: "jwt without remote or local JWKS",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"either remoteJWKS or localJWKS must be specified.",
			},
		},
		{
			desc: "jwt with both optional and failOpen",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Optional: new(true),
						FailOpen: new(true),
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								RemoteJWKS: &egv1a1.RemoteJWKS{
									URI: "https://example.com/jwt/jwks.json",
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"optional and failOpen cannot both be set; failOpen already tolerates a missing JWT",
			},
		},
		{
			desc: "valueRef type of localJWKS without valueRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					JWT: &egv1a1.JWT{
						Providers: []egv1a1.JWTProvider{
							{
								Name: "example",
								LocalJWKS: &egv1a1.LocalJWKS{
									Type: new(egv1a1.LocalJWKSTypeValueRef),
									Inline: new(`{
  "keys": [
    {
      "kid": "1234567890",
      "kty": "RSA",
      "alg": "RS256",
      "n": "n",
      "e": "e"
    }
  ]
}
									`),
								},
							},
						},
					},
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
								Group: "gateway.networking.k8s.io",
								Kind:  "Gateway",
								Name:  "eg",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"Exactly one of inline or valueRef must be set with correct type.",
			},
		},
		{
			desc: "target selectors without targetRefs or targetRef",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
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
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
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
											Kind:  new(gwapiv1.Kind("Backend")),
											Port:  new(gwapiv1.PortNumber(8080)),
											Group: new(gwapiv1.Group("gateway.envoyproxy.io")),
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
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
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
											Kind:  new(gwapiv1.Kind("Backend")),
											Port:  new(gwapiv1.PortNumber(80)),
											Group: new(gwapiv1.Group("gateway.envoyproxy.io")),
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
			desc: "authorization-missing-principal-and-cel",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
							},
						},
					},
				}
			},
			wantErrors: []string{"at least one of principal or cel must be specified"},
		},
		{
			desc: "authorization-empty-principal",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action:    egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{},
							},
						},
					},
				}
			},
			wantErrors: []string{"at least one of clientCIDRs, jwt, headers, clientIPGeoLocations, or clientCert must be specified"},
		},
		{
			desc: "authorization-cel-only",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								CEL:    new(egv1a1.CELExpression("request.path.startsWith('/admin')")),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "authorization-client-ip-geo-locations",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientIPGeoLocations: []egv1a1.ClientIPGeoLocation{
										{Country: new("US")},
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
			desc: "authorization-jwt-claims-without-jwt-authn",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									JWT: &egv1a1.JWTPrincipal{
										Claims: []egv1a1.JWTClaim{
											{
												Name:   "iss",
												Values: []string{"https://example.com"},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"if authorization.rules.principal.jwt is used, jwt must be defined"},
		},
		{
			desc: "authorization-jwt-empty-principal",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									JWT: &egv1a1.JWTPrincipal{},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"at least one of claims or scopes must be specified"},
		},
		{
			desc: "oidc-retry",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							BackendCluster: egv1a1.BackendCluster{
								BackendSettings: &egv1a1.ClusterSettings{
									Retry: &egv1a1.Retry{
										NumRetries: new(int32(3)),
										PerRetry: &egv1a1.PerRetryPolicy{
											BackOff: &egv1a1.BackOffPolicy{
												BaseInterval: new(gwapiv1.Duration("1s")),
												MaxInterval:  new(gwapiv1.Duration("10s")),
											},
										},
										RetryOn: &egv1a1.RetryOn{
											Triggers: []egv1a1.TriggerEnum{
												egv1a1.Error5XX, egv1a1.GatewayError, egv1a1.Reset,
											},
										},
									},
								},
							},
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-retry-unsupported-parameters",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							BackendCluster: egv1a1.BackendCluster{
								BackendSettings: &egv1a1.ClusterSettings{
									Retry: &egv1a1.Retry{
										NumRetries: new(int32(3)),
										PerRetry: &egv1a1.PerRetryPolicy{
											Timeout: new(gwapiv1.Duration("10s")),
										},
										RetryOn: &egv1a1.RetryOn{
											HTTPStatusCodes: []egv1a1.HTTPStatus{500},
										},
									},
								},
							},
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
					},
				}
			},
			wantErrors: []string{"Retry timeout is not supported", "HTTPStatusCodes is not supported"},
		},
		{
			desc: "oidc-issuer-http-scheme",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = securityPolicySpecWithOIDCIssuer("http://keycloak.gateway-conformance-infra/realms/master")
			},
			wantErrors: []string{"should match '^https://"},
		},
		{
			desc: "oidc-issuer-with-query",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = securityPolicySpecWithOIDCIssuer("https://keycloak.gateway-conformance-infra/realms/master?foo=bar")
			},
			wantErrors: []string{"should match '^https://"},
		},
		{
			desc: "oidc-issuer-with-fragment",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = securityPolicySpecWithOIDCIssuer("https://keycloak.gateway-conformance-infra/realms/master#foo")
			},
			wantErrors: []string{"should match '^https://"},
		},
		{
			desc: "oidc-issuer-with-userinfo",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = securityPolicySpecWithOIDCIssuer("https://user@keycloak.gateway-conformance-infra/realms/master")
			},
			wantErrors: []string{"should match '^https://"},
		},
		{
			desc: "oidc-without-clientid",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
					},
				}
			},
			wantErrors: []string{"only one of clientID or clientIDRef must be set"},
		},
		{
			desc: "oidc-two-clientids",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientIDRef: &gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
					},
				}
			},
			wantErrors: []string{"only one of clientID or clientIDRef must be set"},
		},
		{
			desc: "oidc-cookie-domain-single-character-first-label",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
						CookieDomain: new("m.example.com"),
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-cookie-domain-single-character-label",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
						CookieDomain: new("example.m.com"),
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-cookie-domain-invalid-leading-hyphen",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID: new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{
							Name: "secret",
						},
						CookieDomain: new("-example.m.com"),
					},
				}
			},
			wantErrors: []string{"spec.oidc.cookieDomain", "should match"},
		},
		{
			desc: "oidc-forward-id-token-custom-header",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "X-ID-Token",
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-forward-id-token-token-chars-header",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						// Underscore is a valid RFC 7230 token character; it must be
						// accepted by both the CRD and the translation-time validation.
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "X_Id_Token",
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-forward-id-token-authorization",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "Authorization",
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "oidc-forward-id-token-pseudo-header",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: ":path",
						},
					},
				}
			},
			wantErrors: []string{"should match"},
		},
		{
			desc: "oidc-forward-id-token-control-char-header",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "X-Id\nToken",
						},
					},
				}
			},
			wantErrors: []string{"should match"},
		},
		{
			desc: "oidc-forward-id-token-host-header",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:     new("client-id"),
						ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "host",
						},
					},
				}
			},
			wantErrors: []string{"header cannot be the Host header"},
		},
		{
			desc: "oidc-forward-id-token-and-access-token-on-authorization",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group:       new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:        "HTTPRoute",
								MatchLabels: map[string]string{"eg/namespace": "reference-apps"},
							},
						},
					},
					OIDC: &egv1a1.OIDC{
						Provider: egv1a1.OIDCProvider{
							Issuer:                "https://accounts.google.com",
							AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
							TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
						},
						ClientID:           new("client-id"),
						ClientSecret:       gwapiv1b1.SecretObjectReference{Name: "secret"},
						ForwardAccessToken: new(true),
						ForwardIDToken: &egv1a1.OIDCTokenForwarding{
							Header: "authorization",
						},
					},
				}
			},
			wantErrors: []string{"forwardAccessToken cannot be true when forwardIDToken.header is Authorization"},
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

func TestSecurityPolicyAPIKeyAuthExtractFrom(t *testing.T) {
	ctx := context.Background()
	baseSP := egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.SecurityPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				},
			},
			APIKeyAuth: &egv1a1.APIKeyAuth{
				CredentialRefs: []gwapiv1.SecretObjectReference{
					{
						Name: gwapiv1.ObjectName("api-key-secret"),
					},
				},
			},
		},
	}

	cases := []struct {
		desc        string
		extractFrom []*egv1a1.ExtractFrom
		wantErrors  []string
	}{
		{
			desc: "headers source is valid",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Headers: []string{"x-api-key"},
				},
			},
		},
		{
			desc: "params source is valid",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Params: []string{"api_key"},
				},
			},
		},
		{
			desc: "cookies source is valid",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Cookies: []string{"api-key"},
				},
			},
		},
		{
			desc: "no source specified",
			extractFrom: []*egv1a1.ExtractFrom{
				{},
			},
			wantErrors: []string{
				"exactly one of headers, params, or cookies must be specified",
			},
		},
		{
			desc: "multiple sources specified",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Headers: []string{"x-api-key"},
					Params:  []string{"api_key"},
				},
			},
			wantErrors: []string{
				"exactly one of headers, params, or cookies must be specified",
			},
		},
		{
			desc: "empty header source name",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Headers: []string{""},
				},
			},
			wantErrors: []string{
				"spec.apiKeyAuth.extractFrom[0].headers[0]",
				"should be at least 1 chars long",
			},
		},
		{
			desc: "empty param source name",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Params: []string{""},
				},
			},
			wantErrors: []string{
				"spec.apiKeyAuth.extractFrom[0].params[0]",
				"should be at least 1 chars long",
			},
		},
		{
			desc: "empty cookie source name",
			extractFrom: []*egv1a1.ExtractFrom{
				{
					Cookies: []string{""},
				},
			},
			wantErrors: []string{
				"spec.apiKeyAuth.extractFrom[0].cookies[0]",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty extractFrom list",
			extractFrom: []*egv1a1.ExtractFrom{},
			wantErrors: []string{
				"spec.apiKeyAuth.extractFrom",
				"should have at least 1 items",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sp := baseSP.DeepCopy()
			sp.Name = fmt.Sprintf("sp-api-key-auth-%v", time.Now().UnixNano())
			sp.Spec.APIKeyAuth.ExtractFrom = tc.extractFrom

			err := c.Create(ctx, sp)
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

func securityPolicySpecWithOIDCIssuer(issuer string) egv1a1.SecurityPolicySpec {
	return egv1a1.SecurityPolicySpec{
		PolicyTargetReferences: egv1a1.PolicyTargetReferences{
			TargetSelectors: []egv1a1.TargetSelector{
				{
					Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
					Kind:  "HTTPRoute",
					MatchLabels: map[string]string{
						"eg/namespace": "reference-apps",
					},
				},
			},
		},
		OIDC: &egv1a1.OIDC{
			Provider: egv1a1.OIDCProvider{
				Issuer:                issuer,
				AuthorizationEndpoint: new("https://keycloak.gateway-conformance-infra/realms/master/protocol/openid-connect/auth"),
				TokenEndpoint:         new("https://keycloak.gateway-conformance-infra/realms/master/protocol/openid-connect/token"),
			},
			ClientID: new("client-id"),
			ClientSecret: gwapiv1b1.SecretObjectReference{
				Name: "secret",
			},
		},
	}
}

func TestSecurityPolicyClientCert(t *testing.T) {
	ctx := context.Background()
	baseSP := egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.SecurityPolicySpec{},
	}

	cases := []struct {
		desc         string
		mutate       func(sp *egv1a1.SecurityPolicy)
		mutateStatus func(sp *egv1a1.SecurityPolicy)
		wantErrors   []string
	}{
		{
			desc: "valid subject-only clientCert principal",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										Subject: &egv1a1.StringMatch{
											Value: "CN=test",
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
			desc: "valid uris-only clientCert principal",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										SubjectAltNames: &egv1a1.SubjectAltNames{
											URIs: []egv1a1.StringMatch{
												{Value: "spiffe://test"},
											},
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
			desc: "valid subject and uris clientCert principal",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										Subject: &egv1a1.StringMatch{
											Value: "CN=test",
										},
										SubjectAltNames: &egv1a1.SubjectAltNames{
											URIs: []egv1a1.StringMatch{
												{Value: "spiffe://test"},
											},
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
			desc: "invalid clientCert principal with neither subject nor subjectAltNames",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"at least one of subject or subjectAltNames"},
		},
		{
			desc: "invalid clientCert principal with empty subjectAltNames object",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										// subjectAltNames present but empty: must not silently drop the
										// SAN constraint (a principal that also ANDs a CIDR/header/JWT
										// condition would otherwise authorize without matching any
										// certificate identity).
										SubjectAltNames: &egv1a1.SubjectAltNames{},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"subjectAltNames must specify at least one of uris or dnsNames"},
		},
		{
			desc: "invalid clientCert principal with emailAddresses",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										SubjectAltNames: &egv1a1.SubjectAltNames{
											EmailAddresses: []egv1a1.StringMatch{
												{Value: "user@example.com"},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"emailAddresses is not supported"},
		},
		{
			desc: "invalid clientCert principal with ipAddresses",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										SubjectAltNames: &egv1a1.SubjectAltNames{
											IPAddresses: []egv1a1.StringMatch{
												{Value: "192.0.2.1"},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"ipAddresses is not supported"},
		},
		{
			desc: "invalid clientCert principal with otherNames",
			mutate: func(sp *egv1a1.SecurityPolicy) {
				sp.Spec = egv1a1.SecurityPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetSelectors: []egv1a1.TargetSelector{
							{
								Group: new(gwapiv1.Group("gateway.networking.k8s.io")),
								Kind:  "HTTPRoute",
								MatchLabels: map[string]string{
									"eg/namespace": "reference-apps",
								},
							},
						},
					},
					Authorization: &egv1a1.Authorization{
						Rules: []egv1a1.AuthorizationRule{
							{
								Action: egv1a1.AuthorizationActionAllow,
								Principal: &egv1a1.Principal{
									ClientCert: &egv1a1.ClientCertPrincipal{
										SubjectAltNames: &egv1a1.SubjectAltNames{
											OtherNames: []egv1a1.OtherSANMatch{
												{Oid: "1.2.3.4", StringMatch: egv1a1.StringMatch{Value: "test"}},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{"otherNames is not supported"},
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

func TestSecurityPolicyOIDCCookieNames(t *testing.T) {
	ctx := context.Background()
	baseSP := egv1a1.SecurityPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sp",
			Namespace: metav1.NamespaceDefault,
		},
		Spec: egv1a1.SecurityPolicySpec{
			PolicyTargetReferences: egv1a1.PolicyTargetReferences{
				TargetRef: &gwapiv1.LocalPolicyTargetReferenceWithSectionName{
					LocalPolicyTargetReference: gwapiv1.LocalPolicyTargetReference{
						Group: gwapiv1.Group("gateway.networking.k8s.io"),
						Kind:  gwapiv1.Kind("Gateway"),
						Name:  gwapiv1.ObjectName("eg"),
					},
				},
			},
			OIDC: &egv1a1.OIDC{
				Provider: egv1a1.OIDCProvider{
					Issuer:                "https://accounts.google.com",
					AuthorizationEndpoint: new("https://accounts.google.com/o/oauth2/v2/auth"),
					TokenEndpoint:         new("https://oauth2.googleapis.com/token"),
				},
				ClientID:     new("client-id"),
				ClientSecret: gwapiv1b1.SecretObjectReference{Name: "secret"},
			},
		},
	}

	cases := []struct {
		desc        string
		cookieNames *egv1a1.OIDCCookieNames
		wantErrors  []string
	}{
		{
			desc: "all cookie names are valid",
			cookieNames: &egv1a1.OIDCCookieNames{
				AccessToken:  new("access-token"),
				OAuthExpires: new("oauth-expires"),
				OAuthHMAC:    new("oauth-hmac"),
				IDToken:      new("id-token"),
				RefreshToken: new("refresh-token"),
				OAuthNonce:   new("oauth-nonce"),
				CodeVerifier: new("code-verifier"),
			},
		},
		{
			desc:        "empty accessToken",
			cookieNames: &egv1a1.OIDCCookieNames{AccessToken: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.accessToken",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty oauthExpires",
			cookieNames: &egv1a1.OIDCCookieNames{OAuthExpires: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.oauthExpires",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty oauthHmac",
			cookieNames: &egv1a1.OIDCCookieNames{OAuthHMAC: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.oauthHmac",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty idToken",
			cookieNames: &egv1a1.OIDCCookieNames{IDToken: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.idToken",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty refreshToken",
			cookieNames: &egv1a1.OIDCCookieNames{RefreshToken: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.refreshToken",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty oauthNonce",
			cookieNames: &egv1a1.OIDCCookieNames{OAuthNonce: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.oauthNonce",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "empty codeVerifier",
			cookieNames: &egv1a1.OIDCCookieNames{CodeVerifier: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.codeVerifier",
				"should be at least 1 chars long",
			},
		},
		{
			desc: "duplicate cookie names",
			cookieNames: &egv1a1.OIDCCookieNames{
				OAuthHMAC:    new("oauth-cookie"),
				OAuthExpires: new("oauth-cookie"),
			},
			wantErrors: []string{
				"spec.oidc.cookieNames",
				"cookie names must be unique",
			},
		},
		{
			desc:        "accessToken longer than 256 chars",
			cookieNames: &egv1a1.OIDCCookieNames{AccessToken: new(strings.Repeat("a", 257))},
			// The exact wording after "Too long" varies across apiserver versions,
			// so only assert on the stable prefix.
			wantErrors: []string{
				"spec.oidc.cookieNames.accessToken",
				"Too long",
			},
		},
		{
			desc:        "suffix only",
			cookieNames: &egv1a1.OIDCCookieNames{Suffix: new("checkout")},
		},
		{
			desc: "suffix with a per-cookie override",
			cookieNames: &egv1a1.OIDCCookieNames{
				Suffix:  new("checkout"),
				IDToken: new("myapp_id_token"),
			},
		},
		{
			desc:        "empty suffix",
			cookieNames: &egv1a1.OIDCCookieNames{Suffix: new("")},
			wantErrors: []string{
				"spec.oidc.cookieNames.suffix",
				"should be at least 1 chars long",
			},
		},
		{
			desc:        "suffix longer than 238 chars",
			cookieNames: &egv1a1.OIDCCookieNames{Suffix: new(strings.Repeat("a", 239))},
			wantErrors: []string{
				"spec.oidc.cookieNames.suffix",
				"Too long",
			},
		},
		{
			desc:        "suffix is not a valid cookie name token",
			cookieNames: &egv1a1.OIDCCookieNames{Suffix: new("check out")},
			wantErrors: []string{
				"spec.oidc.cookieNames.suffix",
				"should match",
			},
		},
		{
			desc: "override collides with a name derived from the suffix",
			cookieNames: &egv1a1.OIDCCookieNames{
				Suffix:      new("checkout"),
				AccessToken: new("IdToken-checkout"),
			},
			wantErrors: []string{
				"spec.oidc.cookieNames",
				"cookie names must be unique",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.desc, func(t *testing.T) {
			sp := baseSP.DeepCopy()
			sp.Name = fmt.Sprintf("sp-oidc-cookie-names-%v", time.Now().UnixNano())
			sp.Spec.OIDC.CookieNames = tc.cookieNames

			err := c.Create(ctx, sp)
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
