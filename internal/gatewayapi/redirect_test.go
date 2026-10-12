// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestRedirectExtension(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*[]gwapiv1.HTTPRouteFilter, *egv1a1.HTTPRouteFilter)
		err    string
		want   *ir.Redirect
	}{
		{name: "native first"},
		{name: "extension first", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) { slices.Reverse(*filters) }},
		{
			name: "standalone", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) { *filters = (*filters)[1:] },
			want: &ir.Redirect{StatusCode: new(int32(302))},
		},
		{name: "status override only", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			(*filters)[0].RequestRedirect = &gwapiv1.HTTPRequestRedirectFilter{StatusCode: new(301)}
		}, want: &ir.Redirect{StatusCode: new(int32(301))}},
		{name: "invalid native scheme", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			(*filters)[0].RequestRedirect.Scheme = new("ftp")
		}, err: "scheme: ftp is unsupported"},
		{name: "nil native redirect", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			(*filters)[0].RequestRedirect = nil
		}, err: "RequestRedirect filter must specify requestRedirect"},
		{name: "duplicate extension", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append(*filters, (*filters)[1])
		}, err: "only one redirect extension"},
		{name: "duplicate native", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append(*filters, (*filters)[0])
		}, err: "at most one RequestRedirect"},
		{name: "standalone duplicate extension", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append((*filters)[1:], (*filters)[1])
		}, err: "only one redirect extension"},
		{name: "standalone rewrite conflict", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append((*filters)[1:], gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterURLRewrite, URLRewrite: &gwapiv1.HTTPURLRewriteFilter{}})
		}, err: "cannot be combined with URLRewrite"},
		{name: "standalone direct response conflict", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			*filters = (*filters)[1:]
			hrf.Spec.DirectResponse = &egv1a1.HTTPDirectResponseFilter{}
		}, err: "cannot be combined with URLRewrite or DirectResponse"},
		{name: "standalone invalid pattern", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			*filters = (*filters)[1:]
			hrf.Spec.Redirect.Path.ReplaceRegexMatch.Pattern = "("
		}, err: "valid RE2"},
		{name: "standalone unresolved extension", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			*filters = (*filters)[1:]
			hrf.Name = "different"
		}, err: "Unable to translate HTTPRouteFilter"},
		{name: "standalone unresolved additional reference", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append((*filters)[1:], gwapiv1.HTTPRouteFilter{
				Type: gwapiv1.HTTPRouteFilterExtensionRef,
				ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "missing",
				},
			})
		}, err: "Unable to translate HTTPRouteFilter"},
		{name: "native path conflict", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			(*filters)[0].RequestRedirect.Path = &gwapiv1.HTTPPathModifier{Type: gwapiv1.FullPathHTTPPathModifier, ReplaceFullPath: new("/landing")}
		}, err: "cannot be combined with RequestRedirect.path"},
		{name: "native path conflict reversed", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			(*filters)[0].RequestRedirect.Path = &gwapiv1.HTTPPathModifier{Type: gwapiv1.PrefixMatchHTTPPathModifier, ReplacePrefixMatch: new("/")}
			slices.Reverse(*filters)
		}, err: "cannot be combined with RequestRedirect.path"},
		{name: "native rewrite conflict", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append(*filters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterURLRewrite, URLRewrite: &gwapiv1.HTTPURLRewriteFilter{}})
		}, err: "cannot be combined with URLRewrite"},
		{name: "extension rewrite conflict", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			hrf.Spec.URLRewrite = &egv1a1.HTTPURLRewriteFilter{}
		}, err: "cannot be combined with URLRewrite"},
		{name: "direct response conflict", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			hrf.Spec.DirectResponse = &egv1a1.HTTPDirectResponseFilter{}
			slices.Reverse(*filters)
		}, err: "cannot be combined with URLRewrite or DirectResponse"},
		{name: "invalid pattern", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			hrf.Spec.Redirect.Path.ReplaceRegexMatch.Pattern = "("
		}, err: "valid RE2"},
		{name: "missing pattern", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			hrf.Spec.Redirect.Path.ReplaceRegexMatch = nil
		}, err: "non-empty pattern"},
		{name: "invalid capture", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) {
			hrf.Spec.Redirect.Path.ReplaceRegexMatch.Substitution = `/post-\2`
		}, err: "valid RE2 capture references"},
		{
			name: "unresolved extension", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) { hrf.Name = "different" },
			err: "Unable to translate HTTPRouteFilter",
			want: &ir.Redirect{
				Scheme: new("https"), Hostname: new("example.com"), Port: new(uint32(8443)), StatusCode: new(int32(301)),
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			hrf := &egv1a1.HTTPRouteFilter{
				ObjectMeta: metav1.ObjectMeta{Name: "regex-redirect", Namespace: "default"},
				Spec: egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{Path: egv1a1.HTTPPathModifier{
					Type:              egv1a1.RegexHTTPPathModifier,
					ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{Pattern: `^/blogs/([0-9]+)$`, Substitution: `/post-\1`},
				}}},
			}
			filters := []gwapiv1.HTTPRouteFilter{
				{Type: gwapiv1.HTTPRouteFilterRequestRedirect, RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{
					Scheme: new("https"), Hostname: new(gwapiv1.PreciseHostname("example.com")), Port: new(gwapiv1.PortNumber(8443)), StatusCode: new(301),
				}},
				{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "regex-redirect",
				}},
			}
			if tc.mutate != nil {
				tc.mutate(&filters, hrf)
			}
			route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
			translator := &Translator{}
			got, errs := translator.ProcessHTTPFilters(nil, route, filters, 0, &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{hrf}}, nil)
			if tc.err != "" {
				require.NotEmpty(t, errs)
				require.Contains(t, errs[0].Error(), tc.err)
				require.Equal(t, tc.want, got.RedirectResponse)
				return
			}
			require.Empty(t, errs)
			require.NotNil(t, got.RedirectResponse)
			require.NoError(t, got.RedirectResponse.Validate())
			want := tc.want
			if want == nil {
				want = &ir.Redirect{
					Scheme: new("https"), Hostname: new("example.com"), Port: new(uint32(8443)), StatusCode: new(int32(301)),
				}
			}
			want.Path = &ir.ExtendedHTTPPathModifier{RegexMatchReplace: &ir.RegexMatchReplace{
				Pattern: `^/blogs/([0-9]+)$`, Substitution: `/post-\1`,
			}}
			require.Equal(t, want, got.RedirectResponse)
			require.Nil(t, got.URLRewrite)
		})
	}
}

