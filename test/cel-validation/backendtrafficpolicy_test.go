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
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/utils/ptr"
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path: "",
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.http.path: Invalid value: "": spec.healthCheck.healthChecker.http.path in body should be at least 1 chars long`,
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
						UnhealthyThreshold: ptr.To[uint32](0),
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path: "/healthz",
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.unhealthyThreshold: Invalid value: 0: spec.healthCheck.unhealthyThreshold in body should be greater than or equal to 1`,
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
						HealthyThreshold: ptr.To[uint32](0),
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path: "/healthz",
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthyThreshold: Invalid value: 0: spec.healthCheck.healthyThreshold in body should be greater than or equal to 1`,
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							TCP:  &egv1a1.TCPHealthChecker{},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker: Invalid value: "object": If Health Checker type is HTTP, http field needs to be set., spec.healthCheck.healthChecker: Invalid value: "object": If Health Checker type is TCP, tcp field needs to be set`,
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path:             "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatusRange{{Start: 400, End: 200}},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.http.expectedStatuses[0]: Invalid value: "object": start should be not greater than end`,
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path:             "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatusRange{{Start: 100, End: 600}},
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path: "/healthz",
								ExpectedStatuses: []egv1a1.HTTPStatusRange{
									{Start: 600, End: 600},
									{Start: 200, End: 700},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.http.expectedStatuses[0].start: Invalid value: 600: spec.healthCheck.healthChecker.http.expectedStatuses[0].start in body should be less than 600, spec.healthCheck.healthChecker.http.expectedStatuses[1].end: Invalid value: 700: spec.healthCheck.healthChecker.http.expectedStatuses[1].end in body should be less than or equal to 600`,
			},
		},
		{
			desc: "invalid http expected responses",
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeHTTP,
							HTTP: &egv1a1.HTTPHealthChecker{
								Path: "/healthz",
								ExpectedResponses: []egv1a1.HealthCheckPayload{
									{
										Type:   egv1a1.HealthCheckPayloadTypeText,
										Binary: []byte{'f', 'o', 'o'},
									},
									{
										Type: egv1a1.HealthCheckPayloadTypeBinary,
										Text: ptr.To("foo"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.http.expectedResponses[0]: Invalid value: "object": If payload type is Text, text field needs to be set., spec.healthCheck.healthChecker.http.expectedResponses[0]: Invalid value: "object": If payload type is Binary, binary field needs to be set., spec.healthCheck.healthChecker.http.expectedResponses[1]: Invalid value: "object": If payload type is Text, text field needs to be set., spec.healthCheck.healthChecker.http.expectedResponses[1]: Invalid value: "object": If payload type is Binary, binary field needs to be set.`,
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeTCP,
							TCP: &egv1a1.TCPHealthChecker{
								Send: &egv1a1.HealthCheckPayload{
									Type:   egv1a1.HealthCheckPayloadTypeText,
									Binary: []byte{'f', 'o', 'o'},
								},
								Receive: []egv1a1.HealthCheckPayload{
									{
										Type: egv1a1.HealthCheckPayloadTypeText,
										Text: ptr.To("foo"),
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.tcp.send: Invalid value: "object": If payload type is Text, text field needs to be set., spec.healthCheck.healthChecker.tcp.send: Invalid value: "object": If payload type is Binary, binary field needs to be set.`,
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
						HealthChecker: egv1a1.HealthChecker{
							Type: egv1a1.HealthCheckerTypeTCP,
							TCP: &egv1a1.TCPHealthChecker{
								Send: &egv1a1.HealthCheckPayload{
									Type: egv1a1.HealthCheckPayloadTypeText,
									Text: ptr.To("foo"),
								},
								Receive: []egv1a1.HealthCheckPayload{
									{
										Type:   egv1a1.HealthCheckPayloadTypeText,
										Binary: []byte{'f', 'o', 'o'},
									},
								},
							},
						},
					},
				}
			},
			wantErrors: []string{
				`spec.healthCheck.healthChecker.tcp.receive[0]: Invalid value: "object": If payload type is Text, text field needs to be set., spec.healthCheck.healthChecker.tcp.receive[0]: Invalid value: "object": If payload type is Binary, binary field needs to be set.`,
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
