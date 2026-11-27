// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"slices"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	filterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/common/set_filter_state/v3"
	setfilterstatev3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/set_filter_state/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const (
	regexRedirectPathFilter = "envoy.filters.http.set_filter_state/regex_redirect"
	regexRedirectPathKey    = "envoy_gateway.regex_redirect.path"
)

func patchHCMWithRegexRedirectPath(mgr *hcmv3.HttpConnectionManager, listener *ir.HTTPListener) error {
	var pathFilter *hcmv3.HttpFilter
	for _, filter := range mgr.HttpFilters {
		if filter.Name == regexRedirectPathFilter {
			pathFilter = filter
			break
		}
	}
	hasRegexRedirect := false
	for _, route := range listener.Routes {
		if route.DirectResponse != nil || route.Redirect == nil || route.Redirect.Path == nil ||
			route.Redirect.Path.RegexMatchReplace == nil {
			continue
		}
		hasRegexRedirect = true
		break
	}
	if !hasRegexRedirect {
		mgr.HttpFilters = slices.DeleteFunc(mgr.HttpFilters, func(filter *hcmv3.HttpFilter) bool {
			return filter.Name == regexRedirectPathFilter || strings.HasPrefix(filter.Name, regexRedirectPathFilter+"/")
		})
		return nil
	}
	if pathFilter == nil {
		var err error
		pathFilter, err = buildRegexRedirectPathFilter(regexRedirectPathFilter, true)
		if err != nil {
			return err
		}
	}
	filters := slices.DeleteFunc(mgr.HttpFilters, func(filter *hcmv3.HttpFilter) bool {
		return filter.Name == regexRedirectPathFilter || strings.HasPrefix(filter.Name, regexRedirectPathFilter+"/")
	})
	// Populate the guard before any filter can cache a route or apply its policies.
	mgr.HttpFilters = []*hcmv3.HttpFilter{pathFilter}
	for _, filter := range filters {
		mgr.HttpFilters = append(mgr.HttpFilters, filter)
		if isFilterType(filter, egv1a1.EnvoyFilterGRPCJSONTranscoder) {
			// Refresh the path for later route recomputation without unpinning the transcoder's route.
			refresh, err := buildRegexRedirectPathFilter(regexRedirectPathFilter+"/"+filter.Name, false)
			if err != nil {
				return err
			}
			mgr.HttpFilters = append(mgr.HttpFilters, refresh)
		}
	}
	return nil
}

func buildRegexRedirectPathFilter(name string, clearRouteCache bool) (*hcmv3.HttpFilter, error) {
	config, err := proto.ToAnyWithValidation(&setfilterstatev3.Config{
		ClearRouteCache: clearRouteCache,
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
		return nil, err
	}
	return &hcmv3.HttpFilter{
		Name:       name,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{TypedConfig: config},
	}, nil
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
