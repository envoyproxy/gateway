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
	"k8s.io/utils/pointer"
	"k8s.io/utils/ptr"

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
			desc: "sectionName disabled until supported",
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
			wantErrors: []string{
				"spec.targetRef: Invalid value: \"object\": this policy does not yet support the sectionName field",
			},
		},
		{
			desc: "consistentHash field not nil when type is consistentHash",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.ConsistentHashLoadBalancerType,
						ConsistentHash: &egv1a1.ConsistentHash{
							Type: "SourceIP",
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.ConsistentHashLoadBalancerType,
					},
				}
			},
			wantErrors: []string{
				"spec.loadBalancer: Invalid value: \"object\": If LoadBalancer type is consistentHash, consistentHash field needs to be set",
			},
		},
		{
			desc: "leastRequest with ConsistentHash nil",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.LeastRequestLoadBalancerType,
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: "leastRequest with SlowStar is set",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.LeastRequestLoadBalancerType,
						SlowStart: &egv1a1.SlowStart{
							Window: &metav1.Duration{
								Duration: 10000000,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.RoundRobinLoadBalancerType,
						SlowStart: &egv1a1.SlowStart{
							Window: &metav1.Duration{
								Duration: 10000000,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.RandomLoadBalancerType,
						SlowStart: &egv1a1.SlowStart{
							Window: &metav1.Duration{
								Duration: 10000000,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					LoadBalancer: &egv1a1.LoadBalancer{
						Type: egv1a1.ConsistentHashLoadBalancerType,
						SlowStart: &egv1a1.SlowStart{
							Window: &metav1.Duration{
								Duration: 10000000,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
				valMax := pointer.Int64(4294967295)
				valMin := pointer.Int64(0)
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					CircuitBreaker: &egv1a1.CircuitBreaker{
						MaxConnections:      valMax,
						MaxPendingRequests:  valMin,
						MaxParallelRequests: nil,
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: " invalid config: min and max valyues",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				valOverMax := pointer.Int64(4294967296)
				valUnderMin := pointer.Int64(-1)
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					CircuitBreaker: &egv1a1.CircuitBreaker{
						MaxConnections:      valOverMax,
						MaxPendingRequests:  valUnderMin,
						MaxParallelRequests: valOverMax,
					},
				}
			},
			wantErrors: []string{
				"spec.circuitBreaker.maxParallelRequests: Invalid value: 4294967296: spec.circuitBreaker.maxParallelRequests in body should be less than or equal to 4294967295",
				"spec.circuitBreaker.maxPendingRequests: Invalid value: -1: spec.circuitBreaker.maxPendingRequests in body should be greater than or equal to 0",
				"spec.circuitBreaker.maxConnections: Invalid value: 4294967296: spec.circuitBreaker.maxConnections in body should be less than or equal to 4294967295",
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
