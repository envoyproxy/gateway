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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
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
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "no targetRef",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{}
			},
			wantErrors: []string{
				"spec: Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "targetRef unsupported kind",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "targetRefs unsupported kind",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "targetRef unsupported group",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "targetRef unsupported group and kind",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "sectionName disabled until supported",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "consistentHash field not nil when type is consistentHash",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							ConsistentHash: &egv1a1.ConsistentHash{
								Type: "SourceIP",
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "consistentHash field nil when type is consistentHash",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
						},
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer: Invalid value: \"object\": If LoadBalancer type is consistentHash, consistentHash field needs to be set",
			},
		},
		{
			desc: "consistentHash header field not nil when consistentHashType is header",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							ConsistentHash: &egv1a1.ConsistentHash{
								Type: "Header",
								Header: &egv1a1.Header{
									Name: "name",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "consistentHash header field nil when consistentHashType is header",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							ConsistentHash: &egv1a1.ConsistentHash{
								Type: "Header",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer.consistentHash: Invalid value: \"object\": If consistent hash type is header, the header field must be set",
			},
		},
		{
			desc: "consistentHash cookie field not nil when consistentHashType is cookie",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							ConsistentHash: &egv1a1.ConsistentHash{
								Type: "Cookie",
								Cookie: &egv1a1.Cookie{
									Name: "name",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "consistentHash cookie field nil when consistentHashType is cookie",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							ConsistentHash: &egv1a1.ConsistentHash{
								Type: "Cookie",
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer.consistentHash: Invalid value: \"object\": If consistent hash type is cookie, the cookie field must be set",
			},
		},

		{
			desc: "leastRequest with ConsistentHash nil",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.LeastRequestLoadBalancerType,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "leastRequest with SlowStar is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.LeastRequestLoadBalancerType,
							SlowStart: &egv1a1.SlowStart{
								Window: &metav1.Duration{
									Duration: 10000000,
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "roundrobin with SlowStart is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.RoundRobinLoadBalancerType,
							SlowStart: &egv1a1.SlowStart{
								Window: &metav1.Duration{
									Duration: 10000000,
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: " random with SlowStart is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.RandomLoadBalancerType,
							SlowStart: &egv1a1.SlowStart{
								Window: &metav1.Duration{
									Duration: 10000000,
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer: Invalid value: \"object\": Currently SlowStart is only supported for RoundRobin and LeastRequest load balancers.",
			},
		},
		{
			desc: " consistenthash with SlowStart is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						LoadBalancer: &egv1a1.LoadBalancer{
							Type: egv1a1.ConsistentHashLoadBalancerType,
							SlowStart: &egv1a1.SlowStart{
								Window: &metav1.Duration{
									Duration: 10000000,
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer: Invalid value: \"object\": Currently SlowStart is only supported for RoundRobin and LeastRequest load balancers.",
			},
		},
		{
			desc: "Using both httpStatus and grpcStatus in abort fault injection",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{
						Abort: &egv1a1.FaultInjectionAbort{
							HTTPStatus: ptr.To[int32](200),
							GrpcStatus: ptr.To[int32](80),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.faultInjection.abort: Invalid value: \"object\": httpStatus and grpcStatus cannot be simultaneously defined.",
			},
		},
		{
			desc: "Using httpStatus in abort fault injection",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{
						Abort: &egv1a1.FaultInjectionAbort{
							HTTPStatus: ptr.To[int32](200),
							Percentage: ptr.To[float32](80),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "Using grpcStatus in abort fault injection",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{
						Abort: &egv1a1.FaultInjectionAbort{
							GrpcStatus: ptr.To[int32](10),
							Percentage: ptr.To[float32](20),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "httpStatus and grpcStatus are set at least one",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{
						Abort: &egv1a1.FaultInjectionAbort{
							Percentage: ptr.To[float32](20),
						},
					},
				}
			},
			wantErrors: []string{"spec.faultInjection.abort: Invalid value: \"object\": httpStatus and grpcStatus are set at least one."},
		},
		{
			desc: "Neither delay nor abort faults are set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{},
				}
			},
			wantErrors: []string{"spec.faultInjection: Invalid value: \"object\": Delay and abort faults are set at least one."},
		},
		{
			desc: "Using delay fault injection",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					FaultInjection: &egv1a1.FaultInjection{
						Delay: &egv1a1.FaultInjectionDelay{
							FixedDelay: &metav1.Duration{
								Duration: 10000000,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: " valid config: min, max, nil",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				valMax := ptr.To[int64](4294967295)
				valMin := ptr.To[int64](0)
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						CircuitBreaker: &egv1a1.CircuitBreaker{
							MaxConnections:      valMax,
							MaxPendingRequests:  valMin,
							MaxParallelRequests: nil,
							MaxParallelRetries:  nil,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: " invalid config: min and max values",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				valOverMax := ptr.To[int64](4294967296)
				valUnderMin := ptr.To[int64](-1)
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						CircuitBreaker: &egv1a1.CircuitBreaker{
							MaxConnections:           valOverMax,
							MaxPendingRequests:       valUnderMin,
							MaxParallelRequests:      valOverMax,
							MaxRequestsPerConnection: valUnderMin,
							MaxParallelRetries:       valOverMax,
						},
					},
				}
			},
			wantErrors: []string{
				"spec.circuitBreaker.MaxParallelRetries: Invalid value: 4294967296: spec.circuitBreaker.MaxParallelRetries in body should be less than or equal to 4294967295",
				"spec.circuitBreaker.maxRequestsPerConnection: Invalid value: -1: spec.circuitBreaker.maxRequestsPerConnection in body should be greater than or equal to 0",
				"spec.circuitBreaker.maxParallelRequests: Invalid value: 4294967296: spec.circuitBreaker.maxParallelRequests in body should be less than or equal to 4294967295",
				"spec.circuitBreaker.maxPendingRequests: Invalid value: -1: spec.circuitBreaker.maxPendingRequests in body should be greater than or equal to 0",
				"spec.circuitBreaker.maxConnections: Invalid value: 4294967296: spec.circuitBreaker.maxConnections in body should be less than or equal to 4294967295",
			},
		},
		{
			desc: "invalid path of http health checker",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path: "",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.http.path: Invalid value: "": spec.HealthCheck.active.http.path in body should be at least 1 chars long`,
			},
		},
		{
			desc: "invalid unhealthy threshold",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								UnhealthyThreshold: ptr.To[uint32](0),
								Type:               egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path: "/healthz",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.unhealthyThreshold: Invalid value: 0: spec.HealthCheck.active.unhealthyThreshold in body should be greater than or equal to 1`,
			},
		},
		{
			desc: "invalid healthy threshold",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								HealthyThreshold: ptr.To[uint32](0),
								Type:             egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path: "/healthz",
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.healthyThreshold: Invalid value: 0: spec.HealthCheck.active.healthyThreshold in body should be greater than or equal to 1`,
			},
		},
		{
			desc: "invalid health checker type",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								TCP:  &egv1a1.TCPActiveHealthChecker{},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active: Invalid value: "object": If Health Checker type is HTTP, http field needs to be set., spec.HealthCheck.active: Invalid value: "object": If Health Checker type is TCP, tcp field needs to be set`,
			},
		},
		{
			desc: "grpc settings with non-gRPC health checker type",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								GRPC: &egv1a1.GRPCActiveHealthChecker{},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`If Health Checker type is HTTP, http field needs to be set.`,
				`The grpc field can only be set if the Health Checker type is GRPC.`,
			},
		},

		{
			desc: "invalid http expected statuses",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path:             "/healthz",
									ExpectedStatuses: []egv1a1.HTTPStatus{99, 200},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.http.expectedStatuses[0]: Invalid value: 99: spec.HealthCheck.active.http.expectedStatuses[0] in body should be greater than or equal to 100`,
			},
		},
		{
			desc: "valid http expected statuses",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path:             "/healthz",
									ExpectedStatuses: []egv1a1.HTTPStatus{100, 200, 201},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid http expected statuses - out of range",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path:             "/healthz",
									ExpectedStatuses: []egv1a1.HTTPStatus{200, 300, 601},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.http.expectedStatuses[2]: Invalid value: 601: spec.HealthCheck.active.http.expectedStatuses[2] in body should be less than 600`,
			},
		},
		{
			desc: "http expected responses - invalid text payload",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path: "/healthz",
									ExpectedResponse: &egv1a1.ActiveHealthCheckPayload{
										Type:   egv1a1.ActiveHealthCheckPayloadTypeText,
										Binary: []byte{'f', 'o', 'o'},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`[spec.HealthCheck.active.http.expectedResponse: Invalid value: "object": If payload type is Text, text field needs to be set., spec.HealthCheck.active.http.expectedResponse: Invalid value: "object": If payload type is Binary, binary field needs to be set.]`,
			},
		},
		{
			desc: "http expected responses - invalid binary payload",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeHTTP,
								HTTP: &egv1a1.HTTPActiveHealthChecker{
									Path: "/healthz",
									ExpectedResponse: &egv1a1.ActiveHealthCheckPayload{
										Type: egv1a1.ActiveHealthCheckPayloadTypeBinary,
										Text: ptr.To("foo"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`[spec.HealthCheck.active.http.expectedResponse: Invalid value: "object": If payload type is Text, text field needs to be set., spec.HealthCheck.active.http.expectedResponse: Invalid value: "object": If payload type is Binary, binary field needs to be set.]`,
			},
		},
		{
			desc: "invalid tcp send",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeTCP,
								TCP: &egv1a1.TCPActiveHealthChecker{
									Send: &egv1a1.ActiveHealthCheckPayload{
										Type:   egv1a1.ActiveHealthCheckPayloadTypeText,
										Binary: []byte{'f', 'o', 'o'},
									},
									Receive: &egv1a1.ActiveHealthCheckPayload{
										Type: egv1a1.ActiveHealthCheckPayloadTypeText,
										Text: ptr.To("foo"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active.tcp.send: Invalid value: "object": If payload type is Text, text field needs to be set., spec.HealthCheck.active.tcp.send: Invalid value: "object": If payload type is Binary, binary field needs to be set.`,
			},
		},
		{
			desc: "invalid tcp receive",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							Active: &egv1a1.ActiveHealthCheck{
								Type: egv1a1.ActiveHealthCheckerTypeTCP,
								TCP: &egv1a1.TCPActiveHealthChecker{
									Send: &egv1a1.ActiveHealthCheckPayload{
										Type: egv1a1.ActiveHealthCheckPayloadTypeText,
										Text: ptr.To("foo"),
									},
									Receive: &egv1a1.ActiveHealthCheckPayload{
										Type:   egv1a1.ActiveHealthCheckPayloadTypeText,
										Binary: []byte{'f', 'o', 'o'},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`[spec.HealthCheck.active.tcp.receive: Invalid value: "object": If payload type is Text, text field needs to be set., spec.HealthCheck.active.tcp.receive: Invalid value: "object": If payload type is Binary, binary field needs to be set.]`,
			},
		},
		{
			desc: " valid timeout",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				d := gwapiv1.Duration("3s")
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						Timeout: &egv1a1.Timeout{
							TCP: &egv1a1.TCPTimeout{
								ConnectTimeout: &d,
							},
							HTTP: &egv1a1.HTTPTimeout{
								ConnectionIdleTimeout: &d,
								MaxConnectionDuration: &d,
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "valid count of Global rate limit rules items",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				rules := make([]egv1a1.RateLimitRule, 64)
				rule := egv1a1.RateLimitRule{
					Limit: egv1a1.RateLimitValue{
						Requests: 10,
						Unit:     "Minute",
					},
				}
				for i := range rules {
					rules[i] = rule
				}

				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					RateLimit: &egv1a1.RateLimitSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: rules,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid count of Global rate limit rules items",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				rules := make([]egv1a1.RateLimitRule, 65)
				rule := egv1a1.RateLimitRule{
					Limit: egv1a1.RateLimitValue{
						Requests: 10,
						Unit:     "Minute",
					},
				}
				for i := range rules {
					rules[i] = rule
				}

				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					RateLimit: &egv1a1.RateLimitSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: rules,
						},
					},
				}
			},
			wantErrors: []string{
				`[spec.rateLimit.global.rules: Too many: 65: must have at most 64 items, <nil>: Invalid value: "null": some validation rules were not checked because the object was invalid; correct the existing errors to complete validation]`,
			},
		},
		{
			desc: "valid connectionBufferLimitBytes format",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: ptr.To(resource.MustParse("1Mi")),
						},
					},
				}
			},
		},
		{
			desc: "connectionBufferLimitBytes given as a number",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: ptr.To(resource.MustParse("12345678")),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid connectionBufferLimitBytes format",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						Connection: &egv1a1.BackendConnection{
							BufferLimit: ptr.To(resource.MustParse("1m")),
						},
					},
				}
			},
			wantErrors: []string{
				"spec.connection.bufferLimit: Invalid value: \"1m\": spec.connection.bufferLimit in body should match '^[1-9]+[0-9]*([EPTGMK]i|[EPTGMk])?$', <nil>: Invalid value: \"\"",
			},
		},
		{
			desc: "both targetref and targetrefs specified",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
									Kind:  gwapiv1a2.Kind("Gateway"),
									Name:  gwapiv1a2.ObjectName("eg"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				" Invalid value: \"object\": either targetRef or targetRefs must be used",
			},
		},
		{
			desc: "target selectors without targetRefs or targetRef",
			mutate: func(sp *egv1a1.BackendTrafficPolicy) {
				sp.Spec = egv1a1.BackendTrafficPolicySpec{
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
			desc: "custom response or redirect required in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeRange),
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.responseOverride[0]: Invalid value: \"object\": exactly one of response or redirect must be specified",
			},
		},
		{
			desc: "custom response in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeRange),
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
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
			desc: "custom redirect in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeRange),
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
							Redirect: &egv1a1.CustomRedirect{
								Scheme:   ptr.To("https"),
								Hostname: ptr.To(gwapiv1a2.PreciseHostname("redirect.host")),
								Path: &gwapiv1.HTTPPathModifier{
									Type:            "ReplaceFullPath",
									ReplaceFullPath: ptr.To("/redirect"),
								},
								Port:       ptr.To(gwapiv1a2.PortNumber(9090)),
								StatusCode: ptr.To(302),
							},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "custom redirect in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeRange),
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
							Redirect: &egv1a1.CustomRedirect{
								Scheme:   ptr.To("https"),
								Hostname: ptr.To(gwapiv1a2.PreciseHostname("redirect.host")),
								Path: &gwapiv1.HTTPPathModifier{
									Type:               "ReplacePrefixMatch",
									ReplacePrefixMatch: ptr.To("/redirect"),
								},
								Port:       ptr.To(gwapiv1a2.PortNumber(9090)),
								StatusCode: ptr.To(302),
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.responseOverride[0].redirect.path: Invalid value: \"object\": only ReplaceFullPath is supported for path.type",
			},
		},
		{
			desc: "status value required for type in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeValue),
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.responseOverride[0].match.statusCodes[0]: Invalid value: \"object\": value must be set for type Value",
			},
		},
		{
			desc: "status value required for default type in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Range: &egv1a1.StatusCodeRange{
											Start: 100,
											End:   200,
										},
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.responseOverride[0].match.statusCodes[0]: Invalid value: \"object\": value must be set for type Value",
			},
		},
		{
			desc: "status range required for type in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type:  ptr.To(egv1a1.StatusCodeValueTypeRange),
										Value: ptr.To(100),
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"spec.responseOverride[0].match.statusCodes[0]: Invalid value: \"object\": range must be set for type Range",
			},
		},
		{
			desc: "status range invalid in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Type: ptr.To(egv1a1.StatusCodeValueTypeRange),
										Range: &egv1a1.StatusCodeRange{
											Start: 200,
											End:   100,
										},
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"end must be greater than start",
			},
		},
		{
			desc: "default require inline response body in response override",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Value: ptr.To(100),
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("ConfigMap"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"inline must be set for type Inline",
			},
		},
		{
			desc: "both targetref and targetrefs specified",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Value: ptr.To(100),
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type:   ptr.To(egv1a1.ResponseValueTypeValueRef),
									Inline: ptr.To("foo"),
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"valueRef must be set for type ValueRef",
			},
		},
		{
			desc: "both targetref and targetrefs specified",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ResponseOverride: []*egv1a1.ResponseOverride{
						{
							Match: egv1a1.CustomResponseMatch{
								StatusCodes: []egv1a1.StatusCodeMatch{
									{
										Value: ptr.To(100),
									},
								},
							},
							Response: &egv1a1.CustomResponse{
								Body: &egv1a1.CustomResponseBody{
									Type: ptr.To(egv1a1.ResponseValueTypeValueRef),
									ValueRef: &gwapiv1a2.LocalObjectReference{
										Kind: gwapiv1a2.Kind("Foo"),
										Name: gwapiv1a2.ObjectName("eg"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				"only ConfigMap is supported for ValueRe",
			},
		},
		{
			desc: "valid Global rate limit rules with request and response hit addends",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				rules := []egv1a1.RateLimitRule{
					{
						Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
						Cost: &egv1a1.RateLimitCost{
							Request: &egv1a1.RateLimitCostSpecifier{From: egv1a1.RateLimitCostFromNumber, Number: ptr.To[uint64](200)},
						},
					},
					{
						Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
						Cost: &egv1a1.RateLimitCost{
							Response: &egv1a1.RateLimitCostSpecifier{From: egv1a1.RateLimitCostFromNumber, Number: ptr.To[uint64](200)},
						},
					},
					{
						Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
						Cost: &egv1a1.RateLimitCost{
							Request:  &egv1a1.RateLimitCostSpecifier{From: egv1a1.RateLimitCostFromNumber, Number: ptr.To[uint64](200)},
							Response: &egv1a1.RateLimitCostSpecifier{From: egv1a1.RateLimitCostFromNumber, Number: ptr.To[uint64](200)},
						},
					},
					{
						Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
						Cost: &egv1a1.RateLimitCost{
							Request: &egv1a1.RateLimitCostSpecifier{
								From: egv1a1.RateLimitCostFromMetadata,
								Metadata: &egv1a1.RateLimitCostMetadata{
									Namespace: "com.test.my_filter",
									Key:       "on_request_key",
								},
							},
							Response: &egv1a1.RateLimitCostSpecifier{
								From: egv1a1.RateLimitCostFromMetadata,
								Metadata: &egv1a1.RateLimitCostMetadata{
									Namespace: "com.test.my_filter",
									Key:       "on_response_key",
								},
							},
						},
					},
				}

				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					RateLimit: &egv1a1.RateLimitSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: rules,
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "invalid Global rate limit rules with request cost specifying both number and metadata fields",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					RateLimit: &egv1a1.RateLimitSpec{
						Type: egv1a1.GlobalRateLimitType,
						Global: &egv1a1.GlobalRateLimit{
							Rules: []egv1a1.RateLimitRule{
								{
									Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
									Cost: &egv1a1.RateLimitCost{
										Request: &egv1a1.RateLimitCostSpecifier{
											From:     egv1a1.RateLimitCostFromNumber,
											Metadata: &egv1a1.RateLimitCostMetadata{},
											Number:   ptr.To[uint64](200),
										},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.rateLimit.global.rules[0].cost.request: Invalid value: "object": only one of number or metadata can be specified`,
			},
		},
		{
			desc: "invalid count of local rate limit rules specifying costPerResponse",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					RateLimit: &egv1a1.RateLimitSpec{
						Type: egv1a1.GlobalRateLimitType,
						Local: &egv1a1.LocalRateLimit{
							Rules: []egv1a1.RateLimitRule{
								{
									Limit: egv1a1.RateLimitValue{Requests: 10, Unit: "Minute"},
									Cost: &egv1a1.RateLimitCost{
										// This is not supported for LocalRateLimit.
										Response: &egv1a1.RateLimitCostSpecifier{From: egv1a1.RateLimitCostFromNumber, Number: ptr.To[uint64](200)},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{`response cost is not supported for Local Rate Limits`},
		},
		{
			desc: "panicThreshold is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							PanicThreshold: ptr.To[uint32](80),
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "panicThreshold fails validation",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRef: &gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
								Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
								Kind:  gwapiv1a2.Kind("Gateway"),
								Name:  gwapiv1a2.ObjectName("eg"),
							},
						},
					},
					ClusterSettings: egv1a1.ClusterSettings{
						HealthCheck: &egv1a1.HealthCheck{
							PanicThreshold: ptr.To[uint32](200),
						},
					},
				}
			},
			wantErrors: []string{`Invalid value: 200: spec.healthCheck.panicThreshold in body should be less than or equal to 100`},
		},
		{
			desc: "websocket with connect config",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: "gateway.networking.k8s.io",
									Kind:  "Gateway",
									Name:  "eg",
								},
							},
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type:    "websocket",
							Connect: &egv1a1.ConnectConfig{},
						},
					},
				}
			},
			wantErrors: []string{"The connect configuration is only allowed when the type is CONNECT."},
		},
		{
			desc: "http connect config",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: "gateway.networking.k8s.io",
									Kind:  "Gateway",
									Name:  "eg",
								},
							},
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type:    "CONNECT",
							Connect: &egv1a1.ConnectConfig{},
						},
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "http with connect config",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					PolicyTargetReferences: egv1a1.PolicyTargetReferences{
						TargetRefs: []gwapiv1a2.LocalPolicyTargetReferenceWithSectionName{
							{
								LocalPolicyTargetReference: gwapiv1a2.LocalPolicyTargetReference{
									Group: "gateway.networking.k8s.io",
									Kind:  "Gateway",
									Name:  "eg",
								},
							},
						},
					},
					HTTPUpgrade: []*egv1a1.ProtocolUpgradeConfig{
						{
							Type: "CONNECT",
						},
					},
				}
			},
			wantErrors: []string{},
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
