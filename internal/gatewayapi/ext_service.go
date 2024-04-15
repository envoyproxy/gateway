// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"errors"
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
	egv1a2 "sigs.k8s.io/gateway-api/apis/v1alpha2"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// TODO: zhaohuabing combine this function with the one in the route translator
func (t *Translator) processExtServiceDestination(
	backendRef *gwapiv1.BackendObjectReference,
	policyNamespacedName types.NamespacedName,
	policyKind string,
	protocol ir.AppProtocol,
	resources *Resources) (*ir.DestinationSetting, error) {
	var (
		endpoints   []*ir.DestinationEndpoint
		addrType    *ir.DestinationAddressType
		servicePort v1.ServicePort
		backendTLS  *ir.TLSUpstreamConfig
	)

	serviceNamespace := NamespaceDerefOr(backendRef.Namespace, policyNamespacedName.Namespace)
	service := resources.GetService(serviceNamespace, string(backendRef.Name))
	for _, port := range service.Spec.Ports {
		if port.Port == int32(*backendRef.Port) {
			servicePort = port
			break
		}
	}

	if servicePort.AppProtocol != nil &&
		*servicePort.AppProtocol == "kubernetes.io/h2c" {
		protocol = ir.HTTP2
	}

	// Route to endpoints by default
	if !t.EndpointRoutingDisabled {
		endpointSlices := resources.GetEndpointSlicesForBackend(
			serviceNamespace, string(backendRef.Name), KindService)
		endpoints, addrType = getIREndpointsFromEndpointSlices(
			endpointSlices, servicePort.Name, servicePort.Protocol)
	} else {
		// Fall back to Service ClusterIP routing
		ep := ir.NewDestEndpoint(
			service.Spec.ClusterIP,
			uint32(*backendRef.Port))
		endpoints = append(endpoints, ep)
	}

	// TODO: support mixed endpointslice address type for the same backendRef
	if !t.EndpointRoutingDisabled && addrType != nil && *addrType == ir.MIXED {
		return nil, errors.New(
			"mixed endpointslice address type for the same backendRef is not supported")
	}

	backendTLS = t.processBackendTLSPolicy(
		*backendRef,
		serviceNamespace,
		// Gateway is not the appropriate parent reference here because the owner
		// of the BackendRef is the policy, and there is no hierarchy
		// relationship between the policy and a gateway.
		// The owner policy of the BackendRef is used as the parent reference here.
		egv1a2.ParentReference{
			Group:     ptr.To(gwapiv1.Group(egv1a1.GroupName)),
			Kind:      ptr.To(gwapiv1.Kind(policyKind)),
			Namespace: ptr.To(gwapiv1.Namespace(policyNamespacedName.Namespace)),
			Name:      gwapiv1.ObjectName(policyNamespacedName.Name),
		},
		resources)

	return &ir.DestinationSetting{
		Weight:      ptr.To(uint32(1)),
		Protocol:    protocol,
		Endpoints:   endpoints,
		AddressType: addrType,
		TLS:         backendTLS,
	}, nil
}

// TODO: also refer to extension type, as WASM may also introduce destinations
func irIndexedExtServiceDestinationName(policyNamespacedName types.NamespacedName, policyKind string, idx int) string {
	return strings.ToLower(fmt.Sprintf(
		"%s/%s/%s/%d",
		policyKind,
		policyNamespacedName.Namespace,
		policyNamespacedName.Name,
		idx))
}
