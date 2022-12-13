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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

var useUniquePorts = flag.Bool("use-unique-ports", true, "whether to use unique ports")

func TestGatewayAPIConformance(t *testing.T) {
	flag.Parse()

	cfg, err := config.GetConfig()
	require.NoError(t, err)

	client, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	require.NoError(t, v1alpha2.AddToScheme(client.Scheme()))
	require.NoError(t, v1beta1.AddToScheme(client.Scheme()))

	validUniqueListenerPorts := []v1alpha2.PortNumber{
		v1alpha2.PortNumber(int32(80)),
		v1alpha2.PortNumber(int32(81)),
		v1alpha2.PortNumber(int32(82)),
		v1alpha2.PortNumber(int32(83)),
	}

	if !*useUniquePorts {
		validUniqueListenerPorts = []v1alpha2.PortNumber{}
	}

	cSuite := suite.New(suite.Options{
		Client:                   client,
		GatewayClassName:         *flags.GatewayClassName,
		Debug:                    *flags.ShowDebug,
		CleanupBaseResources:     *flags.CleanupBaseResources,
		ValidUniqueListenerPorts: validUniqueListenerPorts,
		SupportedFeatures: map[suite.SupportedFeature]bool{
			suite.SupportHTTPRouteQueryParamMatching:    true,
			suite.SupportReferenceGrant:                 true,
			suite.SupportHTTPResponseHeaderModification: true,
		},
	})
	cSuite.Setup(t)
	egTests := []suite.ConformanceTest{
		tests.HTTPRouteSimpleSameNamespace,
		tests.HTTPRouteRequestHeaderModifier,
		tests.HTTPRouteResponseHeaderModifier,
		tests.HTTPRouteQueryParamMatching,
		tests.HTTPRouteInvalidCrossNamespaceParentRef,
		tests.HTTPExactPathMatching,
		tests.HTTPRouteCrossNamespace,
		tests.HTTPRouteHeaderMatching,
		tests.HTTPRouteMatchingAcrossRoutes,
		tests.HTTPRouteHostnameIntersection,
		tests.HTTPRouteListenerHostnameMatching,
		tests.HTTPRouteInvalidNonExistentBackendRef,
		tests.HTTPRouteInvalidBackendRefUnknownKind,
		tests.HTTPRouteInvalidCrossNamespaceBackendRef,
		tests.GatewaySecretReferenceGrantAllInNamespace,
		tests.GatewaySecretReferenceGrantSpecific,
		tests.GatewaySecretMissingReferenceGrant,
		tests.GatewaySecretInvalidReferenceGrant,
	}
	cSuite.Run(t, egTests)

}
