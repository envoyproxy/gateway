// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build resilience

package tests

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"

	"github.com/envoyproxy/gateway/test/resilience/suite"
)

func init() {
	ResilienceTests = append(ResilienceTests, MultipleExtManagers)
}

var MultipleExtManagers = suite.ResilienceTest{
	ShortName:   "MultipleExtManagers",
	Description: "Multiple extension managers chaining and resource isolation",
	Test: func(t *testing.T, suite *suite.ResilienceTestSuite) {
		ctx := t.Context()
		ap := kubernetes.Applier{
			ManifestFS:     suite.ManifestFS,
			GatewayClass:   suite.GatewayClassName,
			ControllerName: "gateway.envoyproxy.io/gatewayclass-controller",
		}

		// Apply the multi-manager ConfigMap and restart the control plane to pick it up.
		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/config_multiple_ext_managers.yaml", true)
		err := suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 0, time.Minute, false)
		require.NoError(t, err, "Failed to scale down envoy-gateway")
		err = suite.Kube().ScaleDeploymentAndWait(ctx, envoygateway, namespace, 1, time.Minute, false)
		require.NoError(t, err, "Failed to scale up envoy-gateway")

		ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/base.yaml", true)

		localTimeout := suite.TimeoutConfig
		localTimeout.RequiredConsecutiveSuccesses = 2
		localTimeout.MaxTimeToConsistency = time.Minute

		t.Run("chaining applies both VirtualHost mutations in order", func(t *testing.T) {
			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "chaining-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_multiple_ext_managers.yaml", true)
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			t.Log("Verify base route works")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.chain.com",
					Path: "/chain",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})

			t.Log("Verify chained VirtualHost mutations: ext-a appends .ext-a, ext-b appends .ext-b")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.chain.com.ext-a.ext-b",
					Path: "/chain",
				},
				Response: http.Response{
					StatusCodes: []int{200},
				},
				Namespace: ns,
			})
		})

		t.Run("resource isolation: only owning extension receives resources", func(t *testing.T) {
			ns := "gateway-resilience"
			routeNN := types.NamespacedName{Name: "isolation-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "all-namespaces", Namespace: ns}

			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/extension_crds.yaml", true)
			ap.MustApplyWithCleanup(t, suite.Client, suite.TimeoutConfig, "testdata/route_for_resource_isolation.yaml", true)
			gwAddr := kubernetes.GatewayAndRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName, kubernetes.NewGatewayRef(gwNN), &gwapiv1.HTTPRoute{}, false, routeNN)

			t.Log("Verify ext-a processed the route with its extension resource (FooFilter)")
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, localTimeout, gwAddr, http.ExpectedResponse{
				Request: http.Request{
					Host: "www.isolation.com",
					Path: "/isolate",
				},
				Response: http.Response{
					StatusCodes: []int{200},
					Headers: map[string]string{
						"x-ext-server": "ext-a",
					},
				},
				Namespace: ns,
			})
		})
	},
}