func TestRedirectExtensionMirrorConflict(t *testing.T) {
	for _, composed := range []bool{false, true} {
		for _, mirrorFirst := range []bool{false, true} {
			filters := []gwapiv1.HTTPRouteFilter{
				{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
					Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "redirect",
				}},
				{Type: gwapiv1.HTTPRouteFilterRequestMirror, RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
					BackendRef: gwapiv1.BackendObjectReference{Name: "mirror", Port: new(gwapiv1.PortNumber(8080))},
				}},
			}
			if composed {
				filters = append(filters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterRequestRedirect, RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{}})
			}
			if mirrorFirst {
				slices.Reverse(filters)
			}
			resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{
				ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
				Spec: egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{Path: egv1a1.HTTPPathModifier{
					Type:              egv1a1.RegexHTTPPathModifier,
					ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{Pattern: "^/old$", Substitution: "/new"},
				}}},
			}}}
			route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
			path, err := processRedirectExtension(filters, route, resources)
			require.NotNil(t, err, "composed=%t mirrorFirst=%t", composed, mirrorFirst)
			require.Equal(t, gwapiv1.RouteConditionAccepted, err.Type())
			require.Equal(t, gwapiv1.RouteReasonIncompatibleFilters, err.Reason())
			require.Contains(t, err.Error(), requestMirrorRedirectConflictMsg)
			require.Nil(t, path)
		}
	}
}

