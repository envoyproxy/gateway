// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package multiplegc

import (
	"flag"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/test/e2e"
	"github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestMultipleGC(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	// Install all the scheme to kubernetes client.
	e2e.CheckInstallScheme(t, c)

	if flags.RunTest != nil && *flags.RunTest != "" {
		t.Logf("Running E2E test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}
	t.Run("Internet GC Test", func(t *testing.T) {
		t.Parallel()
		internetGatewaySuiteGatewayClassName := "internet"
		internetGatewaySuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
			Client:               c,
			GatewayClassName:     internetGatewaySuiteGatewayClassName,
			Debug:                *flags.ShowDebug,
			CleanupBaseResources: *flags.CleanupBaseResources,
			RunTest:              *flags.RunTest,
			// SupportedFeatures cannot be empty, so we set it to SupportGateway
			// All e2e tests should leave Features empty.
			SupportedFeatures: sets.New[features.SupportedFeature](features.SupportGateway),
			SkipTests:         []string{},
		})
		if err != nil {
			t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
		}

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		internetGatewaySuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
		internetGatewaySuite.Applier.GatewayClass = internetGatewaySuiteGatewayClassName
		internetGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, internetGatewaySuite.Client, internetGatewaySuite.TimeoutConfig, internetGatewaySuite.GatewayClassName)

		t.Logf("Running %d MultipleGC tests", len(tests.MultipleGCTests[internetGatewaySuiteGatewayClassName]))

		err = internetGatewaySuite.Run(t, tests.MultipleGCTests[internetGatewaySuiteGatewayClassName])
		if err != nil {
			t.Fatalf("Failed to run InternetGC tests: %v", err)
		}
	})

	t.Run("Private GC Test", func(t *testing.T) {
		t.Parallel()
		privateGatewaySuiteGatewayClassName := "private"
		privateGatewaySuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
			Client:               c,
			GatewayClassName:     privateGatewaySuiteGatewayClassName,
			Debug:                *flags.ShowDebug,
			CleanupBaseResources: *flags.CleanupBaseResources,
			RunTest:              *flags.RunTest,
			// SupportedFeatures cannot be empty, so we set it to SupportGateway
			// All e2e tests should leave Features empty.
			SupportedFeatures: sets.New[features.SupportedFeature](features.SupportGateway),
			SkipTests:         []string{},
		})
		if err != nil {
			t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
		}

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		privateGatewaySuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
		privateGatewaySuite.Applier.GatewayClass = privateGatewaySuiteGatewayClassName
		privateGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, privateGatewaySuite.Client, privateGatewaySuite.TimeoutConfig, privateGatewaySuite.GatewayClassName)

		t.Logf("Running %d MultipleGC tests", len(tests.MultipleGCTests[privateGatewaySuiteGatewayClassName]))
		err = privateGatewaySuite.Run(t, tests.MultipleGCTests[privateGatewaySuiteGatewayClassName])
		if err != nil {
			t.Fatalf("Failed to run PrivateGC tests: %v", err)
		}
	})
}
