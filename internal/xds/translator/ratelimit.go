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

// routeContainsGlobalRateLimit checks if a route has global rate limit configuration.
// Returns false if any required component is nil.
func routeContainsGlobalRateLimit(irRoute *ir.HTTPRoute) bool {
	return irRoute != nil &&
		irRoute.Traffic != nil &&
		irRoute.Traffic.RateLimit != nil &&
		irRoute.Traffic.RateLimit.Global != nil
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
	if irListener == nil || irListener.Routes == nil {
		return nil
	}

	var filters []*hcmv3.HttpFilter
	domainMap := make(map[string]bool) // Track domains for which filters have been created

	// Iterate over each route in the listener to create rate limit filters.
	for _, route := range irListener.Routes {
		if !routeContainsGlobalRateLimit(route) {
			continue
		}

		// Determine the domain for the rate limit filter. If the rate limit is shared, use a shared domain.
		domain := getRateLimitDomainFilters(irListener, isSharedRateLimit(route), domainMap)

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
	rateLimits := buildRouteRateLimits(irRoute)
	return patchRouteWithRateLimitOnTypedFilterConfig(route, rateLimits)
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

// getSharedDescriptorKey returns the descriptor key used for shared rate limits based on BTP.
func getSharedDescriptorKey(route *ir.HTTPRoute) string {
	if !hasValidBackendTrafficPolicy(route) {
		return ""
	}
	return route.Traffic.BackendTrafficPolicy.Name
}

// getSharedDescriptorValue returns the descriptor value used for shared rate limits based on BTP.
func getSharedDescriptorValue(route *ir.HTTPRoute) string {
	if !hasValidBackendTrafficPolicy(route) {
		return ""
	}
	return route.Traffic.BackendTrafficPolicy.Namespace
}

// buildRouteRateLimits constructs rate limit actions for a given route based on the global rate limit configuration.
func buildRouteRateLimits(route *ir.HTTPRoute) (rateLimits []*routev3.RateLimit) {
	descriptorPrefix := route.Name
	// Ensure route has rate limit config
	if !routeContainsGlobalRateLimit(route) {
		return nil
	}

	// Determine if the rate limit is shared across multiple routes.
	global := route.Traffic.RateLimit.Global
	isShared := isSharedRateLimit(route)

	// Set the descriptor key based on whether the rate limit is shared.
	var descriptorKey, descriptorValue string
	if isShared {
		// For shared rate limits, use BTP name as key and namespace as value
		descriptorKey = getSharedDescriptorKey(route)
		descriptorValue = getSharedDescriptorValue(route)
	} else {
		// For non-shared rate limits, include the route name in the descriptor
		descriptorKey = getRouteDescriptor(descriptorPrefix)
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
						DescriptorKey:   getRouteRuleDescriptor(rIdx, -1),
						DescriptorValue: getRouteRuleDescriptor(rIdx, -1),
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
		}
		// Add the rate limit to the list of rate limits.
		rateLimits = append(rateLimits, rateLimit)

		// Handle response cost by creating a separate rate limit object.
		if c := rule.ResponseCost; c != nil {
			responseRule := &routev3.RateLimit{Actions: rlActions, ApplyOnStreamDone: true}
			responseRule.HitsAddend = rateLimitCostToHitsAddend(c)
			rateLimits = append(rateLimits, responseRule)
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
// It returns a list of configurations, one for each unique domain needed.
func BuildRateLimitServiceConfig(irListener *ir.HTTPListener) []*rlsconfv3.RateLimitConfig {
	// Map to store descriptors for each domain
	domainDescriptors := make(map[string][]*rlsconfv3.RateLimitDescriptor)

	// Process each route to build descriptors
	for _, route := range irListener.Routes {
		if !routeContainsGlobalRateLimit(route) {
			continue
		}

		// Get route rule descriptors within each route
		serviceDescriptors := buildRateLimitServiceDescriptors(route)
		if len(serviceDescriptors) == 0 {
			continue
		}

		// Determine the domain for this route
		domain := irListener.Name // Default domain
		if isSharedRateLimit(route) && route.Traffic.BackendTrafficPolicy != nil &&
			route.Traffic.BackendTrafficPolicy.Name != "" && route.Traffic.BackendTrafficPolicy.Namespace != "" {
			domain = route.Traffic.BackendTrafficPolicy.Name + "-" + route.Traffic.BackendTrafficPolicy.Namespace
		}

		// Handle shared and non-shared rate limits differently
		if isSharedRateLimit(route) {
			addSharedRateLimitDescriptor(route, serviceDescriptors, domain, domainDescriptors)
		} else {
			addRouteSpecificDescriptor(route, serviceDescriptors, domain, domainDescriptors)
		}
	}

	// Convert map to list of RateLimitConfig objects
	return createRateLimitConfigs(domainDescriptors)
}

// addSharedRateLimitDescriptor adds shared rate limit descriptors to the domain descriptor map
func addSharedRateLimitDescriptor(
	route *ir.HTTPRoute,
	serviceDescriptors []*rlsconfv3.RateLimitDescriptor,
	domain string,
	domainDescriptors map[string][]*rlsconfv3.RateLimitDescriptor) {

	// Get BTP details for the shared descriptor
	if !hasValidBackendTrafficPolicy(route) {
		return
	}

	btp := route.Traffic.BackendTrafficPolicy
	sharedKey := btp.Name
	sharedValue := btp.Namespace

	// Check if we already have a descriptor for this key/value pair
	for i, desc := range domainDescriptors[domain] {
		if desc.Key == sharedKey && desc.Value == sharedValue {
			// Found existing shared descriptor, merge new descriptors into it
			domainDescriptors[domain][i].Descriptors = mergeUniqueDescriptors(
				desc.Descriptors, serviceDescriptors)
			return
		}
	}

	// No existing descriptor found, create a new one
	globalDescriptor := &rlsconfv3.RateLimitDescriptor{
		Key:         sharedKey,
		Value:       sharedValue,
		Descriptors: serviceDescriptors,
	}
	domainDescriptors[domain] = append(domainDescriptors[domain], globalDescriptor)
}

// addRouteSpecificDescriptor adds a route-specific descriptor to the domain descriptor map
func addRouteSpecificDescriptor(
	route *ir.HTTPRoute,
	serviceDescriptors []*rlsconfv3.RateLimitDescriptor,
	domain string,
	domainDescriptors map[string][]*rlsconfv3.RateLimitDescriptor) {

	routeDescriptor := &rlsconfv3.RateLimitDescriptor{
		Key:         getRouteDescriptor(route.Name),
		Value:       getRouteDescriptor(route.Name),
		Descriptors: serviceDescriptors,
	}
	domainDescriptors[domain] = append(domainDescriptors[domain], routeDescriptor)
}

// mergeUniqueDescriptors merges two descriptor lists, avoiding duplicates
func mergeUniqueDescriptors(
	existing []*rlsconfv3.RateLimitDescriptor,
	new []*rlsconfv3.RateLimitDescriptor) []*rlsconfv3.RateLimitDescriptor {

	if len(new) == 0 {
		return existing
	}

	result := make([]*rlsconfv3.RateLimitDescriptor, len(existing))
	copy(result, existing)

	// Track descriptors by key/value for faster duplicate checking
	existingMap := make(map[string]bool)
	for _, desc := range existing {
		key := desc.Key + ":" + desc.Value
		existingMap[key] = true
	}

	// Add only non-duplicate descriptors
	for _, desc := range new {
		key := desc.Key + ":" + desc.Value
		if !existingMap[key] {
			result = append(result, desc)
			existingMap[key] = true
		}
	}

	return result
}

// createRateLimitConfigs creates rate limit configs from the domain descriptor map
func createRateLimitConfigs(
	domainDescriptors map[string][]*rlsconfv3.RateLimitDescriptor) []*rlsconfv3.RateLimitConfig {

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

// Helper function to check if a route has a shared rate limit
func isSharedRateLimit(route *ir.HTTPRoute) bool {
	global := route.Traffic.RateLimit.Global
	return global != nil && global.Shared != nil && *global.Shared && len(global.Rules) > 0
}

// buildRateLimitServiceDescriptors creates the rate limit service pb descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(route *ir.HTTPRoute) []*rlsconfv3.RateLimitDescriptor {
	// Safely check that we have a GlobalRateLimit config
	if route == nil || route.Traffic == nil || route.Traffic.RateLimit == nil || route.Traffic.RateLimit.Global == nil {
		return nil
	}
	global := route.Traffic.RateLimit.Global

	pbDescriptors := make([]*rlsconfv3.RateLimitDescriptor, 0, len(global.Rules))
	usedSharedKey := false // Track if the shared key has been used

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

		// 1) Header Matches
		for mIdx, match := range rule.HeaderMatches {
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			// Distinct vs HeaderValueMatch
			if match.Distinct {
				pbDesc.Key = getRouteRuleDescriptor(rIdx, mIdx)
			} else {
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
		// 2) CIDR Match
		if rule.CIDRMatch != nil {
			// MaskedRemoteAddress case
			pbDesc := new(rlsconfv3.RateLimitDescriptor)
			pbDesc.Key = "masked_remote_address"
			pbDesc.Value = rule.CIDRMatch.CIDR

			if cur != nil {
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

			// Determine if we should use the shared rate limit key (BTP-based) or a generic route key
			if isSharedRateLimit(route) && !usedSharedKey && hasValidBackendTrafficPolicy(route) {
				// For shared rate limits, use BTP name and namespace
				btp := route.Traffic.BackendTrafficPolicy
				pbDesc.Key = btp.Name
				pbDesc.Value = btp.Namespace
				usedSharedKey = true
			} else {
				// Use generic key for non-shared rate limits
				pbDesc.Key = getRouteRuleDescriptor(rIdx, -1)
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
		Endpoints: []*ir.DestinationEndpoint{ir.NewDestEndpoint(host, port, false, nil)},
		Name:      destinationSettingName(clusterName),
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

func getRateLimitDomainFilters(irListener *ir.HTTPListener, isShared bool, usedDomains map[string]bool) string {
	// Iterate over the routes to find a domain that hasn't been used yet
	for _, route := range irListener.Routes {
		if hasValidBackendTrafficPolicy(route) {
			domain := route.Traffic.BackendTrafficPolicy.Name + "-" + route.Traffic.BackendTrafficPolicy.Namespace
			if isShared && !usedDomains[domain] {
				return domain
			}
		}
	}
	// Fallback to using the listener's name if no unused domain is found
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

// hasValidBackendTrafficPolicy checks if a route has a non-nil BackendTrafficPolicy with name and namespace.
func hasValidBackendTrafficPolicy(route *ir.HTTPRoute) bool {
	return route != nil &&
		route.Traffic != nil &&
		route.Traffic.BackendTrafficPolicy != nil &&
		route.Traffic.BackendTrafficPolicy.Name != "" &&
		route.Traffic.BackendTrafficPolicy.Namespace != ""
}
