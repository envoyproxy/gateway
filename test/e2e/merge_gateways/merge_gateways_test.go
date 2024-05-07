// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

//go:build e2e
// +build e2e

package mergegateways

import (
	"flag"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	gwapiv1a3 "sigs.k8s.io/gateway-api/apis/v1alpha3"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/kubernetes"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/test/e2e"
	"github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestMergeGateways(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)
	require.NoError(t, gwapiv1a3.AddToScheme(c.Scheme()))
	require.NoError(t, gwapiv1a2.AddToScheme(c.Scheme()))
	require.NoError(t, gwapiv1.AddToScheme(c.Scheme()))
	require.NoError(t, egv1a1.AddToScheme(c.Scheme()))

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
		RunTest:              *flags.RunTest,
		SkipTests:            []string{},
	})
	if err != nil {
		t.Fatalf("Failed to create ConformanceTestSuite: %v", err)
	}

	// Setting up the necessary arguments for the suite instead of calling Suite.Setup method again,
	// since this test suite reuse the base resources of previous test suite.
	cSuite.Applier.ManifestFS = []fs.FS{e2e.Manifests}
	cSuite.Applier.GatewayClass = *flags.GatewayClassName
	cSuite.ControllerName = kubernetes.GWCMustHaveAcceptedConditionTrue(t, cSuite.Client, cSuite.TimeoutConfig, cSuite.GatewayClassName)

	t.Logf("Running %d MergeGateways tests", len(tests.MergeGatewaysTests))
	err = cSuite.Run(t, tests.MergeGatewaysTests)
	if err != nil {
		t.Fatalf("Failed to run MergeGateways tests: %v", err)
	}
}
