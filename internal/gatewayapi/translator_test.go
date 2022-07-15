package gatewayapi

import (
	"sort"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func mustUnmarshal(t *testing.T, val string, out interface{}) {
	require.NoError(t, yaml.UnmarshalStrict([]byte(val), out, yaml.DisallowUnknownFields))
}

func TestTranslate(t *testing.T) {
	tests := map[string]struct {
		resources string
		want      string
	}{
		// Route-Gateway attachment
		"Gateway with one HTTP Listener, HTTPRoute attaching to the Gateway": {
			resources: BasicHTTPRouteAttachingToGatewayIn,
			want:      BasicHTTPRouteAttachingToGatewayOut,
		},

		"Gateway with one HTTP Listener, HTTPRoute attaching to the Listener": {
			resources: BasicHTTPRouteAttachingToListenerIn,
			want:      BasicHTTPRouteAttachingToListenerOut,
		},

		"Gateway that allows HTTPRoutes from the same namespace, HTTPRoute in the same namespace": {
			resources: GatewayAllowsSameNamespaceWithAllowedHTTPRouteIn,
			want:      GatewayAllowsSameNamespaceWithAllowedHTTPRouteOut,
		},

		"Gateway that allows HTTPRoutes from the same namespace, HTTPRoute not in the same namespace": {
			resources: GatewayAllowsSameNamespaceWithDisallowedHTTPRouteIn,
			want:      GatewayAllowsSameNamespaceWithDisallowedHTTPRouteOut,
		},

		"Gateway with two HTTP Listeners, HTTPRoute attaching to the Gateway": {
			resources: HTTPRouteAttachingToGatewayWithTwoListenersIn,
			want:      HTTPRouteAttachingToGatewayWithTwoListenersOut,
		},

		"Gateway with two HTTP Listeners, HTTPRoute attaching to one Listener": {
			resources: HTTPRouteAttachingToListenerOnGatewayWithTwoListenersIn,
			want:      HTTPRouteAttachingToListenerOnGatewayWithTwoListenersOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with matching specific hostname": {
			resources: HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameIn,
			want:      HTTPRouteWithSpecificHostnameAttachingToGatewayWithWildcardHostnameOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with two matching specific hostnames": {
			resources: HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameIn,
			want:      HTTPRouteWithTwoSpecificHostnamesAttachingToGatewayWithWildcardHostnameOut,
		},

		"Gateway with one HTTP Listener with wildcard hostname, HTTPRoute attaching to the Gateway with non-matching specific hostname": {
			resources: HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameIn,
			want:      HTTPRouteWithNonMatchingSpecificHostnameAttachingToGatewayWithWildcardHostnameOut,
		},

		// Gateway/Listener error cases
		"Gateway with one Listener with protocol other than HTTP": {
			resources: GatewayWithListenerWithNonHTTPProtocolIn,
			want:      GatewayWithListenerWithNonHTTPProtocolOut,
		},

		"Gateway with one Listener with missing allowed namespaces selector": {
			resources: GatewayWithListenerWithMissingAllowedNamespacesSelectorIn,
			want:      GatewayWithListenerWithMissingAllowedNamespacesSelectorOut,
		},

		"Gateway with one Listener with invalid allowed namespaces selector": {
			resources: GatewayWithListenerWithInvalidAllowedNamespacesSelectorIn,
			want:      GatewayWithListenerWithInvalidAllowedNamespacesSelectorOut,
		},

		"Gateway with one Listener with invalid allowed routes group": {
			resources: GatewayWithListenerWithInvalidAllowedRoutesGroupIn,
			want:      GatewayWithListenerWithInvalidAllowedRoutesGroupOut,
		},

		"Gateway with one Listener with invalid allowed routes kind": {
			resources: GatewayWithListenerWithInvalidAllowedRoutesKindIn,
			want:      GatewayWithListenerWithInvalidAllowedRoutesKindOut,
		},

		"Gateway with two Listeners with the same port and hostname": {
			resources: GatewayWithTwoListenersWithSamePortAndHostnameIn,
			want:      GatewayWithTwoListenersWithSamePortAndHostnameOut,
		},

		"Gateway with two Listeners with the same port and incompatible protocols": {
			resources: GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsIn,
			want:      GatewayWithTwoListenersWithSamePortAndIncompatibleProtocolsOut,
		},

		// Route matches
		"HTTPRoute with single rule with path prefix and exact header matches": {
			resources: HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesIn,
			want:      HTTPRouteWithSingleRuleWithPathPrefixAndExactHeaderMatchesOut,
		},

		"HTTPRoute with single rule with exact path match": {
			resources: HTTPRouteWithSingleRuleWithExactPathMatchIn,
			want:      HTTPRouteWithSingleRuleWithExactPathMatchOut,
		},

		// Route backends
		"HTTPRoute rule with multiple backends, no weights explicitly specified": {
			resources: HTTPRouteRuleWithMultipleBackendsAndNoWeightsIn,
			want:      HTTPRouteRuleWithMultipleBackendsAndNoWeightsOut,
		},

		"HTTPRoute rule with multiple backends, weights explicitly specified": {
			resources: HTTPRouteRuleWithMultipleBackendsAndWeightsIn,
			want:      HTTPRouteRuleWithMultipleBackendsAndWeightsOut,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			translator := &Translator{
				gatewayClassName: "envoy-gateway-class",
			}

			resources := &Resources{}
			mustUnmarshal(t, tc.resources, resources)

			// Add common test fixtures
			for i := 1; i <= 3; i++ {
				resources.Services = append(resources.Services,
					&v1.Service{
						ObjectMeta: metav1.ObjectMeta{
							Namespace: "default",
							Name:      "service-" + strconv.Itoa(i),
						},
						Spec: v1.ServiceSpec{
							ClusterIP: "7.7.7.7",
							Ports: []v1.ServicePort{
								{Port: 8080},
								{Port: 8443},
							},
						},
					},
				)
			}

			resources.Namespaces = append(resources.Namespaces, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-gateway",
				},
			}, &v1.Namespace{
				ObjectMeta: metav1.ObjectMeta{
					Name: "default",
				},
			})

			want := &TranslateResult{}
			mustUnmarshal(t, tc.want, want)

			got := translator.Translate(resources)

			sort.Slice(got.IR.HTTP, func(i, j int) bool { return got.IR.HTTP[i].Name < got.IR.HTTP[j].Name })

			assert.EqualValues(t, want, got)
		})
	}

}
