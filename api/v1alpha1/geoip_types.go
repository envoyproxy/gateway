// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

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
// +union
// +kubebuilder:validation:XValidation:rule="has(self.local) || has(self.remote)",message="at least one of local or remote must be specified"
type GeoIPDBSource struct {
	// Local is a database source from a local file.
	//
	// +optional
	Local *LocalGeoIPDBSource `json:"local,omitempty"`

	// Remote is a database source fetched from a remote URL.
	// TODO: implement this in the future
	// +notImplementedHide
	// +optional
	Remote *RemoteGeoIPDBSource `json:"remote,omitempty"`
}

// LocalGeoIPDBSource configures a GeoIP database from a local file path.
type LocalGeoIPDBSource struct {
	// Path is the path to the database file.
	//
	// +kubebuilder:validation:Pattern=`^.*\\.mmdb$`
	Path string `json:"path"`
}

// RemoteGeoIPDBSource configures a GeoIP database fetched from a remote URL.
// TODO: implement this in the future
// +notImplementedHide
type RemoteGeoIPDBSource struct{}

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
