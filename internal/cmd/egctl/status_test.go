// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func TestWriteStatus(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name               string
		resourceNamespaced bool
		resourceList       client.ObjectList
		resourceKind       string
		namespace          string
		quiet              bool
		verbose            bool
		allNamespaces      bool
		typedName          bool
		outputs            string
	}{
		{
			name:               "egctl x status gc -v, but no resources",
			resourceList:       &gwapiv1.GatewayClassList{},
			resourceNamespaced: false,
			resourceKind:       resource.KindGatewayClass,
			quiet:              false,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS    REASON    MESSAGE   OBSERVED GENERATION   LAST TRANSITION TIME
`,
		},
		{
			name: "egctl x status gc",
			resourceList: &gwapiv1.GatewayClassList{
				Items: []gwapiv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwapiv1.GatewayClassStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar1",
									Status:             metav1.ConditionStatus("test-status-1"),
									ObservedGeneration: 123456,
									LastTransitionTime: metav1.NewTime(testTime),
									Reason:             "test reason 1",
									Message:            "test message 1",
								},
								{
									Type:               "foobar2",
									Status:             metav1.ConditionStatus("test-status-2"),
									ObservedGeneration: 123457,
									LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
									Reason:             "test reason 2",
									Message:            "test message 2",
								},
							},
						},
					},
				},
			},
			resourceNamespaced: false,
			resourceKind:       resource.KindGatewayClass,
			quiet:              false,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON
gc        foobar2   test-status-2   test reason 2
          foobar1   test-status-1   test reason 1
`,
		},
		{
			name: "egctl x status gc -v",
			resourceList: &gwapiv1.GatewayClassList{
				Items: []gwapiv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwapiv1.GatewayClassStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar1",
									Status:             metav1.ConditionStatus("test-status-1"),
									ObservedGeneration: 123456,
									LastTransitionTime: metav1.NewTime(testTime),
									Reason:             "test reason 1",
									Message:            "test message 1",
								},
								{
									Type:               "foobar2",
									Status:             metav1.ConditionStatus("test-status-2"),
									ObservedGeneration: 123457,
									LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
									Reason:             "test reason 2",
									Message:            "test message 2",
								},
							},
						},
					},
				},
			},
			resourceNamespaced: false,
			resourceKind:       resource.KindGatewayClass,
			quiet:              false,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
gc        foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
          foobar1   test-status-1   test reason 1   test message 1   123456                2024-01-01 00:00:00 +0000 UTC
