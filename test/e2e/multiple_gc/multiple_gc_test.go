// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package multiplegc

import (
	"encoding/json"
	"flag"
	"io/fs"
	"testing"

	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/conformance/utils/tlog"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e"
	"github.com/envoyproxy/gateway/test/e2e/tests"
	kubetest "github.com/envoyproxy/gateway/test/utils/kubernetes"
)

func TestMultipleGC(t *testing.T) {
	flag.Parse()
	c, cfg := kubetest.NewClient(t)
	recorder := e2e.NewTimingRecorder()
	t.Cleanup(func() {
		recorder.Report(t)
	})

	suiteOpts := suite.ConfigurableOptions{}
	flags.ApplyAll(&suiteOpts)
	data, _ := json.MarshalIndent(suiteOpts, "", "  ")
	tlog.Logf(t, "Running MultipleGC tests with options: %s\n", string(data))
	suiteOpts.TimeoutConfig = tests.TimeoutConfig()
	// SupportedFeatures cannot be empty, so we set it to SupportGateway
	// All e2e tests should leave Features empty.
	suiteOpts.SupportedFeatures = []features.FeatureName{features.SupportGateway}
	suiteOpts.SkipTests = []string{}
	suiteOpts.FailFast = true
	suiteOpts.CleanupTestResources = true
	t.Run("Internet GC Test", func(t *testing.T) {
		t.Parallel()
		internetGatewaySuiteGatewayClassName := "internet"
		suiteOpts.GatewayClassName = internetGatewaySuiteGatewayClassName
		internetGatewaySuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
			Client:              c,
			RestConfig:          cfg,
			Hook:                e2e.Hook,
			ConfigurableOptions: suiteOpts,
		})
		if err != nil {
			t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
		}

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		internetGatewaySuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
		internetGatewaySuite.Applier.GatewayClass = internetGatewaySuiteGatewayClassName
		internetGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, internetGatewaySuite.Client, internetGatewaySuite.TimeoutConfig, internetGatewaySuite.GatewayClassName)

		timedTests := e2e.WrapConformanceTestsWithTiming(tests.MultipleGCTests[internetGatewaySuiteGatewayClassName], recorder)
		tlog.Logf(t, "Running %d MultipleGC tests", len(tests.MultipleGCTests[internetGatewaySuiteGatewayClassName]))

		err = internetGatewaySuite.Run(t, timedTests)
		if err != nil {
			t.Fatalf("Failed to run InternetGC tests: %v", err)
		}
	})

	t.Run("Private GC Test", func(t *testing.T) {
		t.Parallel()
		privateGatewaySuiteGatewayClassName := "private"
		suiteOpts.GatewayClassName = privateGatewaySuiteGatewayClassName
		privateGatewaySuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
			Client:              c,
			RestConfig:          cfg,
			ConfigurableOptions: suiteOpts,
			Hook:                e2e.Hook,
		})
		if err != nil {
			t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
		}

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		privateGatewaySuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
		privateGatewaySuite.Applier.GatewayClass = privateGatewaySuiteGatewayClassName
		privateGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, privateGatewaySuite.Client, privateGatewaySuite.TimeoutConfig, privateGatewaySuite.GatewayClassName)

		timedTests := e2e.WrapConformanceTestsWithTiming(tests.MultipleGCTests[privateGatewaySuiteGatewayClassName], recorder)
		tlog.Logf(t, "Running %d MultipleGC tests", len(tests.MultipleGCTests[privateGatewaySuiteGatewayClassName]))
		err = privateGatewaySuite.Run(t, timedTests)
		if err != nil {
			t.Fatalf("Failed to run PrivateGC tests: %v", err)
		}
	})
}
