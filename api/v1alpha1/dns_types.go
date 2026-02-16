// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// DNSLookupFamily defines the behavior of Envoy when resolving DNS for hostnames
// +enum
// +kubebuilder:validation:Enum=IPv4;IPv6;IPv4Preferred;IPv6Preferred;IPv4AndIPv6
type DNSLookupFamily string

const (
	// IPv4DNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv4 family.
	IPv4DNSLookupFamily DNSLookupFamily = "IPv4"
	// IPv6DNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv6 family.
	IPv6DNSLookupFamily DNSLookupFamily = "IPv6"
	// IPv4PreferredDNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv4 family and fallback
	// to a lookup for addresses in the IPv6 family.
	IPv4PreferredDNSLookupFamily DNSLookupFamily = "IPv4Preferred"
	// IPv6PreferredDNSLookupFamily means the DNS resolver will first perform a lookup for addresses in the IPv6 family and fallback
	// to a lookup for addresses in the IPv4 family.
	IPv6PreferredDNSLookupFamily DNSLookupFamily = "IPv6Preferred"
	// IPv4AndIPv6DNSLookupFamily mean the DNS resolver will perform a lookup for both IPv4 and IPv6 families, and return all resolved
	// addresses. When this is used, Happy Eyeballs will be enabled for upstream connections.
	IPv4AndIPv6DNSLookupFamily DNSLookupFamily = "IPv4AndIPv6"
)

type DNS struct {
	// DNSRefreshRate specifies the rate at which DNS records should be refreshed.
	// Defaults to 30 seconds.
	//
	// +optional
	DNSRefreshRate *gwapiv1.Duration `json:"dnsRefreshRate,omitempty"`
	// RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.
	// If the value is set to true, the DNS refresh rate will be set to the resource record’s TTL.
	// Defaults to true.
	//
	// +optional
	RespectDNSTTL *bool `json:"respectDnsTtl,omitempty"`
	// LookupFamily determines how Envoy would resolve DNS for Routes where the backend is specified as a fully qualified domain name (FQDN).
	// If set, this configuration overrides other defaults.
	// +optional
	LookupFamily *DNSLookupFamily `json:"lookupFamily,omitempty"`
}
