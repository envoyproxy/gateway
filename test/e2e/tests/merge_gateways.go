// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"net"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	MergeGatewaysTests = append(MergeGatewaysTests, MergeGatewaysTest)
}

var MergeGatewaysTest = suite.ConformanceTest{
	ShortName:   "BasicMergeGateways",
	Description: "Basic test for MergeGateways feature",
	Manifests:   []string{"testdata/basic-merge-gateways.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		ctx := context.Background()
		ns := "gateway-conformance-infra"

		route1NN := types.NamespacedName{Name: "merged-gateway-route-1", Namespace: ns}
		gw1NN := types.NamespacedName{Name: "merged-gateway-1", Namespace: ns}
		gw1HostPort := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gw1NN), route1NN)

		route2NN := types.NamespacedName{Name: "merged-gateway-route-2", Namespace: ns}
		gw2NN := types.NamespacedName{Name: "merged-gateway-2", Namespace: ns}
		gw2HostPort := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gw2NN), route2NN)

		route3NN := types.NamespacedName{Name: "merged-gateway-route-3", Namespace: ns}
		gw3NN := types.NamespacedName{Name: "merged-gateway-3", Namespace: ns}
		gw3HostPort := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gw3NN), route3NN)

		// Conflicted Gateway and HTTPRoute name
		route4NN := types.NamespacedName{Name: "merged-gateway-route-4", Namespace: ns}
		gw4NN := types.NamespacedName{Name: "merged-gateway-4", Namespace: ns}

		gw1Addr, _, err := net.SplitHostPort(gw1HostPort)
		if err != nil {
			t.Errorf("failed to split hostport %s of gateway %s: %v", gw1HostPort, gw1NN.String(), err)
		}

		gw2Addr, _, err := net.SplitHostPort(gw2HostPort)
		if err != nil {
			t.Errorf("failed to split hostport %s of gateway %s: %v", gw2HostPort, gw2NN.String(), err)
		}

		gw3Addr, _, err := net.SplitHostPort(gw3HostPort)
		if err != nil {
			t.Errorf("failed to split hostport %s of gateway %s: %v", gw3HostPort, gw3NN.String(), err)
		}

		if gw1Addr != gw2Addr || gw2Addr != gw3Addr {
			t.Errorf("inconsistent gateway address %s: %s, %s: %s and %s: %s",
				gw1NN.String(), gw1Addr, gw2NN.String(), gw2Addr, gw3NN.String(), gw3Addr)
			t.FailNow()
		}

		t.Run("merged three gateways under the same namespace with http routes", func(t *testing.T) {
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw1HostPort, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge1", Host: "www.example1.com"},
				Response:  http.Response{StatusCode: 200},
				Namespace: ns,
				Backend:   "infra-backend-v1",
			})

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw2HostPort, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge2", Host: "www.example2.com"},
				Response:  http.Response{StatusCode: 200},
				Namespace: ns,
				Backend:   "infra-backend-v2",
			})

			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw3HostPort, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge3", Host: "www.example3.com"},
				Response:  http.Response{StatusCode: 200},
				Namespace: ns,
				Backend:   "infra-backend-v3",
			})
		})

		t.Run("apply a gateway with conflicted listener and a httproute", func(t *testing.T) {
			// Manually create the conflicted Gateway and HTTPRoute resources,
			// in order to make sure the conflicted status will certainly surface for them.

			// Create Gateway first to make sure there's no HTTPRoute attach to it.
			conflictedGateway := gwapiv1.Gateway{
				TypeMeta: metav1.TypeMeta{
					Kind:       resource.KindGateway,
					APIVersion: gwapiv1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      gw4NN.Name,
					Namespace: gw4NN.Namespace,
				},
				Spec: gwapiv1.GatewaySpec{
					GatewayClassName: gwapiv1.ObjectName(suite.GatewayClassName),
					Listeners: []gwapiv1.Listener{
						{
							AllowedRoutes: &gwapiv1.AllowedRoutes{
								Namespaces: &gwapiv1.RouteNamespaces{
									From: gatewayapi.FromNamespacesPtr(gwapiv1.NamespacesFromSame),
								},
							},
							Name:     "http3",
							Port:     8082,
							Protocol: gwapiv1.HTTPProtocolType,
						},
					},
				},
			}

			if err := suite.Client.Create(ctx, &conflictedGateway); err != nil {
				t.Errorf("failed to create conflicted gateway: %v", err)
			}

			conflictedListenerStatus := []gwapiv1.ListenerStatus{{
				Name: gwapiv1.SectionName("http3"),
				SupportedKinds: []gwapiv1.RouteGroupKind{
					{
						Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
						Kind:  resource.KindHTTPRoute,
					},
					{
						Group: gatewayapi.GroupPtr(gwapiv1.GroupName),
						Kind:  resource.KindGRPCRoute,
					},
				},
				Conditions: []metav1.Condition{{
					Type:   string(gwapiv1.ListenerConditionConflicted),
					Status: metav1.ConditionTrue,
					Reason: string(gwapiv1.ListenerReasonHostnameConflict),
				}},
				AttachedRoutes: 0,
			}}
			kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gw4NN, conflictedListenerStatus)

			// Create HTTPRoute at last to make sure it will be referenced by the conflicted listener in Gateway.
			conflictedHTTPRoute := gwapiv1.HTTPRoute{
				TypeMeta: metav1.TypeMeta{
					Kind:       resource.KindHTTPRoute,
					APIVersion: gwapiv1.GroupVersion.String(),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      route4NN.Name,
					Namespace: route4NN.Namespace,
				},
				Spec: gwapiv1.HTTPRouteSpec{
					CommonRouteSpec: gwapiv1.CommonRouteSpec{
						ParentRefs: []gwapiv1.ParentReference{
							{
								Name: gwapiv1.ObjectName(gw4NN.Name),
							},
						},
					},
					Hostnames: []gwapiv1.Hostname{
						"www.example4.com",
					},
					Rules: []gwapiv1.HTTPRouteRule{
						{
							BackendRefs: []gwapiv1.HTTPBackendRef{
								{
									BackendRef: gwapiv1.BackendRef{
										BackendObjectReference: gwapiv1.BackendObjectReference{
											Group: gatewayapi.GroupPtr(""),
											Kind:  gatewayapi.KindPtr(resource.KindService),
											Name:  "infra-backend-v3",
											Port:  gatewayapi.PortNumPtr(8080),
										},
										Weight: ptr.To[int32](1),
									},
								},
							},
							Matches: []gwapiv1.HTTPRouteMatch{
								{
									Path: &gwapiv1.HTTPPathMatch{
										Type:  ptr.To(gwapiv1.PathMatchPathPrefix),
										Value: ptr.To("/merge4"),
									},
								},
							},
						},
					},
				},
			}

			if err := suite.Client.Create(ctx, &conflictedHTTPRoute); err != nil {
				t.Errorf("failed to create conflicted httproute: %v", err)
				t.FailNow()
			}

			conflictedListenerStatus[0].AttachedRoutes = 1
			kubernetes.GatewayStatusMustHaveListeners(t, suite.Client, suite.TimeoutConfig, gw4NN, conflictedListenerStatus)

			expectedHTTPRouteCondition := metav1.Condition{
				Type:   string(gwapiv1.RouteConditionAccepted),
				Status: metav1.ConditionFalse,
				Reason: "NoReadyListeners",
			}
			kubernetes.HTTPRouteMustHaveCondition(t, suite.Client, suite.TimeoutConfig, route4NN, gw4NN, expectedHTTPRouteCondition)
		})

		t.Run("gateway with conflicted listener cannot be merged", func(t *testing.T) {
			gw4HostPort, err := kubernetes.WaitForGatewayAddress(t, suite.Client, suite.TimeoutConfig, kubernetes.GatewayRef{NamespacedName: gw4NN})
			if err != nil {
				t.Errorf("failed to get the address of gateway %s", gw4NN.String())
			}

			// Even the gateway cannot be merged, it still has the consistent address.
			gw4Addr, _, err := net.SplitHostPort(gw4HostPort)
			if err != nil {
				t.Errorf("failed to split hostport %s of gateway %s: %v", gw4HostPort, gw4NN.String(), err)
				t.FailNow()
			}

			if gw4Addr != gw1Addr {
				t.Errorf("gateway %s has inconsistent address %s with other gateways %s",
					gw4NN.String(), gw4Addr, gw1Addr)
				t.FailNow()
			}

			// Not merged gateway should not receive any traffic.
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gw4HostPort, http.ExpectedResponse{
				Request:   http.Request{Path: "/merge4", Host: "www.example4.com"},
				Response:  http.Response{StatusCode: 404},
				Namespace: ns,
			})
		})

		// Clean-up the conflicted gateway and route resources.
		t.Cleanup(func() {
			if t.Failed() {
				// Collect and dump every config before removing the created resource.
				CollectAndDump(t, suite.RestConfig)
			}

			conflictedGateway := new(gwapiv1.Gateway)
			conflictedHTTPRoute := new(gwapiv1.HTTPRoute)

			if err := suite.Client.Get(ctx, gw4NN, conflictedGateway); err != nil {
				t.Errorf("failed to get conflicted gateway: %v", err)
			}
			if err := suite.Client.Delete(ctx, conflictedGateway); err != nil {
				t.Errorf("failed to delete conflicted gateway: %v", err)
			}

			if err := suite.Client.Get(ctx, route4NN, conflictedHTTPRoute); err != nil {
				t.Errorf("failed to get conflicted httproute: %v", err)
			}
			if err := suite.Client.Delete(ctx, conflictedHTTPRoute); err != nil {
				t.Errorf("failed to delete conflicted httproute: %v", err)
			}
		})
	},
}
