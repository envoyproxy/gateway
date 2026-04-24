// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package translator

import (
	"errors"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	routev3 "github.com/envoyproxy/go-control-plane/envoy/config/route/v3"
	httpgeoipv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/http/geoip/v3"
	hcmv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/filters/network/http_connection_manager/v3"
	commonv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/geoip_providers/common/v3"
	maxmindv3 "github.com/envoyproxy/go-control-plane/envoy/extensions/geoip_providers/maxmind/v3"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils/proto"
	"github.com/envoyproxy/gateway/internal/xds/types"
)

const (
	geoIPMaxMindProviderName     = "envoy.geoip_providers.maxmind"
	geoIPInternalCountryHeader   = "x-eg-internal-geoip-country"
	geoIPInternalRegionHeader    = "x-eg-internal-geoip-region"
	geoIPInternalCityHeader      = "x-eg-internal-geoip-city"
	geoIPInternalASNHeader       = "x-eg-internal-geoip-asn"
	geoIPInternalISPHeader       = "x-eg-internal-geoip-isp"
	geoIPInternalAnonHeader      = "x-eg-internal-geoip-anon"
	geoIPInternalAnonVPNHeader   = "x-eg-internal-geoip-anon-vpn"
	geoIPInternalAnonHostHeader  = "x-eg-internal-geoip-anon-hosting"
	geoIPInternalAnonTorHeader   = "x-eg-internal-geoip-anon-tor"
	geoIPInternalAnonProxyHeader = "x-eg-internal-geoip-anon-proxy"
)

func init() {
	registerHTTPFilter(&geoip{})
}

type geoip struct{}

var _ httpFilter = &geoip{}

type geoIPFieldRequirements struct {
	country   bool
	region    bool
	city      bool
	asn       bool
	isp       bool
	anon      bool
	anonVPN   bool
	anonHost  bool
	anonTor   bool
	anonProxy bool
}

func (r geoIPFieldRequirements) any() bool {
	return r.country || r.region || r.city || r.asn || r.isp || r.anon || r.anonVPN || r.anonHost || r.anonTor || r.anonProxy
}

func (r geoIPFieldRequirements) headers() []string {
	headers := make([]string, 0, 10)
	if r.country {
		headers = append(headers, geoIPInternalCountryHeader)
	}
	if r.region {
		headers = append(headers, geoIPInternalRegionHeader)
	}
	if r.city {
		headers = append(headers, geoIPInternalCityHeader)
	}
	if r.asn {
		headers = append(headers, geoIPInternalASNHeader)
	}
	if r.isp {
		headers = append(headers, geoIPInternalISPHeader)
	}
	if r.anon {
		headers = append(headers, geoIPInternalAnonHeader)
	}
	if r.anonVPN {
		headers = append(headers, geoIPInternalAnonVPNHeader)
	}
	if r.anonHost {
		headers = append(headers, geoIPInternalAnonHostHeader)
	}
	if r.anonTor {
		headers = append(headers, geoIPInternalAnonTorHeader)
	}
	if r.anonProxy {
		headers = append(headers, geoIPInternalAnonProxyHeader)
	}
	return headers
}

func geoIPRequirementsForAuthorization(authorization *ir.Authorization) geoIPFieldRequirements {
	var requirements geoIPFieldRequirements
	if authorization == nil {
		return requirements
	}

	for _, rule := range authorization.Rules {
		if rule == nil {
			continue
		}
		for _, geo := range rule.Principal.ClientIPGeoLocations {
			requirements.country = requirements.country || geo.Country != nil
			requirements.region = requirements.region || geo.Region != nil
			requirements.city = requirements.city || geo.City != nil
			requirements.asn = requirements.asn || geo.ASN != nil
			requirements.isp = requirements.isp || geo.ISP != nil
			if geo.Anonymous == nil {
				continue
			}
			requirements.anon = requirements.anon || geo.Anonymous.IsAnonymous != nil
			requirements.anonVPN = requirements.anonVPN || geo.Anonymous.IsVPN != nil
			requirements.anonHost = requirements.anonHost || geo.Anonymous.IsHosting != nil
			requirements.anonTor = requirements.anonTor || geo.Anonymous.IsTor != nil
			requirements.anonProxy = requirements.anonProxy || geo.Anonymous.IsProxy != nil
		}
	}

	return requirements
}

func geoIPRequirementsForHTTPListener(irListener *ir.HTTPListener) geoIPFieldRequirements {
	var requirements geoIPFieldRequirements
	if irListener == nil {
		return requirements
	}

	for _, route := range irListener.Routes {
		if route == nil || route.Security == nil {
			continue
		}

		routeRequirements := geoIPRequirementsForAuthorization(route.Security.Authorization)
		requirements.country = requirements.country || routeRequirements.country
		requirements.region = requirements.region || routeRequirements.region
		requirements.city = requirements.city || routeRequirements.city
		requirements.asn = requirements.asn || routeRequirements.asn
		requirements.isp = requirements.isp || routeRequirements.isp
		requirements.anon = requirements.anon || routeRequirements.anon
		requirements.anonVPN = requirements.anonVPN || routeRequirements.anonVPN
		requirements.anonHost = requirements.anonHost || routeRequirements.anonHost
		requirements.anonTor = requirements.anonTor || routeRequirements.anonTor
		requirements.anonProxy = requirements.anonProxy || routeRequirements.anonProxy
	}

	return requirements
}

func geoIPHeadersToRemove(irListener *ir.HTTPListener) []string {
	return geoIPRequirementsForHTTPListener(irListener).headers()
}

