// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"net"
	"sync"

	corev1 "k8s.io/api/core/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/status"
)

type nodeDetails struct {
	name      string
	addresses status.NodeAddresses
}

// kubernetesProviderStore holds cached information for the kubernetes provider.
type kubernetesProviderStore struct {
	// nodes holds information required for updating Gateway status with the Node
	// addresses, in case the Gateway is exposed on every Node of the cluster, using
	// Service of type NodePort.
	nodes map[string]nodeDetails
	mu    sync.Mutex
}

func newProviderStore() *kubernetesProviderStore {
	return &kubernetesProviderStore{
		nodes: make(map[string]nodeDetails),
	}
}

func (p *kubernetesProviderStore) addNode(n *corev1.Node) {
	details := nodeDetails{name: n.Name}

	var internalIPs, externalIPs status.NodeAddresses
	for _, addr := range n.Status.Addresses {
		var addrs *status.NodeAddresses
		switch addr.Type {
		case corev1.NodeExternalIP:
			addrs = &externalIPs
		case corev1.NodeInternalIP:
			addrs = &internalIPs
		default:
			continue
		}
		if net.ParseIP(addr.Address).To4() != nil {
			addrs.IPv4 = append(addrs.IPv4, addr.Address)
		} else {
			addrs.IPv6 = append(addrs.IPv6, addr.Address)
		}
	}

	// In certain scenarios (like in local KinD clusters), the Node
	// externalIP is not provided, in that case we default back
	// to the internalIP of the Node.
	if len(externalIPs.IPv4) > 0 || len(externalIPs.IPv6) > 0 {
		details.addresses = externalIPs
	} else if len(internalIPs.IPv4) > 0 || len(internalIPs.IPv6) > 0 {
		details.addresses = internalIPs
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.nodes[n.Name] = details
}

func (p *kubernetesProviderStore) removeNode(n *corev1.Node) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.nodes, n.Name)
}

func (p *kubernetesProviderStore) listNodeAddresses() status.NodeAddresses {
	p.mu.Lock()
	defer p.mu.Unlock()
	addrs := status.NodeAddresses{}
	for _, n := range p.nodes {
		addrs.IPv4 = append(addrs.IPv4, n.addresses.IPv4...)
		addrs.IPv6 = append(addrs.IPv6, n.addresses.IPv6...)
	}
	return addrs
}
