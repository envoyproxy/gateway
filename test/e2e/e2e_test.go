// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package e2e

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

	"github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestE2E(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	// Install all the scheme for kubernetes client.
	CheckInstallScheme(t, c)

	if flags.RunTest != nil && *flags.RunTest != "" {
		t.Logf("Running E2E test %s with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.RunTest, *flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	} else {
		t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t",
			*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug)
	}

	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:               c,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		ManifestFS:           []fs.FS{Manifests},
		RunTest:              *flags.RunTest,
		// SupportedFeatures cannot be empty, so we set it to SupportGateway
		// All e2e tests should leave Features empty.
		SupportedFeatures: sets.New[features.SupportedFeature](features.SupportGateway),
		SkipTests: []string{
			tests.ClientTimeoutTest.ShortName,        // https://github.com/envoyproxy/gateway/issues/2720
			tests.GatewayInfraResourceTest.ShortName, // https://github.com/envoyproxy/gateway/issues/3191
			tests.UseClientProtocolTest.ShortName,    // https://github.com/envoyproxy/gateway/issues/3473
		},
	})
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}

	cSuite.Setup(t, tests.ConformanceTests)
	t.Logf("Running %d E2E tests", len(tests.ConformanceTests))
	err = cSuite.Run(t, tests.ConformanceTests)
	if err != nil {
		t.Fatalf("Failed to run E2E tests: %v", err)
	}

	t.Run("Multiple GatewayClass E2E", func(t *testing.T) {
		internetGatewaySuiteGatewayClassName := "internet"

		t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t\n supported features: [%v]\n exempt features: [%v]",
			internetGatewaySuiteGatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.SupportedFeatures, *flags.ExemptFeatures)

		internetGatewaySuite := suite.New(suite.Options{
			Client:           client,
			GatewayClassName: internetGatewaySuiteGatewayClassName,
			Debug:            *flags.ShowDebug,
		})

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		internetGatewaySuite.Applier.FS = Manifests
		internetGatewaySuite.Applier.GatewayClass = internetGatewaySuiteGatewayClassName
		internetGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, internetGatewaySuite.Client,
			internetGatewaySuite.TimeoutConfig, internetGatewaySuite.GatewayClassName)

		t.Logf("Running %d E2E tests for MergeGateways feature", len(tests.InternetGCTests))
		internetGatewaySuite.Run(t, tests.InternetGCTests)

		privateGatewaySuiteGatewayClassName := "private"

		t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t\n supported features: [%v]\n exempt features: [%v]",
			privateGatewaySuiteGatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.SupportedFeatures, *flags.ExemptFeatures)

		privateGatewaySuite := suite.New(suite.Options{
			Client:           client,
			GatewayClassName: privateGatewaySuiteGatewayClassName,
			Debug:            *flags.ShowDebug,
		})

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		privateGatewaySuite.Applier.FS = Manifests
		privateGatewaySuite.Applier.GatewayClass = privateGatewaySuiteGatewayClassName
		privateGatewaySuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, privateGatewaySuite.Client,
			privateGatewaySuite.TimeoutConfig, privateGatewaySuite.GatewayClassName)

		t.Logf("Running %d E2E tests for MergeGateways feature", len(tests.PrivateGCTests))
		privateGatewaySuite.Run(t, tests.PrivateGCTests)
	})
}
