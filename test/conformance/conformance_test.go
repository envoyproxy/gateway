// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build conformance
// +build conformance

package conformance

import (
	"flag"
	"testing"

	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	internalconf "github.com/envoyproxy/gateway/internal/gatewayapi/conformance"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()

	opts := conformance.DefaultOptions(t)
<<<<<<< HEAD
	opts.SkipTests = internalconf.EnvoyGatewaySuite.SkipTests
	opts.SupportedFeatures = internalconf.EnvoyGatewaySuite.SupportedFeatures
	opts.ExemptFeatures = internalconf.EnvoyGatewaySuite.ExemptFeatures
=======
	opts.SkipTests = []string{
		tests.GatewayStaticAddresses.ShortName,
		tests.GatewayHTTPListenerIsolation.ShortName, // https://github.com/kubernetes-sigs/gateway-api/issues/3049
	}
	opts.SupportedFeatures = features.AllFeatures
	opts.ExemptFeatures = features.MeshCoreFeatures
>>>>>>> 81edb95a (enable backendRef filters conformance tests)

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}
	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
