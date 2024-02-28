// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package e2e

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestE2E(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	client, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	require.NoError(t, gwapiv1a2.AddToScheme(client.Scheme()))
	require.NoError(t, gwapiv1.AddToScheme(client.Scheme()))
	require.NoError(t, egv1a1.AddToScheme(client.Scheme()))

	t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t\n supported features: [%v]\n exempt features: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.SupportedFeatures, *flags.ExemptFeatures)

	cSuite := suite.New(suite.Options{
		Client:               client,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		FS:                   &Manifests,
	})

	cSuite.Setup(t)
	t.Logf("Running %d E2E tests", len(tests.ConformanceTests))
	cSuite.Run(t, tests.ConformanceTests)

	// E2E tests for other GatewayClasses.
	NewE2ETestForGatewayClass(t, "MergeGateways E2E", "merge-gateways", client, tests.MergeGatewaysTests)
}

// NewE2ETestForGatewayClass creates a new e2e test for gateway class based on the base e2e resources.
func NewE2ETestForGatewayClass(t *testing.T, testName, gatewayClassName string, client client.Client, testSet []suite.ConformanceTest) {
	t.Run(testName, func(t *testing.T) {
		t.Logf("Running E2E tests with %s GatewayClass\n cleanup: %t\n debug: %t\n supported features: [%v]\n exempt features: [%v]",
			gatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.SupportedFeatures, *flags.ExemptFeatures)

		newSuite := suite.New(suite.Options{
			Client:           client,
			GatewayClassName: gatewayClassName,
			Debug:            *flags.ShowDebug,
		})

		// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
		// since this test suite reuse the base resources of previous test suite.
		newSuite.Applier.FS = Manifests
		newSuite.Applier.GatewayClass = gatewayClassName
		newSuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t,
			newSuite.Client, newSuite.TimeoutConfig, newSuite.GatewayClassName)

		t.Logf("Running %d E2E tests for GatewayClass: %s", len(testSet), gatewayClassName)
		newSuite.Run(t, testSet)
	})
}
