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
// +kubebuilder:validation:XValidation:rule="self.type in ['Random', 'ConsistentHash'] ? !has(self.slowStart) : true ",message="Currently SlowStart is only supported for RoundRobin and LeastRequest load balancers."
// +kubebuilder:validation:XValidation:rule="self.type == 'ConsistentHash' ? (!has(self.requestDistribution) || !has(self.requestDistribution.preferLocalZone)) : true ",message="ConsistentHash load balancer only supports weightedLocality in requestDistribution; preferLocalZone is not allowed."
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

	// SlowStart defines the configuration related to the slow start load balancer policy.
	// If set, during slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently this is only supported for RoundRobin and LeastRequest load balancers
	//
	// +optional
	SlowStart *SlowStart `json:"slowStart,omitempty"`

	// RequestDistribution defines the configuration related to the distribution of requests between localities.
	//
	// +optional
	// +notImplementedHide
	RequestDistribution *RequestDistribution `json:"requestDistribution,omitempty"`
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

// RequestDistribution defines the configuration related to the distribution of requests between localities.
// Exactly one of PreferLocalZone or WeightedLocality must be specified.
//
// +kubebuilder:validation:XValidation:rule="!(has(self.preferLocalZone) && has(self.weightedLocality))",message="only one of preferLocalZone or weightedLocality may be specified."
// +kubebuilder:validation:XValidation:rule="!(!has(self.preferLocalZone) && !has(self.weightedLocality))",message="one of preferLocalZone or weightedLocality must be specified."
type RequestDistribution struct {
	// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
	//
	// +optional
	// +notImplementedHide
	PreferLocalZone *PreferLocalZone `json:"preferLocalZone,omitempty"`

	// WeightedLocality configures explicit weights for each locality.
	//
	// +optional
	// +notImplementedHide
	WeightedLocality *WeightedLocality `json:"weightedLocality,omitempty"`
}

// PreferLocalZone configures zone-aware routing to prefer sending traffic to the local locality zone.
type PreferLocalZone struct {
	// ForceLocalZone defines override configuration for forcing all traffic to stay local vs Envoy default behavior
	// which maintains equal distribution among upstreams while sending as much traffic as possible locally.
	//
	// +optional
	// +notImplementedHide
	ForceLocalZone *ForceLocalZone `json:"forceLocalZone,omitempty"`

	// MinClusterSize is the minimum number of total upstream hosts across all zones required to enable zone-aware routing.
	//
	// +optional
	// +notImplementedHide
	MinClusterSize *uint64 `json:"minClusterSize,omitempty"`
}

// ForceLocalZone defines override configuration for forcing all traffic to stay local vs Envoy default behavior
// which maintains equal distribution among upstreams while sending as much traffic as possible locally.
type ForceLocalZone struct {
	// Enabled causes Envoy to route all requests to the local zone if there are at least "minZoneSize" healthy hosts
	// available.
	//
	// +optional
	// +notImplementedHide
	Enabled *bool `json:"enabled"`

	// MinZoneSize is the minimum number of upstream hosts in the local zone required to honor the forceLocalZone
	// override. Defaults to 1 if not specified.
	//
	// +optional
	// +notImplementedHide
	MinZoneSize *uint32 `json:"minZoneSize,omitempty"`
}

// WeightedLocality defines explicit traffic weights per locality.
type WeightedLocality struct {
	// Weights specifies a list of localities and their corresponding traffic weights.
	//
	// +notImplementedHide
	Weights []LocalityWeights `json:"weights"`
}

// LocalityWeights associates a locality with a traffic weight.
type LocalityWeights struct {
	// Locality defines locality information for which the weight applies.
	//
	// +notImplementedHide
	Locality Locality `json:"locality"`

	// Weight is the relative weight for traffic distribution to the specified locality.
	//
	// +notImplementedHide
	Weight uint32 `json:"weight"`
}

// Locality specifies the details of a particular locality. Currently only Zone is supported.
type Locality struct {
	// Zone is the name of the locality zone (e.g. topology.kubernetes.io/zone).
	//
	// +notImplementedHide
	Zone string `json:"zone"`
}