func (*geoip) patchHCM(mgr *hcmv3.HttpConnectionManager, irListener *ir.HTTPListener) error {
	if mgr == nil {
		return errors.New("hcm is nil")
	}
	if irListener == nil {
		return errors.New("ir listener is nil")
	}

	requirements := geoIPRequirementsForHTTPListener(irListener)
	if !requirements.any() {
		return nil
	}

	// geoIPProvider could be nil when GeoIP authorization is configured.
	// Just silently ignore the GeoIP filter to avoid blocking xDS generation.
	// The GeoIP provider missing errors are already surfaced to the status of the SecurityPolicy.
	if irListener.GeoIPProvider == nil || irListener.GeoIPProvider.MaxMind == nil {
		return nil
	}

	if hcmContainsFilter(mgr, egv1a1.EnvoyFilterGeoIP.String()) {
		return nil
	}

	filter, err := buildHCMGeoIPFilter(irListener, requirements)
	if err != nil {
		return err
	}

	mgr.HttpFilters = append(mgr.HttpFilters, filter)
	return nil
}

func buildHCMGeoIPFilter(irListener *ir.HTTPListener, requirements geoIPFieldRequirements) (*hcmv3.HttpFilter, error) {
	provider, err := buildGeoIPProviderExtension(irListener.GeoIPProvider, requirements)
	if err != nil {
		return nil, err
	}

	cfg := &httpgeoipv3.Geoip{
		Provider: provider,
	}
	// irListener.ClientIPDetection should never be nil since we've already verified it in the Gateway API translator, just a sanity check
	if irListener.ClientIPDetection != nil {
		switch {
		case irListener.ClientIPDetection.CustomHeader != nil:
			cfg.CustomHeaderConfig = &httpgeoipv3.Geoip_CustomHeaderConfig{
				HeaderName: irListener.ClientIPDetection.CustomHeader.Name,
			}
		case irListener.ClientIPDetection.XForwardedFor != nil && irListener.ClientIPDetection.XForwardedFor.NumTrustedHops != nil:
			cfg.XffConfig = &httpgeoipv3.Geoip_XffConfig{
				XffNumTrustedHops: xffNumTrustedHops(irListener.ClientIPDetection),
			}
		}
	}

	typedConfig, err := proto.ToAnyWithValidation(cfg)
	if err != nil {
		return nil, err
	}

	return &hcmv3.HttpFilter{
		Name: egv1a1.EnvoyFilterGeoIP.String(),
		ConfigType: &hcmv3.HttpFilter_TypedConfig{
			TypedConfig: typedConfig,
		},
	}, nil
}

func buildGeoIPProviderExtension(geoIPProvider *ir.GeoIPProvider, requirements geoIPFieldRequirements) (*corev3.TypedExtensionConfig, error) {
	// geoIPProvider should never be nil since we've already verified it in the Gateway API translator, just a sanity check
	if geoIPProvider == nil || geoIPProvider.MaxMind == nil {
		return nil, errors.New("geoIP provider is nil")
	}

	fieldKeys := &commonv3.CommonGeoipProviderConfig_GeolocationFieldKeys{}
	if requirements.country {
		fieldKeys.Country = geoIPInternalCountryHeader
	}
	if requirements.region {
		fieldKeys.Region = geoIPInternalRegionHeader
	}
	if requirements.city {
		fieldKeys.City = geoIPInternalCityHeader
	}
	if requirements.asn {
		fieldKeys.Asn = geoIPInternalASNHeader
	}
	if requirements.isp {
		fieldKeys.Isp = geoIPInternalISPHeader
	}
	if requirements.anon {
		fieldKeys.Anon = geoIPInternalAnonHeader
	}
	if requirements.anonVPN {
		fieldKeys.AnonVpn = geoIPInternalAnonVPNHeader
	}
	if requirements.anonHost {
		fieldKeys.AnonHosting = geoIPInternalAnonHostHeader
	}
	if requirements.anonTor {
		fieldKeys.AnonTor = geoIPInternalAnonTorHeader
	}
	if requirements.anonProxy {
		fieldKeys.AnonProxy = geoIPInternalAnonProxyHeader
	}

	maxMindConfig := &maxmindv3.MaxMindConfig{
		CityDbPath:    ptr.Deref(geoIPProvider.MaxMind.CityDBPath, ""),
		AsnDbPath:     ptr.Deref(geoIPProvider.MaxMind.ASNDBPath, ""),
		AnonDbPath:    ptr.Deref(geoIPProvider.MaxMind.AnonymousIPDBPath, ""),
		IspDbPath:     ptr.Deref(geoIPProvider.MaxMind.ISPDBPath, ""),
		CountryDbPath: ptr.Deref(geoIPProvider.MaxMind.CountryDBPath, ""),
		CommonProviderConfig: &commonv3.CommonGeoipProviderConfig{
			GeoFieldKeys: fieldKeys,
		},
	}

	typedConfig, err := proto.ToAnyWithValidation(maxMindConfig)
	if err != nil {
		return nil, err
	}

	return &corev3.TypedExtensionConfig{
		Name:        geoIPMaxMindProviderName,
		TypedConfig: typedConfig,
	}, nil
}

func (*geoip) patchRoute(*routev3.Route, *ir.HTTPRoute, *ir.HTTPListener) error {
	return nil
}

func (*geoip) patchResources(*types.ResourceVersionTable, []*ir.HTTPRoute) error {
	return nil
}
