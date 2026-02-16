// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func init() {
	ConformanceTests = append(ConformanceTests, BackendTrafficPolicyMergedTest)
}

var BackendTrafficPolicyMergedTest = suite.ConformanceTest{
	ShortName:   "BackendTrafficPolicyMerged",
	Description: "Test section level policy attach and merged parent policy for BackendTrafficPolicy",
	Manifests:   []string{"testdata/backendtrafficpolicy-section-merged.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("BackendTrafficPolicyMerged", func(t *testing.T) {
			ns := "gateway-conformance-infra"

			sectionRouteNN := types.NamespacedName{Name: "btp-section", Namespace: ns}
			nonsectionRouteNN := types.NamespacedName{Name: "btp-non-section", Namespace: ns}
			gwNN := types.NamespacedName{Name: "btp-section", Namespace: ns}

			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t,
				suite.Client,
				suite.TimeoutConfig,
				suite.ControllerName,
				kubernetes.NewGatewayRef(gwNN),
				sectionRouteNN,
				nonsectionRouteNN,
			)

			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-section-gateway", Namespace: ns},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:        gatewayapi.KindPtr(resource.KindGateway),
					Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:        gwapiv1.ObjectName(gwNN.Name),
					SectionName: ptr.To(gwapiv1.SectionName("listener-1")),
				},
			)

			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-section-route", Namespace: ns},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:        gatewayapi.KindPtr(resource.KindGateway),
					Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:        gwapiv1.ObjectName(gwNN.Name),
					SectionName: ptr.To(gwapiv1.SectionName("listener-1")),
				},
			)

			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-section-route-rule", Namespace: ns},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:        gatewayapi.KindPtr(resource.KindGateway),
					Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:        gwapiv1.ObjectName(gwNN.Name),
					SectionName: ptr.To(gwapiv1.SectionName("listener-1")),
				},
			)

			BackendTrafficPolicyMustBeAccepted(t,
				suite.Client,
				types.NamespacedName{Name: "btp-non-section-route", Namespace: ns},
				suite.ControllerName,
				gwapiv1a2.ParentReference{
					Group:       gatewayapi.GroupPtr(gwapiv1.GroupName),
					Kind:        gatewayapi.KindPtr(resource.KindGateway),
					Namespace:   gatewayapi.NamespacePtr(gwNN.Namespace),
					Name:        gwapiv1.ObjectName(gwNN.Name),
					SectionName: ptr.To(gwapiv1.SectionName("listener-2")),
				},
			)

			// Test 1: HTTPRoute btp-section/rule-2 with section-level FaultInjection (419)
			// This rule has its own BTP without mergeType, so Gateway's ResponseOverride doesn't apply.
			// Expected: StatusCode 419 (FaultInjection abort)
			expectedResponse := http.ExpectedResponse{
				Namespace: ns,
				Request:   http.Request{Host: "listener1.merged.example.com", Path: "/bar"},
				Response:  http.Response{StatusCode: 419},
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Test 2: HTTPRoute btp-section/rule-1 with merged BTP (FaultInjection 418 + ResponseOverride)
			// Route-level FaultInjection (418) with mergeType merges with Gateway's ResponseOverride (418->500).
			// Expected: StatusCode 500 (ResponseOverride transforms 418 to 500)
			expectedResponse = http.ExpectedResponse{
				Namespace: ns,
				Request:   http.Request{Host: "listener1.merged.example.com", Path: "/foo"},
				Response:  http.Response{StatusCode: 500},
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)

			// Test 3: HTTPRoute btp-non-section on listener-2 with merged BTP (FaultInjection 420)
			// This route uses listener-2 which doesn't have Gateway ResponseOverride (listener-1 only).
			// Expected: StatusCode 420 (FaultInjection abort, no ResponseOverride applied)
			expectedResponse = http.ExpectedResponse{
				Namespace: ns,
				Request:   http.Request{Host: "listener2.merged.example.com", Path: "/foo"},
				Response:  http.Response{StatusCode: 420},
			}
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, expectedResponse)
		})
	},
}
