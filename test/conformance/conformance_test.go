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
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	suiteOpts := suite.ConfigurableOptions{}
	flags.ApplyAll(&suiteOpts)
	if suiteOpts.RunTest != "" {
		tlog.Logf(t, "Running Conformance test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			suiteOpts.RunTest, suiteOpts.GatewayClassName, suiteOpts.CleanupBaseResources, suiteOpts.Debug)
	} else {
		tlog.Logf(t, "Running Conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			suiteOpts.GatewayClassName, suiteOpts.CleanupBaseResources, suiteOpts.Debug)
	}

	opts := conformanceOpts(t, &suiteOpts)
	opts.RunTest = suiteOpts.RunTest

	// If focusing on a single test, clear the skip list to ensure it runs.
	if opts.RunTest != "" {
		opts.SkipTests = nil
	}

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}
	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
