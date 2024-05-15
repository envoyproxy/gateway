// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build experimental
// +build experimental

package conformance

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/gateway-api/conformance"
	conformancev1 "sigs.k8s.io/gateway-api/conformance/apis/v1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/yaml"

	internalconf "github.com/envoyproxy/gateway/internal/gatewayapi/conformance"
)

func TestExperimentalConformance(t *testing.T) {
	flag.Parse()

	opts := conformance.DefaultOptions(t)
	opts.SkipTests = internalconf.EnvoyGatewaySuite.SkipTests
	opts.SupportedFeatures = internalconf.EnvoyGatewaySuite.SupportedFeatures
	opts.ExemptFeatures = internalconf.EnvoyGatewaySuite.ExemptFeatures
	opts.ConformanceProfiles = sets.New(
		suite.GatewayHTTPConformanceProfileName,
		suite.GatewayTLSConformanceProfileName,
	)

	t.Logf("Running experimental conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t\n enable all features: %t \n conformance profiles: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, opts.ConformanceProfiles)

	cSuite, err := suite.NewConformanceTestSuite(opts)
	if err != nil {
		t.Fatalf("error creating experimental conformance test suite: %v", err)
	}

	cSuite.Setup(t, tests.ConformanceTests)
	err = cSuite.Run(t, tests.ConformanceTests)
	if err != nil {
		t.Fatalf("error running conformance profile report: %v", err)
	}
	report, err := cSuite.Report()
	if err != nil {
		t.Fatalf("error generating conformance profile report: %v", err)
	}

	err = experimentalConformanceReport(t.Logf, *report, *flags.ReportOutput)
	require.NoError(t, err)
}

func experimentalConformanceReport(logf func(string, ...any), report conformancev1.ConformanceReport, output string) error {
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
