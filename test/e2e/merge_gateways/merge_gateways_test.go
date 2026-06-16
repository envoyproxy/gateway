// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package mergegateways

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

func TestMergeGateways(t *testing.T) {
	// Skip the entire test suite if we're in Gateway Namespace Mode
	if tests.IsGatewayNamespaceMode() {
		t.Skip("MergeGateways tests are not supported in Gateway Namespace Mode")
	}
	flag.Parse()

	c, cfg := kubetest.NewClient(t)
	suiteOpts := suite.ConfigurableOptions{}
	flags.ApplyAll(&suiteOpts)
	data, _ := json.MarshalIndent(suiteOpts, "", "  ")
	tlog.Logf(t, "Running MergeGateways tests with options: %s\n", string(data))
	suiteOpts.TimeoutConfig = tests.TimeoutConfig()
	// SupportedFeatures cannot be empty, so we set it to SupportGateway
	// All e2e tests should leave Features empty.
	suiteOpts.SupportedFeatures = []features.FeatureName{features.SupportGateway}
	suiteOpts.SkipTests = []string{}
	suiteOpts.FailFast = true
	suiteOpts.CleanupTestResources = true
	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:              c,
		RestConfig:          cfg,
		ConfigurableOptions: suiteOpts,
	})
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}

	// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
	// since this test suite reuse the base resources of previous test suite.
	cSuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
	cSuite.Applier.GatewayClass = suiteOpts.GatewayClassName
	cSuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, cSuite.Client, cSuite.TimeoutConfig, cSuite.GatewayClassName)

	recorder := e2e.NewTimingRecorder()
	t.Cleanup(func() {
		recorder.Report(t)
	})
	timedTests := e2e.WrapConformanceTestsWithTiming(tests.MergeGatewaysTests, recorder)
	tlog.Logf(t, "Running %d MergeGateways tests", len(tests.MergeGatewaysTests))
	err = cSuite.Run(t, timedTests)
	if err != nil {
		t.Fatalf("Failed to run MergeGateways tests: %v", err)
	}
}
