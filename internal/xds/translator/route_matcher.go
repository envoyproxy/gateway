// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strings"

	cncfcorev3 "github.com/cncf/xds/go/xds/core/v3"
	xdsmatcherv3 "github.com/cncf/xds/go/xds/type/matcher/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	envoymatcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
)

const pathHeaderName = ":path"

// pathRouteGroup holds a path key and the corresponding IR routes and built xDS routes.
type pathRouteGroup struct {
	pathKey   string
	irRoutes  []*ir.HTTPRoute
	xdsRoutes []*routev3.Route
}

// buildVirtualHostMatcher builds an xDS Matcher tree for sublinear route matching on :path.
// The Generic Matcher API supports exact_match_map (O(1) hash) and prefix_match_map (trie, longest-prefix);
// we use exact_match_map for exact path. Path prefixes are grouped in fallback (linear RouteList) because
// Gateway API uses path-segment semantics (PathSeparatedPrefix), which prefix_match_map does not provide.
// Order: exact path match first, then prefix tree if used, then fallback RouteList (regex/default/prefix).
func buildVirtualHostMatcher(
	exactGroups []pathRouteGroup,
	prefixGroups []pathRouteGroup,
	fallbackRoutes []*routev3.Route,
) (*xdsmatcherv3.Matcher, error) {
	pathInput, err := buildPathMatchInput()
	if err != nil {
		return nil, err
	}

	exactMap := make(map[string]*xdsmatcherv3.Matcher_OnMatch)
	for _, g := range exactGroups {
		onMatch, err := buildRouteOnMatch(g.xdsRoutes, g.irRoutes)
		if err != nil {
			return nil, err
		}
		exactMap[g.pathKey] = onMatch
	}

	prefixMap := make(map[string]*xdsmatcherv3.Matcher_OnMatch)
	for _, g := range prefixGroups {
		onMatch, err := buildRouteOnMatch(g.xdsRoutes, g.irRoutes)
		if err != nil {
			return nil, err
		}
		prefixMap[g.pathKey] = onMatch
	}

	// Innermost on_no_match: fallback RouteList (linear matching for regex/default routes)
	var fallbackOnMatch *xdsmatcherv3.Matcher_OnMatch
	if len(fallbackRoutes) > 0 {
		fallbackAction, err := routeListActionConfig(fallbackRoutes)
		if err != nil {
			return nil, err
		}
		fallbackOnMatch = &xdsmatcherv3.Matcher_OnMatch{
			OnMatch: &xdsmatcherv3.Matcher_OnMatch_Action{Action: fallbackAction},
		}
	}

	// When we have exact match, on_no_match is prefix tree (then fallback). When only prefix, on_no_match is fallback.
	var prefixThenFallback *xdsmatcherv3.Matcher_OnMatch
	if len(prefixMap) > 0 {
		prefixTree := &xdsmatcherv3.Matcher_MatcherTree{
			Input: pathInput,
			TreeType: &xdsmatcherv3.Matcher_MatcherTree_PrefixMatchMap{
				PrefixMatchMap: &xdsmatcherv3.Matcher_MatcherTree_MatchMap{Map: prefixMap},
			},
		}
		prefixThenFallback = &xdsmatcherv3.Matcher_OnMatch{
			OnMatch: &xdsmatcherv3.Matcher_OnMatch_Matcher{
				Matcher: &xdsmatcherv3.Matcher{
					MatcherType: &xdsmatcherv3.Matcher_MatcherTree_{MatcherTree: prefixTree},
					OnNoMatch:   fallbackOnMatch,
				},
			},
		}
	} else {
		prefixThenFallback = fallbackOnMatch
	}

	// Top level: exact_match_map (if any), then prefix or fallback. Envoy requires at least one map entry.
	if len(exactMap) > 0 {
		topTree := &xdsmatcherv3.Matcher_MatcherTree{
			Input: pathInput,
			TreeType: &xdsmatcherv3.Matcher_MatcherTree_ExactMatchMap{
				ExactMatchMap: &xdsmatcherv3.Matcher_MatcherTree_MatchMap{Map: exactMap},
			},
		}
		return &xdsmatcherv3.Matcher{
			MatcherType: &xdsmatcherv3.Matcher_MatcherTree_{MatcherTree: topTree},
			OnNoMatch:   prefixThenFallback,
		}, nil
	}
	// Only prefix and/or fallback: use prefix tree as top level.
	if len(prefixMap) > 0 {
		topTree := &xdsmatcherv3.Matcher_MatcherTree{
			Input: pathInput,
			TreeType: &xdsmatcherv3.Matcher_MatcherTree_PrefixMatchMap{
				PrefixMatchMap: &xdsmatcherv3.Matcher_MatcherTree_MatchMap{Map: prefixMap},
			},
		}
		return &xdsmatcherv3.Matcher{
			MatcherType: &xdsmatcherv3.Matcher_MatcherTree_{MatcherTree: topTree},
			OnNoMatch:   fallbackOnMatch,
		}, nil
	}
	// Fallback: exactMap and prefixMap are both empty (all routes were default "/", regex, or no path).
	// Envoy requires at least one map entry; caller should use linear vHost.Routes instead of matcher.
	return nil, nil
}

func buildPathMatchInput() (*cncfcorev3.TypedExtensionConfig, error) {
	input := &envoymatcherv3.HttpRequestHeaderMatchInput{
		HeaderName: pathHeaderName,
	}
	anyInput, err := anypb.New(input)
	if err != nil {
		return nil, err
	}
	return &cncfcorev3.TypedExtensionConfig{
		Name:        "request-headers",
		TypedConfig: anyInput,
	}, nil
}