func TestRedirectExtensionReferenceErrors(t *testing.T) {
	translator := &Translator{TranslatorContext: &TranslatorContext{}}
	translator.SetServices([]*corev1.Service{{
		ObjectMeta: metav1.ObjectMeta{Name: "valid-mirror", Namespace: "default"},
		Spec:       corev1.ServiceSpec{Ports: []corev1.ServicePort{{Port: 8080, Protocol: corev1.ProtocolTCP}}},
	}})
	matches := []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "cohort", Value: "redirect"}}}}
	resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
			Spec:       egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{}, Matches: matches},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "deny", Namespace: "default"},
			Spec:       egv1a1.HTTPRouteFilterSpec{DirectResponse: &egv1a1.HTTPDirectResponseFilter{StatusCode: new(403)}},
		},
	}}
	for _, routeKind := range []gwapiv1.Kind{resource.KindHTTPRoute, resource.KindGRPCRoute} {
		t.Run(string(routeKind), func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				refs    []gwapiv1.ObjectName
				missing []string
				mirror  gwapiv1.ObjectName
			}{
				{name: "all references resolve", refs: []gwapiv1.ObjectName{"redirect", "deny"}},
				{name: "redirect first", refs: []gwapiv1.ObjectName{"redirect", "missing"}, missing: []string{"missing"}},
				{name: "missing reference first", refs: []gwapiv1.ObjectName{"missing", "redirect"}, missing: []string{"missing"}},
				{name: "direct response first", refs: []gwapiv1.ObjectName{"deny", "redirect", "missing"}, missing: []string{"missing"}},
				{name: "multiple missing references", refs: []gwapiv1.ObjectName{"missing", "deny", "redirect", "other"}, missing: []string{"missing", "other"}},
				{name: "missing mirror", refs: []gwapiv1.ObjectName{"redirect"}, mirror: "missing-mirror", missing: []string{"missing-mirror"}},
				{name: "direct response before missing mirror", refs: []gwapiv1.ObjectName{"deny", "redirect"}, mirror: "missing-mirror", missing: []string{"missing-mirror"}},
				{name: "valid mirror without endpoints", refs: []gwapiv1.ObjectName{"redirect"}, mirror: "valid-mirror"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					var got *HTTPFiltersContext
					var errs []status.Error
					mirror := &gwapiv1.HTTPRequestMirrorFilter{BackendRef: gwapiv1.BackendObjectReference{Name: tc.mirror, Port: new(gwapiv1.PortNumber(8080))}}
					if routeKind == resource.KindHTTPRoute {
						filters := make([]gwapiv1.HTTPRouteFilter, 0, len(tc.refs))
						for _, name := range tc.refs {
							filters = append(filters, gwapiv1.HTTPRouteFilter{
								Type:         gwapiv1.HTTPRouteFilterExtensionRef,
								ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: name},
							})
						}
						if tc.mirror != "" {
							filters = append(filters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterRequestMirror, RequestMirror: mirror})
						}
						route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
						got, errs = translator.ProcessHTTPFilters(nil, route, filters, 0, resources, nil)
					} else {
						filters := make([]gwapiv1.GRPCRouteFilter, 0, len(tc.refs))
						for _, name := range tc.refs {
							filters = append(filters, gwapiv1.GRPCRouteFilter{
								Type:         gwapiv1.GRPCRouteFilterExtensionRef,
								ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: name},
							})
						}
						if tc.mirror != "" {
							filters = append(filters, gwapiv1.GRPCRouteFilter{Type: gwapiv1.GRPCRouteFilterRequestMirror, RequestMirror: mirror})
						}
						route := &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
						got, errs = translator.ProcessGRPCFilters(nil, route, filters, resources, nil)
					}
					require.Equal(t, &HTTPFilterIR{Matches: matches}, got.HTTPFilterIR)
					conditionErrors := make(map[gwapiv1.RouteConditionType][]string)
					for _, err := range errs {
						conditionErrors[err.Type()] = append(conditionErrors[err.Type()], err.Error())
					}
					require.NotEmpty(t, conditionErrors[gwapiv1.RouteConditionAccepted])
					if len(tc.missing) == 0 {
						require.Empty(t, conditionErrors[gwapiv1.RouteConditionResolvedRefs])
					} else {
						require.NotEmpty(t, conditionErrors[gwapiv1.RouteConditionResolvedRefs])
						for _, name := range tc.missing {
							require.Contains(t, strings.Join(conditionErrors[gwapiv1.RouteConditionResolvedRefs], "; "), "default/"+name)
						}
					}
				})
			}
		})
	}
}

