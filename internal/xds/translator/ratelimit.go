// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimitfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	tlsv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/transport_sockets/tls/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	rlsconfv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/envoyproxy/ratelimit/src/config"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	goyaml "gopkg.in/yaml.v3" // nolint: depguard
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	// rateLimitClientTLSCertFilename is the ratelimit tls cert file.
	rateLimitClientTLSCertFilename = "/certs/tls.crt"
	// rateLimitClientTLSKeyFilename is the ratelimit key file.
	rateLimitClientTLSKeyFilename = "/certs/tls.key"
	// rateLimitClientTLSCACertFilename is the ratelimit ca cert file.
	rateLimitClientTLSCACertFilename = "/certs/ca.crt"
	// sharedDescriptorKey is the descriptor key used for shared rate limits.
	sharedDescriptorKey = "global-limit"
)

// patchHCMWithRateLimit builds and appends the Rate Limit Filter to the HTTP connection manager
// if applicable and it does not already exist.
func (t *Translator) patchHCMWithRateLimit(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) {
	// Return early if rate limits dont exist
	if !t.isRateLimitPresent(irListener) {
		return
	}

	// Return early if filter already exists.
	for _, httpFilter := range mgr.HttpFilters {
		if httpFilter.Name == wellknown.HTTPRateLimit {
			return
		}
	}

	rateLimitFilters := t.buildRateLimitFilter(irListener)
	mgr.HttpFilters = append(rateLimitFilters, mgr.HttpFilters...)
}

func routeContainsGlobalRateLimit(irRoute *ir.HTTPRoute) bool {
	if irRoute == nil ||
		irRoute.Traffic == nil ||
		irRoute.Traffic.RateLimit == nil ||
		irRoute.Traffic.RateLimit.Global == nil {
		return false
	}

	return true
}

// isRateLimitPresent returns true if rate limit config exists for the listener.
func (t *Translator) isRateLimitPresent(irListener *ir.HTTPListener) bool {
	// Return false if global ratelimiting is disabled.
	if t.GlobalRateLimit == nil {
		return false
	}
	// Return true if rate limit config exists.
	for _, route := range irListener.Routes {
		if routeContainsGlobalRateLimit(route) {
			return true
		}
	}
	return false
}

// buildRateLimitFilter constructs a list of HTTP filters for rate limiting based on the provided HTTP listener configuration.
func (t *Translator) buildRateLimitFilter(irListener *ir.HTTPListener) []*hcmv3.HttpFilter {
	var filters []*hcmv3.HttpFilter
	domainMap := make(map[string]bool) // Track domains for which filters have been created

	// Iterate over each route in the listener to create rate limit filters.
	for _, route := range irListener.Routes {
		global := route.Traffic.RateLimit.Global
		if global == nil {
			// Skip routes without global rate limit configuration.
			continue
		}

		// Determine the domain for the rate limit filter. If the rate limit is shared, use a shared domain.
		domain := getRateLimitDomain(irListener, global.Shared != nil && *global.Shared)

		// Check if a filter for this domain already exists to avoid duplicates.
		if _, exists := domainMap[domain]; exists {
			continue
		}

		// Create a new rate limit filter configuration.
		rateLimitFilterProto := &ratelimitfilterv3.RateLimit{
			Domain: domain,
			RateLimitService: &ratelimitv3.RateLimitServiceConfig{
				GrpcService: &corev3.GrpcService{
					TargetSpecifier: &corev3.GrpcService_EnvoyGrpc_{
						EnvoyGrpc: &corev3.GrpcService_EnvoyGrpc{
							ClusterName: getRateLimitServiceClusterName(),
						},
					},
				},
				TransportApiVersion: corev3.ApiVersion_V3,
			},
		}

		// Set the timeout for the rate limit service if specified.
		if t.GlobalRateLimit.Timeout > 0 {
			rateLimitFilterProto.Timeout = durationpb.New(t.GlobalRateLimit.Timeout)
		}

		// Configure the X-RateLimit headers based on the listener's header settings.
		if irListener.Headers != nil && irListener.Headers.DisableRateLimitHeaders {
			rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_OFF
		} else {
			rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_DRAFT_VERSION_03
		}

		// Set the failure mode to deny if the global rate limit is configured to fail closed.
		if t.GlobalRateLimit.FailClosed {
			rateLimitFilterProto.FailureModeDeny = t.GlobalRateLimit.FailClosed
		}

		// Convert the rate limit filter configuration to a protobuf Any type.
		rateLimitFilterAny, err := anypb.New(rateLimitFilterProto)
		if err != nil {
			// Skip this filter if there is an error in conversion.
			continue
		}

		// Create the HTTP filter with the rate limit configuration.
		rateLimitFilter := &hcmv3.HttpFilter{
			Name: egv1a1.EnvoyFilterRateLimit.String(),
			ConfigType: &hcmv3.HttpFilter_TypedConfig{
				TypedConfig: rateLimitFilterAny,
			},
		}

		// Add the filter to the list and mark the domain as having a filter.
		filters = append(filters, rateLimitFilter)
		domainMap[domain] = true
	}

	return filters
}

