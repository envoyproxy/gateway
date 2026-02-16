// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package tests

import (
	"net/http"
	"testing"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func init() {
	ConformanceTests = append(ConformanceTests, BTPTimeoutTest)
}

var BTPTimeoutTest = suite.ConformanceTest{
	ShortName:   "BTPTimeout",
	Description: "Test BackendTrafficPolicy timeout",
	Manifests:   []string{"testdata/btp-timeout.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("HTTP", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-btp-timeout", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				SameNamespaceGatewayRef, routeNN)

			ExpectRequestTimeout(t, suite, gwAddr, "/btp-timeout", "", http.StatusOK)
			// Timeout is 2s, so deplay 1s will return 200
			ExpectRequestTimeout(t, suite, gwAddr, "/btp-timeout", "delay=1s", http.StatusOK)
			// Timeout is 2s, so deplay 4s will return 504
			ExpectRequestTimeout(t, suite, gwAddr, "/btp-timeout", "delay=4s", http.StatusGatewayTimeout)
		})

		// TODO: add test for TCP
	},
}