`,
		},
		{
			name: "egctl x status gc -v -q",
			resourceList: &gwapiv1.GatewayClassList{
				Items: []gwapiv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwapiv1.GatewayClassStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar1",
									Status:             metav1.ConditionStatus("test-status-1"),
									ObservedGeneration: 123456,
									LastTransitionTime: metav1.NewTime(testTime),
									Reason:             "test reason 1",
									Message:            "test message 1",
								},
								{
									Type:               "foobar2",
									Status:             metav1.ConditionStatus("test-status-2"),
									ObservedGeneration: 123457,
									LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
									Reason:             "test reason 2",
									Message:            "test message 2",
								},
							},
						},
					},
				},
			},
			resourceNamespaced: false,
			resourceKind:       resource.KindGatewayClass,
			quiet:              true,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
gc        foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
`,
		},
		{
			name:               "egctl x status gtw -v -A, no resources",
			resourceList:       &gwapiv1.GatewayList{},
			resourceNamespaced: true,
			resourceKind:       resource.KindGateway,
			quiet:              false,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS    REASON    MESSAGE   OBSERVED GENERATION   LAST TRANSITION TIME
`,
		},
		{
			name: "egctl x status gtw -v -A",
			resourceList: &gwapiv1.GatewayList{
				Items: []gwapiv1.Gateway{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gtw",
							Namespace: "default",
						},
						Status: gwapiv1.GatewayStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar1",
									Status:             metav1.ConditionStatus("test-status-1"),
									ObservedGeneration: 123456,
									LastTransitionTime: metav1.NewTime(testTime),
									Reason:             "test reason 1",
									Message:            "test message 1",
								},
								{
									Type:               "foobar2",
									Status:             metav1.ConditionStatus("test-status-2"),
									ObservedGeneration: 123457,
									LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
									Reason:             "test reason 2",
									Message:            "test message 2",
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindGateway,
			quiet:              false,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
default     gtw       foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
                      foobar1   test-status-1   test reason 1   test message 1   123456                2024-01-01 00:00:00 +0000 UTC
`,
		},
		{
			name: "egctl x status gtw -v -q -A",
			resourceList: &gwapiv1.GatewayList{
				Items: []gwapiv1.Gateway{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gtw1",
							Namespace: "default1",
						},
						Status: gwapiv1.GatewayStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar1",
									Status:             metav1.ConditionStatus("test-status-1"),
									ObservedGeneration: 123456,
									LastTransitionTime: metav1.NewTime(testTime),
									Reason:             "test reason 1",
									Message:            "test message 1",
								},
								{
									Type:               "foobar2",
									Status:             metav1.ConditionStatus("test-status-2"),
									ObservedGeneration: 123457,
									LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
									Reason:             "test reason 2",
									Message:            "test message 2",
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gtw2",
							Namespace: "default2",
						},
						Status: gwapiv1.GatewayStatus{
							Conditions: []metav1.Condition{
								{
									Type:               "foobar3",
									Status:             metav1.ConditionStatus("test-status-3"),
									ObservedGeneration: 123458,
									LastTransitionTime: metav1.NewTime(testTime.Add(2 * time.Hour)),
									Reason:             "test reason 3",
									Message:            "test message 3",
								},
								{
									Type:               "foobar4",
									Status:             metav1.ConditionStatus("test-status-4"),
									ObservedGeneration: 123459,
									LastTransitionTime: metav1.NewTime(testTime.Add(3 * time.Hour)),
									Reason:             "test reason 4",
									Message:            "test message 4",
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindGateway,
			quiet:              true,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
default1    gtw1      foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
default2    gtw2      foobar4   test-status-4   test reason 4   test message 4   123459                2024-01-01 03:00:00 +0000 UTC
`,
		},
		{
			name: "egctl x status httproute -A",
			resourceList: &gwapiv1.HTTPRouteList{
				Items: []gwapiv1.HTTPRoute{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http1",
							Namespace: "default1",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-1"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar1",
												Status:             metav1.ConditionStatus("test-status-1"),
												ObservedGeneration: 123456,
												LastTransitionTime: metav1.NewTime(testTime),
												Reason:             "test reason 1",
												Message:            "test message 1",
											},
											{
												Type:               "foobar2",
												Status:             metav1.ConditionStatus("test-status-2"),
												ObservedGeneration: 123457,
												LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
												Reason:             "test reason 2",
												Message:            "test message 2",
											},
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http2",
							Namespace: "default2",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-2"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar3",
												Status:             metav1.ConditionStatus("test-status-3"),
												ObservedGeneration: 123458,
												LastTransitionTime: metav1.NewTime(testTime.Add(2 * time.Hour)),
												Reason:             "test reason 3",
												Message:            "test message 3",
											},
											{
												Type:               "foobar4",
												Status:             metav1.ConditionStatus("test-status-4"),
												ObservedGeneration: 123459,
												LastTransitionTime: metav1.NewTime(testTime.Add(3 * time.Hour)),
												Reason:             "test reason 4",
												Message:            "test message 4",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindHTTPRoute,
			quiet:              false,
			verbose:            false,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      PARENT           TYPE      STATUS          REASON
default1    http1     gateway/test-1   foobar2   test-status-2   test reason 2
                                       foobar1   test-status-1   test reason 1
default2    http2     gateway/test-2   foobar4   test-status-4   test reason 4
                                       foobar3   test-status-3   test reason 3
`,
		},
		{
			name: "egctl x status httproute -q -n default1",
			resourceList: &gwapiv1.HTTPRouteList{
				Items: []gwapiv1.HTTPRoute{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http1",
							Namespace: "default1",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-1"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar1",
												Status:             metav1.ConditionStatus("test-status-1"),
												ObservedGeneration: 123456,
												LastTransitionTime: metav1.NewTime(testTime),
												Reason:             "test reason 1",
												Message:            "test message 1",
											},
											{
												Type:               "foobar2",
												Status:             metav1.ConditionStatus("test-status-2"),
												ObservedGeneration: 123457,
												LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
												Reason:             "test reason 2",
												Message:            "test message 2",
											},
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http2",
							Namespace: "default2",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-2"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar3",
												Status:             metav1.ConditionStatus("test-status-3"),
												ObservedGeneration: 123458,
												LastTransitionTime: metav1.NewTime(testTime.Add(2 * time.Hour)),
												Reason:             "test reason 3",
												Message:            "test message 3",
											},
											{
												Type:               "foobar4",
												Status:             metav1.ConditionStatus("test-status-4"),
												ObservedGeneration: 123459,
												LastTransitionTime: metav1.NewTime(testTime.Add(3 * time.Hour)),
												Reason:             "test reason 4",
												Message:            "test message 4",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindHTTPRoute,
			quiet:              true,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			namespace:          "default1",
			outputs: `NAME      PARENT           TYPE      STATUS          REASON
http1     gateway/test-1   foobar2   test-status-2   test reason 2
http2     gateway/test-2   foobar4   test-status-4   test reason 4
`,
		},
		{
			name: "egctl x status btlspolicy",
			resourceList: &gwapiv1.BackendTLSPolicyList{
				Items: []gwapiv1.BackendTLSPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "btls",
							Namespace: "default",
						},
						Status: gwapiv1.PolicyStatus{
							Ancestors: []gwapiv1.PolicyAncestorStatus{
								{
									AncestorRef: gwapiv1.ParentReference{
										Kind: gatewayapi.KindPtr(resource.KindGateway),
										Name: gwapiv1.ObjectName("test"),
									},
									Conditions: []metav1.Condition{
										{
											Type:               "foobar1",
											Status:             metav1.ConditionStatus("test-status-1"),
											ObservedGeneration: 123456,
											LastTransitionTime: metav1.NewTime(testTime),
											Reason:             "test reason 1",
											Message:            "test message 1",
										},
										{
											Type:               "foobar2",
											Status:             metav1.ConditionStatus("test-status-2"),
											ObservedGeneration: 123457,
											LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
											Reason:             "test reason 2",
											Message:            "test message 2",
										},
									},
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindBackendTLSPolicy,
			quiet:              false,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      ANCESTOR REFERENCE   TYPE      STATUS          REASON
btls      gateway/test         foobar2   test-status-2   test reason 2
                               foobar1   test-status-1   test reason 1
`,
		},
		{
			name: "multiple httproutes with multiple parents",
			resourceList: &gwapiv1.HTTPRouteList{
				Items: []gwapiv1.HTTPRoute{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http1",
							Namespace: "default1",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-1"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar1",
												Status:             metav1.ConditionStatus("test-status-1"),
												ObservedGeneration: 123456,
												LastTransitionTime: metav1.NewTime(testTime),
												Reason:             "test reason 1",
												Message:            "test message 1",
											},
											{
												Type:               "foobar2",
												Status:             metav1.ConditionStatus("test-status-2"),
												ObservedGeneration: 123457,
												LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
												Reason:             "test reason 2",
												Message:            "test message 2",
											},
										},
									},
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-2"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar3",
												Status:             metav1.ConditionStatus("test-status-3"),
												ObservedGeneration: 123456,
												LastTransitionTime: metav1.NewTime(testTime),
												Reason:             "test reason 3",
												Message:            "test message 3",
											},
											{
												Type:               "foobar4",
												Status:             metav1.ConditionStatus("test-status-4"),
												ObservedGeneration: 123457,
												LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
												Reason:             "test reason 4",
												Message:            "test message 4",
											},
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http2",
							Namespace: "default2",
						},
						Status: gwapiv1.HTTPRouteStatus{
							RouteStatus: gwapiv1.RouteStatus{
								Parents: []gwapiv1.RouteParentStatus{
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-3"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar5",
												Status:             metav1.ConditionStatus("test-status-5"),
												ObservedGeneration: 123458,
												LastTransitionTime: metav1.NewTime(testTime.Add(2 * time.Hour)),
												Reason:             "test reason 5",
												Message:            "test message 5",
											},
											{
												Type:               "foobar6",
												Status:             metav1.ConditionStatus("test-status-6"),
												ObservedGeneration: 123459,
												LastTransitionTime: metav1.NewTime(testTime.Add(3 * time.Hour)),
												Reason:             "test reason 6",
												Message:            "test message 6",
											},
										},
									},
									{
										ParentRef: gwapiv1.ParentReference{
											Kind: gatewayapi.KindPtr(resource.KindGateway),
											Name: gwapiv1.ObjectName("test-4"),
										},
										Conditions: []metav1.Condition{
											{
												Type:               "foobar7",
												Status:             metav1.ConditionStatus("test-status-7"),
												ObservedGeneration: 123458,
												LastTransitionTime: metav1.NewTime(testTime.Add(2 * time.Hour)),
												Reason:             "test reason 7",
												Message:            "test message 7",
											},
											{
												Type:               "foobar8",
												Status:             metav1.ConditionStatus("test-status-8"),
												ObservedGeneration: 123459,
												LastTransitionTime: metav1.NewTime(testTime.Add(3 * time.Hour)),
												Reason:             "test reason 8",
												Message:            "test message 8",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindHTTPRoute,
			quiet:              false,
			verbose:            false,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      PARENT           TYPE      STATUS          REASON
default1    http1     gateway/test-1   foobar2   test-status-2   test reason 2
                                       foobar1   test-status-1   test reason 1
                      gateway/test-2   foobar4   test-status-4   test reason 4
                                       foobar3   test-status-3   test reason 3
default2    http2     gateway/test-3   foobar6   test-status-6   test reason 6
                                       foobar5   test-status-5   test reason 5
                      gateway/test-4   foobar8   test-status-8   test reason 8
                                       foobar7   test-status-7   test reason 7
`,
		},
		{
			name: "multiple backendtrafficpolicy with multiple ancestors",
			resourceList: &egv1a1.BackendTrafficPolicyList{
				Items: []egv1a1.BackendTrafficPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "btp-1",
							Namespace: "default",
						},
						Status: gwapiv1.PolicyStatus{
							Ancestors: []gwapiv1.PolicyAncestorStatus{
								{
									AncestorRef: gwapiv1.ParentReference{
										Kind: gatewayapi.KindPtr(resource.KindGateway),
										Name: gwapiv1.ObjectName("test-1"),
									},
									Conditions: []metav1.Condition{
										{
											Type:               "foobar1",
											Status:             metav1.ConditionStatus("test-status-1"),
											ObservedGeneration: 123456,
											LastTransitionTime: metav1.NewTime(testTime),
											Reason:             "test reason 1",
											Message:            "test message 1",
										},
										{
											Type:               "foobar2",
											Status:             metav1.ConditionStatus("test-status-2"),
											ObservedGeneration: 123457,
											LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
											Reason:             "test reason 2",
											Message:            "test message 2",
										},
									},
								},
								{
									AncestorRef: gwapiv1.ParentReference{
										Kind: gatewayapi.KindPtr(resource.KindHTTPRoute),
										Name: gwapiv1.ObjectName("test-2"),
									},
									Conditions: []metav1.Condition{
										{
											Type:               "foobar3",
											Status:             metav1.ConditionStatus("test-status-3"),
											ObservedGeneration: 123456,
											LastTransitionTime: metav1.NewTime(testTime),
											Reason:             "test reason 3",
											Message:            "test message 3",
										},
										{
											Type:               "foobar4",
											Status:             metav1.ConditionStatus("test-status-4"),
											ObservedGeneration: 123457,
											LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
											Reason:             "test reason 4",
											Message:            "test message 4",
										},
									},
								},
							},
						},
					},
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "btp-2",
							Namespace: "default",
						},
						Status: gwapiv1.PolicyStatus{
							Ancestors: []gwapiv1.PolicyAncestorStatus{
								{
									AncestorRef: gwapiv1.ParentReference{
										Kind: gatewayapi.KindPtr(resource.KindGateway),
										Name: gwapiv1.ObjectName("test-3"),
									},
									Conditions: []metav1.Condition{
										{
											Type:               "foobar5",
											Status:             metav1.ConditionStatus("test-status-5"),
											ObservedGeneration: 123456,
											LastTransitionTime: metav1.NewTime(testTime),
											Reason:             "test reason 5",
											Message:            "test message 5",
										},
										{
											Type:               "foobar6",
											Status:             metav1.ConditionStatus("test-status-6"),
											ObservedGeneration: 123457,
											LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
											Reason:             "test reason 6",
											Message:            "test message 6",
										},
									},
								},
								{
									AncestorRef: gwapiv1.ParentReference{
										Kind: gatewayapi.KindPtr(resource.KindGRPCRoute),
										Name: gwapiv1.ObjectName("test-4"),
									},
									Conditions: []metav1.Condition{
										{
											Type:               "foobar7",
											Status:             metav1.ConditionStatus("test-status-7"),
											ObservedGeneration: 123456,
											LastTransitionTime: metav1.NewTime(testTime),
											Reason:             "test reason 7",
											Message:            "test message 7",
										},
										{
											Type:               "foobar8",
											Status:             metav1.ConditionStatus("test-status-8"),
											ObservedGeneration: 123457,
											LastTransitionTime: metav1.NewTime(testTime.Add(1 * time.Hour)),
											Reason:             "test reason 8",
											Message:            "test message 8",
										},
									},
								},
							},
						},
					},
				},
			},
			resourceNamespaced: true,
			resourceKind:       resource.KindBackendTrafficPolicy,
			quiet:              false,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      ANCESTOR REFERENCE   TYPE      STATUS          REASON
btp-1     gateway/test-1       foobar2   test-status-2   test reason 2
                               foobar1   test-status-1   test reason 1
          httproute/test-2     foobar4   test-status-4   test reason 4
                               foobar3   test-status-3   test reason 3
btp-2     gateway/test-3       foobar6   test-status-6   test reason 6
                               foobar5   test-status-5   test reason 5
          grpcroute/test-4     foobar8   test-status-8   test reason 8
                               foobar7   test-status-7   test reason 7
`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			tab := newStatusTableWriter(&out)

			needNamespace := tc.allNamespaces && tc.resourceNamespaced
			header := fetchStatusHeader(tc.resourceKind, tc.verbose, needNamespace)
			body := fetchStatusBody(tc.resourceList, tc.resourceKind, tc.quiet, tc.verbose, needNamespace, tc.typedName)
			writeStatusTable(tab, header, body)
			err := tab.Flush()
			require.NoError(t, err)

			require.Equal(t, tc.outputs, out.String())
		})
	}
}