// patchRouteWithRateLimit builds rate limit actions and appends to the route.
func patchRouteWithRateLimit(route *routev3.Route, irRoute *ir.HTTPRoute) error { //nolint:unparam
	// Return early if no rate limit config exists.
	xdsRouteAction := route.GetRoute()
	if !routeContainsGlobalRateLimit(irRoute) || xdsRouteAction == nil {
		return nil
	}
	global := irRoute.Traffic.RateLimit.Global
	rateLimits, costSpecified := buildRouteRateLimits(irRoute.Name, global)
	if costSpecified {
		// PerRoute global rate limit configuration via typed_per_filter_config can have its own rate routev3.RateLimit that overrides the route level rate limits.
		// Per-descriptor level hits_addend can only be configured there: https://github.com/envoyproxy/envoy/pull/37972
		// vs the "legacy" core route-embedded rate limits doesn't support the feature due to the "technical debt".
		//
		// This branch is only reached when the response cost is specified which allows us to assume that
		// users are using Envoy >= v1.33.0 which also supports the typed_per_filter_config.
		//
		// https://www.envoyproxy.io/docs/envoy/latest/api-v3/extensions/filters/http/ratelimit/v3/rate_limit.proto#extensions-filters-http-ratelimit-v3-ratelimitperroute
		//
		// Though this is not explicitly documented, the rate limit functionality is the same as the core route-embedded rate limits.
		// Only code path different is in the following code which is identical for both core and typed_per_filter_config
		// as we are not using virtual_host level rate limits except that when typed_per_filter_config is used, per-descriptor
		// level hits_addend is correctly resolved.
		//
		// https://github.com/envoyproxy/envoy/blob/47f99c5aacdb582606a48c85c6c54904fd439179/source/extensions/filters/http/ratelimit/ratelimit.cc#L93-L114
		return patchRouteWithRateLimitOnTypedFilterConfig(route, rateLimits)
	}
	xdsRouteAction.RateLimits = rateLimits
	return nil
}

// patchRouteWithRateLimitOnTypedFilterConfig builds rate limit actions and appends to the route via
// the TypedPerFilterConfig field.
func patchRouteWithRateLimitOnTypedFilterConfig(route *routev3.Route, rateLimits []*routev3.RateLimit) error { //nolint:unparam
	filterCfg := route.TypedPerFilterConfig
	if filterCfg == nil {
		filterCfg = make(map[string]*anypb.Any)
		route.TypedPerFilterConfig = filterCfg
	}
	if _, ok := filterCfg[egv1a1.EnvoyFilterRateLimit.String()]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf(
			"route already contains global rate limit filter config: %s", route.Name)
	}

	g, err := anypb.New(&ratelimitfilterv3.RateLimitPerRoute{RateLimits: rateLimits})
	if err != nil {
		return fmt.Errorf("failed to marshal per-route ratelimit filter config: %w", err)
	}
	filterCfg[egv1a1.EnvoyFilterRateLimit.String()] = g
	return nil
}

