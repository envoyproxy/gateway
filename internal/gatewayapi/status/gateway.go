// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"slices"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func UpdateGatewayStatusNotAccepted(gw *gwapiv1.Gateway, reason gwapiv1.GatewayConditionReason, msg string) *gwapiv1.Gateway {
	cond := newCondition(string(gwapiv1.GatewayConditionAccepted), metav1.ConditionFalse, string(reason), msg, gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
	return gw
}

func UpdateGatewayStatusAccepted(gw *gwapiv1.Gateway) *gwapiv1.Gateway {
	cond := newCondition(string(gwapiv1.GatewayConditionAccepted), metav1.ConditionTrue,
		string(gwapiv1.GatewayReasonAccepted), "The Gateway has been scheduled by Envoy Gateway", gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
	return gw
}

func UpdateGatewayStatusResolvedRefsCondition(gw *gwapiv1.Gateway, status metav1.ConditionStatus, reason gwapiv1.GatewayConditionReason, msg string) {
	cond := newCondition(string(gwapiv1.GatewayConditionResolvedRefs), status, string(reason), msg, gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
}

func UpdateGatewayStatusCondition(gw *gwapiv1.Gateway, conditionType gwapiv1.GatewayConditionType, status metav1.ConditionStatus, reason gwapiv1.GatewayConditionReason, msg string) {
	cond := newCondition(string(conditionType), status, string(reason), msg, gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
}

func GatewayNotAccepted(gw *gwapiv1.Gateway) bool {
	for _, c := range gw.Status.Conditions {
		if c.Type == string(gwapiv1.GatewayConditionAccepted) && c.Status == metav1.ConditionFalse {
			return true
		}
	}
	return false
}

func GatewayAccepted(gw *gwapiv1.Gateway) bool {
	return !GatewayNotAccepted(gw)
}

type NodeAddresses struct {
	IPv4 []string
	IPv6 []string
}

// UpdateGatewayStatusProgrammedCondition updates the status addresses for the provided gateway
// based on the status IP/Hostname of svc and updates the Programmed condition based on the
// service and deployment or daemonset state.
func UpdateGatewayStatusProgrammedCondition(gw *gwapiv1.Gateway, svc *corev1.Service, envoyObj client.Object, nodeAddresses NodeAddresses) {
	var addresses, hostnames []string
	var addressNotUsable bool

	// When the Service doesn't exist yet but spec.Addresses is set, populate
	// status addresses from spec so they are visible as soon as conditions are
	// reconciled. The Programmed condition will indicate AddressNotUsable until
	// the backing Service is created.
	if svc == nil && len(gw.Spec.Addresses) > 0 {
		addresses, hostnames = specAddressesToSlices(gw.Spec.Addresses)
		gw.Status.Addresses = buildGatewayAddresses(addresses, hostnames)
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonAddressNotUsable),
				messageAddressNotUsable, gw.Generation))
		return
	}

	// Update the status addresses field.
	if svc != nil {
		// If the addresses is explicitly set in the Gateway spec by the user, use it
		// to populate the Status
		if len(gw.Spec.Addresses) > 0 {
			// Make sure the addresses have been populated into ExternalIPs/ClusterIPs
			// and use that value
			if len(svc.Spec.ExternalIPs) > 0 {
				addresses = append(addresses, svc.Spec.ExternalIPs...)
			} else {
				addresses = filterNoneClusterIPs(svc.Spec.ClusterIPs)
			}
			// Hostname addresses in spec cannot be assigned via ExternalIPs/LB;
			// report AddressNotUsable if any are present.
			addressNotUsable = hasHostnameAddress(gw.Spec.Addresses)
		} else {
			switch svc.Spec.Type {
			case corev1.ServiceTypeLoadBalancer:
				addresses, hostnames = collectLoadBalancerAddresses(svc)
			case corev1.ServiceTypeClusterIP:
				addresses = filterNoneClusterIPs(svc.Spec.ClusterIPs)
			case corev1.ServiceTypeNodePort:
				if slices.Contains(svc.Spec.IPFamilies, corev1.IPv4Protocol) {
					addresses = append(addresses, nodeAddresses.IPv4...)
				}
				if slices.Contains(svc.Spec.IPFamilies, corev1.IPv6Protocol) {
					addresses = append(addresses, nodeAddresses.IPv6...)
				}
			}
		}

		gw.Status.Addresses = buildGatewayAddresses(addresses, hostnames)
	} else {
		gw.Status.Addresses = nil
	}

	// If any requested address cannot be used (e.g. hostname addresses),
	// report AddressNotUsable before checking deployment readiness.
	if addressNotUsable {
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonAddressNotUsable),
				messageAddressNotUsable, gw.Generation))
		return
	}

	// Update the programmed condition.
	updateGatewayProgrammedCondition(gw, envoyObj)
}

func collectLoadBalancerAddresses(svc *corev1.Service) (addresses, hostnames []string) {
	for i := range svc.Status.LoadBalancer.Ingress {
		switch {
		case len(svc.Status.LoadBalancer.Ingress[i].IP) > 0:
			addresses = append(addresses, svc.Status.LoadBalancer.Ingress[i].IP)
		case len(svc.Status.LoadBalancer.Ingress[i].Hostname) > 0:
			hostnames = append(hostnames, svc.Status.LoadBalancer.Ingress[i].Hostname)
		}
	}
	return addresses, hostnames
}

