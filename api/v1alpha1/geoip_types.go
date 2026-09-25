// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import (
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// GeoIPProvider defines provider-specific settings.
// +kubebuilder:validation:XValidation:rule="self.type == 'MaxMind' ? has(self.maxMind) : true",message="maxMind must be set when type is MaxMind"
type GeoIPProvider struct {
	// +kubebuilder:validation:Enum=MaxMind
	// +kubebuilder:validation:Required
	Type GeoIPProviderType `json:"type"`

	// MaxMind configures the MaxMind provider.
	//
	// +optional
	MaxMind *GeoIPMaxMind `json:"maxMind,omitempty"`
}

// GeoIPProviderType enumerates GeoIP providers supported by Envoy Gateway.
type GeoIPProviderType string

const (
	// GeoIPProviderTypeMaxMind configures Envoy with the MaxMind provider pointing to local files.
	GeoIPProviderTypeMaxMind GeoIPProviderType = "MaxMind"
)

// GeoIPMaxMind configures the MaxMind provider.
// These database files are expected to be mounted into the Envoy container, and a sidecar container can be used to update the database files.
// +kubebuilder:validation:XValidation:rule="has(self.cityDbSource) || has(self.countryDbSource) || has(self.asnDbSource) || has(self.ispDbSource) || has(self.anonymousIpDbSource)",message="at least one MaxMind database source must be specified"
type GeoIPMaxMind struct {
	// CityDBSource configures the City database source.
	//
	// +optional
	CityDBSource *GeoIPDBSource `json:"cityDbSource,omitempty"`

	// CountryDBSource configures the Country database source.
	//
	// +optional
	CountryDBSource *GeoIPDBSource `json:"countryDbSource,omitempty"`

	// ASNDBSource configures the ASN database source.
	//
	// +optional
	ASNDBSource *GeoIPDBSource `json:"asnDbSource,omitempty"`

	// ISPDBSource configures the ISP database source.
	//
	// +optional
	ISPDBSource *GeoIPDBSource `json:"ispDbSource,omitempty"`

	// AnonymousIPDBSource configures the Anonymous IP database source.
	//
	// +optional
	AnonymousIPDBSource *GeoIPDBSource `json:"anonymousIpDbSource,omitempty"`
}

// GeoIPDBSource defines where a GeoIP .mmdb database can be loaded from.
type GeoIPDBSource struct {
	// Local is a database source from a local file.
	Local LocalGeoIPDBSource `json:"local"`
}

// LocalGeoIPDBSource configures a GeoIP database from a local file path.
type LocalGeoIPDBSource struct {
	// Path is the path to the database file.
	//
	// +kubebuilder:validation:Pattern=`^.*\.mmdb$`
	Path string `json:"path"`
}

// GeoIPAnonymousMatch matches anonymous network signals emitted by the GeoIP provider.
// If multiple fields are specified, all specified fields must match.
// These signals are not mutually exclusive. A single IP may satisfy multiple
// flags at the same time (for example, a commercial VPN exit IP may also be
// classified as a public proxy, so both IsVPN and IsProxy can be true).
//
// +kubebuilder:validation:XValidation:rule="has(self.isAnonymous) || has(self.isVPN) || has(self.isHosting) || has(self.isTor) || has(self.isProxy)",message="at least one of isAnonymous, isVPN, isHosting, isTor, or isProxy must be specified"
type GeoIPAnonymousMatch struct {
	// IsAnonymous matches whether the client IP is considered anonymous.
	//
	// +optional
	IsAnonymous *bool `json:"isAnonymous,omitempty"`

	// IsVPN matches whether the client IP is detected as VPN.
	//
	// +optional
	IsVPN *bool `json:"isVPN,omitempty"`

	// IsHosting matches whether the client IP belongs to a hosting provider.
	//
	// +optional
	IsHosting *bool `json:"isHosting,omitempty"`

	// IsTor matches whether the client IP belongs to a Tor exit node.
	//
	// +optional
	IsTor *bool `json:"isTor,omitempty"`

	// IsProxy matches whether the client IP belongs to a public proxy.
	//
	// +optional
	IsProxy *bool `json:"isProxy,omitempty"`
}

// GeoIPEnrichment configures request enrichment with GeoIP-derived data.
// When set, Envoy Gateway inserts the envoy.filters.http.geoip filter into the
// listener's HTTP filter chain and populates the configured request headers
// from the MaxMind provider defined on the EnvoyProxy resource.
//
// The MaxMind provider exposes GeoIP attributes only as request headers, so the
// enriched headers are both forwarded to the backend and available to access
// logs (reference them with %REQ(<header>)% in the EnvoyProxy access log
// format).
type GeoIPEnrichment struct {
	// RequestHeaders adds GeoIP-derived HTTP request headers. Only the fields
	// set here are populated, and each requires the corresponding MaxMind
	// database source on the provider (for example, region and city require
	// cityDbSource). Header values are set by Envoy Gateway and overwrite any
	// client-supplied header of the same name, so backends can treat them as
	// trusted.
	RequestHeaders GeoIPRequestHeaders `json:"requestHeaders"`
}

// GeoIPRequestHeaders maps GeoIP fields to the HTTP request header names they
// are written to. Only fields that are set are populated.
//
// +kubebuilder:validation:XValidation:rule="has(self.country) || has(self.region) || has(self.city) || has(self.asn) || has(self.isp)",message="at least one GeoIP field must be set"
type GeoIPRequestHeaders struct {
	// Country is the header name to which the ISO 3166-1 country code is
	// written. Requires countryDbSource or cityDbSource on the provider.
	//
	// +optional
	Country *gwapiv1.HTTPHeaderName `json:"country,omitempty"`

	// Region is the header name to which the region/subdivision ISO code is
	// written. Requires cityDbSource on the provider.
	//
	// +optional
	Region *gwapiv1.HTTPHeaderName `json:"region,omitempty"`

	// City is the header name to which the city name is written.
	// Requires cityDbSource on the provider.
	//
	// +optional
	City *gwapiv1.HTTPHeaderName `json:"city,omitempty"`

	// ASN is the header name to which the autonomous system number is written.
	// Requires asnDbSource on the provider.
	//
	// +optional
	ASN *gwapiv1.HTTPHeaderName `json:"asn,omitempty"`

	// ISP is the header name to which the ISP name is written.
	// Requires ispDbSource on the provider.
	//
	// +optional
	ISP *gwapiv1.HTTPHeaderName `json:"isp,omitempty"`
}