// buildRouteRateLimits constructs rate limit actions for a given route based on the global rate limit configuration.
func buildRouteRateLimits(descriptorPrefix string, global *ir.GlobalRateLimit) (rateLimits []*routev3.RateLimit, costSpecified bool) {
	// Determine if the rate limit is shared across multiple routes.
	isShared := global.Shared != nil && *global.Shared

	// Set the descriptor key based on whether the rate limit is shared.
	descriptorKey := getRouteDescriptor(descriptorPrefix)
	if isShared {
		descriptorKey = sharedDescriptorKey // Use the constant for shared limits.
	}

	// Create a generic key action for the route descriptor.
	routeDescriptor := &routev3.RateLimit_Action{
		ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
			GenericKey: &routev3.RateLimit_Action_GenericKey{
				DescriptorKey:   descriptorKey,
				DescriptorValue: descriptorKey,
			},
		},
	}

	// Iterate over each rule in the global rate limit configuration.
	for rIdx, rule := range global.Rules {
		// Initialize a list of rate limit actions for the current rule.
		rlActions := []*routev3.RateLimit_Action{routeDescriptor}

		// Process each header match in the rule.
		for mIdx, match := range rule.HeaderMatches {
			var action *routev3.RateLimit_Action

			// Handle distinct matches by setting up request header actions.
			if match.Distinct {
				descriptorKey := getRouteRuleDescriptor(rIdx, mIdx)
				action = &routev3.RateLimit_Action{
					ActionSpecifier: &routev3.RateLimit_Action_RequestHeaders_{
						RequestHeaders: &routev3.RateLimit_Action_RequestHeaders{
							HeaderName:    match.Name,
							DescriptorKey: descriptorKey,
						},
					},
				}
			} else {
				// Handle non-distinct matches by setting up header value match actions.
				descriptorKey := getRouteRuleDescriptor(rIdx, mIdx)
				descriptorVal := getRouteRuleDescriptor(rIdx, mIdx)
				headerMatcher := &routev3.HeaderMatcher{
					Name: match.Name,
					HeaderMatchSpecifier: &routev3.HeaderMatcher_StringMatch{
						StringMatch: buildXdsStringMatcher(match),
					},
				}
				expectMatch := true
				if match.Invert != nil && *match.Invert {
					expectMatch = false
				}
				action = &routev3.RateLimit_Action{
					ActionSpecifier: &routev3.RateLimit_Action_HeaderValueMatch_{
						HeaderValueMatch: &routev3.RateLimit_Action_HeaderValueMatch{
							DescriptorKey:   descriptorKey,
							DescriptorValue: descriptorVal,
							ExpectMatch: &wrapperspb.BoolValue{
								Value: expectMatch,
							},
							Headers: []*routev3.HeaderMatcher{headerMatcher},
						},
					},
				}
			}
			// Add the action to the list of rate limit actions.
			rlActions = append(rlActions, action)
		}

		// To be able to rate limit each individual IP, we need to use a nested descriptors structure in the configuration
		// of the rate limit server:
		// * the outer layer is a masked_remote_address descriptor that catches all the source IPs inside a specified CIDR.
		// * the inner layer is a remote_address descriptor that sets the limit for individual IP.
		//
		// An example of rate limit server configuration looks like this:
		//
		//	descriptors:
		//	  - key: masked_remote_address //catch all the source IPs inside a CIDR
		//	    value: 192.168.0.0/16
		//	    descriptors:
		//	      - key: remote_address //set limit for individual IP
		//	        rate_limit:
		//	          unit: second
		//	          requests_per_unit: 100
		//
		// Please refer to [Rate Limit Service Descriptor list definition](https://github.com/envoyproxy/ratelimit#descriptor-list-definition) for details.		
		// If a CIDR match is specified, add MaskedRemoteAddress and RemoteAddress descriptors.
		if rule.CIDRMatch != nil {
			// Setup MaskedRemoteAddress action.
			mra := &routev3.RateLimit_Action_MaskedRemoteAddress{}
			maskLen := &wrapperspb.UInt32Value{Value: rule.CIDRMatch.MaskLen}
			if rule.CIDRMatch.IsIPv6 {
				mra.V6PrefixMaskLen = maskLen
			} else {
				mra.V4PrefixMaskLen = maskLen
			}
			action := &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_MaskedRemoteAddress_{
					MaskedRemoteAddress: mra,
				},
			}
			rlActions = append(rlActions, action)

			// Setup RemoteAddress action if distinct match is set.
			if rule.CIDRMatch.Distinct {
				action = &routev3.RateLimit_Action{
					ActionSpecifier: &routev3.RateLimit_Action_RemoteAddress_{
						RemoteAddress: &routev3.RateLimit_Action_RemoteAddress{},
					},
				}
				rlActions = append(rlActions, action)
			}
		}

		// If no matches are set, apply rate limiting to all traffic.
		if !rule.IsMatchSet() {
			action := &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
					GenericKey: &routev3.RateLimit_Action_GenericKey{
						DescriptorKey:   descriptorKey,
						DescriptorValue: descriptorKey,
					},
				},
			}
			rlActions = append(rlActions, action)
		}

		// Create a rate limit object for the current rule.
		rateLimit := &routev3.RateLimit{Actions: rlActions}
		if c := rule.RequestCost; c != nil {
			// Set the hits addend for the request cost if specified.
			rateLimit.HitsAddend = rateLimitCostToHitsAddend(c)
			costSpecified = true
		}
		// Add the rate limit to the list of rate limits.
		rateLimits = append(rateLimits, rateLimit)

		// Handle response cost by creating a separate rate limit object.
		if c := rule.ResponseCost; c != nil {
			responseRule := &routev3.RateLimit{Actions: rlActions, ApplyOnStreamDone: true}
			responseRule.HitsAddend = rateLimitCostToHitsAddend(c)
			rateLimits = append(rateLimits, responseRule)
			costSpecified = true
		}
	}
	return
}

