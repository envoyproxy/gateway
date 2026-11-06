// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	"reflect"
	"sort"

	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

type Key struct {
	Kind, Namespace, Name string
}

func BackendKey(backend ir.EndpointBackend) Key {
	return Key{Kind: backend.Kind, Namespace: backend.Namespace, Name: backend.Name}
}

type Port struct {
	Name     string
	Protocol corev1.Protocol
}

// BackendEndpoints contains immutable Kubernetes endpoint data for one backend revision.
// EndpointsByPort is derived from EndpointSlices; both represent this revision.
// Consumers may share this data without copying it, but must not modify it.
type BackendEndpoints struct {
	Backend         ir.EndpointBackend
	Revision        uint64
	EndpointSlices  []*discoveryv1.EndpointSlice
	AddressType     *ir.DestinationAddressType
	EndpointsByPort map[Port][]*ir.DestinationEndpoint
}

// DeepCopy shares immutable data across the provider resource boundary.
func (in *BackendEndpoints) DeepCopy() *BackendEndpoints { return in }

func (in *BackendEndpoints) DeepCopyInto(out *BackendEndpoints) { *out = *in }

func (in *BackendEndpoints) Reference(portName string, protocol corev1.Protocol) *ir.EndpointSource {
	return &ir.EndpointSource{Backend: in.Backend, Revision: in.Revision, PortName: portName, Protocol: string(protocol)}
}

// ForPort returns this revision's endpoints, address type, and source for one port.
func (in *BackendEndpoints) ForPort(portName string, protocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType, *ir.EndpointSource) {
	return in.EndpointsByPort[Port{Name: portName, Protocol: protocol}], in.AddressType, in.Reference(portName, protocol)
}

func NewBackendEndpoints(backend ir.EndpointBackend, slices []*discoveryv1.EndpointSlice) *BackendEndpoints {
	slices = append([]*discoveryv1.EndpointSlice(nil), slices...)
	sort.SliceStable(slices, func(i, j int) bool {
		if !slices[i].CreationTimestamp.Equal(&slices[j].CreationTimestamp) {
			return slices[i].CreationTimestamp.Before(&slices[j].CreationTimestamp)
		}
		return slices[i].Namespace+"/"+slices[i].Name < slices[j].Namespace+"/"+slices[j].Name
	})
	backendEndpoints := &BackendEndpoints{Backend: backend, EndpointSlices: slices, EndpointsByPort: make(map[Port][]*ir.DestinationEndpoint)}
	for _, slice := range slices {
		for _, port := range slice.Ports {
			key := Port{Name: portName(port), Protocol: portProtocol(port)}
			if _, found := backendEndpoints.EndpointsByPort[key]; !found {
				backendEndpoints.EndpointsByPort[key], _ = FromSlices(slices, key.Name, key.Protocol)
			}
		}
	}
	backendEndpoints.AddressType = addressType(slices)
	return backendEndpoints
}

// sameEndpointData compares backend identity and derived endpoints. Revisions
// and raw EndpointSlice metadata do not determine whether the endpoint data changed.
func sameEndpointData(previous, current *BackendEndpoints) bool {
	return previous != nil && previous.Backend == current.Backend &&
		reflect.DeepEqual(previous.AddressType, current.AddressType) &&
		reflect.DeepEqual(previous.EndpointsByPort, current.EndpointsByPort)
}