func TestRedirectFilterOrdering(t *testing.T) {
	for _, routeKind := range []gwapiv1.Kind{resource.KindHTTPRoute, resource.KindGRPCRoute} {
		for _, invalidRedirect := range []bool{false, true} {
			for _, composed := range []bool{false, true} {
				if composed && routeKind == resource.KindGRPCRoute {
					continue
				}
				for _, tc := range []struct {
					name     string
					refs     []string
					accepted bool
					missing  []string
				}{
					{name: "valid", refs: []string{"redirect"}},
					{name: "missing first", refs: []string{"missing", "redirect"}, missing: []string{"missing"}},
					{name: "missing last", refs: []string{"redirect", "missing"}, missing: []string{"missing"}},
					{name: "multiple missing", refs: []string{"missing", "redirect", "other"}, missing: []string{"missing", "other"}},
					{name: "direct first", refs: []string{"deny", "missing", "redirect"}, accepted: true, missing: []string{"missing"}},
					{name: "direct last", refs: []string{"redirect", "missing", "deny"}, accepted: true, missing: []string{"missing"}},
					{name: "duplicate matches first", refs: []string{"cookies", "missing", "redirect"}, accepted: true, missing: []string{"missing"}},
					{name: "duplicate matches last", refs: []string{"redirect", "missing", "cookies"}, accepted: true, missing: []string{"missing"}},
					{name: "unsupported first", refs: []string{"unsupported", "redirect", "missing"}, accepted: true, missing: []string{"missing"}},
					{name: "unsupported last", refs: []string{"missing", "redirect", "unsupported"}, accepted: true, missing: []string{"missing"}},
				} {
					t.Run(fmt.Sprintf("%s/invalid=%t/composed=%t/%s", routeKind, invalidRedirect, composed, tc.name), func(t *testing.T) {
						matches := []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "cohort", Value: "redirect"}}}}
						redirect := &egv1a1.HTTPRedirectFilter{Path: egv1a1.HTTPPathModifier{
							Type: egv1a1.RegexHTTPPathModifier, ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{Pattern: "^/old$", Substitution: "/new"},
						}}
						if invalidRedirect {
							redirect.Path.ReplaceRegexMatch.Pattern = "("
						}
						resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
							{ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{Redirect: redirect, Matches: matches}},
							{ObjectMeta: metav1.ObjectMeta{Name: "cookies", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{Matches: matches}},
							{ObjectMeta: metav1.ObjectMeta{Name: "deny", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{DirectResponse: &egv1a1.HTTPDirectResponseFilter{StatusCode: new(403)}}},
						}}
						httpFilters := make([]gwapiv1.HTTPRouteFilter, 0, len(tc.refs)+1)
						grpcFilters := make([]gwapiv1.GRPCRouteFilter, 0, len(tc.refs))
						if composed {
							httpFilters = append(httpFilters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterRequestRedirect, RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{}})
						}
						for _, name := range tc.refs {
							ref := &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: gwapiv1.ObjectName(name)}
							if name == "unsupported" {
								ref.Group = "unsupported.example.com"
							}
							httpFilters = append(httpFilters, gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: ref})
							grpcFilters = append(grpcFilters, gwapiv1.GRPCRouteFilter{Type: gwapiv1.GRPCRouteFilterExtensionRef, ExtensionRef: ref})
						}
						translator := &Translator{}
						var got *HTTPFiltersContext
						var errs []status.Error
						if routeKind == resource.KindHTTPRoute {
							route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
							got, errs = translator.ProcessHTTPFilters(nil, route, httpFilters, 0, resources, nil)
						} else {
							route := &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
							got, errs = translator.ProcessGRPCFilters(nil, route, grpcFilters, resources, nil)
						}
						require.Equal(t, matches, got.Matches)
						conditionErrors := make(map[gwapiv1.RouteConditionType][]string)
						for _, err := range errs {
							conditionErrors[err.Type()] = append(conditionErrors[err.Type()], err.Error())
						}
						wantAcceptedError := tc.accepted || invalidRedirect || routeKind == resource.KindGRPCRoute
						require.Equal(t, wantAcceptedError, len(conditionErrors[gwapiv1.RouteConditionAccepted]) > 0)
						for _, name := range tc.missing {
							require.Contains(t, strings.Join(conditionErrors[gwapiv1.RouteConditionResolvedRefs], ";"), "default/"+name)
						}
						if len(tc.missing) == 0 {
							require.Empty(t, conditionErrors[gwapiv1.RouteConditionResolvedRefs])
						}
						if !wantAcceptedError && len(tc.missing) == 0 {
							require.Empty(t, errs)
							require.NotNil(t, got.RedirectResponse)
						}
					})
				}
			}
		}
	}
}

