// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// LoadBalancer defines the load balancer policy to be applied.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'ConsistentHash' ? has(self.consistentHash) : !has(self.consistentHash)",message="If LoadBalancer type is consistentHash, consistentHash field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'HostOverride' ? has(self.hostOverrideSettings) : !has(self.hostOverrideSettings)",message="If LoadBalancer type is HostOverride, hostOverrideSettings field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type in ['Random', 'ConsistentHash', 'HostOverride'] ? !has(self.slowStart) : true ",message="Currently SlowStart is only supported for RoundRobin and LeastRequest load balancers."
// +kubebuilder:validation:XValidation:rule="self.type in ['ConsistentHash', 'HostOverride'] ? !has(self.zoneAware) : true ",message="Currently ZoneAware is only supported for LeastRequest, Random, and RoundRobin load balancers."
type LoadBalancer struct {
	// Type decides the type of Load Balancer policy.
	// Valid LoadBalancerType values are
	// "ConsistentHash",
	// "LeastRequest",
	// "Random",
	// "RoundRobin",
	// "HostOverride".
	//
	// +unionDiscriminator
	Type LoadBalancerType `json:"type"`
	// ConsistentHash defines the configuration when the load balancer type is
	// set to ConsistentHash
	//
	// +optional
	ConsistentHash *ConsistentHash `json:"consistentHash,omitempty"`

	// HostOverrideSettings defines the configuration when the load balancer type is
	// set to HostOverride
	//
	// +optional
	HostOverrideSettings *HostOverrideSettings `json:"hostOverrideSettings,omitempty"`

	// SlowStart defines the configuration related to the slow start load balancer policy.
	// If set, during slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently this is only supported for RoundRobin and LeastRequest load balancers
	//
	// +optional
	SlowStart *SlowStart `json:"slowStart,omitempty"`

	// ZoneAware defines the configuration related to the distribution of requests between locality zones.
	//
	// +optional
	// +notImplementedHide
	ZoneAware *ZoneAware `json:"zoneAware,omitempty"`
}

// LoadBalancerType specifies the types of LoadBalancer.
// +kubebuilder:validation:Enum=ConsistentHash;LeastRequest;Random;RoundRobin;HostOverride
type LoadBalancerType string

const (
	// ConsistentHashLoadBalancerType load balancer policy.
	ConsistentHashLoadBalancerType LoadBalancerType = "ConsistentHash"
	// LeastRequestLoadBalancerType load balancer policy.
	LeastRequestLoadBalancerType LoadBalancerType = "LeastRequest"
	// RandomLoadBalancerType load balancer policy.
	RandomLoadBalancerType LoadBalancerType = "Random"
	// RoundRobinLoadBalancerType load balancer policy.
	RoundRobinLoadBalancerType LoadBalancerType = "RoundRobin"
	// HostOverrideLoadBalancerType load balancer policy.
	HostOverrideLoadBalancerType LoadBalancerType = "HostOverride"
)

// ConsistentHash defines the configuration related to the consistent hash
// load balancer policy.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Header' ? has(self.header) : !has(self.header)",message="If consistent hash type is header, the header field must be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Cookie' ? has(self.cookie) : !has(self.cookie)",message="If consistent hash type is cookie, the cookie field must be set."
type ConsistentHash struct {
	// ConsistentHashType defines the type of input to hash on. Valid Type values are
	// "SourceIP",
	// "Header",
	// "Cookie".
	//
	// +unionDiscriminator
	Type ConsistentHashType `json:"type"`

	// Header configures the header hash policy when the consistent hash type is set to Header.
	//
	// +optional
	Header *Header `json:"header,omitempty"`

	// Cookie configures the cookie hash policy when the consistent hash type is set to Cookie.
	//
	// +optional
	Cookie *Cookie `json:"cookie,omitempty"`

	// The table size for consistent hashing, must be prime number limited to 5000011.
	//
	// +kubebuilder:validation:Minimum=2
	// +kubebuilder:validation:Maximum=5000011
	// +kubebuilder:default=65537
	// +optional
	TableSize *uint64 `json:"tableSize,omitempty"`
}

// Header defines the header hashing configuration for consistent hash based
// load balancing.
type Header struct {
	// Name of the header to hash.
	Name string `json:"name"`
}

