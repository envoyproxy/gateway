package translator

import (
	core "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	route "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcher "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	wrappers "github.com/golang/protobuf/ptypes/wrappers"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsAddedRequestHeaders(headersToAdd []ir.AddHeader) []*core.HeaderValueOption {
	ret := make([]*core.HeaderValueOption, len(headersToAdd))

	for i, header := range headersToAdd {
		ret[i] = &core.HeaderValueOption{
			Header: &core.HeaderValue{
				Key:   header.Name,
				Value: header.Value,
			},
			Append:         &wrappers.BoolValue{Value: header.Append},
			KeepEmptyValue: header.AllowEmpty,
		}
	}

	return ret
}

func buildXdsRoute(httpRoute *ir.HTTPRoute) (*route.Route, error) {
	ret := &route.Route{
		Match: buildXdsRouteMatch(httpRoute.PathMatch, httpRoute.HeaderMatches, httpRoute.QueryParamMatches),
	}
	if httpRoute.AddRequestHeaders != nil {
		ret.RequestHeadersToAdd = buildXdsAddedRequestHeaders(*httpRoute.AddRequestHeaders)
	}
	if httpRoute.RemoveRequestHeaders != nil {
		ret.RequestHeadersToRemove = *httpRoute.RemoveRequestHeaders
	}

	switch {
	case httpRoute.DirectResponse != nil:
		ret.Action = &route.Route_DirectResponse{DirectResponse: buildXdsDirectResponseAction(httpRoute.DirectResponse)}
	case httpRoute.Redirect != nil:
		ret.Action = &route.Route_Redirect{Redirect: buildXdsRedirectAction(httpRoute.Redirect)}
	default:
		ret.Action = &route.Route_Route{Route: buildXdsRouteAction(httpRoute.Name)}
	}

	return ret, nil
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

func buildXdsRedirectAction(redirection *ir.Redirect) *route.RedirectAction {
	ret := &route.RedirectAction{}

	if redirection.Scheme != nil {
		ret.SchemeRewriteSpecifier = &route.RedirectAction_SchemeRedirect{
			SchemeRedirect: *redirection.Scheme,
		}
	}
	if redirection.Path != nil {
		if redirection.Path.FullReplace != nil {
			ret.PathRewriteSpecifier = &route.RedirectAction_PathRedirect{
				PathRedirect: *redirection.Path.FullReplace,
			}
		} else if redirection.Path.PrefixMatchReplace != nil {
			ret.PathRewriteSpecifier = &route.RedirectAction_PrefixRewrite{
				PrefixRewrite: *redirection.Path.PrefixMatchReplace,
			}
		}
	}
	if redirection.Hostname != nil {
		ret.HostRedirect = *redirection.Hostname
	}
	if redirection.Port != nil {
		ret.PortRedirect = *redirection.Port
	}
	if redirection.StatusCode != nil {
		ret.ResponseCode = route.RedirectAction_RedirectResponseCode(*redirection.StatusCode)
	} else {
		ret.ResponseCode = 302 // default
	}

	return ret
}

func buildXdsDirectResponseAction(res *ir.DirectResponse) *route.DirectResponseAction {
	ret := &route.DirectResponseAction{Status: res.StatusCode}

	if res.Body != nil {
		ret.Body = &core.DataSource{
			Specifier: &core.DataSource_InlineString{
				InlineString: *res.Body,
			},
		}
	}

	return ret
}
