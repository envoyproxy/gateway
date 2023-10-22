// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

func buildXdsRoute(httpRoute *ir.HTTPRoute) *routev3.Route {
	router := &routev3.Route{
		Name:  httpRoute.Name,
		Match: buildXdsRouteMatch(httpRoute.PathMatch, httpRoute.HeaderMatches, httpRoute.QueryParamMatches),
	}

	if len(httpRoute.AddRequestHeaders) > 0 {
		router.RequestHeadersToAdd = buildXdsAddedHeaders(httpRoute.AddRequestHeaders)
	}
	if len(httpRoute.RemoveRequestHeaders) > 0 {
		router.RequestHeadersToRemove = httpRoute.RemoveRequestHeaders
	}

	if len(httpRoute.AddResponseHeaders) > 0 {
		router.ResponseHeadersToAdd = buildXdsAddedHeaders(httpRoute.AddResponseHeaders)
	}
	if len(httpRoute.RemoveResponseHeaders) > 0 {
		router.ResponseHeadersToRemove = httpRoute.RemoveResponseHeaders
	}

	switch {
	case httpRoute.DirectResponse != nil:
		router.Action = &routev3.Route_DirectResponse{DirectResponse: buildXdsDirectResponseAction(httpRoute.DirectResponse)}
	case httpRoute.Redirect != nil:
		router.Action = &routev3.Route_Redirect{Redirect: buildXdsRedirectAction(httpRoute.Redirect)}
	case httpRoute.URLRewrite != nil:
		routeAction := buildXdsURLRewriteAction(httpRoute.Destination.Name, httpRoute.URLRewrite)
		if httpRoute.Mirrors != nil {
			routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
		}

		router.Action = &routev3.Route_Route{Route: routeAction}
	default:
		if httpRoute.BackendWeights.Invalid != 0 {
			// If there are invalid backends then a weighted cluster is required for the route
			routeAction := buildXdsWeightedRouteAction(httpRoute)
			if httpRoute.Mirrors != nil {
				routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
			}
			router.Action = &routev3.Route_Route{Route: routeAction}
		} else {
			routeAction := buildXdsRouteAction(httpRoute.Destination.Name)
			if httpRoute.Mirrors != nil {
				routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
			}
			router.Action = &routev3.Route_Route{Route: routeAction}
		}
	}

	// Hash Policy
	if router.GetRoute() != nil {
		router.GetRoute().HashPolicy = buildHashPolicy(httpRoute)
	}

	// Timeouts
	if router.GetRoute() != nil && httpRoute.Timeout != nil {
		router.GetRoute().Timeout = durationpb.New(httpRoute.Timeout.Duration)
	}

	// Add per route filter configs to the route, if needed.
	if err := patchRouteWithFilters(router, httpRoute); err != nil {
		return nil
	}

	return router
}

func buildXdsRouteMatch(pathMatch *ir.StringMatch, headerMatches []*ir.StringMatch, queryParamMatches []*ir.StringMatch) *routev3.RouteMatch {
	outMatch := &routev3.RouteMatch{}

	// Add a prefix match to '/' if no matches are specified
	if pathMatch == nil {
		// Setup default path specifier. It may be overwritten by :host:.
		outMatch.PathSpecifier = &routev3.RouteMatch_Prefix{
			Prefix: "/",
		}
	} else {
		// Path match
		//nolint:gocritic
		if pathMatch.Exact != nil {
			outMatch.PathSpecifier = &routev3.RouteMatch_Path{
				Path: *pathMatch.Exact,
			}
		} else if pathMatch.Prefix != nil {
			// when the prefix ends with "/", use RouteMatch_Prefix
			if strings.HasSuffix(*pathMatch.Prefix, "/") {
				outMatch.PathSpecifier = &routev3.RouteMatch_Prefix{
					Prefix: *pathMatch.Prefix,
				}
			} else {
				outMatch.PathSpecifier = &routev3.RouteMatch_PathSeparatedPrefix{
					PathSeparatedPrefix: *pathMatch.Prefix,
				}
			}
		} else if pathMatch.SafeRegex != nil {
			outMatch.PathSpecifier = &routev3.RouteMatch_SafeRegex{
				SafeRegex: &matcherv3.RegexMatcher{
					Regex: *pathMatch.SafeRegex,
				},
			}
		}
	}
	// Header matches
	for _, headerMatch := range headerMatches {
		stringMatcher := buildXdsStringMatcher(headerMatch)

		headerMatcher := &routev3.HeaderMatcher{
			Name: headerMatch.Name,
			HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
				StringMatch: stringMatcher,
			},
		}
		outMatch.Headers = append(outMatch.Headers, headerMatcher)
	}

	// Query param matches
	for _, queryParamMatch := range queryParamMatches {
		stringMatcher := buildXdsStringMatcher(queryParamMatch)

		queryParamMatcher := &routev3.QueryParameterMatcher{
			Name: queryParamMatch.Name,
			QueryParameterMatchSpecifier: &routev3.QueryParameterMatcher_StringMatch{
				StringMatch: stringMatcher,
			},
		}
		outMatch.QueryParameters = append(outMatch.QueryParameters, queryParamMatcher)
	}

	return outMatch
}

