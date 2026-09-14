// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

// Package endpoints converts Kubernetes EndpointSlices into IR destination
// endpoints. It is shared by the Gateway API translator (full translation)
// and the xDS runner's endpoint fast path, so both compute endpoints
// identically.
package endpoints

import (
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	mcsapiv1a1 "sigs.k8s.io/mcs-api/pkg/apis/v1alpha1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

// BackendForSlice returns the backend an EndpointSlice belongs to, derived from
// its service-name labels. A slice carrying both labels belongs to the Service:
// that precedence decides which backend's endpoints a slice contributes to, so
// the full translation and the endpoint fast path must apply it identically.
// ok is false when the slice carries neither label.
func BackendForSlice(slice *discoveryv1.EndpointSlice) (kind, name string, ok bool) {
	if name, ok := slice.Labels[discoveryv1.LabelServiceName]; ok {
		return resource.KindService, name, true
	}
	if name, ok := slice.Labels[mcsapiv1a1.LabelServiceName]; ok {
		return resource.KindServiceImport, name, true
	}
	return "", "", false
}

// EndpointsFromSlices converts the given EndpointSlices into IR destination
// endpoints, keeping only ports matching portName and portProtocol, and
// returns them together with the resulting address type (IP, FQDN, or MIXED
// when the slices disagree).
func EndpointsFromSlices(endpointSlices []*discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol) ([]*ir.DestinationEndpoint, *ir.DestinationAddressType) {
	var (
		dstEndpoints []*ir.DestinationEndpoint
		dstAddrType  *ir.DestinationAddressType
	)

	addrTypeMap := make(map[ir.DestinationAddressType]int)
	for _, endpointSlice := range endpointSlices {
		if endpointSlice.AddressType == discoveryv1.AddressTypeFQDN {
			addrTypeMap[ir.FQDN]++
		} else {
			addrTypeMap[ir.IP]++
		}
		endpoints := endpointsFromSlice(endpointSlice, portName, portProtocol)
		dstEndpoints = append(dstEndpoints, endpoints...)
	}

	for addrTypeState, addrTypeCounts := range addrTypeMap {
		if addrTypeCounts == len(endpointSlices) {
			dstAddrType = new(addrTypeState)
			break
		}
	}

	if len(addrTypeMap) > 0 && dstAddrType == nil {
		dstAddrType = new(ir.MIXED)
	}

	return dstEndpoints, dstAddrType
}

func endpointsFromSlice(endpointSlice *discoveryv1.EndpointSlice, portName string, portProtocol corev1.Protocol) []*ir.DestinationEndpoint {
	var endpoints []*ir.DestinationEndpoint
	for _, endpoint := range endpointSlice.Endpoints {
		for _, endpointPort := range endpointSlice.Ports {
			// Check if the endpoint port matches the service port
			if *endpointPort.Name != portName || *endpointPort.Protocol != portProtocol {
				continue
			}
			conditions := endpoint.Conditions

			// Unknown Serving/Terminating (nil) should fall-back to Ready, see https://pkg.go.dev/k8s.io/api/discovery/v1#EndpointConditions
			// So drain the endpoint if:
			// 1. Both `Terminating` and `Serving` are != null, and either `Terminating=true` or `Serving=false`
			// 2. Or `Ready=false`
			var draining bool
			if conditions.Serving != nil && conditions.Terminating != nil {
				draining = *conditions.Terminating || !*conditions.Serving
			} else {
				draining = conditions.Ready != nil && !*conditions.Ready
			}

			for _, address := range endpoint.Addresses {
				ep := ir.NewDestEndpoint(nil, address, uint32(*endpointPort.Port), draining, endpoint.Zone)
				endpoints = append(endpoints, ep)
			}

		}
	}

	return endpoints
}
