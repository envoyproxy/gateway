// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"
	"fmt"
	"strings"
	"time"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	previoushost "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/host/previous_hosts/v3"
	previouspriority "github.com/envoyproxy/go-control-plane/envoy/extensions/retry/priority/previous_priorities/v3"
	matcherv3 "github.com/envoyproxy/go-control-plane/envoy/type/matcher/v3"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/utils/fractionalpercent"
)

const (
	retryDefaultRetryOn             = "connect-failure,refused-stream,unavailable,cancelled,retriable-status-codes"
	retryDefaultRetriableStatusCode = 503
	retryDefaultNumRetries          = 2

	websocketUpgradeType = "websocket"
)

// Allow websocket upgrades for HTTP 1.1
// Reference: https://developer.mozilla.org/en-US/docs/Web/HTTP/Protocol_upgrade_mechanism
var defaultUpgradeConfig = []*routev3.RouteAction_UpgradeConfig{
	{
		UpgradeType: websocketUpgradeType,
	},
}

func buildXdsRoute(httpRoute *ir.HTTPRoute, httpListener *ir.HTTPListener) (*routev3.Route, error) {
	router := &routev3.Route{
		Name:     httpRoute.Name,
		Match:    buildXdsRouteMatch(httpRoute.PathMatch, httpRoute.HeaderMatches, httpRoute.QueryParamMatches),
		Metadata: buildXdsMetadata(httpRoute.Metadata),
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
			routeAction.UpgradeConfigs = buildUpgradeConfig(httpRoute.Traffic)
		}

		router.Action = &routev3.Route_Route{Route: routeAction}
	default:
		backendWeights := httpRoute.Destination.ToBackendWeights()
		routeAction := buildXdsRouteAction(backendWeights, httpRoute.Destination)
		routeAction.IdleTimeout = idleTimeout(httpRoute)

		if httpRoute.Mirrors != nil {
			routeAction.RequestMirrorPolicies = buildXdsRequestMirrorPolicies(httpRoute.Mirrors)
		}
		if !httpRoute.IsHTTP2 {
			routeAction.UpgradeConfigs = buildUpgradeConfig(httpRoute.Traffic)
		}
		router.Action = &routev3.Route_Route{Route: routeAction}
	}

	// Hash Policy
	if router.GetRoute() != nil {
		router.GetRoute().HashPolicy = buildHashPolicy(httpRoute)
	}

	// Timeouts
	if router.GetRoute() != nil {
		rt := getEffectiveRequestTimeout(httpRoute)
		if rt != nil {
			router.GetRoute().Timeout = durationpb.New(rt.Duration)
		}
	}

	// Retries
	if router.GetRoute() != nil &&
		httpRoute.GetRetry() != nil {
		if rp, err := buildRetryPolicy(httpRoute); err == nil {
			router.GetRoute().RetryPolicy = rp
		} else {
			return nil, err
		}
	}

	// Telemetry
	if httpRoute.Traffic != nil && httpRoute.Traffic.Telemetry != nil {
		if tracingCfg, err := buildRouteTracing(httpRoute); err == nil {
			router.Tracing = tracingCfg
		} else {
			return nil, err
		}
	}

	// Add per route filter configs to the route, if needed.
	if err := patchRouteWithPerRouteConfig(router, httpRoute, httpListener); err != nil {
		return nil, err
	}

	return router, nil
}

func buildUpgradeConfig(trafficFeatures *ir.TrafficFeatures) []*routev3.RouteAction_UpgradeConfig {
	if trafficFeatures == nil || trafficFeatures.HTTPUpgrade == nil {
		return defaultUpgradeConfig
	}

	upgradeConfigs := make([]*routev3.RouteAction_UpgradeConfig, 0, len(trafficFeatures.HTTPUpgrade))
	for _, protocol := range trafficFeatures.HTTPUpgrade {
		upgradeConfigs = append(upgradeConfigs, &routev3.RouteAction_UpgradeConfig{
			UpgradeType: protocol,
		})
	}

	return upgradeConfigs
}

