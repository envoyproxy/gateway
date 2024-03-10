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

	"github.com/envoyproxy/gateway/test/e2e/utils/certificate"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	gwapiv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/test/e2e/tests"
)

func TestE2E(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	c, err := client.New(cfg, client.Options{})
	require.NoError(t, err)
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

	cSuite := suite.New(suite.Options{
		Client:               c,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		FS:                   &Manifests,
		RunTest:              *flags.RunTest,
	})

	cSuite.Setup(t)
	addAdditionalResourcesToSuite(t, cSuite)
	t.Logf("Running %d E2E tests", len(tests.ConformanceTests))
	cSuite.Run(t, tests.ConformanceTests)
}

// set up additional resources that are created and cleaned up programmatically like certificates
func addAdditionalResourcesToSuite(t *testing.T, testSuite *suite.ConformanceTestSuite) {
	secret, configmap := certificate.MustCreateSelfSignedCAConfigmapAndCertSecret(t, "gateway-conformance-infra", "backend-tls-checks-certificate", []string{"example.com"})
	testSuite.Applier.MustApplyObjectsWithCleanup(t, testSuite.Client, testSuite.TimeoutConfig, []client.Object{secret}, testSuite.Cleanup)
	testSuite.Applier.MustApplyObjectsWithCleanup(t, testSuite.Client, testSuite.TimeoutConfig, []client.Object{configmap}, testSuite.Cleanup)
}
