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
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
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

	require.NoError(t, v1alpha2.AddToScheme(client.Scheme()))
	require.NoError(t, v1beta1.AddToScheme(client.Scheme()))
	require.NoError(t, egv1a1.AddToScheme(client.Scheme()))

	t.Logf("Running conformance tests with %s GatewayClass\n cleanup: %t\n debug: %t\n enable all features: %t \n supported features: [%v]\n exempt features: [%v]",
		*flags.GatewayClassName, *flags.CleanupBaseResources, *flags.ShowDebug, *flags.EnableAllSupportedFeatures, *flags.SupportedFeatures, *flags.ExemptFeatures)

	cSuite := suite.New(suite.Options{
		Client:                     client,
		GatewayClassName:           *flags.GatewayClassName,
		Debug:                      *flags.ShowDebug,
		CleanupBaseResources:       *flags.CleanupBaseResources,
		EnableAllSupportedFeatures: *flags.EnableAllSupportedFeatures,
		FS:                         Manifests,
	})

	cSuite.Setup(t)
	t.Logf("Running %d conformance tests", len(tests.ConformanceTests))
	cSuite.Run(t, tests.ConformanceTests)
}