func buildXdsRouteMatch(pathMatch *ir.StringMatch, headerMatches, queryParamMatches []*ir.StringMatch) *routev3.RouteMatch {
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

func buildXdsRouteAction(backendWeights *ir.BackendWeights, dest *ir.RouteDestination) *routev3.RouteAction {
	// only use weighted cluster when there are invalid weights
	if dest.NeedsClusterPerSetting() || backendWeights.Invalid != 0 {
		return buildXdsWeightedRouteAction(backendWeights, dest.Settings)
	}

	return &routev3.RouteAction{
		ClusterSpecifier: &routev3.RouteAction_Cluster{
			Cluster: backendWeights.Name,
		},
	}
}

func buildXdsWeightedRouteAction(backendWeights *ir.BackendWeights, settings []*ir.DestinationSetting) *routev3.RouteAction {
	weightedClusters := make([]*routev3.WeightedCluster_ClusterWeight, 0, len(settings))
	if backendWeights.Invalid > 0 {
		invalidCluster := &routev3.WeightedCluster_ClusterWeight{
			Name:   "invalid-backend-cluster",
			Weight: &wrapperspb.UInt32Value{Value: backendWeights.Invalid},
		}
		weightedClusters = append(weightedClusters, invalidCluster)
	}

	for _, destinationSetting := range settings {
		if len(destinationSetting.Endpoints) > 0 || destinationSetting.IsDynamicResolver { // Dynamic resolver has no endpoints
			validCluster := &routev3.WeightedCluster_ClusterWeight{
				Name:   destinationSetting.Name,
				Weight: &wrapperspb.UInt32Value{Value: *destinationSetting.Weight},
			}

			if destinationSetting.Filters != nil {
				if len(destinationSetting.Filters.AddRequestHeaders) > 0 {
					validCluster.RequestHeadersToAdd = append(validCluster.RequestHeadersToAdd, buildXdsAddedHeaders(destinationSetting.Filters.AddRequestHeaders)...)
				}

				if len(destinationSetting.Filters.RemoveRequestHeaders) > 0 {
					validCluster.RequestHeadersToRemove = append(validCluster.RequestHeadersToRemove, destinationSetting.Filters.RemoveRequestHeaders...)
				}

				if len(destinationSetting.Filters.AddResponseHeaders) > 0 {
					validCluster.ResponseHeadersToAdd = append(validCluster.ResponseHeadersToAdd, buildXdsAddedHeaders(destinationSetting.Filters.AddResponseHeaders)...)
				}

				if len(destinationSetting.Filters.RemoveResponseHeaders) > 0 {
					validCluster.ResponseHeadersToRemove = append(validCluster.ResponseHeadersToRemove, destinationSetting.Filters.RemoveResponseHeaders...)
				}
			}

			weightedClusters = append(weightedClusters, validCluster)
		}
	}

	return &routev3.RouteAction{
		// Intentionally route to a non-existent cluster and return a 500 error when it is not found
		ClusterNotFoundResponseCode: routev3.RouteAction_INTERNAL_SERVER_ERROR,
		ClusterSpecifier: &routev3.RouteAction_WeightedClusters{
			WeightedClusters: &routev3.WeightedCluster{
				Clusters: weightedClusters,
			},
		},
	}
}

func getEffectiveRequestTimeout(httpRoute *ir.HTTPRoute) *metav1.Duration {
	// gateway-api timeout takes precedence
	if httpRoute.Timeout != nil {
		return httpRoute.Timeout
	}

	if httpRoute.Traffic != nil &&
		httpRoute.Traffic.Timeout != nil &&
		httpRoute.Traffic.Timeout.HTTP != nil &&
		httpRoute.Traffic.Timeout.HTTP.RequestTimeout != nil {
		return httpRoute.Traffic.Timeout.HTTP.RequestTimeout
	}

	return nil
}

func idleTimeout(httpRoute *ir.HTTPRoute) *durationpb.Duration {
	rt := getEffectiveRequestTimeout(httpRoute)
	timeout := time.Hour // Default to 1 hour
	if rt != nil {
		// Ensure is not less than the request timeout
		if timeout < rt.Duration {
			timeout = rt.Duration
		}

		// Disable idle timeout when request timeout is disabled
		if rt.Duration == 0 {
			timeout = 0
		}

		return durationpb.New(timeout)
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
	// Ignore the redirect port if it is a well-known port number, in order to
	// prevent the port be added in the response's location header.
	if redirection.Port != nil && *redirection.Port != 80 && *redirection.Port != 443 {
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
		switch {
		case urlRewrite.Path.FullReplace != nil:
			routeAction.RegexRewrite = &matcherv3.RegexMatchAndSubstitute{
				Pattern: &matcherv3.RegexMatcher{
					Regex: "^/.*$",
				},
				Substitution: *urlRewrite.Path.FullReplace,
			}
		case urlRewrite.Path.PrefixMatchReplace != nil:
			// Circumvent the case of "//" when the replace string is "/"
			// An empty replace string does not seem to solve the issue so we are using
			// a regex match and replace instead
			// Remove this workaround once https://github.com/envoyproxy/envoy/issues/26055 is fixed
			if useRegexRewriteForPrefixMatchReplace(pathMatch, *urlRewrite.Path.PrefixMatchReplace) {
				routeAction.RegexRewrite = prefix2RegexRewrite(*pathMatch.Prefix)
			} else {
				// remove trailing / to fix #3989
				// when the pathMath.Prefix has suffix / but EG has removed it,
				// and the urlRewrite.Path.PrefixMatchReplace suffix with / the upstream will get unwanted /
				routeAction.PrefixRewrite = strings.TrimSuffix(*urlRewrite.Path.PrefixMatchReplace, "/")
			}
		case urlRewrite.Path.RegexMatchReplace != nil:
			routeAction.RegexRewrite = &matcherv3.RegexMatchAndSubstitute{
				Pattern: &matcherv3.RegexMatcher{
					Regex: urlRewrite.Path.RegexMatchReplace.Pattern,
				},
				Substitution: urlRewrite.Path.RegexMatchReplace.Substitution,
			}
		}
	}

	if urlRewrite.Host != nil {

		switch {
		case urlRewrite.Host.Name != nil:
			routeAction.HostRewriteSpecifier = &routev3.RouteAction_HostRewriteLiteral{
				HostRewriteLiteral: *urlRewrite.Host.Name,
			}
		case urlRewrite.Host.Header != nil:
			routeAction.HostRewriteSpecifier = &routev3.RouteAction_HostRewriteHeader{
				HostRewriteHeader: *urlRewrite.Host.Header,
			}
		case urlRewrite.Host.Backend != nil:
			routeAction.HostRewriteSpecifier = &routev3.RouteAction_AutoHostRewrite{
				AutoHostRewrite: wrapperspb.Bool(true),
			}
		}

		routeAction.AppendXForwardedHost = true
	}

	return routeAction
}

func buildXdsDirectResponseAction(res *ir.CustomResponse) *routev3.DirectResponseAction {
	routeAction := &routev3.DirectResponseAction{}
	if res.StatusCode != nil {
		routeAction.Status = *res.StatusCode
	}

	if res.Body != nil && *res.Body != "" {
		routeAction.Body = &corev3.DataSource{
			Specifier: &corev3.DataSource_InlineString{
				InlineString: *res.Body,
			},
		}
	}

	return routeAction
}

func buildXdsRequestMirrorPolicies(mirrorPolicies []*ir.MirrorPolicy) []*routev3.RouteAction_RequestMirrorPolicy {
	var xdsMirrorPolicies []*routev3.RouteAction_RequestMirrorPolicy

	for _, policy := range mirrorPolicies {
		if mp := mirrorPercentByPolicy(policy); mp != nil && policy.Destination != nil {
			xdsMirrorPolicies = append(xdsMirrorPolicies, &routev3.RouteAction_RequestMirrorPolicy{
				Cluster:         policy.Destination.Name,
				RuntimeFraction: mp,
			})
		}
	}

	return xdsMirrorPolicies
}

// mirrorPercentByPolicy computes the mirror percent to be used based on ir.MirrorPolicy.
func mirrorPercentByPolicy(mirror *ir.MirrorPolicy) *corev3.RuntimeFractionalPercent {
	switch {
	case mirror.Percentage != nil:
		if p := *mirror.Percentage; p > 0 {
			return &corev3.RuntimeFractionalPercent{
				DefaultValue: fractionalpercent.FromFloat32(p),
			}
		}
		// If zero percent is provided explicitly, we should not mirror.
		return nil
	default:
		// Default to 100 percent if percent is not given.
		return &corev3.RuntimeFractionalPercent{
			DefaultValue: fractionalpercent.FromIn32(100),
		}
	}
}

func buildXdsAddedHeaders(headersToAdd []ir.AddHeader) []*corev3.HeaderValueOption {
	headerValueOptions := []*corev3.HeaderValueOption{}

	for _, header := range headersToAdd {
		var appendAction corev3.HeaderValueOption_HeaderAppendAction

		if header.Append {
			appendAction = corev3.HeaderValueOption_APPEND_IF_EXISTS_OR_ADD
		} else {
			appendAction = corev3.HeaderValueOption_OVERWRITE_IF_EXISTS_OR_ADD
		}
		// Allow empty headers to be set, but don't add the config to do so unless necessary
		if len(header.Value) == 0 {
			headerValueOptions = append(headerValueOptions, &corev3.HeaderValueOption{
				Header: &corev3.HeaderValue{
					Key: header.Name,
				},
				AppendAction:   appendAction,
				KeepEmptyValue: true,
			})
		} else {
			for _, val := range header.Value {
				headerValueOptions = append(headerValueOptions, &corev3.HeaderValueOption{
					Header: &corev3.HeaderValue{
						Key:   header.Name,
						Value: val,
					},
					AppendAction:   appendAction,
					KeepEmptyValue: val == "",
				})
			}
		}
	}

	return headerValueOptions
}

func buildHashPolicy(httpRoute *ir.HTTPRoute) []*routev3.RouteAction_HashPolicy {
	// Return early
	if httpRoute == nil ||
		httpRoute.Traffic == nil ||
		httpRoute.Traffic.LoadBalancer == nil ||
		httpRoute.Traffic.LoadBalancer.ConsistentHash == nil {
		return nil
	}

	ch := httpRoute.Traffic.LoadBalancer.ConsistentHash

	switch {
	case ch.Header != nil:
		hashPolicy := &routev3.RouteAction_HashPolicy{
			PolicySpecifier: &routev3.RouteAction_HashPolicy_Header_{
				Header: &routev3.RouteAction_HashPolicy_Header{
					HeaderName: ch.Header.Name,
				},
			},
		}
		return []*routev3.RouteAction_HashPolicy{hashPolicy}
	case ch.Cookie != nil:
		hashPolicy := &routev3.RouteAction_HashPolicy{
			PolicySpecifier: &routev3.RouteAction_HashPolicy_Cookie_{
				Cookie: &routev3.RouteAction_HashPolicy_Cookie{
					Name: ch.Cookie.Name,
				},
			},
		}
		if ch.Cookie.TTL != nil {
			hashPolicy.GetCookie().Ttl = durationpb.New(ch.Cookie.TTL.Duration)
		}
		if ch.Cookie.Attributes != nil {
			attributes := make([]*routev3.RouteAction_HashPolicy_CookieAttribute, 0, len(ch.Cookie.Attributes))
			for name, value := range ch.Cookie.Attributes {
				attributes = append(attributes, &routev3.RouteAction_HashPolicy_CookieAttribute{
					Name:  name,
					Value: value,
				})
			}
			hashPolicy.GetCookie().Attributes = attributes
		}
		return []*routev3.RouteAction_HashPolicy{hashPolicy}
	case ch.SourceIP != nil:
		if !*ch.SourceIP {
			return nil
		}
		hashPolicy := &routev3.RouteAction_HashPolicy{
			PolicySpecifier: &routev3.RouteAction_HashPolicy_ConnectionProperties_{
				ConnectionProperties: &routev3.RouteAction_HashPolicy_ConnectionProperties{
					SourceIp: true,
				},
			},
		}
		return []*routev3.RouteAction_HashPolicy{hashPolicy}
	default:
		return nil
	}
}

func buildRetryPolicy(route *ir.HTTPRoute) (*routev3.RetryPolicy, error) {
	rr := route.GetRetry()
	anyCfg, err := proto.ToAnyWithValidation(&previoushost.PreviousHostsPredicate{})
	if err != nil {
		return nil, err
	}
	rp := &routev3.RetryPolicy{
		RetryOn:              retryDefaultRetryOn,
		RetriableStatusCodes: []uint32{retryDefaultRetriableStatusCode},
		NumRetries:           &wrapperspb.UInt32Value{Value: retryDefaultNumRetries},
		RetryHostPredicate: []*routev3.RetryPolicy_RetryHostPredicate{
			{
				Name: "envoy.retry_host_predicates.previous_hosts",
				ConfigType: &routev3.RetryPolicy_RetryHostPredicate_TypedConfig{
					TypedConfig: anyCfg,
				},
			},
		},
		HostSelectionRetryMaxAttempts: 5,
	}

	if rr.NumRetries != nil {
		rp.NumRetries = &wrapperspb.UInt32Value{Value: *rr.NumRetries}
	}

	if rr.NumAttemptsPerPriority != nil && *rr.NumAttemptsPerPriority > 0 {
		anyCfgPriority, err := proto.ToAnyWithValidation(&previouspriority.PreviousPrioritiesConfig{
			UpdateFrequency: *rr.NumAttemptsPerPriority,
		})
		if err != nil {
			return nil, err
		}
		rp.RetryPriority = &routev3.RetryPolicy_RetryPriority{
			Name: "envoy.retry_priorities.previous_priorities",
			ConfigType: &routev3.RetryPolicy_RetryPriority_TypedConfig{
				TypedConfig: anyCfgPriority,
			},
		}
	}

	if rr.RetryOn != nil {
		if len(rr.RetryOn.Triggers) > 0 {
			if ro, err := buildRetryOn(rr.RetryOn.Triggers); err == nil {
				rp.RetryOn = ro
			} else {
				return nil, err
			}
		}

		rp.RetriableStatusCodes = buildRetryStatusCodes(rr.RetryOn.HTTPStatusCodes)
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

func buildRouteTracing(httpRoute *ir.HTTPRoute) (*routev3.Tracing, error) {
	if httpRoute == nil || httpRoute.Traffic == nil ||
		httpRoute.Traffic.Telemetry == nil ||
		httpRoute.Traffic.Telemetry.Tracing == nil {
		return nil, nil
	}

	tracing := httpRoute.Traffic.Telemetry.Tracing
	tags, err := buildTracingTags(tracing.CustomTags)
	if err != nil {
		return nil, fmt.Errorf("failed to build route tracing tags:%w", err)
	}

	return &routev3.Tracing{
		RandomSampling: fractionalpercent.FromFraction(tracing.SamplingFraction),
		CustomTags:     tags,
	}, nil
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
