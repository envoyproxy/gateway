// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"strings"
	"time"

	"github.com/envoyproxy/gateway/internal/utils/protocov"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	previoushost "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/host/previous_hosts/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	retryDefaultRetryOn             = "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes"
	retryDefaultRetriableStatusCode = 503
	retryDefaultNumRetries          = 2
)

func buildXdsRoute(httpRoute *ir.HTTPRoute) (*routev3.Route, error) {
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
		router.Action = &routev3.Route_Redirect{Redirect: buildXdsRedirectAction(httpRoute)}
	case httpRoute.URLRewrite != nil:
		routeAction := buildXdsURLRewriteAction(httpRoute.Destination.Name, httpRoute.URLRewrite, httpRoute.PathMatch)
		if httpRoute.Mirrors != nil {
			routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
		}

		if !httpRoute.IsHTTP2 {
			// Allow websocket upgrades for HTTP 1.1
			// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism
			routeAction.UpgradeConfigs = []*routev3.RouteAction_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
			}
		}

		router.Action = &routev3.Route_Route{Route: routeAction}
	default:
		var routeAction *routev3.RouteAction
		if httpRoute.BackendWeights.Invalid != 0 {
			// If there are invalid backends then a weighted cluster is required for the route
			routeAction = buildXdsWeightedRouteAction(httpRoute)
		} else {
			routeAction = buildXdsRouteAction(httpRoute)
		}
		if httpRoute.Mirrors != nil {
			routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
		}
		if !httpRoute.IsHTTP2 {
			// Allow websocket upgrades for HTTP 1.1
			// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism
			routeAction.UpgradeConfigs = []*routev3.RouteAction_UpgradeConfig{
				{
					UpgradeType: "websocket",
				},
			}
		}
		router.Action = &routev3.Route_Route{Route: routeAction}
	}

	// Hash Policy
	if router.GetRoute() != nil {
		router.GetRoute().HashPolicy = buildHashPolicy(httpRoute)
	}

	// Timeouts
	if router.GetRoute() != nil && httpRoute.Timeout != nil && httpRoute.Timeout.HTTP != nil &&
		httpRoute.Timeout.HTTP.RequestTimeout != nil {
		router.GetRoute().Timeout = durationpb.New(httpRoute.Timeout.HTTP.RequestTimeout.Duration)
	}

	// Retries
	if router.GetRoute() != nil && httpRoute.Retry != nil {
		if rp, err := buildRetryPolicy(httpRoute); err == nil {
			router.GetRoute().RetryPolicy = rp
		} else {
			return nil, err
		}
	}

	// Add per route filter configs to the route, if needed.
	if err := patchRouteWithPerRouteConfig(router, httpRoute); err != nil {
		return nil, err
	}

	return router, nil
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
			if *pathMatch.Prefix == "/" {
				outMatch.PathSpecifier = &routev3.RouteMatch_Prefix{
					Prefix: "/",
				}
			} else {
				// Remove trailing /
				trimmedPrefix := strings.TrimSuffix(*pathMatch.Prefix, "/")
				outMatch.PathSpecifier = &routev3.RouteMatch_PathSeparatedPrefix{
					PathSeparatedPrefix: trimmedPrefix,
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

func buildXdsRouteAction(httpRoute *ir.HTTPRoute) *routev3.RouteAction {
	return &routev3.RouteAction{
		ClusterSpecifier: &routev3.RouteAction_Cluster{
			Cluster: httpRoute.Destination.Name,
		},
		IdleTimeout: idleTimeout(httpRoute),
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
		IdleTimeout: idleTimeout(httpRoute),
	}
}

func idleTimeout(httpRoute *ir.HTTPRoute) *durationpb.Duration {
	if httpRoute.Timeout != nil && httpRoute.Timeout.HTTP != nil {
		if httpRoute.Timeout.HTTP.RequestTimeout != nil {
			timeout := time.Hour // Default to 1 hour

			// Ensure is not less than the request timeout
			if timeout < httpRoute.Timeout.HTTP.RequestTimeout.Duration {
				timeout = httpRoute.Timeout.HTTP.RequestTimeout.Duration
			}

			// Disable idle timeout when request timeout is disabled
			if httpRoute.Timeout.HTTP.RequestTimeout.Duration == 0 {
				timeout = 0
			}

			return durationpb.New(timeout)
		}
	}
	return nil
}

func buildXdsRedirectAction(httpRoute *ir.HTTPRoute) *routev3.RedirectAction {
	var (
		redirection = httpRoute.Redirect
		routeAction = &routev3.RedirectAction{}
	)

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
			if useRegexRewriteForPrefixMatchReplace(httpRoute.PathMatch, *redirection.Path.PrefixMatchReplace) {
				routeAction.PathRewriteSpecifier = &routev3.RedirectAction_RegexRewrite{
					RegexRewrite: prefix2RegexRewrite(*httpRoute.PathMatch.Prefix),
				}
			} else {
				routeAction.PathRewriteSpecifier = &routev3.RedirectAction_PrefixRewrite{
					PrefixRewrite: *redirection.Path.PrefixMatchReplace,
				}
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

// useRegexRewriteForPrefixMatchReplace checks if the regex rewrite should be used for prefix match replace
// due to the issue with Envoy not handling the case of "//" when the replace string is "/".
// See: https://github.com/envoyproxy/envoy/issues/26055
func useRegexRewriteForPrefixMatchReplace(pathMatch *ir.StringMatch, prefixMatchReplace string) bool {
	return pathMatch != nil &&
		pathMatch.Prefix != nil &&
		(prefixMatchReplace == "" || prefixMatchReplace == "/")
}

func prefix2RegexRewrite(prefix string) *matcherv3.RegexMatchAndSubstitute {
	return &matcherv3.RegexMatchAndSubstitute{
		Pattern: &matcherv3.RegexMatcher{
			Regex: "^" + prefix + `\/*`,
		},
		Substitution: "/",
	}
}

func buildXdsURLRewriteAction(destName string, urlRewrite *ir.URLRewrite, pathMatch *ir.StringMatch) *routev3.RouteAction {
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
			// Circumvent the case of "//" when the replace string is "/"
			// An empty replace string does not seem to solve the issue so we are using
			// a regex match and replace instead
			// Remove this workaround once https://github.com/envoyproxy/envoy/issues/26055 is fixed
			if useRegexRewriteForPrefixMatchReplace(pathMatch, *urlRewrite.Path.PrefixMatchReplace) {
				routeAction.RegexRewrite = prefix2RegexRewrite(*pathMatch.Prefix)
			} else {
				routeAction.PrefixRewrite = *urlRewrite.Path.PrefixMatchReplace
			}
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
		var appendAction corev3.HeaderValueOption_HeaderAppendAction

		if header.Append {
			appendAction = corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		} else {
			appendAction = corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		}

		headerValueOptions[i] = &corev3.HeaderValueOption{
			Header: &corev3.HeaderValue{
				Key:   header.Name,
				Value: header.Value,
			},
			AppendAction: appendAction,
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

func buildRetryPolicy(route *ir.HTTPRoute) (*routev3.RetryPolicy, error) {
	if route.Retry != nil {
		rr := route.Retry
		rp := &routev3.RetryPolicy{
			RetryOn:              retryDefaultRetryOn,
			RetriableStatusCodes: []uint32{retryDefaultRetriableStatusCode},
			NumRetries:           &wrapperspb.UInt32Value{Value: retryDefaultNumRetries},
			RetryHostPredicate: []*routev3.RetryPolicy_RetryHostPredicate{
				{
					Name: "envoy.retry_host_predicates.previous_hosts",
					ConfigType: &routev3.RetryPolicy_RetryHostPredicate_TypedConfig{
						TypedConfig: protocov.ToAny(&previoushost.PreviousHostsPredicate{}),
					},
				},
			},
			HostSelectionRetryMaxAttempts: 5,
		}

		if rr.NumRetries != nil {
			rp.NumRetries = &wrapperspb.UInt32Value{Value: *rr.NumRetries}
		}

		if rr.RetryOn != nil {
			if rr.RetryOn.Triggers != nil && len(rr.RetryOn.Triggers) > 0 {
				if ro, err := buildRetryOn(rr.RetryOn.Triggers); err == nil {
					rp.RetryOn = ro
				} else {
					return nil, err
				}
			}

			if rr.RetryOn.HTTPStatusCodes != nil && len(rr.RetryOn.HTTPStatusCodes) > 0 {
				rp.RetriableStatusCodes = buildRetryStatusCodes(rr.RetryOn.HTTPStatusCodes)
			}
		}

		if rr.PerRetry != nil {
			if rr.PerRetry.Timeout != nil {
				rp.PerTryTimeout = durationpb.New(rr.PerRetry.Timeout.Duration)
			}

			if rr.PerRetry.BackOff != nil {
				bbo := false
				rbo := &routev3.RetryPolicy_RetryBackOff{}
				if rr.PerRetry.BackOff.BaseInterval != nil {
					rbo.BaseInterval = durationpb.New(rr.PerRetry.BackOff.BaseInterval.Duration)
					bbo = true
				}

				if rr.PerRetry.BackOff.MaxInterval != nil {
					rbo.MaxInterval = durationpb.New(rr.PerRetry.BackOff.MaxInterval.Duration)
					bbo = true
				}

				if bbo {
					rp.RetryBackOff = rbo
				}
			}
		}
		return rp, nil
	}

	return nil, nil
}

func buildRetryStatusCodes(codes []ir.HTTPStatus) []uint32 {
	ret := make([]uint32, len(codes))
	for i, c := range codes {
		ret[i] = uint32(c)
	}
	return ret
}

// buildRetryOn concatenates triggers to a comma-delimited string.
// An error is returned if a trigger is not in the list of supported values (not likely, due to prior validations).
func buildRetryOn(triggers []ir.TriggerEnum) (string, error) {
	if len(triggers) == 0 {
		return "", nil
	}

	lookup := map[ir.TriggerEnum]string{
		ir.Error5XX:             "5xx",
		ir.GatewayError:         "gateway-error",
		ir.Reset:                "reset",
		ir.ConnectFailure:       "connect-failure",
		ir.Retriable4XX:         "retriable-4xx",
		ir.RefusedStream:        "refused-stream",
		ir.RetriableStatusCodes: "retriable-status-codes",
		ir.Cancelled:            "cancelled",
		ir.DeadlineExceeded:     "deadline-exceeded",
		ir.Internal:             "internal",
		ir.ResourceExhausted:    "resource-exhausted",
		ir.Unavailable:          "unavailable",
	}

	var b strings.Builder

	if t, found := lookup[triggers[0]]; found {
		b.WriteString(t)
	} else {
		return "", errors.New("unsupported RetryOn trigger")
	}

	for _, v := range triggers[1:] {
		if t, found := lookup[v]; found {
			b.WriteString(",")
			b.WriteString(t)
		} else {
			return "", errors.New("unsupported RetryOn trigger")
		}
	}

	return b.String(), nil
}
