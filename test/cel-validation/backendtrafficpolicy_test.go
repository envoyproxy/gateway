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
	"k8s.io/utils/pointer"
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
						MaxParallelRetries:  nil,
					},
				}
			},
			wantErrors: []string{},
		},
		{
			desc: " invalid config: min and max values",
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
						MaxConnections:           valOverMax,
						MaxPendingRequests:       valUnderMin,
						MaxParallelRequests:      valOverMax,
						MaxRequestsPerConnection: valUnderMin,
						MaxParallelRetries:       valOverMax,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							Type: egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path: "",
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							UnhealthyThreshold: ptr.To[uint32](0),
							Type:               egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path: "/healthz",
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							HealthyThreshold: ptr.To[uint32](0),
							Type:             egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path: "/healthz",
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							Type: egv1a1.ActiveHealthCheckerTypeHTTP,
							TCP:  &egv1a1.TCPActiveHealthChecker{},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.HealthCheck.active: Invalid value: "object": If Health Checker type is HTTP, http field needs to be set., spec.HealthCheck.active: Invalid value: "object": If Health Checker type is TCP, tcp field needs to be set`,
			},
		},
		{
			desc: "invalid http expected statuses",
			mutate: func(btp *egv1a1.BackendTrafficPolicy) {
				btp.Spec = egv1a1.BackendTrafficPolicySpec{
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							Type: egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path:             "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatus{99, 200},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							Type: egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path:             "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatus{100, 200, 201},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					HealthCheck: &egv1a1.HealthCheck{
						Active: &egv1a1.ActiveHealthCheck{
							Type: egv1a1.ActiveHealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPActiveHealthChecker{
								Path:             "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatus{200, 300, 601},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
						},
					},
					Timeout: &egv1a1.Timeout{
						TCP: &egv1a1.TCPTimeout{
							ConnectTimeout: &d,
						},
						HTTP: &egv1a1.HTTPTimeout{
							ConnectionIdleTimeout: &d,
							MaxConnectionDuration: &d,
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
					TargetRef: gwapiv1a2.PolicyTargetReferenceWithSectionName{
						PolicyTargetReference: gwapiv1a2.PolicyTargetReference{
							Group: gwapiv1a2.Group("gateway.networking.k8s.io"),
							Kind:  gwapiv1a2.Kind("Gateway"),
							Name:  gwapiv1a2.ObjectName("eg"),
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
