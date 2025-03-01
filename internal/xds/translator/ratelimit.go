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

	rateLimitFilter := t.buildRateLimitFilter(irListener)
	mgr.HttpFilters = append([]*hcmv3.HttpFilter{rateLimitFilter}, mgr.HttpFilters...)
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

func (t *Translator) buildRateLimitFilter(irListener *ir.HTTPListener) *hcmv3.HttpFilter {
	rateLimitFilterProto := &ratelimitfilterv3.RateLimit{
		Domain: getRateLimitDomain(irListener),
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
	if t.GlobalRateLimit.Timeout > 0 {
		rateLimitFilterProto.Timeout = durationpb.New(t.GlobalRateLimit.Timeout)
	}

	if irListener.Headers != nil && irListener.Headers.DisableRateLimitHeaders {
		rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_OFF
	} else {
		rateLimitFilterProto.EnableXRatelimitHeaders = ratelimitfilterv3.RateLimit_DRAFT_VERSION_03
	}

	if t.GlobalRateLimit.FailClosed {
		rateLimitFilterProto.FailureModeDeny = t.GlobalRateLimit.FailClosed
	}

	rateLimitFilterAny, err := anypb.New(rateLimitFilterProto)
	if err != nil {
		return nil
	}

	rateLimitFilter := &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterRateLimit.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: rateLimitFilterAny,
		},
	}
	return rateLimitFilter
}

// patchRouteWithRateLimit builds rate limit actions and appends to the route.
func patchRouteWithRateLimit(route *routev3.Route, irRoute *ir.HTTPRoute) error { //nolint:unparam
	// Return early if no rate limit config exists.
	xdsRouteAction := route.GetRoute()
	if !routeContainsGlobalRateLimit(irRoute) || xdsRouteAction == nil {
		return nil
	}
	global := irRoute.Traffic.RateLimit.Global
	rateLimits := buildRouteRateLimits(irRoute.Name, global)
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

func buildRouteRateLimits(descriptorPrefix string, global *ir.GlobalRateLimit) (rateLimits []*routev3.RateLimit) {
	// Route descriptor for each route rule action
	routeDescriptor := &routev3.RateLimit_Action{
		ActionSpecifier: &routev3.RateLimit_Action_GenericKey_{
			GenericKey: &routev3.RateLimit_Action_GenericKey{
				DescriptorKey:   getRouteDescriptor(descriptorPrefix),
				DescriptorValue: getRouteDescriptor(descriptorPrefix),
			},
		},
	}

	// Rules are ORed
	for rIdx, rule := range global.Rules {
		// Matches are ANDed
		rlActions := []*routev3.RateLimit_Action{routeDescriptor}
		for mIdx, match := range rule.HeaderMatches {
			var action *routev3.RateLimit_Action
			// Case for distinct match
			if match.Distinct {
				// Setup RequestHeader actions
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
				// Setup HeaderValueMatch actions
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
		if rule.CIDRMatch != nil {
			// Setup MaskedRemoteAddress action
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

			// Setup RemoteAddress action if distinct match is set
			if rule.CIDRMatch.Distinct {
				// Setup RemoteAddress action
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
		if !rule.IsMatchSet() {
			// Setup GenericKey action
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

		rateLimit := &routev3.RateLimit{Actions: rlActions}
		if c := rule.RequestCost; c != nil {
			rateLimit.HitsAddend = rateLimitCostToHitsAddend(c)
		}
		rateLimits = append(rateLimits, rateLimit)
		if c := rule.ResponseCost; c != nil {
			// To apply the cost to the response, we need to set ApplyOnStreamDone to true which is per Rule option,
			// so we need to create a new RateLimit for the response with the option set.
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

// BuildRateLimitServiceConfig builds the rate limit service configuration based on
// https://github.com/envoyproxy/ratelimit#the-configuration-format
func BuildRateLimitServiceConfig(irListener *ir.HTTPListener) *rlsconfv3.RateLimitConfig {
	pbDescriptors := make([]*rlsconfv3.RateLimitDescriptor, 0, len(irListener.Routes))

	for _, route := range irListener.Routes {
		if routeContainsGlobalRateLimit(route) {
			serviceDescriptors := buildRateLimitServiceDescriptors(route.Traffic.RateLimit.Global)

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
			routeDescriptor := &rlsconfv3.RateLimitDescriptor{
				Key:         getRouteDescriptor(route.Name),
				Value:       getRouteDescriptor(route.Name),
				Descriptors: serviceDescriptors,
			}
			pbDescriptors = append(pbDescriptors, routeDescriptor)
		}
	}

	if len(pbDescriptors) == 0 {
		return nil
	}

	domain := getRateLimitDomain(irListener)
	return &rlsconfv3.RateLimitConfig{
		Name:        domain,
		Domain:      domain,
		Descriptors: pbDescriptors,
	}
}

// buildRateLimitServiceDescriptors creates the rate limit service pb descriptors based on the global rate limit IR config.
func buildRateLimitServiceDescriptors(global *ir.GlobalRateLimit) []*rlsconfv3.RateLimitDescriptor {
	pbDescriptors := make([]*rlsconfv3.RateLimitDescriptor, 0, len(global.Rules))

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
			pbDesc.Key = getRouteRuleDescriptor(rIdx, -1)
			pbDesc.Value = getRouteRuleDescriptor(rIdx, -1)
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

func getRateLimitDomain(irListener *ir.HTTPListener) string {
	// Use IR listener name as domain
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