func buildXdsStringMatcher(irMatch *ir.StringMatch) *matcherv3.StringMatcher {
	stringMatcher := new(matcherv3.StringMatcher)

	//nolint:gocritic
	if irMatch.Exact != nil {
		stringMatcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Exact{
				Exact: *irMatch.Exact,
			},
		}
	} else if irMatch.Prefix != nil {
		stringMatcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Prefix{
				Prefix: *irMatch.Prefix,
			},
		}
	} else if irMatch.Suffix != nil {
		stringMatcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_Suffix{
				Suffix: *irMatch.Suffix,
			},
		}
	} else if irMatch.SafeRegex != nil {
		stringMatcher = &matcherv3.StringMatcher{
			MatchPattern: &matcherv3.StringMatcher_SafeRegex{
				SafeRegex: &matcherv3.RegexMatcher{
					Regex: *irMatch.SafeRegex,
				},
			},
		}
	}

	return stringMatcher
}

func buildXdsRouteAction(destName string) *routev3.RouteAction {
	return &routev3.RouteAction{
		ClusterSpecifier: &routev3.RouteAction_Cluster{
			Cluster: destName,
		},
	}
}

func buildXdsWeightedRouteAction(httpRoute *ir.HTTPRoute) *routev3.RouteAction {
	clusters := []*routev3.WeightedCluster_ClusterWeight{
		{
			Name:   "invalid-backend-cluster",
			Weight: &wrapperspb.UInt32Value{Value: httpRoute.BackendWeights.Invalid},
		},
	}

	if httpRoute.BackendWeights.Valid > 0 {
		validCluster := &routev3.WeightedCluster_ClusterWeight{
			Name:   httpRoute.Destination.Name,
			Weight: &wrapperspb.UInt32Value{Value: httpRoute.BackendWeights.Valid},
		}
		clusters = append(clusters, validCluster)
	}

	return &routev3.RouteAction{
		// Intentionally route to a non-existent cluster and return a 500 error when it is not found
		ClusterNotFoundResponseCode: routev3.RouteAction_INTERNAL_SERVER_ERROR,
		ClusterSpecifier: &routev3.RouteAction_WeightedClusters{
			WeightedClusters: &routev3.WeightedCluster{
				Clusters: clusters,
			},
		},
	}
}

