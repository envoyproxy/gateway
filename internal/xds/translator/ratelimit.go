// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"bytes"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	ratelimitv3 "github.com/envoyproxy/go-control-plane/envoy/config/ratelimit/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	ratelimitfilterv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/ratelimit/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	"github.com/envoyproxy/go-control-plane/pkg/wellknown"
	rlsconfv3 "github.com/envoyproxy/go-control-plane/ratelimit/config/ratelimit/v3"
	"github.com/envoyproxy/ratelimit/src/config"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/wrapperspb"
	goyaml "gopkg.in/yaml.v3" // nolint: depguard

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
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
	// Only append if we have filters to add
	if len(rateLimitFilters) > 0 {
		mgr.HttpFilters = append(rateLimitFilters, mgr.HttpFilters...)
	}
}

// isRateLimitPresent returns true if rate limit config exists for the listener.
func (t *Translator) isRateLimitPresent(irListener *ir.HTTPListener) bool {
	// Return false if global ratelimiting is disabled.
	if t.GlobalRateLimit == nil {
		return false
	}
	// Return true if rate limit config exists.
	for _, route := range irListener.Routes {
		if isValidGlobalRateLimit(route) {
			return true
		}
	}
	return false
}

// buildRateLimitFilter constructs a list of HTTP filters for rate limiting based on the provided HTTP listener configuration.
// It creates separate filters for shared and non-shared rate limits.
func (t *Translator) buildRateLimitFilter(irListener *ir.HTTPListener) []*hcmv3.HttpFilter {
	if irListener == nil || irListener.Routes == nil {
		return nil
	}

	filters := []*hcmv3.HttpFilter{}
	created := make(map[string]bool)

	for _, route := range irListener.Routes {
		if !isValidGlobalRateLimit(route) {
			continue
		}

		hasShared, hasNonShared := false, false
		var sharedRuleName string

		for _, rule := range route.Traffic.RateLimit.Global.Rules {
			if isRuleShared(rule) {
				hasShared = true
				sharedRuleName = stripRuleIndexSuffix(rule.Name)
			} else {
				hasNonShared = true
			}
		}

		if hasShared {
			sharedDomain := sharedRuleName
			if !created[sharedDomain] {
				filterName := fmt.Sprintf("%s/%s", egv1a1.EnvoyFilterRateLimit.String(), sharedRuleName)
				if filter := createRateLimitFilter(t, irListener, sharedDomain, filterName); filter != nil {
					filters = append(filters, filter)
				}
				created[sharedDomain] = true
			}
		}

		if hasNonShared {
			nonSharedDomain := irListener.Name
			if !created[nonSharedDomain] {
				filterName := egv1a1.EnvoyFilterRateLimit.String()
				if filter := createRateLimitFilter(t, irListener, nonSharedDomain, filterName); filter != nil {
					filters = append(filters, filter)
				}
				created[nonSharedDomain] = true
			}
		}
	}

	return filters
}

// createRateLimitFilter creates a single rate limit filter for the given domain
func createRateLimitFilter(t *Translator, irListener *ir.HTTPListener, domain, filterName string) *hcmv3.HttpFilter {
	// Create a new rate limit filter configuration
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

	// Set the timeout for the rate limit service if specified
	if t.GlobalRateLimit.Timeout > 0 {
		rateLimitFilterProto.Timeout = durationpb.New(t.GlobalRateLimit.Timeout)
	}

	// Configure the X-RateLimit headers based on the listener's header settings
	if irListener.Headers != nil && irListener.Headers.DisableRateLimitHeaders {
		rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_OFF
	} else {
		rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_DRAFT_VERSION_03
	}

	// Set the failure mode to deny if the global rate limit is configured to fail closed
	if t.GlobalRateLimit.FailClosed {
		rateLimitFilterProto.FailureModeDeny = t.GlobalRateLimit.FailClosed
	}

	// Convert the rate limit filter configuration to a protobuf Any type
	rateLimitFilterAny, err := anypb.New(rateLimitFilterProto)
	if err != nil {
		// Return nil if there is an error in conversion
		return nil
	}

	// Create the HTTP filter with the rate limit configuration
	return &hcmv3.HttpFilter{
		Name: filterName,
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rateLimitFilterAny,
		},
	}
}

