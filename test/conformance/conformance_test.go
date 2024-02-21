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
	v1 "sigs.k8s.io/gateway-api/apis/v1"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	client, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	clientset, err := kubernetes.NewForConfig(cfg)
	require.NoError(t, err)

	require.NoError(t, v1alpha2.AddToScheme(client.Scheme()))
	require.NoError(t, v1beta1.AddToScheme(client.Scheme()))
	require.NoError(t, v1.AddToScheme(client.Scheme()))

	cSuite := suite.New(suite.Options{
		Client:               client,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		Clientset:            clientset,
		CleanupBaseResources: *flags.CleanupBaseResources,
		SupportedFeatures:    suite.AllFeatures,
		SkipTests: []string{
			tests.GatewayStaticAddresses.ShortName,
		},
		ExemptFeatures: suite.MeshCoreFeatures,
	})
	cSuite.Setup(t)
	cSuite.Run(t, tests.ConformanceTests)

}