func specAddressesToSlices(specAddresses []gwapiv1.GatewaySpecAddress) (addresses, hostnames []string) {
	for _, specAddr := range specAddresses {
		addrType := gwapiv1.IPAddressType
		if specAddr.Type != nil {
			addrType = *specAddr.Type
		}
		switch addrType {
		case gwapiv1.IPAddressType:
			addresses = append(addresses, specAddr.Value)
		case gwapiv1.HostnameAddressType:
			hostnames = append(hostnames, specAddr.Value)
		}
	}
	return addresses, hostnames
}

func filterNoneClusterIPs(clusterIPs []string) []string {
	var out []string
	for _, ip := range clusterIPs {
		if ip != "" && ip != "None" {
			out = append(out, ip)
		}
	}
	return out
}

// Important: do not use this function directly, use listener.SetCondition instead so that listeners from ListenerSet can be updated correctly
func SetGatewayListenerStatusCondition(gateway *gwapiv1.Gateway, listenerStatusIdx int,
	conditionType gwapiv1.ListenerConditionType, status metav1.ConditionStatus, reason gwapiv1.ListenerConditionReason, message string,
) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: gateway.Generation,
	}
	gateway.Status.Listeners[listenerStatusIdx].Conditions = MergeConditions(gateway.Status.Listeners[listenerStatusIdx].Conditions, cond)
}

const (
	messageAddressNotAssigned  = "No addresses have been assigned to the Gateway"
	messageAddressNotUsable    = "One or more addresses requested in spec.addresses cannot be used"
	messageFmtTooManyAddresses = "Too many addresses (%d) have been assigned to the Gateway; only the first 16 are included in the status."
	messageNoResources         = "Envoy replicas unavailable"
	messageFmtProgrammed       = "Address assigned to the Gateway, %d/%d envoy replicas available"
)

// updateGatewayProgrammedCondition computes the Gateway Programmed status condition.
// Programmed condition surfaces true when the Envoy Deployment or DaemonSet status is ready.
func updateGatewayProgrammedCondition(gw *gwapiv1.Gateway, envoyObj client.Object) {
	if len(gw.Status.Addresses) == 0 {
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonAddressNotAssigned),
				messageAddressNotAssigned, gw.Generation))
		return
	}

	if len(gw.Status.Addresses) > 16 {
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue, string(gwapiv1.GatewayReasonProgrammed),
				fmt.Sprintf(messageFmtTooManyAddresses, len(gw.Status.Addresses)), gw.Generation))

		// Truncate the addresses to 16
		// so that the status can be updated successfully.
		gw.Status.Addresses = gw.Status.Addresses[:16]
		return
	}

	// Check for available Envoy replicas and if found mark the gateway as ready.
	switch obj := envoyObj.(type) {
	case *appsv1.Deployment:
		if obj != nil && obj.Status.AvailableReplicas > 0 {
			gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
				newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue, string(gwapiv1.GatewayConditionProgrammed),
					fmt.Sprintf(messageFmtProgrammed, obj.Status.AvailableReplicas, obj.Status.Replicas), gw.Generation))
			return
		}
	case *appsv1.DaemonSet:
		if obj != nil && obj.Status.NumberAvailable > 0 {
			gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
				newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue, string(gwapiv1.GatewayConditionProgrammed),
					fmt.Sprintf(messageFmtProgrammed, obj.Status.NumberAvailable, obj.Status.CurrentNumberScheduled), gw.Generation))
			return
		}
	}

	// If there are no available replicas for the Envoy Deployment or
	// Envoy DaemonSet, don't mark the Gateway as ready yet.
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
		newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonNoResources),
			messageNoResources, gw.Generation))
}

// GetGatewayListenerStatusConditions returns the status conditions for a specific listener in the gateway status.
func GetGatewayListenerStatusConditions(gateway *gwapiv1.Gateway, listenerStatusIdx int) []metav1.Condition {
	if gateway == nil || listenerStatusIdx < 0 || listenerStatusIdx >= len(gateway.Status.Listeners) {
		return nil
	}
	return gateway.Status.Listeners[listenerStatusIdx].Conditions
}

// hasHostnameAddress returns true if any address in the slice uses HostnameAddressType.
func hasHostnameAddress(addrs []gwapiv1.GatewaySpecAddress) bool {
	for i := range addrs {
		if addrs[i].Type != nil && *addrs[i].Type == gwapiv1.HostnameAddressType {
			return true
		}
	}
	return false
}

// buildGatewayAddresses creates a status address slice from IP and hostname slices.
func buildGatewayAddresses(ips, hostnames []string) []gwapiv1.GatewayStatusAddress {
	out := make([]gwapiv1.GatewayStatusAddress, 0, len(ips)+len(hostnames))
	for i := range ips {
		out = append(out, gwapiv1.GatewayStatusAddress{
			Type:  new(gwapiv1.IPAddressType),
			Value: ips[i],
		})
	}
	for i := range hostnames {
		out = append(out, gwapiv1.GatewayStatusAddress{
			Type:  new(gwapiv1.HostnameAddressType),
			Value: hostnames[i],
		})
	}
	return out
}
