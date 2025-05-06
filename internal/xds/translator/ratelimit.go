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
// It creates at most one filter per domain to avoid duplicates.
func (t *Translator) buildRateLimitFilter(irListener *ir.HTTPListener) []*hcmv3.HttpFilter {
	if irListener == nil || irListener.Routes == nil {
		return nil
	}

	var filters []*hcmv3.HttpFilter
	// Map to track which domains we've already created filters for, prevents creating duplicate filters for the same domain
	processedDomains := make(map[string]bool)

	// Iterate over each route in the listener to create rate limit filters for shared rate limits.
	for _, route := range irListener.Routes {
		if !routeContainsGlobalRateLimit(route) {
			continue
		}

		var domain string
		filterName := getRateLimitFilterName(route)
		if isSharedRateLimit(route) {
			// For shared rate limits, use the domain derived from the traffic policy
			domain = getDomainName(route)
		} else {
			// For non-shared rate limits, use the listener domain
			domain = irListener.Name
		}

		// Skip if we've already created a filter for this domain
		if processedDomains[domain] {
			continue
		}
		processedDomains[domain] = true

		// Create a filter for this domain
		filter := createRateLimitFilter(t, irListener, domain, filterName)
		if filter != nil {
			filters = append(filters, filter)
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
	if !routeContainsGlobalRateLimit(irRoute) || xdsRouteAction == nil {
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
	if !routeContainsGlobalRateLimit(route) {
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
		if rule.Shared != nil && *rule.Shared {
			// For shared rule, use traffic policy name
			descriptorKey = route.Traffic.Name
			descriptorValue = route.Traffic.Name
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

			// We DO NOT add another generic key action here
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
	// Map to store descriptors for each domain
	domainDescriptors := make(map[string][]*rlsconfv3.RateLimitDescriptor)
	// Map to track which domains we've already created filters for, prevents creating duplicate filters for the same domain
	processedSharedDomains := make(map[string]bool)

	// Process each listener
	for _, irListener := range irListeners {
		// Process each route to build descriptors
		for _, route := range irListener.Routes {
			if !routeContainsGlobalRateLimit(route) {
				continue
			}

			domain := irListener.Name
			if isSharedRateLimit(route) {
				domain = getDomainName(route)

				// Skip if we've already processed this shared domain
				if processedSharedDomains[domain] {
					continue
				}
				processedSharedDomains[domain] = true
			}

			// Get route rule descriptors within each route
			serviceDescriptors := buildRateLimitServiceDescriptors(route)
			if len(serviceDescriptors) == 0 {
				continue
			}

			// Add the rate limit descriptor (handles both shared and non-shared)
			addRateLimitDescriptor(route, serviceDescriptors, domain, domainDescriptors)
		}
	}

	configs := createRateLimitConfigs(domainDescriptors)
	return configs
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
) {
	var key, value string

	if isSharedRateLimit(route) {
		// For shared rate limits, use traffic policy name key/value
		key = route.Traffic.Name
		value = route.Traffic.Name
	} else {
		// For non-shared rate limits, use route descriptor key/value
		key = getRouteDescriptor(route.Name)
		value = getRouteDescriptor(route.Name)
	}

	descriptor := &rlsconfv3.RateLimitDescriptor{
		Key:         key,
		Value:       value,
		Descriptors: serviceDescriptors,
	}
	domainDescriptors[domain] = append(domainDescriptors[domain], descriptor)
}

// Helper function to check if a route has a shared rate limit
func isSharedRateLimit(route *ir.HTTPRoute) bool {
	if route == nil || route.Traffic == nil || route.Traffic.RateLimit == nil || route.Traffic.RateLimit.Global == nil {
		return false
	}

	global := route.Traffic.RateLimit.Global
	if len(global.Rules) == 0 {
		return false
	}

	// Check if any rule has shared=true
	for _, rule := range global.Rules {
		if rule.Shared != nil && *rule.Shared {
			return true
		}
	}

	return false
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

		// Create root descriptor based on rule's shared attribute
		rootDesc := new(rlsconfv3.RateLimitDescriptor)
		if rule.Shared != nil && *rule.Shared {
			rootDesc.Key = route.Traffic.Name
			rootDesc.Value = route.Traffic.Name
		} else {
			rootDesc.Key = getRouteRuleDescriptor(rIdx, -1)
			rootDesc.Value = rootDesc.Key
		}

		// We'll work with the current descriptor pointer
		currDesc := rootDesc

		// 1) Header Matches - add as nested descriptors if needed
		if len(rule.HeaderMatches) > 0 {
			for mIdx, match := range rule.HeaderMatches {
				headerDesc := new(rlsconfv3.RateLimitDescriptor)
				// Distinct vs HeaderValueMatch
				if match.Distinct {
					// RequestHeader case
					headerDesc.Key = getRouteRuleDescriptor(rIdx, mIdx)
				} else {
					// HeaderValueMatch case
					headerDesc.Key = getRouteRuleDescriptor(rIdx, mIdx)
					headerDesc.Value = getRouteRuleDescriptor(rIdx, mIdx)
				}

				if mIdx == 0 {
					// First header - add to root
					currDesc.Descriptors = []*rlsconfv3.RateLimitDescriptor{headerDesc}
				} else {
					// Chain to previous header
					currDesc.Descriptors = []*rlsconfv3.RateLimitDescriptor{headerDesc}
				}
				currDesc = headerDesc
			}
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
			cidrDesc := new(rlsconfv3.RateLimitDescriptor)
			cidrDesc.Key = "masked_remote_address"
			cidrDesc.Value = rule.CIDRMatch.CIDR

			// Add CIDR descriptor to the chain
			if len(rule.HeaderMatches) > 0 {
				currDesc.Descriptors = []*rlsconfv3.RateLimitDescriptor{cidrDesc}
			} else {
				rootDesc.Descriptors = []*rlsconfv3.RateLimitDescriptor{cidrDesc}
			}
			currDesc = cidrDesc

			// Add remote_address descriptor for distinct CIDRs
			if rule.CIDRMatch.Distinct {
				remoteAddrDesc := new(rlsconfv3.RateLimitDescriptor)
				remoteAddrDesc.Key = "remote_address"
				currDesc.Descriptors = []*rlsconfv3.RateLimitDescriptor{remoteAddrDesc}
				currDesc = remoteAddrDesc
			}
		}

		// Add rate limit policy to the last descriptor in the chain
		currDesc.RateLimit = rateLimitPolicy
		pbDescriptors = append(pbDescriptors, rootDesc)
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

func getDomainName(route *ir.HTTPRoute) string {
	return strings.Replace(route.Traffic.Name, "/", "-", 1)
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

// For shared rate limits, it appends the traffic policy name to the base filter name.
// For non-shared rate limits, it returns just the base filter name.
func getRateLimitFilterName(route *ir.HTTPRoute) string {
	filterName := egv1a1.EnvoyFilterRateLimit.String()
	if isSharedRateLimit(route) {
		filterName = fmt.Sprintf("%s/%s", filterName, route.Traffic.Name)
	}
	return filterName
}
