// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
)

func TestRedirectExtension(t *testing.T) {
	for _, tc := range []struct {
		name   string
		mutate func(*[]gwapiv1.HTTPRouteFilter, *egv1a1.HTTPRouteFilter)
		err    string
	}{
		{name: "native first"},
		{name: "extension first", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) { slices.Reverse(*filters) }},
		{name: "missing redirect", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) { *filters = (*filters)[1:] }, err: "requires exactly one RequestRedirect"},
		{name: "duplicate extension", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append(*filters, (*filters)[1])
		}, err: "only one redirect extension"},
		{name: "duplicate native", mutate: func(filters *[]gwapiv1.HTTPRouteFilter, _ *egv1a1.HTTPRouteFilter) {
			*filters = append(*filters, (*filters)[0])
		}, err: "requires exactly one RequestRedirect"},
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
		{name: "unresolved extension", mutate: func(_ *[]gwapiv1.HTTPRouteFilter, hrf *egv1a1.HTTPRouteFilter) { hrf.Name = "different" }, err: "Unable to translate HTTPRouteFilter"},
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
				return
			}
			require.Empty(t, errs)
			require.NotNil(t, got.RedirectResponse)
			require.NoError(t, got.RedirectResponse.Validate())
			require.Equal(t, "https", *got.RedirectResponse.Scheme)
			require.Equal(t, "example.com", *got.RedirectResponse.Hostname)
			require.Equal(t, uint32(8443), *got.RedirectResponse.Port)
			require.Equal(t, int32(301), *got.RedirectResponse.StatusCode)
			require.Equal(t, `^/blogs/([0-9]+)$`, got.RedirectResponse.Path.RegexMatchReplace.Pattern)
			require.Equal(t, `/post-\1`, got.RedirectResponse.Path.RegexMatchReplace.Substitution)
			require.Nil(t, got.URLRewrite)
		})
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
	_, errs := translator.ProcessGRPCFilters(nil, route, []gwapiv1.GRPCRouteFilter{{
		Type:         gwapiv1.GRPCRouteFilterExtensionRef,
		ExtensionRef: &gwapiv1.LocalObjectReference{Group: egv1a1.GroupName, Kind: egv1a1.KindHTTPRouteFilter, Name: "redirect"},
	}}, &resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{
		ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
		Spec:       egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{}},
	}}}, nil)
	require.Len(t, errs, 1)
	require.Contains(t, errs[0].Error(), "only supported on HTTPRoute rules")
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
	_, err := translator.processDestinationFilters(resource.KindHTTPRoute, BackendRefWithFilters{Filters: filters}, nil, route,
		&resource.Resources{HTTPRouteFilters: []*egv1a1.HTTPRouteFilter{{
			ObjectMeta: metav1.ObjectMeta{Name: "redirect", Namespace: "default"},
			Spec: egv1a1.HTTPRouteFilterSpec{Redirect: &egv1a1.HTTPRedirectFilter{Path: egv1a1.HTTPPathModifier{
				Type:              egv1a1.RegexHTTPPathModifier,
				ReplaceRegexMatch: &egv1a1.ReplaceRegexMatch{Pattern: "old", Substitution: "new"},
			}}},
		}}}, nil)
	require.ErrorContains(t, err, "not supported on backendRefs")
}