func rateLimitCostToHitsAddend(c *ir.RateLimitCost) *routev3.RateLimit_HitsAddend {
	ret := &routev3.RateLimit_HitsAddend{}
	if c.Number != nil {
		ret.Number = &wrapperspb.UInt64Value{Value: *c.Number}
	}
	if c.Format != nil {
		ret.Format = *c.Format
	}
	return ret
}

// GetRateLimitServiceConfigStr returns the PB string for the rate limit service configuration.
func GetRateLimitServiceConfigStr(pbCfg *rlsconfv3.RateLimitConfig) (string, error) {
	var buf bytes.Buffer
	enc := goyaml.NewEncoder(&buf)
	enc.SetIndent(2)
	// Translate pb config to yaml
	yamlRoot := config.ConfigXdsProtoToYaml(pbCfg)
	rateLimitConfig := &struct {
		Name        string
		Domain      string
		Descriptors []config.YamlDescriptor
	}{
		Name:        pbCfg.Name,
		Domain:      yamlRoot.Domain,
		Descriptors: yamlRoot.Descriptors,
	}
	err := enc.Encode(rateLimitConfig)
	return buf.String(), err
}

// BuildRateLimitServiceConfig builds the rate limit service configuration based on
// https://github.com/envoyproxy/ratelimit#the-configuration-format
func BuildRateLimitServiceConfig(irListener *ir.HTTPListener) *rlsconfv3.RateLimitConfig {
	log.Printf("DEBUG: Starting BuildRateLimitServiceConfig for listener: %s", irListener.Name)

	var routeDescriptors []*rlsconfv3.RateLimitDescriptor
	var globalServiceDescriptors []*rlsconfv3.RateLimitDescriptor // For shared/global rate limits

	// Process each route to build descriptors.
	for _, route := range irListener.Routes {
		if !routeContainsGlobalRateLimit(route) {
			log.Printf("DEBUG: Route %s does not contain a global rate limit, skipping.", route.Name)
			continue
		}

		// Get route rule descriptors within each route.
		//
		// An example of route descriptor looks like this:
		//
		// descriptors:
		//   - key:   ${RouteDescriptor}
		//     value: ${RouteDescriptor}
		//     descriptors:
		//       - key:   ${RouteRuleDescriptor}
		//         value: ${RouteRuleDescriptor}
		//       - ...
		//		
		serviceDescriptors := buildRateLimitServiceDescriptors(route.Traffic.RateLimit.Global)

		if isSharedRateLimit(route) {
			// For shared limits, merge the service descriptors into the global slice,
			// ensuring we don't add duplicates.
			for _, sd := range serviceDescriptors {
				duplicate := false
				for _, existing := range globalServiceDescriptors {
					if existing.Key == sd.Key && existing.Value == sd.Value {
						duplicate = true
						break
					}
				}
				if !duplicate {
					globalServiceDescriptors = append(globalServiceDescriptors, sd)
				}
			}
			log.Printf("DEBUG: Added shared rate limit from route: %s", route.Name)
		} else {
			// For non-shared (per-route) limits, create a descriptor keyed to the route.
			routeDescriptor := &rlsconfv3.RateLimitDescriptor{
				Key:         getRouteDescriptor(route.Name),
				Value:       getRouteDescriptor(route.Name),
				Descriptors: serviceDescriptors,
			}
			routeDescriptors = append(routeDescriptors, routeDescriptor)
			log.Printf("DEBUG: Added route-level descriptor for route: %s", route.Name)
		}
	}

	// If no descriptors were built, return nil.
	if len(routeDescriptors) == 0 && len(globalServiceDescriptors) == 0 {
		log.Printf("DEBUG: No descriptors built, returning nil RateLimitConfig for listener: %s", irListener.Name)
		return nil
	}

	// Determine the domain—this may factor in whether a global limit was set.
	domain := getRateLimitDomain(irListener, len(globalServiceDescriptors) > 0)
	log.Printf("DEBUG: Final domain set to: %s", domain)

	// Build the final list of descriptors.
	var finalDescriptors []*rlsconfv3.RateLimitDescriptor
	if len(globalServiceDescriptors) > 0 {
		// Create a single global descriptor that aggregates all shared limits.
		globalDescriptor := &rlsconfv3.RateLimitDescriptor{
			Key:         sharedDescriptorKey, // Use the constant
			Value:       sharedDescriptorKey, // Use the constant
			Descriptors: globalServiceDescriptors,
		}
		finalDescriptors = append(finalDescriptors, globalDescriptor)
		log.Printf("DEBUG: Using global-limit descriptor with %d sub-descriptors.", len(globalServiceDescriptors))
	}

	// Append all route-specific descriptors.
	finalDescriptors = append(finalDescriptors, routeDescriptors...)
	log.Printf("DEBUG: Final RateLimitServiceConfig will use domain: %s with %d descriptors.", domain, len(finalDescriptors))

	return &rlsconfv3.RateLimitConfig{
		Name:        domain,
		Domain:      domain,
		Descriptors: finalDescriptors,
	}
}

