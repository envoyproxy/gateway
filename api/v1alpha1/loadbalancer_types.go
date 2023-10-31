// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package v1alpha1

// LoadBalancer defines the load balancer policy to be applied.
// +union
//
// +kubebuilder:validation:XValidation:rule="self.type == 'ConsistentHash' ? has(self.consistentHash) : !has(self.consistentHash)",message="If LoadBalancer type is consistentHash, consistentHash field needs to be set."
type LoadBalancer struct {
	// Type decides the type of Load Balancer policy.
	// Valid RateLimitType values are
	// "ConsistentHash",
	// "LeastRequest",
	// "Random",
	// "RoundRobin",
	//
	// +unionDiscriminator
	Type LoadBalancerType `json:"type"`
	// ConsistentHash defines the configuration when the load balancer type is
	// set to ConsistentHash
	//
	// +optional
	ConsistentHash *ConsistentHash `json:"consistentHash,omitempty"`
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
// load balancer policy
type ConsistentHash struct {
	Type ConsistentHashType `json:"type"`
}

// ConsistentHashType defines the type of input to hash on.
// +kubebuilder:validation:Enum=SourceIP
type ConsistentHashType string

const (
	// SourceIPConsistentHashType hashes based on the source IP address.
	SourceIPConsistentHashType ConsistentHashType = "SourceIP"
)
