// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build conformance

package conformance

import (
	"flag"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/conformance"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"

	internalconf "github.com/envoyproxy/gateway/internal/gatewayapi/conformance"
	"github.com/envoyproxy/gateway/test/e2e"
	ege2etest "github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	if flags.RunTest != nil && *flags.RunTest != "" {
		tlog.Logf(t, "Running Conformance test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		tlog.Logf(t, "Running Conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}

	gatewayNamespaceMode := ege2etest.IsGatewayNamespaceMode()
	internalSuite := internalconf.EnvoyGatewaySuite(gatewayNamespaceMode)

	opts := conformance.DefaultOptions(t)
	opts.SkipTests = internalSuite.SkipTests
	opts.SupportedFeatures = internalSuite.SupportedFeatures
	opts.ExemptFeatures = internalSuite.ExemptFeatures
	opts.RunTest = *flags.RunTest
	opts.Hook = e2e.Hook

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}
	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
