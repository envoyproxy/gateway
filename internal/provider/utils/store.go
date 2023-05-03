// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package utils

import (
	"sync"

	corev1 "k8s.io/api/core/v1"
)

type nodeDetails struct {
	nodeName   string
	externalIP string
}

type providerStore struct {
	// nodes holds information required for updating Gateway status with the Node
	// addresses, in case the Gateway is exposed by a Service of type NodePort.
	nodes map[string]nodeDetails
}

var providerStoreInstance *providerStore
var o sync.Once

func ProviderStore() *providerStore {
	o.Do(func() {
		providerStoreInstance = &providerStore{
			nodes: make(map[string]nodeDetails),
		}
	})
	return providerStoreInstance
}

func (p *providerStore) AddNode(n *corev1.Node) {
	details := nodeDetails{nodeName: n.Name}
	var internalIP, externalIP string
	for _, addr := range n.Status.Addresses {
		if addr.Type == corev1.NodeExternalIP {
			externalIP = addr.Address
		}
		if addr.Type == corev1.NodeInternalIP {
			internalIP = addr.Address
		}
	}
	if externalIP != "" {
		details.externalIP = externalIP
		return
	}
	details.externalIP = internalIP
	p.nodes[n.Name] = details
}

func (p *providerStore) RemoveNode(n *corev1.Node) {
	delete(p.nodes, n.Name)
}

func (p *providerStore) ListNodeAddresses() []string {
	addrs := []string{}
	for _, n := range p.nodes {
		addrs = append(addrs, n.externalIP)
	}
	return addrs
}