// patchRouteWithRateLimit builds rate limit actions and appends to the route.
func patchRouteWithRateLimit(route *routev3.Route, irRoute *ir.HTTPRoute) error { //nolint:unparam
	// Return early if no rate limit config exists.
	xdsRouteAction := route.GetRoute()
	if !isValidGlobalRateLimit(irRoute) || xdsRouteAction == nil {
		return nil
	}
	rateLimits, costSpecified := buildRouteRateLimits(irRoute)
	if costSpecified {
		return patchRouteWithRateLimitOnTypedFilterConfig(route, rateLimits, irRoute)
	}
	xdsRouteAction.RateLimits = rateLimits
	return nil
}

// patchRouteWithRateLimitOnTypedFilterConfig builds rate limit actions and appends to the route via
// the TypedPerFilterConfig field.
func patchRouteWithRateLimitOnTypedFilterConfig(route *routev3.Route, rateLimits []*routev3.RateLimit, irRoute *ir.HTTPRoute) error { //nolint:unparam
	filterCfg := route.TypedPerFilterConfig
	if filterCfg == nil {
		filterCfg = make(map[string]*anypb.Any)
		route.TypedPerFilterConfig = filterCfg
	}

	filterName := getRateLimitFilterName(irRoute)

	if _, ok := filterCfg[filterName]; ok {
		// This should not happen since this is the only place where the filter
		// config is added in a route.
		return fmt.Errorf(
			"route already contains global rate limit filter config: %s", route.Name)
	}

	g, err := anypb.New(&ratelimitfilterv3.RateLimitPerRoute{RateLimits: rateLimits})
	if err != nil {
		return fmt.Errorf("failed to marshal per-route ratelimit filter config: %w", err)
	}
	filterCfg[filterName] = g
	return nil
}