func buildRouteOnMatch(xdsRoutes []*routev3.Route, _ []*ir.HTTPRoute) (*xdsmatcherv3.Matcher_OnMatch, error) {
	var action *cncfcorev3.TypedExtensionConfig
	if len(xdsRoutes) == 1 {
		var err error
		action, err = routeActionConfig(xdsRoutes[0])
		if err != nil {
			return nil, err
		}
	} else {
		var err error
		action, err = routeListActionConfig(xdsRoutes)
		if err != nil {
			return nil, err
		}
	}
	return &xdsmatcherv3.Matcher_OnMatch{
		OnMatch: &xdsmatcherv3.Matcher_OnMatch_Action{Action: action},
	}, nil
}

func routeActionConfig(route *routev3.Route) (*cncfcorev3.TypedExtensionConfig, error) {
	anyRoute, err := proto.ToAnyWithValidation(route)
	if err != nil {
		return nil, err
	}
	return &cncfcorev3.TypedExtensionConfig{
		Name:        "route",
		TypedConfig: anyRoute,
	}, nil
}

func routeListActionConfig(routes []*routev3.Route) (*cncfcorev3.TypedExtensionConfig, error) {
	list := &routev3.RouteList{Routes: routes}
	anyList, err := proto.ToAnyWithValidation(list)
	if err != nil {
		return nil, err
	}
	return &cncfcorev3.TypedExtensionConfig{
		Name:        "route_list",
		TypedConfig: anyList,
	}, nil
}

// groupRoutesByPath groups (irRoute, xdsRoute) pairs by path for matcher tree construction.
// Returns exact path groups, prefix path groups, and fallback routes.
// Uses exact_match_map for exact path and prefix_match_map for path prefix (with keys that preserve
// path-segment semantics: prefix key is trimmed+"/" so /v2/ matches /v2/foo but not /v2example; exact key
// trimmed is used so request path /v2 also matches).
func groupRoutesByPath(irRoutes []*ir.HTTPRoute, xdsRoutes []*routev3.Route) (
	exactGroups []pathRouteGroup,
	prefixGroups []pathRouteGroup,
	fallback []*routev3.Route,
) {
	exactKeyToGroup := make(map[string]*pathRouteGroup)
	prefixKeyToGroup := make(map[string]*pathRouteGroup)

	for i := range irRoutes {
		irRoute := irRoutes[i]
		xdsRoute := xdsRoutes[i]
		pathMatch := irRoute.PathMatch

		// Fallback: no path match (default "/") — matched in order by linear RouteList.
		if pathMatch == nil {
			fallback = append(fallback, xdsRoute)
			continue
		}
		// Fallback: regex path — not supported in sublinear exact/prefix maps.
		if pathMatch.SafeRegex != nil {
			fallback = append(fallback, xdsRoute)
			continue
		}
		// Fallback: path prefix "/" only — preserve order with other routes.
		if pathMatch.Prefix != nil && strings.TrimSuffix(*pathMatch.Prefix, "/") == "" {
			fallback = append(fallback, xdsRoute)
			continue
		}

		// Exact path → sublinear exact_match_map (O(1)).
		if pathMatch.Exact != nil {
			key := *pathMatch.Exact
			if g := exactKeyToGroup[key]; g != nil {
				g.irRoutes = append(g.irRoutes, irRoute)
				g.xdsRoutes = append(g.xdsRoutes, xdsRoute)
			} else {
				exactKeyToGroup[key] = &pathRouteGroup{
					pathKey:   key,
					irRoutes:  []*ir.HTTPRoute{irRoute},
					xdsRoutes: []*routev3.Route{xdsRoute},
				}
			}
			continue
		}

		// Path prefix (e.g. /v2/) → sublinear prefix_match_map with key "/v2/" (trie matches /v2/foo, not /v2example).
		// Also register exact key "/v2" so request path /v2 matches (PathSeparatedPrefix semantics).
		if pathMatch.Prefix != nil {
			trimmed := strings.TrimSuffix(*pathMatch.Prefix, "/")
			prefixKey := trimmed + "/"
			if g := prefixKeyToGroup[prefixKey]; g != nil {
				g.irRoutes = append(g.irRoutes, irRoute)
				g.xdsRoutes = append(g.xdsRoutes, xdsRoute)
			} else {
				grp := &pathRouteGroup{
					pathKey:   prefixKey,
					irRoutes:  []*ir.HTTPRoute{irRoute},
					xdsRoutes: []*routev3.Route{xdsRoute},
				}
				prefixKeyToGroup[prefixKey] = grp
				if exactKeyToGroup[trimmed] == nil {
					exactKeyToGroup[trimmed] = grp
				}
			}
			continue
		}

		// Fallback: unknown path match type — matched in order by linear RouteList.
		fallback = append(fallback, xdsRoute)
	}

	for exactKey, g := range exactKeyToGroup {
		exactGroups = append(exactGroups, pathRouteGroup{
			pathKey:   exactKey,
			irRoutes:  g.irRoutes,
			xdsRoutes: g.xdsRoutes,
		})
	}
	for _, g := range prefixKeyToGroup {
		prefixGroups = append(prefixGroups, *g)
	}
	return exactGroups, prefixGroups, fallback
}
