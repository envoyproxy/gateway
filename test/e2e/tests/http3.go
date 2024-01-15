// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package tests

import (
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"testing"
)

func init() {
	ConformanceTests = append(ConformanceTests, Http3Test)
}

var Http3Test = suite.ConformanceTest{
	ShortName:   "Http3",
	Description: "Testing http3 request",
	Manifests:   []string{"testdata/http3.yaml"},
	Test: func(t *testing.T, suite *suite.ConformanceTestSuite) {
		t.Run("Send http3 request", func(t *testing.T) {
			namespace := "gateway-conformance-http3"
			routeNN := types.NamespacedName{Name: "http3-route", Namespace: namespace}
			gwNN := types.NamespacedName{Name: "http3-gateway", Namespace: namespace}

		})
	},
}
