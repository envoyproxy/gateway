// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/utils/http"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func init() {
	ConformanceTests = append(ConformanceTests, EnvoyProxyHPATest)
}

var EnvoyProxyHPATest = suite.ConformanceTest{
	ShortName:   "EnvoyProxyHPA",
	Description: "Test running Envoy with HPA",
	Manifests:   []string{"testdata/envoyproxy-hpa.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("RunAndDelete", func(t *testing.T) {
			ns := "gateway-conformance-infra"
			routeNN := types.NamespacedName{Name: "foo-route", Namespace: ns}
			gwNN := types.NamespacedName{Name: "eg-hpa", Namespace: ns}
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

			expectedNs := GetGatewayResourceNamespace()

			// Make sure there's a deployment/HPA for the gateway
			ExpectEnvoyProxyDeploymentCount(t, suite, gwNN, expectedNs, 1)
			ExpectEnvoyProxyHPACount(t, suite, gwNN, expectedNs, 1)

			// Send a request to a valid path and expect a successful response
			http.MakeRequestAndExpectEventuallyConsistentResponse(t, suite.RoundTripper, suite.TimeoutConfig, gwAddr, OkResp)

			require.NoError(t, wait.PollUntilContextTimeout(t.Context(), time.Second, suite.TimeoutConfig.MaxTimeToConsistency, true, func(ctx context.Context) (bool, error) {
				// Update the Gateway to without HPA
				gw := &gwapiv1.Gateway{}
				err := suite.Client.Get(ctx, gwNN, gw)
				if err != nil {
					tlog.Logf(t, "failed to get Gateway %s", gwNN)
					return false, nil
				}
				gw.Spec.Infrastructure = nil
				err = suite.Client.Update(ctx, gw)
				if err != nil {
					tlog.Logf(t, "failed to update Gateway %s: %v", gwNN, err)
					return false, nil
				}

				return true, nil
			}))

			ExpectEnvoyProxyDeploymentCount(t, suite, gwNN, expectedNs, 1)
			ExpectEnvoyProxyHPACount(t, suite, gwNN, expectedNs, 0)
		})
	},
}
