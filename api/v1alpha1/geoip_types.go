// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// GeoIP defines GeoIP enrichment and access control configuration.
type GeoIP struct {
	// Source configures how the client IP is extracted before being passed to the provider.
	// If unset, Envoy falls back to using the immediate downstream connection address.
	//
	// +optional
	Source *GeoIPSource `json:"source,omitempty"`

	// Provider defines the GeoIP provider configuration.
	Provider GeoIPProvider `json:"provider"`

	// Access defines the GeoIP based access control configuration.
	//
	// +optional
	Access *GeoIPAccessControl `json:"access,omitempty"`
}

// GeoIPSource configures how Envoy determines the client IP address that is passed to the provider.
// +kubebuilder:validation:XValidation:rule="self.type == 'XFF' ? has(self.xff) && !has(self.header) : self.type == 'Header' ? has(self.header) && !has(self.xff) : true",message="When type is XFF, xff must be set (and header unset). When type is Header, header must be set (and xff unset)."
type GeoIPSource struct {
	// +kubebuilder:validation:Enum=XFF;Header
	// +kubebuilder:validation:Required
	Type GeoIPSourceType `json:"type"`

	// XFF configures extraction based on the X-Forwarded-For header chain.
	//
	// +optional
	XFF *GeoIPXFFSource `json:"xff,omitempty"`

	// Header configures extraction from a custom header.
	//
	// +optional
	Header *GeoIPHeaderSource `json:"header,omitempty"`
}

// GeoIPSourceType enumerates supported client IP sources.
type GeoIPSourceType string

const (
	// GeoIPSourceTypeXFF instructs Envoy to honor the X-Forwarded-For header count.
	GeoIPSourceTypeXFF GeoIPSourceType = "XFF"
	// GeoIPSourceTypeHeader instructs Envoy to read a custom request header.
	GeoIPSourceTypeHeader GeoIPSourceType = "Header"
)

// GeoIPXFFSource configures trusted hop count for XFF parsing.
type GeoIPXFFSource struct {
	// TrustedHops defines the number of trusted hops from the right side of XFF.
	// Defaults to 0 when unset.
	//
	// +optional
	TrustedHops *uint32 `json:"trustedHops,omitempty"`
}

// GeoIPHeaderSource configures extraction from a custom header.
type GeoIPHeaderSource struct {
	// HeaderName is the HTTP header that carries the client IP.
	//
	// +kubebuilder:validation:MinLength=1
	HeaderName string `json:"headerName"`
}

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
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
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

// GeoIPAccessControl defines GeoIP-based allow/deny lists.
type GeoIPAccessControl struct {
	// DefaultAction defines how to handle requests that do not match any rule or lack GeoIP data.
	// Defaults to Allow when unset.
	//
	// +optional
	DefaultAction *AuthorizationAction `json:"defaultAction,omitempty"`

	// Rules evaluated in order. The first matching rule's action applies.
	//
	// +optional
	Rules []GeoIPRule `json:"rules,omitempty"`
}

// GeoIPRule defines a single GeoIP allow/deny rule.
// +kubebuilder:validation:XValidation:rule="has(self.countries) || has(self.regions) || has(self.cities)",message="At least one of countries, regions, or cities must be specified"
type GeoIPRule struct {
	// Action is reused from Authorization rules (Allow or Deny).
	Action AuthorizationAction `json:"action"`

	// Countries is a list of ISO 3166-1 alpha-2 country codes.
	//
	// +optional
	Countries []string `json:"countries,omitempty"`

	// Regions refines matching to ISO 3166-2 subdivisions.
	//
	// +optional
	Regions []GeoIPRegion `json:"regions,omitempty"`

	// Cities refines matching to specific city names.
	//
	// +optional
	Cities []GeoIPCity `json:"cities,omitempty"`
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
