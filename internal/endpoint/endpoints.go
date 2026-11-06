// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package endpoint

import (
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"

	"github.com/envoyproxy/gateway/internal/ir"
)

// FromSlices is the shared full/fast endpoint conversion. Callers supply slices
// in provider order; nil ports cannot identify a usable endpoint.
func FromSlices(slices []*discoveryv1.EndpointSlice, name string, protocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType) {
	var capacity int
	for _, slice := range slices {
		capacity += len(slice.Endpoints)
	}
	endpoints := make([]*ir.DestinationEndpoint, 0, capacity)
	for _, slice := range slices {
		endpoints = append(endpoints, FromSlice(slice, name, protocol)...)
	}
	if len(endpoints) == 0 {
		endpoints = nil
	}
	return endpoints, addressType(slices)
}

func FromSlice(slice *discoveryv1.EndpointSlice, name string, protocol corev1.Protocol) []*ir.DestinationEndpoint {
	var endpoints []*ir.DestinationEndpoint
	for _, endpoint := range slice.Endpoints {
		for _, port := range slice.Ports {
			if portName(port) != name || portProtocol(port) != protocol || port.Port == nil {
				continue
			}
			conditions := endpoint.Conditions
			draining := conditions.Ready != nil && !*conditions.Ready
			if conditions.Serving != nil && conditions.Terminating != nil {
				draining = *conditions.Terminating || !*conditions.Serving
			}
			for _, address := range endpoint.Addresses {
				endpoints = append(endpoints, ir.NewDestEndpoint(nil, address, uint32(*port.Port), draining, endpoint.Zone))
			}
		}
	}
	return endpoints
}

func portName(port discoveryv1.EndpointPort) string {
	if port.Name == nil {
		return ""
	}
	return *port.Name
}

func portProtocol(port discoveryv1.EndpointPort) corev1.Protocol {
	if port.Protocol == nil {
		return corev1.ProtocolTCP
	}
	return *port.Protocol
}

func addressType(slices []*discoveryv1.EndpointSlice) *ir.DestinationAddressType {
	var addressType *ir.DestinationAddressType
	for _, slice := range slices {
		kind := ir.IP
		if slice.AddressType == discoveryv1.AddressTypeFQDN {
			kind = ir.FQDN
		}
		if addressType == nil {
			addressType = new(kind)
		} else if *addressType != kind {
			addressType = new(ir.MIXED)
		}
	}
	return addressType
}