// Helper function to check if a route has a shared rate limit
func isSharedRateLimit(route *ir.HTTPRoute) bool {
	global := route.Traffic.RateLimit.Global
	return global != nil && global.Shared != nil && *global.Shared && len(global.Rules) > 0
}

// buildRateLimitServiceDescriptors creates the rate limit service pb descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(global *ir.GlobalRateLimit) []*rlsconfv3.RateLimitDescriptor {
	pbDescriptors := make([]*rlsconfv3.RateLimitDescriptor, 0, len(global.Rules))
	usedGlobalLimit := false // Track if the global-limit key has been used

	// The order in which matching descriptors are built is consistent with
	// the order in which ratelimit actions are built:
	//  1) Header Matches
	//  2) CIDR Match
	//  3) No Match
	for rIdx, rule := range global.Rules {
		rateLimitPolicy := &rlsconfv3.RateLimitPolicy{
			RequestsPerUnit: uint32(rule.Limit.Requests),
			Unit:            rlsconfv3.RateLimitUnit(rlsconfv3.RateLimitUnit_value[strings.ToUpper(string(rule.Limit.Unit))]),
		}

		// We use a chain structure to describe the matching descriptors for one rule.
		// The RateLimitPolicy should be added to the last descriptor in the chain.
		var head, cur *rlsconfv3.RateLimitDescriptor

		for mIdx, match := range rule.HeaderMatches {
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			// Case for distinct match
			if match.Distinct {
				// RequestHeader case
				pbDesc.Key = getRouteRuleDescriptor(rIdx, mIdx)
			} else {
				// HeaderValueMatch case
				pbDesc.Key = getRouteRuleDescriptor(rIdx, mIdx)
				pbDesc.Value = getRouteRuleDescriptor(rIdx, mIdx)
			}

			if mIdx == 0 {
				head = pbDesc
			} else {
				cur.Descriptors = []*rlsconfv3.RateLimitDescriptor{pbDesc}
			}

			cur = pbDesc

			// Do not add the RateLimitPolicy to the last header match descriptor yet,
			// as it is also possible that CIDR match descriptor also exist.
		}

		// EG supports two kinds of rate limit descriptors for the source IP: exact and distinct.
		// * exact means that all IP Addresses within the specified Source IP CIDR share the same rate limit bucket.
		// * distinct means that each IP Address within the specified Source IP CIDR has its own rate limit bucket.
		//
		// To be able to rate limit each individual IP, we need to use a nested descriptors structure in the configuration
		// of the rate limit server:
		// * the outer layer is a masked_remote_address descriptor that catches all the source IPs inside a specified CIDR.
		// * the inner layer is a remote_address descriptor that sets the limit for individual IP.
		//
		// An example of rate limit server configuration looks like this:
		//
		//	descriptors:
		//	  - key: masked_remote_address //catch all the source IPs inside a CIDR
		//	    value: 192.168.0.0/16
		//	    descriptors:
		//	      - key: remote_address //set limit for individual IP
		//	        rate_limit:
		//	          unit: second
		//	          requests_per_unit: 100
		//
		// Please refer to [Rate Limit Service Descriptor list definition](https://github.com/envoyproxy/ratelimit#descriptor-list-definition) for details.
		if rule.CIDRMatch != nil {
			// MaskedRemoteAddress case
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			pbDesc.Key = "masked_remote_address"
			pbDesc.Value = rule.CIDRMatch.CIDR

			if cur != nil {
				// The header match descriptor chain exist, add current
				// descriptor to the chain.
				cur.Descriptors = []*rlsconfv3.RateLimitDescriptor{pbDesc}
			} else {
				head = pbDesc
			}
			cur = pbDesc

			if rule.CIDRMatch.Distinct {
				pbDesc := new(rlsconfv3.RateLimitDescriptor)
				pbDesc.Key = "remote_address"
				cur.Descriptors = []*rlsconfv3.RateLimitDescriptor{pbDesc}
				cur = pbDesc
			}
		}

		// Case when both header and cidr match are not set and the ratelimit
		// will be applied to all traffic.
		if !rule.IsMatchSet() {
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			// GenericKey case
			if global.Shared != nil && *global.Shared && !usedGlobalLimit {
				pbDesc.Key = sharedDescriptorKey // Use the constant
				usedGlobalLimit = true
				log.Printf("DEBUG: Using global-limit key for rule index %d", rIdx)
			} else {
				pbDesc.Key = getRouteRuleDescriptor(rIdx, -1)
				log.Printf("DEBUG: Using unique key for rule index %d: %s", rIdx, pbDesc.Key)
			}
			pbDesc.Value = pbDesc.Key
			head = pbDesc
			cur = head
		}

		// Add the ratelimit policy to the last descriptor of chain.
		cur.RateLimit = rateLimitPolicy
		pbDescriptors = append(pbDescriptors, head)
	}

	return pbDescriptors
}

