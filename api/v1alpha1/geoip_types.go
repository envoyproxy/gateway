// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// GeoIPProvider defines provider-specific settings.
// +kubebuilder:validation:XValidation:rule="self.type == 'MaxMind' ? has(self.MaxMind) : true",message="MaxMind must be set when type is MaxMind"
type GeoIPProvider struct {
	// +kubebuilder:validation:Enum=MaxMind
	// +kubebuilder:validation:Required
	Type GeoIPProviderType `json:"type"`

	// MaxMind configures the MaxMind provider.
	//
	// +optional
	MaxMind *GeoIPMaxMind `json:"MaxMind,omitempty"`
}

// GeoIPProviderType enumerates GeoIP providers supported by Envoy Gateway.
type GeoIPProviderType string

const (
	// GeoIPProviderTypeMaxMind configures Envoy with the MaxMind provider pointing to local files.
	GeoIPProviderTypeMaxMind GeoIPProviderType = "MaxMind"
)

// GeoIPMaxMind configures the MaxMind provider.
// These database files are expected to be mounted into the Envoy container, and a sidecar container can be used to update the database files.
// +kubebuilder:validation:XValidation:rule="has(self.cityDbPath) || has(self.countryDbPath) || has(self.asnDbPath) || has(self.ispDbPath) || has(self.anonymousIpDbPath)",message="At least one MaxMind database path must be specified"
type GeoIPMaxMind struct {
	// CityDBPath is the path to the City database (.mmdb).
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
	CityDBPath *string `json:"cityDbPath,omitempty"`

	// CountryDBPath is the path to the Country database (.mmdb).
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
	CountryDBPath *string `json:"countryDbPath,omitempty"`

	// ASNDBPath is the path to the ASN database (.mmdb).
	//
	// +optional
	// +kubebuilder:validation:P	attern=`^.*\\.mmdb$`
	ASNDBPath *string `json:"asnDbPath,omitempty"`

	// ISPDBPath is the path to the ISP database (.mmdb).
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
	ISPDBPath *string `json:"ispDbPath,omitempty"`

	// AnonymousIPDBPath is the path to the Anonymous IP database (.mmdb).
	//
	// +optional
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
	AnonymousIPDBPath *string `json:"anonymousIpDbPath,omitempty"`
}

// GeoIPRegion selects a region within a country.
// +kubebuilder:validation:XValidation:rule="has(self.countryCode) && has(self.regionCode)",message="countryCode and regionCode must both be set"
type GeoIPRegion struct {
	// CountryCode is the ISO 3166-1 alpha-2 country code.
	//
	// +kubebuilder:validation:Pattern=`^[A-Z]{2}$`
	CountryCode string `json:"countryCode"`

	// RegionCode is the ISO 3166-2 subdivision code (without country prefix).
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=32
	RegionCode string `json:"regionCode"`
}

// GeoIPCity selects a city, optionally scoped to a region.
// +kubebuilder:validation:XValidation:rule="has(self.countryCode) && has(self.cityName)",message="countryCode and cityName must be set"
type GeoIPCity struct {
	// CountryCode is the ISO 3166-1 alpha-2 country code.
	//
	// +kubebuilder:validation:Pattern=`^[A-Z]{2}$`
	CountryCode string `json:"countryCode"`

	// RegionCode optionally scopes the city to a subdivision (ISO 3166-2 without country prefix).
	//
	// +optional
	// +kubebuilder:validation:MaxLength=32
	RegionCode *string `json:"regionCode,omitempty"`

	// CityName is the city name.
	//
	// +kubebuilder:validation:MinLength=1
	CityName string `json:"cityName"`
}

// GeoIPAnonymousMatch matches anonymous network signals emitted by the GeoIP provider.
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
