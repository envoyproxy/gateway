// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

func TestFilterReferencesAfterDirectResponse(t *testing.T) {
	for _, grpc := range []bool{false, true} {
		for _, reverse := range []bool{false, true} {
			for _, kind := range []string{"resolved", "missing", "mirror", "unsupported", "credential", "body", "duplicate matches"} {
				t.Run(fmt.Sprintf("grpc=%t/reverse=%t/%s", grpc, reverse, kind), func(t *testing.T) {
					translator := &Translator{TranslatorContext: &TranslatorContext{}}
					matches := []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "cohort", Value: "deny"}}}}
					resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
						{ObjectMeta: metav1.ObjectMeta{Name: "deny", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{DirectResponse: &egv1a1.HTTPDirectResponseFilter{StatusCode: new(403)}}},
						{ObjectMeta: metav1.ObjectMeta{Name: "cookies", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{Matches: matches}},
					}}
					ref := func(name string) gwapiv1.HTTPRouteFilter {
						return gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
							Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: gwapiv1.ObjectName(name),
						}}
					}
					filters := []gwapiv1.HTTPRouteFilter{ref("deny"), ref("cookies")}
					condition := gwapiv1.RouteConditionResolvedRefs
					switch kind {
					case "missing":
						filters = append(filters, ref("missing"))
					case "mirror":
						filters = append(filters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterRequestMirror, RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
							BackendRef: gwapiv1.BackendObjectReference{Name: "missing", Port: new(gwapiv1.PortNumber(8080))},
						}})
					case "unsupported":
						filters = append(filters, gwapiv1.HTTPRouteFilter{Type: "Unsupported"})
						condition = gwapiv1.RouteConditionAccepted
					case "credential", "body":
						condition = gwapiv1.RouteConditionAccepted
						invalid := &egv1a1.HTTPRouteFilter{ObjectMeta: metav1.ObjectMeta{Name: "invalid", Namespace: "default"}}
						if kind == "credential" {
							invalid.Spec.CredentialInjection = &egv1a1.HTTPCredentialInjectionFilter{Credential: egv1a1.InjectedCredential{ValueRef: gwapiv1.SecretObjectReference{Name: "missing"}}}
						} else {
							invalid.Spec.DirectResponse = &egv1a1.HTTPDirectResponseFilter{Body: &egv1a1.CustomResponseBody{
								Type: new(egv1a1.ResponseValueTypeValueRef), ValueRef: &gwapiv1.LocalObjectReference{Kind: "ConfigMap", Name: "missing"},
							}}
						}
						resources.HTTPRouteFilters = append(resources.HTTPRouteFilters, invalid)
						filters = append(filters, ref("invalid"))
					case "duplicate matches":
						filters = append(filters, ref("cookies"))
						condition = gwapiv1.RouteConditionAccepted
					}
					if reverse {
						slices.Reverse(filters)
					}
					var got *HTTPFiltersContext
					var errs []status.Error
					if grpc {
						grpcFilters := make([]gwapiv1.GRPCRouteFilter, 0, len(filters))
						for _, filter := range filters {
							grpcFilters = append(grpcFilters, gwapiv1.GRPCRouteFilter{Type: gwapiv1.GRPCRouteFilterType(filter.Type), ExtensionRef: filter.ExtensionRef, RequestMirror: filter.RequestMirror})
						}
						got, errs = translator.ProcessGRPCFilters(nil, &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}, grpcFilters, resources, nil)
					} else {
						got, errs = translator.ProcessHTTPFilters(nil, &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}, filters, 0, resources, nil)
					}
					require.Equal(t, matches, got.Matches)
					require.Empty(t, got.Mirrors)
					require.Empty(t, got.CredentialInjection)
					if kind == "resolved" {
						require.Empty(t, errs)
						require.Equal(t, new(uint32(403)), got.DirectResponse.StatusCode)
					} else {
						require.Len(t, errs, 1)
						require.Equal(t, condition, errs[0].Type())
						if kind == "mirror" && reverse {
							// The route translator converts the earlier mirror error to
							// a 500 after applying the filter context.
							require.Equal(t, new(uint32(403)), got.DirectResponse.StatusCode)
						} else if kind != "duplicate matches" {
							require.Equal(t, new(uint32(500)), got.DirectResponse.StatusCode)
						}
					}
				})
			}
		}
	}
}

func TestFilterReferenceExternalGroup(t *testing.T) {
	translator := &Translator{TranslatorContext: &TranslatorContext{}, ExtensionGroupKinds: []schema.GroupKind{{Group: "example.com", Kind: egv1a1.KindHTTPRouteFilter}}}
	resources := &resource.Resources{
		HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{ObjectMeta: metav1.ObjectMeta{Name: "external", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{
			Matches: []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "unrelated", Value: "cookie"}}}},
		}}},
		ExtensionRefFilters: []unstructured.Unstructured{{Object: map[string]interface{}{
			"apiVersion": "example.com/v1", "kind": egv1a1.KindHTTPRouteFilter, "metadata": map[string]interface{}{"name": "external", "namespace": "default"},
		}}},
	}
	filters := []gwapiv1.HTTPRouteFilter{
		{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "missing"}},
		{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{Group: "example.com", Kind: egv1a1.KindHTTPRouteFilter, Name: "external"}},
	}
	got, errs := translator.ProcessHTTPFilters(nil, &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}, filters, 0, resources, nil)
	require.Len(t, errs, 1)
	require.Equal(t, gwapiv1.RouteConditionResolvedRefs, errs[0].Type())
	require.Empty(t, got.Matches)
	require.Empty(t, got.ExtensionRefs)
	require.Empty(t, translator.ExtensionResourceMap)
}