// buildRateLimitTLSocket builds the TLS socket for the rate limit service.
func buildRateLimitTLSocket() (*corev3.TransportSocket, error) {
	tlsCtx := &tlsv3.UpstreamTlsContext{
		CommonTlsContext: &tlsv3.CommonTlsContext{
			TlsCertificates: []*tlsv3.TlsCertificate{},
			ValidationContextType: &tlsv3.CommonTlsContext_ValidationContext{
				ValidationContext: &tlsv3.CertificateValidationContext{
					TrustedCa: &corev3.DataSource{
						Specifier: &corev3.DataSource_Filename{Filename: rateLimitClientTLSCACertFilename},
					},
				},
			},
		},
	}

	tlsCert := &tlsv3.TlsCertificate{
		CertificateChain: &corev3.DataSource{
			Specifier: &corev3.DataSource_Filename{Filename: rateLimitClientTLSCertFilename},
		},
		PrivateKey: &corev3.DataSource{
			Specifier: &corev3.DataSource_Filename{Filename: rateLimitClientTLSKeyFilename},
		},
	}
	tlsCtx.CommonTlsContext.TlsCertificates = append(tlsCtx.CommonTlsContext.TlsCertificates, tlsCert)

	tlsCtxAny, err := anypb.New(tlsCtx)
	if err != nil {
		return nil, err
	}

	return &corev3.TransportSocket{
		Name: wellknown.TransportSocketTls,
		ConfigType: &corev3.TransportSocket_TypedConfig{
			TypedConfig: tlsCtxAny,
		},
	}, nil
}

