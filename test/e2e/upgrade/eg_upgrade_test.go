// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package upgrade

import (
	"flag"
	"io/fs"
	"os"
	"testing"

	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e"
	"github.com/envoyproxy/gateway/test/e2e/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestEGUpgrade(t *testing.T) {
	flag.Parse()
	log.SetLogger(zap.New(zap.WriteTo(os.Stderr), zap.UseDevMode(true)))

	c, cfg := kubetest.NewClient(t)

	suiteOpts := suite.ConfigurableOptions{}
	flags.ApplyAll(&suiteOpts)
	if suiteOpts.RunTest != "" {
		tlog.Logf(t, "Running E2E test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			suiteOpts.RunTest, suiteOpts.GatewayClassName, suiteOpts.CleanupBaseResources, suiteOpts.Debug)
	} else {
		tlog.Logf(t, "Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			suiteOpts.GatewayClassName, suiteOpts.CleanupBaseResources, suiteOpts.Debug)
	}
	suiteOpts.TimeoutConfig = tests.TimeoutConfig()
	suiteOpts.SupportedFeatures = []features.FeatureName{features.SupportGateway}
	suiteOpts.FailFast = true

	var skipTests []string
	// previous did not support ipv6, so skip upgrade tests for ipv6
	if tests.IPFamily == "ipv6" {
		skipTests = append(skipTests,
			tests.EGUpgradeTest.ShortName,
		)
	}
	suiteOpts.SkipTests = skipTests

	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:              c,
		RestConfig:          cfg,
		ManifestFS:          []fs.FS{e2e.UpgradeManifests},
		BaseManifests:       "upgrade/manifests.yaml",
		Hook:                e2e.Hook,
		ConfigurableOptions: suiteOpts,
	})
	if err != nil {
		t.Fatalf("Failed to create test suite: %v", err)
	}

	// upgrade tests should be executed in a specific order
	tests.UpgradeTests = []suite.ConformanceTest{
		tests.EnvoyShutdownTest,
		tests.EGUpgradeTest,
	}

	recorder := e2e.NewTimingRecorder()
	t.Cleanup(func() {
		recorder.Report(t)
	})
	timedTests := e2e.WrapConformanceTestsWithTiming(tests.UpgradeTests, recorder)
	tlog.Logf(t, "Running %d Upgrade tests", len(tests.UpgradeTests))
	cSuite.Setup(t, timedTests)

	err = cSuite.Run(t, timedTests)
	if err != nil {
		t.Fatalf("Failed to run tests: %v", err)
	}
}
