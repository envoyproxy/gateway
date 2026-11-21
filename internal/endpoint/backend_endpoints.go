// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	"sort"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

// Port identifies a Service port by name and protocol.
type Port struct {
	Name     string
	Protocol corev1.Protocol
}

// BackendEndpoints contains immutable Kubernetes endpoint data for one backend revision.
// EndpointsByPort is derived from EndpointSlices; both represent this revision.
type BackendEndpoints struct {
	Backend         ir.ServiceKey
	Revision        uint64
	EndpointSlices  []*discoveryv1.EndpointSlice
	EndpointsByPort map[Port][]*ir.DestinationEndpoint
}

// DeepCopy shares immutable data across the provider resource boundary.
func (in *BackendEndpoints) DeepCopy() *BackendEndpoints { return in }

func (in *BackendEndpoints) DeepCopyInto(out *BackendEndpoints) { *out = *in }

// sourceForPort returns where one port's endpoints come from: this backend, the
// port and this revision.
func (in *BackendEndpoints) sourceForPort(portName string, protocol corev1.Protocol) *ir.EndpointSource {
	return &ir.EndpointSource{Service: in.Backend, Revision: in.Revision, PortName: portName, Protocol: string(protocol)}
}

// ForPort returns this revision's endpoints, address type, and source for one port.
func (in *BackendEndpoints) ForPort(portName string, protocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType, *ir.EndpointSource) {
	return in.EndpointsByPort[Port{Name: portName, Protocol: protocol}], addressType(in.EndpointSlices), in.sourceForPort(portName, protocol)
}

func NewBackendEndpoints(backend ir.ServiceKey, slices []*discoveryv1.EndpointSlice) *BackendEndpoints {
	// The informer cache lists slices in no fixed order. Sort them the way
	// Resources.Sort does, so the same endpoints always build the same assignment
	// and an unchanged one isn't pushed to Envoy again.
	sort.SliceStable(slices, func(i, j int) bool {
		if !slices[i].CreationTimestamp.Equal(&slices[j].CreationTimestamp) {
			return slices[i].CreationTimestamp.Before(&slices[j].CreationTimestamp)
		}
		return slices[i].Namespace+"/"+slices[i].Name < slices[j].Namespace+"/"+slices[j].Name
	})

	// Build the endpoints for each service port.
	backendEndpoints := &BackendEndpoints{Backend: backend, EndpointSlices: slices, EndpointsByPort: make(map[Port][]*ir.DestinationEndpoint)}
	for _, slice := range slices {
		for _, port := range slice.Ports {
			key := Port{Name: portName(port), Protocol: portProtocol(port)}
			if _, found := backendEndpoints.EndpointsByPort[key]; !found {
				backendEndpoints.EndpointsByPort[key], _ = FromSlices(slices, key.Name, key.Protocol)
			}
		}
	}
	return backendEndpoints
}
