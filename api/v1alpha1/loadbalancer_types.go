// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

import gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

// LoadBalancer defines the load balancer policy to be applied.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'ConsistentHash' ? has(self.consistentHash) : !has(self.consistentHash)",message="If LoadBalancer type is consistentHash, consistentHash field needs to be set."
// +kubebuilder:validation:XValidation:rule="self.type in ['Random', 'ConsistentHash'] ? !has(self.slowStart) : true ",message="Currently SlowStart is only supported for RoundRobin and LeastRequest load balancers."
// +kubebuilder:validation:XValidation:rule="self.type == 'ConsistentHash' && has(self.zoneAware) ? !has(self.zoneAware.preferLocal) : true",message="PreferLocal zone-aware routing is not supported for ConsistentHash load balancers. Use weightedZones instead."
// +kubebuilder:validation:XValidation:rule="has(self.zoneAware) ? !(has(self.zoneAware.preferLocal) && has(self.zoneAware.weightedZones)) : true",message="ZoneAware PreferLocal and WeightedZones cannot be specified together."
type LoadBalancer struct {
	// Type decides the type of Load Balancer policy.
	// Valid LoadBalancerType values are
	// "ConsistentHash",
	// "LeastRequest",
	// "Random",
	// "RoundRobin".
	//
	// +unionDiscriminator
	Type LoadBalancerType `json:"type"`
	// ConsistentHash defines the configuration when the load balancer type is
	// set to ConsistentHash
	//
	// +optional
	ConsistentHash *ConsistentHash `json:"consistentHash,omitempty"`

	// EndpointOverride defines the configuration for endpoint override.
	// When specified, the load balancer will attempt to route requests to endpoints
	// based on the override information extracted from request headers or metadata.
	//  If the override endpoints are not available, the configured load balancer policy will be used as fallback.
	//
	// +optional
	EndpointOverride *EndpointOverride `json:"endpointOverride,omitempty"`

	// SlowStart defines the configuration related to the slow start load balancer policy.
	// If set, during slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently this is only supported for RoundRobin and LeastRequest load balancers
	//
	// +optional
	SlowStart *SlowStart `json:"slowStart,omitempty"`

	// ZoneAware defines the configuration related to the distribution of requests between locality zones.
	//
	// +optional
	ZoneAware *ZoneAware `json:"zoneAware,omitempty"`
}

// LoadBalancerType specifies the types of LoadBalancer.
// +kubebuilder:validation:Enum=ConsistentHash;LeastRequest;Random;RoundRobin
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
)

