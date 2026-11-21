// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	filterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/set_filter_state/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const (
	regexRedirectPathFilter = "envoy.filters.http.set_filter_state/regex_redirect"
	regexRedirectPathKey    = "envoy_gateway.regex_redirect.path"
)

func patchHCMWithRegexRedirectPath(mgr *hcmv3.HttpConnectionManager, listener *ir.HTTPListener) error {
	for i, filter := range mgr.HttpFilters {
		if filter.Name == regexRedirectPathFilter {
			mgr.HttpFilters = append([]*hcmv3.HttpFilter{filter}, slices.Delete(mgr.HttpFilters, i, i+1)...)
			return nil
		}
	}
	for _, route := range listener.Routes {
		if route.DirectResponse != nil || route.Redirect == nil || route.Redirect.Path == nil ||
			route.Redirect.Path.RegexMatchReplace == nil {
			continue
		}
		config, err := proto.ToAnyWithValidation(&setfilterstatev3.Config{
			ClearRouteCache: true,
			OnRequestHeaders: []*filterstatev3.FilterStateValue{{
				Key:        &filterstatev3.FilterStateValue_ObjectKey{ObjectKey: regexRedirectPathKey},
				FactoryKey: "envoy.string",
				Value: &filterstatev3.FilterStateValue_FormatString{
					FormatString: &corev3.SubstitutionFormatString{
						Format: &corev3.SubstitutionFormatString_TextFormatSource{
							TextFormatSource: &corev3.DataSource{
								Specifier: &corev3.DataSource_InlineString{
									InlineString: "%PATH(NQ:PATH)%",
								},
							},
						},
					},
				},
			}},
		})
		if err != nil {
			return err
		}
		// Populate the guard before any filter can cache a route or apply its policies.
		mgr.HttpFilters = append([]*hcmv3.HttpFilter{{
			Name:       regexRedirectPathFilter,
			ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: config},
		}}, mgr.HttpFilters...)
		return nil
	}
	return nil
}

func addRegexRedirectMatch(match *routev3.RouteMatch, pattern string) {
	match.FilterState = append(match.FilterState, &matcherv3.FilterStateMatcher{
		Key: regexRedirectPathKey,
		Matcher: &matcherv3.FilterStateMatcher_StringMatch{
			StringMatch: &matcherv3.StringMatcher{
				MatchPattern: &matcherv3.StringMatcher_SafeRegex{
					SafeRegex: &matcherv3.RegexMatcher{
						// Rewrites use partial matching, whereas StringMatcher uses full matching.
						Regex: "(?s:.*)(?:" + pattern + ")(?s:.*)",
					},
				},
			},
		},
	})
}