// Cookie defines the cookie hashing configuration for consistent hash based
// load balancing.
type Cookie struct {
	// Name of the cookie to hash.
	// If this cookie does not exist in the request, Envoy will generate a cookie and set
	// the TTL on the response back to the client based on Layer 4
	// attributes of the backend endpoint, to ensure that these future requests
	// go to the same backend endpoint. Make sure to set the TTL field for this case.
	Name string `json:"name"`
	// TTL of the generated cookie if the cookie is not present. This value sets the
	// Max-Age attribute value.
	//
	// +optional
	TTL *metav1.Duration `json:"ttl,omitempty"`
	// Additional Attributes to set for the generated cookie.
	//
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ConsistentHashType defines the type of input to hash on.
// +kubebuilder:validation:Enum=SourceIP;Header;Cookie
type ConsistentHashType string

const (
	// SourceIPConsistentHashType hashes based on the source IP address.
	SourceIPConsistentHashType ConsistentHashType = "SourceIP"
	// HeaderConsistentHashType hashes based on a request header.
	HeaderConsistentHashType ConsistentHashType = "Header"
	// CookieConsistentHashType hashes based on a cookie.
	CookieConsistentHashType ConsistentHashType = "Cookie"
)

// SlowStart defines the configuration related to the slow start load balancer policy.
type SlowStart struct {
	// Window defines the duration of the warm up period for newly added host.
	// During slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently only supports linear growth of traffic. For additional details,
	// see https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-slowstartconfig
	// +kubebuilder:validation:Required
	Window *metav1.Duration `json:"window"`
	// TODO: Add support for non-linear traffic increases based on user usage.
}

// ZoneAware defines the configuration related to the distribution of requests between locality zones.
type ZoneAware struct {
	// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
	//
	// +optional
	// +notImplementedHide
	PreferLocal *PreferLocalZone `json:"preferLocal,omitempty"`
}

// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
type PreferLocalZone struct {
	// ForceLocalZone defines override configuration for forcing all traffic to stay within the local zone instead of the default behavior
	// which maintains equal distribution among upstream endpoints while sending as much traffic as possible locally.
	//
	// +optional
	// +notImplementedHide
	Force *ForceLocalZone `json:"force,omitempty"`

	// MinEndpointsThreshold is the minimum number of total upstream endpoints across all zones required to enable zone-aware routing.
	//
	// +optional
	// +notImplementedHide
	MinEndpointsThreshold *uint64 `json:"minEndpointsThreshold,omitempty"`
}

// ForceLocalZone defines override configuration for forcing all traffic to stay within the local zone instead of the default behavior
// which maintains equal distribution among upstream endpoints while sending as much traffic as possible locally.
type ForceLocalZone struct {
	// MinEndpointsInZoneThreshold is the minimum number of upstream endpoints in the local zone required to honor the forceLocalZone
	// override. This is useful for protecting zones with fewer endpoints.
	//
	// +optional
	// +notImplementedHide
	MinEndpointsInZoneThreshold *uint32 `json:"minEndpointsInZoneThreshold,omitempty"`
}

// HostOverrideSettings defines the configuration for the Host Override load balancer policy.
// This policy allows endpoint picking to be implemented in downstream HTTP filters.
// It extracts selected override hosts from a list of OverrideHostSource (request headers, metadata, etc.).
// If no valid host in the override host list, then the specified fallback load balancing policy is used.
type HostOverrideSettings struct {
	// OverrideHostSources defines a list of sources to get host addresses from.
	// The host sources are searched in the order specified.
	// The request is forwarded to the first address and subsequent addresses are used for request retries or hedging.
	// Note that if an overridden host address is not present in the current endpoint set, it is skipped and the next found address is used.
	// If there are not enough overridden addresses to satisfy all retry attempts the fallback load balancing policy is used to pick a host.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	OverrideHostSources []OverrideHostSource `json:"overrideHostSources"`

	// FallbackPolicy defines the child LB policy to use in case neither header nor metadata with selected hosts is present.
	// If not specified, defaults to LeastRequest.
	//
	// +optional
	// +kubebuilder:default="LeastRequest"
	FallbackPolicy *LoadBalancerType `json:"fallbackPolicy,omitempty"`
}

// OverrideHostSource defines a source to get override host addresses from.
// +union
//
// +kubebuilder:validation:XValidation:rule="(has(self.header) && !has(self.metadata)) || (!has(self.header) && has(self.metadata))",message="Exactly one of header or metadata must be set."
type OverrideHostSource struct {
	// Header defines the header to get the override host addresses.
	// The header value must specify at least one host in `IP:Port` format or multiple hosts in `IP:Port,IP:Port,...` format.
	// For example `10.0.0.5:8080` or `[2600:4040:5204::1574:24ae]:80`.
	// The IPv6 address is enclosed in square brackets.
	//
	// +optional
	Header *string `json:"header,omitempty"`

	// Metadata defines the metadata key to get the override host addresses from the request dynamic metadata.
	// If set this field then it will take precedence over the header field.
	//
	// +optional
	Metadata *MetadataKey `json:"metadata,omitempty"`
}

// MetadataKey defines the metadata key configuration for host override.
type MetadataKey struct {
	// Key defines the metadata key.
	//
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`

	// Path defines the path within the metadata to extract the host addresses.
	// Each path element represents a key in nested metadata structure.
	//
	// +optional
	Path []MetadataKeyPath `json:"path,omitempty"`
}

// MetadataKeyPath defines a path element in the metadata structure.
type MetadataKeyPath struct {
	// Key defines the key name in the metadata structure.
	//
	// +kubebuilder:validation:MinLength=1
	Key string `json:"key"`
}
