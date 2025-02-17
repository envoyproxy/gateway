// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// DNSLookupFamily defines the behavior of Envoy when resolving DNS for hostnames
// +enum
// +kubebuilder:validation:Enum=IPv4Only;IPv6Only;IPv4Preferred;IPv6Preferred;IPv4AndIPv6
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
	DNSRefreshRate *metav1.Duration `json:"dnsRefreshRate,omitempty"`
	// RespectDNSTTL indicates whether the DNS Time-To-Live (TTL) should be respected.
	// If the value is set to true, the DNS refresh rate will be set to the resource record’s TTL.
	// Defaults to true.
	RespectDNSTTL *bool `json:"respectDnsTtl,omitempty"`
	// LookupFamily determines how Envoy would resolve DNS for. If set, this configuration overrides other default
	// value that Envoy Gateway configures based on attributes of the backends, such Service resource IPFamilies.
	// +optional
	// +notImplementedHide
	LookupFamily *DNSLookupFamily `json:"lookupFamily,omitempty"`
}
