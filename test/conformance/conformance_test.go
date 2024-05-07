//go:build conformance
// +build conformance

// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package conformance

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
	"sigs.k8s.io/gateway-api/pkg/features"

	"github.com/envoyproxy/gateway/internal/envoygateway"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()

	clientCfg, err := config.GetConfig()
	require.NoError(t, err)

	c, err := client.New(clientCfg, client.Options{Scheme: envoygateway.GetScheme()})
	require.NoError(t, err)

	cs, err := kubernetes.NewForConfig(clientCfg)
	require.NoError(t, err)

	cSuite, err := suite.NewConformanceTestSuite(suite.ConformanceOptions{
		Client:               c,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		Clientset:            cs,
		CleanupBaseResources: *flags.CleanupBaseResources,
		SupportedFeatures:    features.AllFeatures,
		SkipTests: []string{
			tests.GatewayStaticAddresses.ShortName,
		},
		ExemptFeatures: features.MeshCoreFeatures,
	})
	if err != nil {
		t.Fatalf("Error creating conformance test suite: %v", err)
	}
	cSuite.Setup(t, tests.ConformanceTests)
	if err := cSuite.Run(t, tests.ConformanceTests); err != nil {
		t.Fatalf("Error running conformance tests: %v", err)
	}
}
