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

type Port struct {
	Name     string
	Protocol corev1.Protocol
}

// BackendEndpoints contains immutable Kubernetes endpoint data for one backend revision.
// EndpointsByPort is derived from EndpointSlices; both represent this revision.
// Consumers may share this data without copying it, but must not modify it.
type BackendEndpoints struct {
	Backend        ir.ServiceKey
	Revision       uint64
	EndpointSlices []*discoveryv1.EndpointSlice
	// AddressType is IP or FQDN when every slice has that type, MIXED when they
	// differ, and nil when there are no slices. It decides the cluster type (EDS
	// for IP, DNS for FQDN) and whether the backend is valid, so the fast path
	// leaves an update that changes it to full translation.
	AddressType     *ir.DestinationAddressType
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
	return in.EndpointsByPort[Port{Name: portName, Protocol: protocol}], in.AddressType, in.sourceForPort(portName, protocol)
}

func NewBackendEndpoints(backend ir.ServiceKey, slices []*discoveryv1.EndpointSlice) *BackendEndpoints {
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