// ConsistentHash defines the configuration related to the consistent hash
// load balancer policy.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'Header' ? has(self.header) : !has(self.header)",message="If consistent hash type is header, the header field must be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Headers' ? has(self.headers) : !has(self.headers)",message="If consistent hash type is headers, the headers field must be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'Cookie' ? has(self.cookie) : !has(self.cookie)",message="If consistent hash type is cookie, the cookie field must be set."
// +kubebuilder:validation:XValidation:rule="self.type == 'QueryParams' ? has(self.queryParams) : !has(self.queryParams)",message="If consistent hash type is queryParams, the queryParams field must be set."
type ConsistentHash struct {
	// ConsistentHashType defines the type of input to hash on. Valid Type values are
	// "SourceIP",
	// "Header",
	// "Headers",
	// "Cookie".
	// "QueryParams".
	//
	// +unionDiscriminator
	Type ConsistentHashType `json:"type"`

	// Header configures the header hash policy when the consistent hash type is set to Header.
	//
	// Deprecated: use Headers instead
	// +optional
	Header *Header `json:"header,omitempty"`

	// Headers configures the header hash policy for each header, when the consistent hash type is set to Headers.
	//
	// +optional
	Headers []*Header `json:"headers,omitempty"`

	// Cookie configures the cookie hash policy when the consistent hash type is set to Cookie.
	//
	// +optional
	Cookie *Cookie `json:"cookie,omitempty"`

	// QueryParams configures the query parameter hash policy when the consistent hash type is set to QueryParams.
	//
	// +optional
	QueryParams []*QueryParam `json:"queryParams,omitempty"`

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

// QueryParam defines the query parameter name hashing configuration for consistent hash based
// load balancing.
type QueryParam struct {
	// Name of the query param to hash.
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
	TTL *gwapiv1.Duration `json:"ttl,omitempty"`
	// Additional Attributes to set for the generated cookie.
	//
	// +optional
	Attributes map[string]string `json:"attributes,omitempty"`
}

// ConsistentHashType defines the type of input to hash on.
// +kubebuilder:validation:Enum=SourceIP;Header;Headers;Cookie;QueryParams
type ConsistentHashType string

const (
	// SourceIPConsistentHashType hashes based on the source IP address.
	SourceIPConsistentHashType ConsistentHashType = "SourceIP"
	// HeaderConsistentHashType hashes based on a request header.
	//
	// Deprecated: use HeadersConsistentHashType instead
	HeaderConsistentHashType ConsistentHashType = "Header"
	// HeadersConsistentHashType hashes based on multiple request headers.
	HeadersConsistentHashType ConsistentHashType = "Headers"
	// CookieConsistentHashType hashes based on a cookie.
	CookieConsistentHashType ConsistentHashType = "Cookie"
	// QueryParamsConsistentHashType hashes based on a multiple query parameter.
	QueryParamsConsistentHashType ConsistentHashType = "QueryParams"
)

// SlowStart defines the configuration related to the slow start load balancer policy.
type SlowStart struct {
	// Window defines the duration of the warm up period for newly added host.
	// During slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently only supports linear growth of traffic. For additional details,
	// see https://www.envoyproxy.io/docs/envoy/latest/api-v3/config/cluster/v3/cluster.proto#config-cluster-v3-cluster-slowstartconfig
	// +kubebuilder:validation:Required
	Window *gwapiv1.Duration `json:"window"`
	// TODO: Add support for non-linear traffic increases based on user usage.
}

// ZoneAware defines the configuration related to the distribution of requests between locality zones.
//
// +kubebuilder:validation:XValidation:rule="!has(self.weightedZones) || self.weightedZones.all(z, self.weightedZones.exists_one(z2, z2.zone == z.zone))",message="Duplicate zone names are not allowed in weightedZones."
type ZoneAware struct {
	// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
	//
	// +optional
	PreferLocal *PreferLocalZone `json:"preferLocal,omitempty"`

	// WeightedZones configures weight-based traffic distribution across locality zones.
	// Traffic is distributed proportionally based on the sum of all zone weights.
	//
	// +optional
	WeightedZones []WeightedZoneConfig `json:"weightedZones,omitempty"`
}

// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
type PreferLocalZone struct {
	// ForceLocalZone defines override configuration for forcing all traffic to stay within the local zone instead of the default behavior
	// which maintains equal distribution among upstream endpoints while sending as much traffic as possible locally.
	//
	// +optional
	Force *ForceLocalZone `json:"force,omitempty"`

	// MinEndpointsThreshold is the minimum number of total upstream endpoints across all zones required to enable zone-aware routing.
	//
	// +optional
	MinEndpointsThreshold *uint64 `json:"minEndpointsThreshold,omitempty"`

	// Configures percentage of requests that will be considered for zone aware routing if zone aware routing is configured. If not specified, Envoy defaults to 100%.
	//
	// +kubebuilder:validation:Minimum=0
	// +kubebuilder:validation:Maximum=100
	// +optional
	PercentageEnabled *uint32 `json:"percentageEnabled,omitempty"`
}

// ForceLocalZone defines override configuration for forcing all traffic to stay within the local zone instead of the default behavior
// which maintains equal distribution among upstream endpoints while sending as much traffic as possible locally.
type ForceLocalZone struct {
	// MinEndpointsInZoneThreshold is the minimum number of upstream endpoints in the local zone required to honor the forceLocalZone
	// override. This is useful for protecting zones with fewer endpoints.
	//
	// +optional
	MinEndpointsInZoneThreshold *uint32 `json:"minEndpointsInZoneThreshold,omitempty"`
}

// WeightedZoneConfig defines the weight for a specific locality zone.
type WeightedZoneConfig struct {
	// Zone specifies the topology zone this weight applies to.
	// The value should match the topology.kubernetes.io/zone label
	// of the nodes where endpoints are running.
	// Zones not listed in the configuration receive a default weight of 1.
	Zone string `json:"zone"`

	// Weight defines the weight for this locality.
	// Higher values receive more traffic. The actual traffic distribution
	// is proportional to this value relative to other localities.
	Weight uint32 `json:"weight"`
}

// EndpointOverride defines the configuration for endpoint override.
// This allows endpoint picking to be implemented based on request headers or metadata.
// It extracts selected override endpoints from the specified sources (request headers, metadata, etc.).
// If no valid endpoint in the override list, then the configured load balancing policy is used as fallback.
type EndpointOverride struct {
	// ExtractFrom defines the sources to extract endpoint override information from.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=10
	ExtractFrom []EndpointOverrideExtractFrom `json:"extractFrom"`
}

// EndpointOverrideExtractFrom defines a source to extract endpoint override information from.
type EndpointOverrideExtractFrom struct {
	// Header defines the header to get the override endpoint addresses.
	// The header value must specify at least one endpoint in `IP:Port` format or multiple endpoints in `IP:Port,IP:Port,...` format.
	// For example `10.0.0.5:8080` or `[2600:4040:5204::1574:24ae]:80`.
	// The IPv6 address is enclosed in square brackets.
	//
	// +optional
	Header *string `json:"header,omitempty"`
}
