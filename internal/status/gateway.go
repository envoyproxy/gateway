// Copyright 2022 Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	gwapiv1b1 "sigs.k8s.io/gateway-api/apis/v1beta1"

	"github.com/envoyproxy/gateway/internal/gatewayapi"
)

// UpdateGatewayScheduledCondition updates the status condition for the provided Gateway based on the scheduled state.
func UpdateGatewayStatusScheduledCondition(gw *gwapiv1b1.Gateway, scheduled bool) *gwapiv1b1.Gateway {
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayScheduledCondition(gw, scheduled))
	return gw
}

// UpdateGatewayStatusAddrs updates the status addresses for the provided gateway
// based on the status IP/Hostname of svc and updates the Ready condition based on the
// service and deployment state.
func UpdateGatewayStatusReadyCondition(gw *gwapiv1b1.Gateway, svc *corev1.Service, deployment *appsv1.Deployment) {
	var addrs, hostnames []string
	// Update the status addresses field.
	if svc != nil {
		for i := range svc.Status.LoadBalancer.Ingress {
			switch {
			case len(svc.Status.LoadBalancer.Ingress[i].IP) > 0:
				addrs = append(addrs, svc.Status.LoadBalancer.Ingress[i].IP)
			case len(svc.Status.LoadBalancer.Ingress[i].Hostname) > 0:
				// Remove when the following supports the hostname address type:
				// https://github.com/kubernetes-sigs/gateway-api/blob/v0.5.0/conformance/utils/kubernetes/helpers.go#L201-L207
				if svc.Status.LoadBalancer.Ingress[i].Hostname == "localhost" {
					addrs = append(addrs, "127.0.0.1")
				}
				hostnames = append(hostnames, svc.Status.LoadBalancer.Ingress[i].Hostname)
			}
		}

		var gwAddrs []gwapiv1b1.GatewayAddress
		for i := range addrs {
			addr := gwapiv1b1.GatewayAddress{
				Type:  gatewayapi.GatewayAddressTypePtr(gwapiv1b1.IPAddressType),
				Value: addrs[i],
			}
			gwAddrs = append(gwAddrs, addr)
		}

		for i := range hostnames {
			addr := gwapiv1b1.GatewayAddress{
				Type:  gatewayapi.GatewayAddressTypePtr(gwapiv1b1.HostnameAddressType),
				Value: hostnames[i],
			}
			gwAddrs = append(gwAddrs, addr)
		}

		gw.Status.Addresses = gwAddrs
	}
	// Update the ready condition.
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, computeGatewayReadyCondition(gw, deployment))
}
