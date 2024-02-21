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
type LoadBalancer struct {
	// Type decides the type of Load Balancer policy.
	// Valid LoadBalancerType values are
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

	// SlowStart defines the configuration related to the slow start load balancer policy.
	// If set, during slow start window, traffic sent to the newly added hosts will gradually increase.
	// Currently this is only supported for RoundRobin and LeastRequest load balancers
	//
	// +optional
	SlowStart *SlowStart `json:"slowStart,omitempty"`
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
