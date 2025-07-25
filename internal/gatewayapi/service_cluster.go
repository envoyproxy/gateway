// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/utils/ptr"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
)

func (t *Translator) ProcessServiceCluster(acceptedGateways []*GatewayContext, resources *resource.Resources, xdsIR resource.XdsIRMap) {
	for _, g := range acceptedGateways {
		if g == nil || g.Gateway == nil {
			continue
		}

		irKey := t.getIRKey(g.Gateway)

		if xdsIR[irKey].ProxyServiceCluster != nil {
			continue
		}

		var svcName string
		if t.MergeGateways {
			svcName = t.expectedServiceName(string(t.GatewayClassName))
		} else {
			svcName = t.expectedServiceName(fmt.Sprintf("%s/%s", g.Namespace, g.Name))
		}

		svc := resources.GetService(t.ControllerNamespace, svcName)
		if svc == nil {
			return
		}

		ds := t.processServiceClusterDestinationSetting(svc.Name, svc, resources)
		ds.IPFamily = getServiceIPFamily(svc)

		if xdsIR[irKey].GlobalResources == nil {
			xdsIR[irKey].GlobalResources = &ir.GlobalResources{}
		}
		xdsIR[irKey].ProxyServiceCluster = &ir.ProxyServiceCluster{
			Name:        svc.Name,
			Destination: ds,
		}

		if t.MergeGateways {
			return
		}
	}
}

// expectedServiceName returns expected Kubernetes Service name referring to the proxy Service Cluster.
func (t *Translator) expectedServiceName(name string) string {
	hashedName := utils.GetHashedName(name, 48)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, hashedName)
}

func (t *Translator) processServiceClusterDestinationSetting(
	name string,
	service *corev1.Service,
	resources *resource.Resources,
) *ir.DestinationSetting {
	var (
		endpoints []*ir.DestinationEndpoint
		addrType  *ir.DestinationAddressType
	)

	endpointSlices := resources.GetEndpointSlicesForBackend(service.Namespace, service.Name, resource.KindService)
	endpoints, addrType = getIREndpointsFromEndpointSlices(endpointSlices, service.Spec.Ports[0].Name, service.Spec.Ports[0].Protocol)

	return &ir.DestinationSetting{
		Name:        name,
		Protocol:    ir.HTTP,
		Endpoints:   endpoints,
		AddressType: addrType,
		// Use Zone Aware Lb so locality info is injected for endpoints
		PreferLocal: &ir.PreferLocalZone{MinEndpointsThreshold: ptr.To[uint64](1)},
	}
}