func TestRedirectExternalFilterGroup(t *testing.T) {
	translator := &Translator{
		TranslatorContext:   &TranslatorContext{},
		ExtensionGroupKinds: []schema.GroupKind{{Group: "example.com", Kind: egv1a1.KindHTTPRouteFilter}},
	}
	resources := &resource.Resources{
		HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{
			ObjectMeta: metav1.ObjectMeta{Name: "external", Namespace: "default"},
			Spec:       egv1a1.HTTPRouteFilterSpec{Matches: []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "unrelated", Value: "cookie"}}}}},
		}},
		ExtensionRefFilters: []unstructured.Unstructured{{Object: map[string]interface{}{
			"apiVersion": "example.com/v1", "kind": egv1a1.KindHTTPRouteFilter,
			"metadata": map[string]interface{}{"name": "external", "namespace": "default"},
		}}},
	}
	filters := []gwapiv1.HTTPRouteFilter{{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
		Group: "example.com", Kind: egv1a1.KindHTTPRouteFilter, Name: "external",
	}}}
	route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
	got, errs := translator.ProcessHTTPFilters(nil, route, filters, 0, resources, nil)
	require.Empty(t, errs)
	require.Empty(t, got.Matches)
	require.Len(t, got.ExtensionRefs, 1)
	require.Equal(t, "example.com/v1", got.ExtensionRefs[0].Object.GetAPIVersion())

	// Reference-only recovery must find the external resource without publishing
	// it or accidentally loading a same-named built-in filter's matches.
	filters = append([]gwapiv1.HTTPRouteFilter{{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
		Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "missing",
	}}}, filters...)
	got, errs = translator.ProcessHTTPFilters(nil, route, filters, 0, resources, nil)
	require.Len(t, errs, 1)
	require.Equal(t, gwapiv1.RouteConditionResolvedRefs, errs[0].Type())
	require.Empty(t, got.Matches)
	require.Empty(t, got.ExtensionRefs)
	require.Empty(t, translator.ExtensionResourceMap)
}

func TestFilterReferencesAfterDirectResponse(t *testing.T) {
	for _, routeKind := range []gwapiv1.Kind{resource.KindHTTPRoute, resource.KindGRPCRoute} {
		for _, mirror := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/mirror=%t", routeKind, mirror), func(t *testing.T) {
				translator := &Translator{TranslatorContext: &TranslatorContext{}}
				matches := []egv1a1.HTTPRouteMatchFilter{{Cookies: []egv1a1.HTTPCookieMatch{{Name: "cohort", Value: "deny"}}}}
				resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
					{ObjectMeta: metav1.ObjectMeta{Name: "deny", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{DirectResponse: &egv1a1.HTTPDirectResponseFilter{StatusCode: new(403)}}},
					{ObjectMeta: metav1.ObjectMeta{Name: "cookies", Namespace: "default"}, Spec: egv1a1.HTTPRouteFilterSpec{Matches: matches}},
				}}
				httpFilters := []gwapiv1.HTTPRouteFilter{
					{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "deny"}},
					{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "cookies"}},
					{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "missing"}},
				}
				if mirror {
					httpFilters[2] = gwapiv1.HTTPRouteFilter{Type: gwapiv1.HTTPRouteFilterRequestMirror, RequestMirror: &gwapiv1.HTTPRequestMirrorFilter{
						BackendRef: gwapiv1.BackendObjectReference{Name: "missing", Port: new(gwapiv1.PortNumber(8080))},
					}}
				}
				var got *HTTPFiltersContext
				var errs []status.Error
				if routeKind == resource.KindHTTPRoute {
					route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
					got, errs = translator.ProcessHTTPFilters(nil, route, httpFilters, 0, resources, nil)
				} else {
					grpcFilters := make([]gwapiv1.GRPCRouteFilter, 0, len(httpFilters))
					for _, filter := range httpFilters {
						grpcFilters = append(grpcFilters, gwapiv1.GRPCRouteFilter{Type: gwapiv1.GRPCRouteFilterType(filter.Type), ExtensionRef: filter.ExtensionRef, RequestMirror: filter.RequestMirror})
					}
					route := &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
					got, errs = translator.ProcessGRPCFilters(nil, route, grpcFilters, resources, nil)
				}
				require.Equal(t, matches, got.Matches)
				require.Len(t, errs, 1)
				require.Equal(t, gwapiv1.RouteConditionResolvedRefs, errs[0].Type())
				require.Contains(t, errs[0].Error(), "default/missing")
				require.Equal(t, new(uint32(500)), got.DirectResponse.StatusCode)
				require.Empty(t, got.Mirrors)
			})
		}
	}
}

