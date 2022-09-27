//go:build conformance
// +build conformance

package conformance

import (
	"testing"

	"github.com/stretchr/testify/require"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/gateway-api/apis/v1alpha2"
	"sigs.k8s.io/gateway-api/conformance/tests"
	"sigs.k8s.io/gateway-api/conformance/utils/flags"
	"sigs.k8s.io/gateway-api/conformance/utils/suite"
)

func TestGatewayAPIConformance(t *testing.T) {
	cfg, err := config.GetConfig()
	require.NoError(t, err)

	client, err := client.New(cfg, client.Options{})
	require.NoError(t, err)

	require.NoError(t, v1alpha2.AddToScheme(client.Scheme()))

	cSuite := suite.New(suite.Options{
		Client:               client,
		GatewayClassName:     *flags.GatewayClassName,
		Debug:                *flags.ShowDebug,
		CleanupBaseResources: *flags.CleanupBaseResources,
		ValidUniqueListenerPorts: []v1alpha2.PortNumber{
			v1alpha2.PortNumber(int32(80)),
			v1alpha2.PortNumber(int32(81)),
			v1alpha2.PortNumber(int32(82)),
			v1alpha2.PortNumber(int32(83)),
			v1alpha2.PortNumber(int32(84)),
		},
	})
	cSuite.Setup(t)
	egTests := []suite.ConformanceTest{tests.HTTPRouteSimpleSameNamespace}
	cSuite.Run(t, egTests)

}
