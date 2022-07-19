package translator

import (
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsRoute(httpRoute *ir.HTTPRoute) (*route.Route, error) {
	return &route.Route{
		Match:  buildXdsRouteMatch(httpRoute.PathMatch, httpRoute.HeaderMatches, httpRoute.QueryParamMatches),
		Action: &route.Route_Route{Route: buildXdsRouteAction(httpRoute.Name)},
	}, nil
}

func buildXdsRouteMatch(pathMatch *ir.StringMatch, headerMatches []*ir.StringMatch, queryParamMatches []*ir.StringMatch) *route.RouteMatch {
	outMatch := &route.RouteMatch{}

	// Return early with a prefix match to '/' if no matches are specified
	if pathMatch == nil && len(headerMatches) == 0 && len(queryParamMatches) == 0 {
		// Setup default path specifier. It may be overwritten by :host:.
		outMatch.PathSpecifier = &route.RouteMatch_Prefix{
			Prefix: "/",
		}
		return outMatch
	}

	// Path match
	//nolint:gocritic
	if pathMatch != nil {
		if pathMatch.Exact != nil {
			outMatch.PathSpecifier = &route.RouteMatch_Path{
				Path: *pathMatch.Exact,
			}
		} else if pathMatch.Prefix != nil {
			outMatch.PathSpecifier = &route.RouteMatch_Prefix{
				Prefix: *pathMatch.Prefix,
			}
		} else if pathMatch.SafeRegex != nil {
			outMatch.PathSpecifier = &route.RouteMatch_SafeRegex{
				SafeRegex: &matcher.RegexMatcher{
					EngineType: &matcher.RegexMatcher_GoogleRe2{},
					Regex:      *pathMatch.SafeRegex,
				},
			}
		}
	}

	// Header matches
	for _, headerMatch := range headerMatches {
		stringMatcher := buildXdsStringMatcher(headerMatch)

		headerMatcher := &route.HeaderMatcher{
			Name: headerMatch.Name,
			HeaderMatchSpecifier: &route.HeaderMatcher_StringMatch{
				StringMatch: stringMatcher,
			},
		}
		outMatch.Headers = append(outMatch.Headers, headerMatcher)
	}

	// Query param matches
	for _, queryParamMatch := range queryParamMatches {
		stringMatcher := buildXdsStringMatcher(queryParamMatch)

		queryParamMatcher := &route.QueryParameterMatcher{
			Name: queryParamMatch.Name,
			QueryParameterMatchSpecifier: &route.QueryParameterMatcher_StringMatch{
				StringMatch: stringMatcher,
			},
		}
		outMatch.QueryParameters = append(outMatch.QueryParameters, queryParamMatcher)
	}

	return outMatch
}

func buildXdsStringMatcher(irMatch *ir.StringMatch) *matcher.StringMatcher {
	stringMatcher := new(matcher.StringMatcher)

	//nolint:gocritic
	if irMatch.Exact != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_Exact{
				Exact: *irMatch.Exact,
			},
		}
	} else if irMatch.Prefix != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_Prefix{
				Prefix: *irMatch.Prefix,
			},
		}
	} else if irMatch.SafeRegex != nil {
		stringMatcher = &matcher.StringMatcher{
			MatchPattern: &matcher.StringMatcher_SafeRegex{
				SafeRegex: &matcher.RegexMatcher{
					Regex: *irMatch.SafeRegex,
					EngineType: &matcher.RegexMatcher_GoogleRe2{
						GoogleRe2: &matcher.RegexMatcher_GoogleRE2{},
					},
				},
			},
		}
	}

	return stringMatcher
}

func buildXdsRouteAction(routeName string) *route.RouteAction {
	return &route.RouteAction{
		ClusterSpecifier: &route.RouteAction_Cluster{
			Cluster: getXdsClusterName(routeName),
		},
	}
}
