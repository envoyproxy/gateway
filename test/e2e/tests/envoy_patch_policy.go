// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyPatchPolicyTest)
}

var EnvoyPatchPolicyTest = suite.ConformanceTest{
	ShortName:   "EnvoyPatchPolicy",
	Description: "update xds using EnvoyPatchPolicy",
	Manifests:   []string{"testdata/envoy-patch-policy.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		for _, gtw := range []string{"same-namespace", "all-namespaces"} {
			t.Run(gtw, func(t *testing.T) {
				testEnvoyPatchPolicy(t, suite, gtw)
			})
		}

		t.Run("Stale status must be cleaned up", func(t *testing.T) {
			eppNN := types.NamespacedName{Name: "custom-response-patch-policy", Namespace: "gateway-conformance-infra"}
			for _, gtw := range []string{"same-namespace", "all-namespaces"} {
				EnvoyPatchPolicyMustBeAccepted(t, suite,
					eppNN,
					suite.ControllerName, gwapiv1.ParentReference{
						Name:      gwapiv1.ObjectName(gtw),
						Namespace: new(gwapiv1.Namespace("gateway-conformance-infra")),
						Kind:      new(gwapiv1.Kind("Gateway")),
						Group:     new(gwapiv1.Group("gateway.networking.k8s.io")),
					})
			}

			// let's remove same-namespace from targetRefs and check if the status is updated accordingly
			err := wait.PollUntilContextTimeout(t.Context(), suite.TimeoutConfig.DefaultPollInterval, suite.TimeoutConfig.MaxTimeToConsistency, true, func(_ context.Context) (done bool, err error) {
				epp := &egv1a1.EnvoyPatchPolicy{}
				if err := suite.Client.Get(t.Context(), eppNN, epp); err != nil {
					return false, err
				}

				epp.Spec.TargetRefs = []gwapiv1.LocalPolicyTargetReference{
					{
						Group: "gateway.networking.k8s.io",
						Kind:  "Gateway",
						Name:  "same-namespace",
					},
				}

				if err := suite.Client.Update(t.Context(), epp); err != nil {
					return false, err
				}
				return true, nil
			})
			require.NoErrorf(t, err, "failed to update EnvoyPatchPolicy to remove all-namespaces from targetRefs: %v", err)

			// check if the status is updated and there is no stale status for all-namespaces
			EnvoyPatchPolicyMustBeAccepted(t, suite,
				eppNN,
				suite.ControllerName, gwapiv1.ParentReference{
					Name:      "same-namespace",
					Namespace: new(gwapiv1.Namespace("gateway-conformance-infra")),
					Kind:      new(gwapiv1.Kind("Gateway")),
					Group:     new(gwapiv1.Group("gateway.networking.k8s.io")),
				})

			EnvoyPatchPolicyMustNotHaveAncestor(t, suite,
				eppNN,
				suite.ControllerName, gwapiv1.ParentReference{
					Name:      "all-namespaces",
					Namespace: new(gwapiv1.Namespace("gateway-conformance-infra")),
					Kind:      new(gwapiv1.Kind("Gateway")),
					Group:     new(gwapiv1.Group("gateway.networking.k8s.io")),
				})
		})
	},
}

func testEnvoyPatchPolicy(t *testing.T, suite *suite.ConformanceTestSuite, gtwName string) {
	ns := "gateway-conformance-infra"
	routeNN := types.NamespacedName{Name: "http-envoy-patch-policy", Namespace: ns}
	gwNN := types.NamespacedName{Name: gtwName, Namespace: ns}
	gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)
	OkResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/foo",
		},
		Response: http.Response{
			StatusCodes: []int{200},
		},
		Namespace: ns,
	}

	// Send a request to a valid path and expect a successful response
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

	customResp := http.ExpectedResponse{
		Request: http.Request{
			Path: "/bar",
		},
		Response: http.Response{
			StatusCodes: []int{406},
		},
		Namespace: ns,
	}

	// Send a request to an invalid path and expect a custom response
	http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, customResp)
}
