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
