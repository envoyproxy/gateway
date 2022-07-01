package gatewayapi

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/gateway-api/apis/v1beta1"
	"sigs.k8s.io/yaml"
)

type StaticCache struct {
	Gateways   []*v1beta1.Gateway
	HTTPRoutes []*v1beta1.HTTPRoute
}

func (c *StaticCache) ListGateways() []*v1beta1.Gateway {
	return c.Gateways
}

func (c *StaticCache) ListHTTPRoutes() []*v1beta1.HTTPRoute {
	return c.HTTPRoutes
}

func (c *StaticCache) GetNamespace(name string) *v1.Namespace {
	return &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
}

func (c *StaticCache) GetService(namespace, name string) *v1.Service {
	return &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1.ServiceSpec{
			ClusterIP: "7.7.7.7",
			Ports: []v1.ServicePort{
				{Port: 8080},
				{Port: 8443},
			},
		},
	}
}

func mustUnmarshal(t *testing.T, val string, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict([]byte(val), out, yaml.DisallowUnknownFields))
}

func TestTranslate(t *testing.T) {
	tests := map[string]struct {
		cache string
		want  string
	}{
		// Route-Gateway attachment
		"Gateway with one HTTP Listener, HTTPRoute attaching to the Gateway": {
			cache: BasicHTTPRouteAttachingToGatewayIn,
			want:  BasicHTTPRouteAttachingToGatewayOut,
		},

		"Gateway with one HTTP Listener, HTTPRoute attaching to the Listener": {
			cache: BasicHTTPRouteAttachingToListenerIn,
			want:  BasicHTTPRouteAttachingToListenerOut,
		},

		"Gateway that allows HTTPRoutes from the same namespace, HTTPRoute in the same namespace": {
			cache: GatewayAllowsSameNamespaceWithAllowedHTTPRouteIn,
			want:  GatewayAllowsSameNamespaceWithAllowedHTTPRouteOut,
		},

		"Gateway that allows HTTPRoutes from the same namespace, HTTPRoute not in the same namespace": {
			cache: GatewayAllowsSameNamespaceWithDisallowedHTTPRouteIn,
			want:  GatewayAllowsSameNamespaceWithDisallowedHTTPRouteOut,
		},

		"Gateway with two HTTP Listeners, HTTPRoute attaching to the Gateway": {
			cache: HTTPRouteAttachingToGatewayWithTwoListenersIn,
			want:  HTTPRouteAttachingToGatewayWithTwoListenersOut,
		},

		"Gateway with two HTTP Listeners, HTTPRoute attaching to one Listener": {
			cache: HTTPRouteAttachingToListenerOnGatewayWithTwoListenersIn,
			want:  HTTPRouteAttachingToListenerOnGatewayWithTwoListenersOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with matching specific hostname": {
			cache: HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameIn,
			want:  HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with two matching specific hostnames": {
			cache: HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameIn,
			want:  HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with non-matching specific hostname": {
			cache: HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameIn,
			want:  HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameOut,
		},

		// Gateway/Listener error cases
		"Gateway with one Listener with protocol other than HTTP": {
			cache: GatewayWithListenerWithNonHTTPProtocolIn,
			want:  GatewayWithListenerWithNonHTTPProtocolOut,
		},

		"Gateway with one Listener with missing allowed namespaces selector": {
			cache: GatewayWithListenerWithMissingAllowedNamespacesSelectorIn,
			want:  GatewayWithListenerWithMissingAllowedNamespacesSelectorOut,
		},

		"Gateway with one Listener with invalid allowed namespaces selector": {
			cache: GatewayWithListenerWithInvalidAllowedNamespacesSelectorIn,
			want:  GatewayWithListenerWithInvalidAllowedNamespacesSelectorOut,
		},

		"Gateway with one Listener with invalid allowed routes group": {
			cache: GatewayWithListenerWithInvalidAllowedRoutesGroupIn,
			want:  GatewayWithListenerWithInvalidAllowedRoutesGroupOut,
		},

		"Gateway with one Listener with invalid allowed routes kind": {
			cache: GatewayWithListenerWithInvalidAllowedRoutesKindIn,
			want:  GatewayWithListenerWithInvalidAllowedRoutesKindOut,
		},

		"Gateway with two Listeners with the same port and hostname": {
			cache: GatewayWithTwoListenersWithSamePortAndHostnameIn,
			want:  GatewayWithTwoListenersWithSamePortAndHostnameOut,
		},

		"Gateway with two Listeners with the same port and incompatible protocols": {
			cache: GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsIn,
			want:  GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsOut,
		},

		// Route matches
		"HTTPRoute with single rule with path prefix and exact header matches": {
			cache: HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesIn,
			want:  HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesOut,
		},

		"HTTPRoute with single rule with exact path match": {
			cache: HTTPRouteWithSingleRuleWithExactPathMatchIn,
			want:  HTTPRouteWithSingleRuleWithExactPathMatchOut,
		},

		// Route backends
		"HTTPRoute rule with multiple backends, no weights explicitly specified": {
			cache: HTTPRouteRuleWithMultipleBackendsAndNoWeightsIn,
			want:  HTTPRouteRuleWithMultipleBackendsAndNoWeightsOut,
		},

		"HTTPRoute rule with multiple backends, weights explicitly specified": {
			cache: HTTPRouteRuleWithMultipleBackendsAndWeightsIn,
			want:  HTTPRouteRuleWithMultipleBackendsAndWeightsOut,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			translator := &Translator{
				gatewayClassName: "envoy-gateway-class",
			}

			cache := &StaticCache{}
			mustUnmarshal(t, tc.cache, cache)

			want := &TranslateResult{}
			mustUnmarshal(t, tc.want, want)

			got := translator.Translate(cache)

			sort.Slice(got.IR.HTTP, func(i, j int) bool { return got.IR.HTTP[i].Name < got.IR.HTTP[j].Name })

			assert.EqualValues(t, want, got)
		})
	}

}