func TestValidateRedirectSubstitution(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{`/post-\1`, true},
		{`/\0`, true},
		{`/literal\\slash`, true},
		{`/`, true},
		{``, false},
		{`/\2`, false},
		{`/\x`, false},
		{"/\\", false},
		{"/new\r\nLocation: other", false},
		{"/\x00", false},
		{`/new?query`, false},
		{`/new#fragment`, false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			err := validateRedirectSubstitution(tc.value, 1)
			if tc.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestRedirectExtensionGRPCRoute(t *testing.T) {
	translator := &Translator{}
	route := &GRPCRouteContext{GRPCRoute: &gwapiv1.GRPCRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
	resources := &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{
		{
			ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
			Spec:       egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{}},
		},
		{
			ObjectMeta: metav1.ObjectMeta{Name: "deny", Namespace: "default"},
			Spec:       egv1a1.HTTPRouteFilterSpec{DirectResponse: &egv1a1.HTTPDirectResponseFilter{StatusCode: new(403)}},
		},
	}}
	for _, tc := range []struct {
		name    string
		refs    []gwapiv1.ObjectName
		invalid bool
	}{
		{name: "redirect alone", refs: []gwapiv1.ObjectName{"redirect"}, invalid: true},
		{name: "direct response first", refs: []gwapiv1.ObjectName{"deny", "redirect"}, invalid: true},
		{name: "redirect first", refs: []gwapiv1.ObjectName{"redirect", "deny"}, invalid: true},
		{name: "direct response alone", refs: []gwapiv1.ObjectName{"deny"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			filters := make([]gwapiv1.GRPCRouteFilter, 0, len(tc.refs))
			for _, name := range tc.refs {
				filters = append(filters, gwapiv1.GRPCRouteFilter{
					Type:         gwapiv1.GRPCRouteFilterExtensionRef,
					ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: name},
				})
			}
			got, errs := translator.ProcessGRPCFilters(nil, route, filters, resources, nil)
			require.Nil(t, got.RedirectResponse)
			if tc.invalid {
				require.Len(t, errs, 1)
				require.Contains(t, errs[0].Error(), "only supported on HTTPRoute rules")
				require.Equal(t, gwapiv1.RouteConditionAccepted, errs[0].Type())
				require.Equal(t, gwapiv1.RouteReasonUnsupportedValue, errs[0].Reason())
				require.Nil(t, got.DirectResponse)
				return
			}
			require.Empty(t, errs)
			require.NotNil(t, got.DirectResponse)
			require.Equal(t, new(uint32(403)), got.DirectResponse.StatusCode)
		})
	}
}

func TestRedirectExtensionBackendRef(t *testing.T) {
	translator := &Translator{}
	route := &HTTPRouteContext{HTTPRoute: &gwapiv1.HTTPRoute{ObjectMeta: metav1.ObjectMeta{Namespace: "default"}}}
	filters := []gwapiv1.HTTPRouteFilter{
		{Type: gwapiv1.HTTPRouteFilterRequestRedirect, RequestRedirect: &gwapiv1.HTTPRequestRedirectFilter{Scheme: new("https")}},
		{Type: gwapiv1.HTTPRouteFilterExtensionRef, ExtensionRef: &gwapiv1.LocalObjectReference{
			Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "redirect",
		}},
	}
	for _, tc := range []struct {
		name    string
		filters []gwapiv1.HTTPRouteFilter
	}{
		{name: "composed", filters: filters},
		{name: "standalone", filters: filters[1:]},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := translator.processDestinationFilters(resource.KindHTTPRoute, BackendRefWithFilters{Filters: tc.filters}, nil, route,
				&resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{
					ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
					Spec: egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{Path: egv1a1.HTTPPathModifier{
						Type:              egv1a1.RegexHTTPPathModifier,
						ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{Pattern: "old", Substitution: "new"},
					}}},
				}}}, nil)
			require.ErrorContains(t, err, "not supported on backendRefs")
		})
	}
}
