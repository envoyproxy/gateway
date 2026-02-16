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
	ConformanceTests = append(ConformanceTests, ClientTimeoutTest)
}

var ClientTimeoutTest = suite.ConformanceTest{
	ShortName:   "ClientTimeout",
	Description: "Test that the ClientTrafficPolicy API implementation supports client timeout",
	Manifests:   []string{"testdata/client-timeout.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("http client timeout", func(t *testing.T) {
			routeNN := types.NamespacedName{Name: "http-client-timeout", Namespace: ConformanceInfraNamespace}
			gwAddr := kubernetes.GatewayAndHTTPRoutesMustBeAccepted(t, suite.Client, suite.TimeoutConfig, suite.ControllerName,
				SameNamespaceGatewayRef, routeNN)

			ExpectRequestTimeout(t, suite, gwAddr, "/client-timeout", "", http.StatusGatewayTimeout)
		})
	},
}
