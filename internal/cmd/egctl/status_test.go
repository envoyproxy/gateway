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
	gwv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
)

func TestWriteStatus(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	testCases := []struct {
		name               string
		resourceNamespaced bool
		resourceList       client.ObjectList
		resourceType       string
		namespace          string
		quiet              bool
		verbose            bool
		allNamespaces      bool
		typedName          bool
		outputs            string
		expect             bool
	}{
		{
			name:               "egctl x status gc -v, but no resources",
			resourceList:       &gwv1.GatewayClassList{},
			resourceNamespaced: false,
			resourceType:       "gatewayclass",
			quiet:              false,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS    REASON    MESSAGE   OBSERVED GENERATION   LAST TRANSITION TIME
`,
			expect: true,
		},
		{
			name: "egctl x status gc",
			resourceList: &gwv1.GatewayClassList{
				Items: []gwv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwv1.GatewayClassStatus{
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
			resourceType:       "gatewayclass",
			quiet:              false,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON
gc        foobar2   test-status-2   test reason 2
          foobar1   test-status-1   test reason 1
`,
			expect: true,
		},
		{
			name: "egctl x status gc -v",
			resourceList: &gwv1.GatewayClassList{
				Items: []gwv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwv1.GatewayClassStatus{
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
			resourceType:       "gatewayclass",
			quiet:              false,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
gc        foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
          foobar1   test-status-1   test reason 1   test message 1   123456                2024-01-01 00:00:00 +0000 UTC
`,
			expect: true,
		},
		{
			name: "egctl x status gc -v -q",
			resourceList: &gwv1.GatewayClassList{
				Items: []gwv1.GatewayClass{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name: "gc",
						},
						Status: gwv1.GatewayClassStatus{
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
			resourceType:       "gatewayclass",
			quiet:              true,
			verbose:            true,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
gc        foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
`,
			expect: true,
		},
		{
			name:               "egctl x status gtw -v -A, no resources",
			resourceList:       &gwv1.GatewayList{},
			resourceNamespaced: true,
			resourceType:       "gateway",
			quiet:              false,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS    REASON    MESSAGE   OBSERVED GENERATION   LAST TRANSITION TIME
`,
			expect: true,
		},
		{
			name: "egctl x status gtw -v -A",
			resourceList: &gwv1.GatewayList{
				Items: []gwv1.Gateway{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gtw",
							Namespace: "default",
						},
						Status: gwv1.GatewayStatus{
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
			resourceType:       "gateway",
			quiet:              false,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
default     gtw       foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
                      foobar1   test-status-1   test reason 1   test message 1   123456                2024-01-01 00:00:00 +0000 UTC
`,
			expect: true,
		},
		{
			name: "egctl x status gtw -v -q -A",
			resourceList: &gwv1.GatewayList{
				Items: []gwv1.Gateway{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "gtw1",
							Namespace: "default1",
						},
						Status: gwv1.GatewayStatus{
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
						Status: gwv1.GatewayStatus{
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
			resourceType:       "gateway",
			quiet:              true,
			verbose:            true,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS          REASON          MESSAGE          OBSERVED GENERATION   LAST TRANSITION TIME
default1    gtw1      foobar2   test-status-2   test reason 2   test message 2   123457                2024-01-01 01:00:00 +0000 UTC
default2    gtw2      foobar4   test-status-4   test reason 4   test message 4   123459                2024-01-01 03:00:00 +0000 UTC
`,
			expect: true,
		},
		{
			name: "egctl x status httproute -A",
			resourceList: &gwv1.HTTPRouteList{
				Items: []gwv1.HTTPRoute{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http1",
							Namespace: "default1",
						},
						Status: gwv1.HTTPRouteStatus{
							RouteStatus: gwv1.RouteStatus{
								Parents: []gwv1.RouteParentStatus{
									{
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
						Status: gwv1.HTTPRouteStatus{
							RouteStatus: gwv1.RouteStatus{
								Parents: []gwv1.RouteParentStatus{
									{
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
			resourceType:       "httproute",
			quiet:              false,
			verbose:            false,
			allNamespaces:      true,
			typedName:          false,
			outputs: `NAMESPACE   NAME      TYPE      STATUS          REASON
default1    http1     foobar2   test-status-2   test reason 2
                      foobar1   test-status-1   test reason 1
default2    http2     foobar4   test-status-4   test reason 4
                      foobar3   test-status-3   test reason 3
`,
			expect: true,
		},
		{
			name: "egctl x status httproute -q -n default1",
			resourceList: &gwv1.HTTPRouteList{
				Items: []gwv1.HTTPRoute{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "http1",
							Namespace: "default1",
						},
						Status: gwv1.HTTPRouteStatus{
							RouteStatus: gwv1.RouteStatus{
								Parents: []gwv1.RouteParentStatus{
									{
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
						Status: gwv1.HTTPRouteStatus{
							RouteStatus: gwv1.RouteStatus{
								Parents: []gwv1.RouteParentStatus{
									{
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
			resourceType:       "httproute",
			quiet:              true,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			namespace:          "default1",
			outputs: `NAME      TYPE      STATUS          REASON
http1     foobar2   test-status-2   test reason 2
http2     foobar4   test-status-4   test reason 4
`,
			expect: true,
		},
		{
			name: "egctl x status btlspolicy",
			resourceList: &gwv1a2.BackendTLSPolicyList{
				Items: []gwv1a2.BackendTLSPolicy{
					{
						ObjectMeta: metav1.ObjectMeta{
							Name:      "btls",
							Namespace: "default",
						},
						Status: gwv1a2.PolicyStatus{
							Ancestors: []gwv1a2.PolicyAncestorStatus{
								{
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
			resourceType:       "btlspolicy",
			quiet:              false,
			verbose:            false,
			allNamespaces:      false,
			typedName:          false,
			outputs: `NAME      TYPE      STATUS          REASON
btls      foobar2   test-status-2   test reason 2
          foobar1   test-status-1   test reason 1
`,
			expect: true,
		},
		// TODO(sh2): add a policy status test for egctl x status cmd
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var out bytes.Buffer
			tab := newStatusTableWriter(&out)

			needNamespace := tc.allNamespaces && tc.resourceNamespaced
			headers := fetchStatusHeaders(tc.verbose, needNamespace)
			bodies, err := fetchStatusBodies(tc.resourceList, tc.resourceType, tc.quiet, tc.verbose, needNamespace, tc.typedName)
			if tc.expect {
				require.NoError(t, err)

				writeStatusTable(tab, headers, bodies)
				err = tab.Flush()
				require.NoError(t, err)

				require.Equal(t, tc.outputs, out.String())
			} else {
				require.EqualError(t, err, tc.outputs)
			}
		})
	}
}