// buildRouteRateLimits constructs rate limit actions for a given route based on the global rate limit configuration.
func buildRouteRateLimits(route *ir.HTTPRoute) (rateLimits []*routev3.RateLimit, costSpecified bool) {
	// Ensure route has rate limit config
	if !isValidGlobalRateLimit(route) {
		return nil, false
	}

	// Get the global rate limit configuration
	global := route.Traffic.RateLimit.Global

	// Iterate over each rule in the global rate limit configuration.
	for rIdx, rule := range global.Rules {
		// Create a list of rate limit actions for the current rule.
		var rlActions []*routev3.RateLimit_Action

		// Create the route descriptor using the rule's shared attribute
		var descriptorKey, descriptorValue string
		if isRuleShared(rule) {
			// For shared rule, use full rule name
			descriptorKey = rule.Name
			descriptorValue = rule.Name
		} else {
			// For non-shared rule, use route name in descriptor
			descriptorKey = getRouteDescriptor(route.Name)
			descriptorValue = descriptorKey
		}

		// Create a generic key action for the route descriptor.
		routeDescriptor := &routev3.RateLimit_Action{
			ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
				GenericKey: &routev3.RateLimit_Action_GenericKey{
					DescriptorKey:   descriptorKey,
					DescriptorValue: descriptorValue,
				},
			},
		}

		// Add the generic key action
		rlActions = append(rlActions, routeDescriptor)

		// Calculate the domain-specific rule index (0-based for each domain)
		ruleIsShared := isRuleShared(rule)
		domainRuleIdx := getDomainRuleIndex(global.Rules, rIdx, ruleIsShared)

		// Process each header match in the rule.
		for mIdx, match := range rule.HeaderMatches {
			var action *routev3.RateLimit_Action

			// Handle distinct matches by setting up request header actions.
			if match.Distinct {
				descriptorKey := getRouteRuleDescriptor(domainRuleIdx, mIdx)
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
				descriptorKey := getRouteRuleDescriptor(domainRuleIdx, mIdx)
				descriptorVal := getRouteRuleDescriptor(domainRuleIdx, mIdx)
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
		//  descriptors:
		//    - key: masked_remote_address //catch all the source IPs inside a CIDR
		//      value: 192.168.0.0/16
		//      descriptors:
		//        - key: remote_address //set limit for individual IP
		//          rate_limit:
		//            unit: second
		//            requests_per_unit: 100
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

		// Case when both header and cidr match are not set and the ratelimit
		// will be applied to all traffic.
		// 3) No Match (apply to all traffic)
		if !rule.IsMatchSet() {
			action := &routev3.RateLimit_Action{
				ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
					GenericKey: &routev3.RateLimit_Action_GenericKey{
						DescriptorKey:   getRouteRuleDescriptor(domainRuleIdx, -1),
						DescriptorValue: getRouteRuleDescriptor(domainRuleIdx, -1),
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

// BuildRateLimitServiceConfig builds the rate limit service configurations based on
// https://github.com/envoyproxy/ratelimit#the-configuration-format
// It returns a list of unique configurations, one for each domain needed across all listeners.
// For shared rate limits, it ensures we only process each shared domain once to improve efficiency.
func BuildRateLimitServiceConfig(irListeners []*ir.HTTPListener) []*rlsconfv3.RateLimitConfig {
	// Map to store rate limit descriptors by domain name
	domainDesc := make(map[string][]*rlsconfv3.RateLimitDescriptor)

	// Process each listener
	for _, irListener := range irListeners {
		// Process each route in the listener
		for _, route := range irListener.Routes {
			// Skip routes without valid global rate limit configuration
			if !isValidGlobalRateLimit(route) {
				continue
			}

			// Build all descriptors for this route in a single pass to maintain consistent indices
			descriptors := buildRateLimitServiceDescriptors(route)

			// Skip if no descriptors were created
			if len(descriptors) == 0 {
				continue
			}

			// Process shared rules - add to traffic policy domain
			sharedDomain := getDomainSharedName(route)
			addRateLimitDescriptor(route, descriptors, sharedDomain, domainDesc, true)

			// Process non-shared rules - add to listener-specific domain
			listenerDomain := irListener.Name
			addRateLimitDescriptor(route, descriptors, listenerDomain, domainDesc, false)
		}
	}

	// Convert domain descriptor map to list of rate limit configurations
	return createRateLimitConfigs(domainDesc)
}

// createRateLimitConfigs creates rate limit configs from the domain descriptor map
func createRateLimitConfigs(
	domainDescriptors map[string][]*rlsconfv3.RateLimitDescriptor,
) []*rlsconfv3.RateLimitConfig {
	var configs []*rlsconfv3.RateLimitConfig
	for domain, descriptors := range domainDescriptors {
		if len(descriptors) > 0 {
			configs = append(configs, &rlsconfv3.RateLimitConfig{
				Name:        domain,
				Domain:      domain,
				Descriptors: descriptors,
			})
		}
	}
	return configs
}

// Helper to recursively compare two RateLimitDescriptors for equality
func descriptorsEqual(a, b *rlsconfv3.RateLimitDescriptor) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Key != b.Key || a.Value != b.Value {
		return false
	}
	if len(a.Descriptors) != len(b.Descriptors) {
		return false
	}
	for i := range a.Descriptors {
		if !descriptorsEqual(a.Descriptors[i], b.Descriptors[i]) {
			return false
		}
	}
	return true
}

// addRateLimitDescriptor adds rate limit descriptors to the domain descriptor map.
// Handles both shared and route-specific rate limits.
//
// An example of route descriptor looks like this:
// descriptors:
//   - key:   ${RouteDescriptor}
//     value: ${RouteDescriptor}
//     descriptors:
//   - key:   ${RouteRuleDescriptor}
//     value: ${RouteRuleDescriptor}
//   - ...
func addRateLimitDescriptor(
	route *ir.HTTPRoute,
	serviceDescriptors []*rlsconfv3.RateLimitDescriptor,
	domain string,
	domainDescriptors map[string][]*rlsconfv3.RateLimitDescriptor,
	includeShared bool,
) {
	if !isValidGlobalRateLimit(route) || len(serviceDescriptors) == 0 {
		return
	}

	for i, rule := range route.Traffic.RateLimit.Global.Rules {
		if i >= len(serviceDescriptors) || (includeShared != isRuleShared(rule)) {
			continue
		}

		var descriptorKey string
		if isRuleShared(rule) {
			descriptorKey = rule.Name
		} else {
			descriptorKey = getRouteDescriptor(route.Name)
		}

		// Find or create descriptor in domainDescriptors[domain]
		var descriptorRule *rlsconfv3.RateLimitDescriptor
		found := false
		for _, d := range domainDescriptors[domain] {
			if d.Key == descriptorKey {
				descriptorRule = d
				found = true
				break
			}
		}
		if !found {
			descriptorRule = &rlsconfv3.RateLimitDescriptor{Key: descriptorKey, Value: descriptorKey}
			domainDescriptors[domain] = append(domainDescriptors[domain], descriptorRule)
		}

		// Ensure no duplicate descriptors
		alreadyExists := false
		for _, existing := range descriptorRule.Descriptors {
			if descriptorsEqual(existing, serviceDescriptors[i]) {
				alreadyExists = true
				break
			}
		}
		if !alreadyExists {
			descriptorRule.Descriptors = append(descriptorRule.Descriptors, serviceDescriptors[i])
		}
	}
}

// isSharedRateLimit checks if a route has at least one shared rate limit rule.
// It returns true if any rule in the global rate limit configuration is marked as shared.
// If no rules are shared or there's no global rate limit configuration, it returns false.
func isSharedRateLimit(route *ir.HTTPRoute) bool {
	if !isValidGlobalRateLimit(route) {
		return false
	}

	global := route.Traffic.RateLimit.Global
	if len(global.Rules) == 0 {
		return false
	}

	// Check if any rule has shared=true
	for _, rule := range global.Rules {
		if isRuleShared(rule) {
			return true
		}
	}

	return false
}

// Helper function to check if a specific rule is shared
func isRuleShared(rule *ir.RateLimitRule) bool {
	return rule != nil && rule.Shared != nil && *rule.Shared
}

// Helper function to check if a specific rule in a route is shared
func isRuleAtIndexShared(route *ir.HTTPRoute, ruleIndex int) bool {
	if route == nil || route.Traffic == nil || route.Traffic.RateLimit == nil ||
		route.Traffic.RateLimit.Global == nil || len(route.Traffic.RateLimit.Global.Rules) <= ruleIndex || ruleIndex < 0 {
		return false
	}

	return isRuleShared(route.Traffic.RateLimit.Global.Rules[ruleIndex])
}

// Helper function to map a global rule index to a domain-specific rule index
// This ensures that both shared and non-shared rules have indices starting from 0 in their own domains.
func getDomainRuleIndex(rules []*ir.RateLimitRule, globalRuleIdx int, ruleIsShared bool) int {
	if globalRuleIdx < 0 || globalRuleIdx >= len(rules) {
		return 0
	}

	// Count how many rules of the same "shared" status came before this one
	count := 0
	for i := 0; i < globalRuleIdx; i++ {
		// If we're looking for shared rules, count shared ones; otherwise count non-shared ones
		if (ruleIsShared && isRuleShared(rules[i])) || (!ruleIsShared && !isRuleShared(rules[i])) {
			count++
		}
	}
	return count
}

// buildRateLimitServiceDescriptors creates the rate limit service pb descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(route *ir.HTTPRoute) []*rlsconfv3.RateLimitDescriptor {
	// Safely check that we have a GlobalRateLimit config
	if route == nil || route.Traffic == nil || route.Traffic.RateLimit == nil || route.Traffic.RateLimit.Global == nil {
		return nil
	}
	global := route.Traffic.RateLimit.Global

	pbDescriptors := make([]*rlsconfv3.RateLimitDescriptor, 0, len(global.Rules))

	// The order in which matching descriptors are built is consistent with
	// the order in which ratelimit actions are built:
	//  1) Header Matches
	//  2) CIDR Match
	//  3) No Match

	for rIdx, rule := range global.Rules {
		rateLimitPolicy := &rlsconfv3.RateLimitPolicy{
			RequestsPerUnit: uint32(rule.Limit.Requests),
			Unit: rlsconfv3.RateLimitUnit(
				rlsconfv3.RateLimitUnit_value[strings.ToUpper(string(rule.Limit.Unit))]),
		}

		// We use a chain structure to describe the matching descriptors for one rule.
		var head, cur *rlsconfv3.RateLimitDescriptor

		// Calculate the domain-specific rule index (0-based for each domain)
		ruleIsShared := isRuleShared(rule)
		domainRuleIdx := getDomainRuleIndex(global.Rules, rIdx, ruleIsShared)

		// 1) Header Matches
		for mIdx, match := range rule.HeaderMatches {
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			// Distinct vs HeaderValueMatch
			if match.Distinct {
				// RequestHeader case
				pbDesc.Key = getRouteRuleDescriptor(domainRuleIdx, mIdx)
			} else {
				// HeaderValueMatch case
				pbDesc.Key = getRouteRuleDescriptor(domainRuleIdx, mIdx)
				pbDesc.Value = getRouteRuleDescriptor(domainRuleIdx, mIdx)
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
		//  descriptors:
		//    - key: masked_remote_address //catch all the source IPs inside a CIDR
		//      value: 192.168.0.0/16
		//      descriptors:
		//        - key: remote_address //set limit for individual IP
		//          rate_limit:
		//            unit: second
		//            requests_per_unit: 100
		//
		// Please refer to [Rate Limit Service Descriptor list definition](https://github.com/envoyproxy/ratelimit#descriptor-list-definition) for details.
		// 2) CIDR Match
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
		// 3) No Match (apply to all traffic)
		if !rule.IsMatchSet() {
			pbDesc := new(rlsconfv3.RateLimitDescriptor)

			// Determine if we should use the shared rate limit key (rule-based) or a generic route key
			if isRuleAtIndexShared(route, rIdx) {
				// For shared rate limits, use rule name
				pbDesc.Key = rule.Name
				pbDesc.Value = rule.Name
			} else {
				// Use generic key for non-shared rate limits, with prefix for uniqueness
				descriptorKey := getRouteRuleDescriptor(domainRuleIdx, -1)
				pbDesc.Key = descriptorKey
				pbDesc.Value = pbDesc.Key
			}

			head = pbDesc
			cur = head
		}

		// Finalize rate-limit policy on the last descriptor in the chain
		cur.RateLimit = rateLimitPolicy
		pbDescriptors = append(pbDescriptors, head)
	}

	return pbDescriptors
}

// getDomainSharedName returns the shared domain (stripped policy name), stripRuleIndexSuffix is used to remove the rule index suffix.
func getDomainSharedName(route *ir.HTTPRoute) string {
	return stripRuleIndexSuffix(route.Traffic.RateLimit.Global.Rules[0].Name)
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

// getRateLimitFilterName gets the filter name for rate limits.
// If any rule in the route is shared, it appends the rule name to the base filter name.
// For non-shared rate limits, it returns just the base filter name.
// Note: This function is primarily used for route-level filter configuration, not for HTTP filters at the listener level.
func getRateLimitFilterName(route *ir.HTTPRoute) string {
	filterName := egv1a1.EnvoyFilterRateLimit.String()
	// If any rule is shared, include the rule name in the filter name
	if isSharedRateLimit(route) {
		// Find the first shared rule to use its name
		for _, rule := range route.Traffic.RateLimit.Global.Rules {
			if isRuleShared(rule) {
				filterName = fmt.Sprintf("%s/%s", filterName, stripRuleIndexSuffix(rule.Name))
				break
			}
		}
	}
	return filterName
}

// Helper to strip /rule/<index> from a rule name in order to use shared http filter
func stripRuleIndexSuffix(name string) string {
	if i := strings.LastIndex(name, "/rule/"); i != -1 {
		return name[:i]
	}
	return strings.Replace(name, "/", "-", 1)
}

// Helper to check if a route has a valid global rate limit config
func isValidGlobalRateLimit(route *ir.HTTPRoute) bool {
	return route != nil &&
		route.Traffic != nil &&
		route.Traffic.RateLimit != nil &&
		route.Traffic.RateLimit.Global != nil
}