func (t *Translator) createRateLimitServiceCluster(tCtx *types.ResourceVersionTable, irListener *ir.HTTPListener, metrics *ir.Metrics) error {
	// Return early if rate limits don't exist.
	if !t.isRateLimitPresent(irListener) {
		return nil
	}
	clusterName := getRateLimitServiceClusterName()
	// Create cluster if it does not exist
	host, port := t.getRateLimitServiceGrpcHostPort()
	ds := &ir.DestinationSetting{
		Weight:    ptr.To[uint32](1),
		Protocol:  ir.GRPC,
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(host, port, false)},
	}

	tSocket, err := buildRateLimitTLSocket()
	if err != nil {
		return err
	}

	return addXdsCluster(tCtx, &xdsClusterArgs{
		name:         clusterName,
		settings:     []*ir.DestinationSetting{ds},
		tSocket:      tSocket,
		endpointType: EndpointTypeDNS,
		metrics:      metrics,
	})
}

func getRouteRuleDescriptor(ruleIndex, matchIndex int) string {
	return "rule-" + strconv.Itoa(ruleIndex) + "-match-" + strconv.Itoa(matchIndex)
}

func getRouteDescriptor(routeName string) string {
	return routeName
}

func getRateLimitServiceClusterName() string {
	return "ratelimit_cluster"
}

func getRateLimitDomain(irListener *ir.HTTPListener, isShared bool) string {
	if isShared && irListener.GatewayName != "" {
		return irListener.GatewayName
	}
	return irListener.Name
}

func (t *Translator) getRateLimitServiceGrpcHostPort() (string, uint32) {
	u, err := url.Parse(t.GlobalRateLimit.ServiceURL)
	if err != nil {
		panic(err)
	}
	p, err := strconv.ParseUint(u.Port(), 10, 32)
	if err != nil {
		panic(err)
	}
	return u.Hostname(), uint32(p)
}
