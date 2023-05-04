// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	corev1 "k8s.io/api/core/v1"
)

type nodeDetails struct {
	name    string
	address string
}

// kubernetesProviderStore holds cached information for the kubernetes provider.
type kubernetesProviderStore struct {
	// nodes holds information required for updating Gateway status with the Node
	// addresses, in case the Gateway is exposed on every Node of the cluster, using
	// Service of type NodePort.
	nodes map[string]nodeDetails
}

func newProviderStore() *kubernetesProviderStore {
	return &kubernetesProviderStore{
		nodes: make(map[string]nodeDetails),
	}
}

func (p *kubernetesProviderStore) addNode(n *corev1.Node) {
	details := nodeDetails{name: n.Name}

	var internalIP, externalIP string
	for _, addr := range n.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
	}

	// In certain scenarios (like in local KinD clusters), the Node
	// externalIP is not provided, in that case we default back
	// to the internalIP of the Node.
	if externalIP != "" {
		details.address = externalIP
	} else if internalIP != "" {
		details.address = internalIP
	}
	p.nodes[n.Name] = details
}

func (p *kubernetesProviderStore) removeNode(n *corev1.Node) {
	delete(p.nodes, n.Name)
}

func (p *kubernetesProviderStore) listNodeAddresses() []string {
	addrs := []string{}
	for _, n := range p.nodes {
		if n.address != "" {
			addrs = append(addrs, n.address)
		}
	}
	return addrs
}
