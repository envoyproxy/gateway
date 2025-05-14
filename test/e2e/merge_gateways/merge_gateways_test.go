// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e

package mergegateways

import (
	"flag"
	"io/fs"
	"testing"

	"k8s.io/apimachinery/pkg/util/sets"
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
	flag.Parse()

	c, cfg := kubetest.NewClient(t)

	if flags.RunTest != nil && *flags.RunTest != "" {
		tlog.Logf(t, "Running E2E test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		tlog.Logf(t, "Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}

	var skipTests []string
	if tests.IsGatewayNamespaceMode() {
		// Skip MergeGateways test because it is not supported in GatewayNamespaceMode
		skipTests = append(skipTests, tests.MergeGatewaysTest.ShortName)
	}

	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:               c,
		RestConfig:           cfg,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		RunTest:              *flags.RunTest,
		// SupportedFeatures cannot be empty, so we set it to SupportGateway
		// All e2e tests should leave Features empty.
		SupportedFeatures: sets.New(features.SupportGateway),
		SkipTests:         skipTests,
	})
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}

	// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
	// since this test suite reuse the base resources of previous test suite.
	cSuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
	cSuite.Applier.GatewayClass = *flags.GatewayClassName
	cSuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, cSuite.Client, cSuite.TimeoutConfig, cSuite.GatewayClassName)

	tlog.Logf(t, "Running %d MergeGateways tests", len(tests.MergeGatewaysTests))
	err = cSuite.Run(t, tests.MergeGatewaysTests)
	if err != nil {
		t.Fatalf("Failed to run MergeGateways tests: %v", err)
	}
}