func buildXdsRedirectAction(redirection *ir.Redirect) *routev3.RedirectAction {
	routeAction := &routev3.RedirectAction{}

	if redirection.Scheme != nil {
		routeAction.SchemeRewriteSpecifier = &routev3.RedirectAction_SchemeRedirect{
			SchemeRedirect: *redirection.Scheme,
		}
	}
	if redirection.Path != nil {
		if redirection.Path.FullReplace != nil {
			routeAction.PathRewriteSpecifier = &routev3.RedirectAction_PathRedirect{
				PathRedirect: *redirection.Path.FullReplace,
			}
		} else if redirection.Path.PrefixMatchReplace != nil {
			routeAction.PathRewriteSpecifier = &routev3.RedirectAction_PrefixRewrite{
				PrefixRewrite: *redirection.Path.PrefixMatchReplace,
			}
		}
	}
	if redirection.Hostname != nil {
		routeAction.HostRedirect = *redirection.Hostname
	}
	if redirection.Port != nil {
		routeAction.PortRedirect = *redirection.Port
	}
	if redirection.StatusCode != nil {
		if *redirection.StatusCode == 302 {
			routeAction.ResponseCode = routev3.RedirectAction_FOUND
		} // no need to check for 301 since Envoy will use 301 as the default if the field is not configured
	}

	return routeAction
}

func buildXdsURLRewriteAction(destName string, urlRewrite *ir.URLRewrite) *routev3.RouteAction {
	routeAction := &routev3.RouteAction{
		ClusterSpecifier: &routev3.RouteAction_Cluster{
			Cluster: destName,
		},
	}

	if urlRewrite.Path != nil {
		if urlRewrite.Path.FullReplace != nil {
			routeAction.RegexRewrite = &matcherv3.RegexMatchAndSubstitute{
				Pattern: &matcherv3.RegexMatcher{
					Regex: "/.+",
				},
				Substitution: *urlRewrite.Path.FullReplace,
			}
		} else if urlRewrite.Path.PrefixMatchReplace != nil {
			routeAction.PrefixRewrite = *urlRewrite.Path.PrefixMatchReplace
		}
	}

	if urlRewrite.Hostname != nil {
		routeAction.HostRewriteSpecifier = &routev3.RouteAction_HostRewriteLiteral{
			HostRewriteLiteral: *urlRewrite.Hostname,
		}

		routeAction.AppendXForwardedHost = true
	}

	return routeAction
}

func buildXdsDirectResponseAction(res *ir.DirectResponse) *routev3.DirectResponseAction {
	routeAction := &routev3.DirectResponseAction{Status: res.StatusCode}

	if res.Body != nil {
		routeAction.Body = &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{
				InlineString: *res.Body,
			},
		}
	}

	return routeAction
}

func buildXdsRequestMirrorPolicies(mirrorDestinations []*ir.RouteDestination) []*routev3.RouteAction_RequestMirrorPolicy {
	var mirrorPolicies []*routev3.RouteAction_RequestMirrorPolicy

	for _, mirrorDest := range mirrorDestinations {
		mirrorPolicies = append(mirrorPolicies, &routev3.RouteAction_RequestMirrorPolicy{
			Cluster: mirrorDest.Name,
		})
	}

	return mirrorPolicies
}

func buildXdsAddedHeaders(headersToAdd []ir.AddHeader) []*corev3.HeaderValueOption {
	headerValueOptions := make([]*corev3.HeaderValueOption, len(headersToAdd))

	for i, header := range headersToAdd {
		headerValueOptions[i] = &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   header.Name,
				Value: header.Value,
			},
			Append: &wrapperspb.BoolValue{Value: header.Append},
		}

		// Allow empty headers to be set, but don't add the config to do so unless necessary
		if header.Value == "" {
			headerValueOptions[i].KeepEmptyValue = true
		}
	}

	return headerValueOptions
}

func buildHashPolicy(httpRoute *ir.HTTPRoute) []*routev3.RouteAction_HashPolicy {
	// Return early
	if httpRoute == nil || httpRoute.LoadBalancer == nil || httpRoute.LoadBalancer.ConsistentHash == nil {
		return nil
	}

	if httpRoute.LoadBalancer.ConsistentHash.SourceIP != nil && *httpRoute.LoadBalancer.ConsistentHash.SourceIP {
		hashPolicy := &routev3.RouteAction_HashPolicy{
			PolicySpecifier: &routev3.RouteAction_HashPolicy_ConnectionProperties_{
				ConnectionProperties: &routev3.RouteAction_HashPolicy_ConnectionProperties{
					SourceIp: true,
				},
			},
		}
		return []*routev3.RouteAction_HashPolicy{hashPolicy}
	}

	return nil
}
