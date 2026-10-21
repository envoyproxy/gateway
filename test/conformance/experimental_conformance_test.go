// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build experimental

package conformance

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	conformancev1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"
)

func TestExperimentalConformance(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	opts := conformanceOpts(t)

	opts.ConformanceProfiles = []suite.ConformanceProfileName{
		suite.GatewayHTTPConformanceProfileName,
		suite.GatewayTLSConformanceProfileName,
		suite.GatewayGRPCConformanceProfileName,
	}

	// If focusing on a single test, clear the skip list to ensure it runs.
	if opts.RunTest != "" {
		opts.SkipTests = nil
	}
	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("error creating experimental conformance test suite: %v", err)
	}

	// Since gateway-api conformance v1.6.2, Run() marks the suite as not running only from a
	// t.Cleanup it registers itself, because parallel subtests only execute after this function
	// returns. Cleanups run last-in-first-out, so register the report generation first so that
	// it runs last, after every subtest has completed. Calling Report() inline after Run() always
	// fails with "the test suite is currently running".
	t.Cleanup(func() {
		report, err := cSuite.Report()
		require.NoError(t, err, "error generating conformance profile report")
		require.NoError(t, experimentalConformanceReport(t.Logf, report, opts.ReportOutputPath))
	})

	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("error running conformance tests: %v", err)
	}
}

func experimentalConformanceReport(logf func(string, ...any), report *conformancev1.ConformanceReport, output string) error {
	rawReport, err := yaml.Marshal(report)
	if err != nil {
		return err
	}

	if output != "" {
		if err = os.WriteFile(output, rawReport, 0o600); err != nil {
			return err
		}
	}
	logf("Conformance report:\n%s", string(rawReport))

	return nil
}
