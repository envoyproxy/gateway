// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package status

import (
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func UpdateGatewayStatusNotAccepted(gw *gwapiv1.Gateway, reason gwapiv1.GatewayConditionReason, msg string) *gwapiv1.Gateway {
	cond := newCondition(string(gwapiv1.GatewayConditionAccepted), metav1.ConditionFalse, string(reason), msg, time.Now(), gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
	return gw
}

func UpdateGatewayStatusAccepted(gw *gwapiv1.Gateway) *gwapiv1.Gateway {
	cond := newCondition(string(gwapiv1.GatewayConditionAccepted), metav1.ConditionTrue,
		string(gwapiv1.GatewayReasonAccepted), "The Gateway has been scheduled by Envoy Gateway", time.Now(), gw.Generation)
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions, cond)
	return gw
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

// UpdateGatewayStatusProgrammedCondition updates the status addresses for the provided gateway
// based on the status IP/Hostname of svc and updates the Programmed condition based on the
// service and deployment or daemonset state.
func UpdateGatewayStatusProgrammedCondition(gw *gwapiv1.Gateway, svc *corev1.Service, envoyObj client.Object, nodeAddresses ...string) {
	var addresses, hostnames []string
	// Update the status addresses field.
	if svc != nil {
		// If the addresses is explicitly set in the Gateway spec by the user, use it
		// to populate the Status
		if len(gw.Spec.Addresses) > 0 {
			// Make sure the addresses have been populated into ExternalIPs/ClusterIPs
			// and use that value
			if len(svc.Spec.ExternalIPs) > 0 {
				addresses = append(addresses, svc.Spec.ExternalIPs...)
			} else if len(svc.Spec.ClusterIPs) > 0 {
				addresses = append(addresses, svc.Spec.ClusterIPs...)
			}
		} else {
			if svc.Spec.Type == corev1.ServiceTypeLoadBalancer {
				for i := range svc.Status.LoadBalancer.Ingress {
					switch {
					case len(svc.Status.LoadBalancer.Ingress[i].IP) > 0:
						addresses = append(addresses, svc.Status.LoadBalancer.Ingress[i].IP)
					case len(svc.Status.LoadBalancer.Ingress[i].Hostname) > 0:
						// Remove when the following supports the hostname address type:
						// https://github.com/kubernetes-sigs/gateway-api/blob/v0.5.0/conformance/utils/kubernetes/helpers.go#L201-L207
						if svc.Status.LoadBalancer.Ingress[i].Hostname == "localhost" {
							addresses = append(addresses, "127.0.0.1")
						}
						hostnames = append(hostnames, svc.Status.LoadBalancer.Ingress[i].Hostname)
					}
				}
			}

			if svc.Spec.Type == corev1.ServiceTypeClusterIP {
				for i := range svc.Spec.ClusterIPs {
					if svc.Spec.ClusterIPs[i] != "" {
						addresses = append(addresses, svc.Spec.ClusterIPs[i])
					}
				}
			}

			if svc.Spec.Type == corev1.ServiceTypeNodePort {
				addresses = nodeAddresses
			}
		}

		gwAddresses := make([]gwapiv1.GatewayStatusAddress, 0, len(addresses)+len(hostnames))
		for i := range addresses {
			addr := gwapiv1.GatewayStatusAddress{
				Type:  ptr.To(gwapiv1.IPAddressType),
				Value: addresses[i],
			}
			gwAddresses = append(gwAddresses, addr)
		}

		for i := range hostnames {
			addr := gwapiv1.GatewayStatusAddress{
				Type:  ptr.To(gwapiv1.HostnameAddressType),
				Value: hostnames[i],
			}
			gwAddresses = append(gwAddresses, addr)
		}

		gw.Status.Addresses = gwAddresses
	} else {
		gw.Status.Addresses = nil
	}

	// Update the programmed condition.
	updateGatewayProgrammedCondition(gw, envoyObj)
}

func SetGatewayListenerStatusCondition(gateway *gwapiv1.Gateway, listenerStatusIdx int,
	conditionType gwapiv1.ListenerConditionType, status metav1.ConditionStatus, reason gwapiv1.ListenerConditionReason, message string,
) {
	cond := metav1.Condition{
		Type:               string(conditionType),
		Status:             status,
		Reason:             string(reason),
		Message:            message,
		ObservedGeneration: gateway.Generation,
		LastTransitionTime: metav1.NewTime(time.Now()),
	}
	gateway.Status.Listeners[listenerStatusIdx].Conditions = MergeConditions(gateway.Status.Listeners[listenerStatusIdx].Conditions, cond)
}

const (
	messageAddressNotAssigned  = "No addresses have been assigned to the Gateway"
	messageFmtTooManyAddresses = "Too many addresses (%d) have been assigned to the Gateway, the maximum number of addresses is 16"
	messageNoResources         = "Envoy replicas unavailable"
	messageFmtProgrammed       = "Address assigned to the Gateway, %d/%d envoy replicas available"
)

// updateGatewayProgrammedCondition computes the Gateway Programmed status condition.
// Programmed condition surfaces true when the Envoy Deployment or DaemonSet status is ready.
func updateGatewayProgrammedCondition(gw *gwapiv1.Gateway, envoyObj client.Object) {
	if len(gw.Status.Addresses) == 0 {
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonAddressNotAssigned),
				messageAddressNotAssigned, time.Now(), gw.Generation))
		return
	}

	if len(gw.Status.Addresses) > 16 {
		gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
			newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonInvalid),
				fmt.Sprintf(messageFmtTooManyAddresses, len(gw.Status.Addresses)), time.Now(), gw.Generation))

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
					fmt.Sprintf(messageFmtProgrammed, obj.Status.AvailableReplicas, obj.Status.Replicas), time.Now(), gw.Generation))
			return
		}
	case *appsv1.DaemonSet:
		if obj != nil && obj.Status.NumberAvailable > 0 {
			gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
				newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionTrue, string(gwapiv1.GatewayConditionProgrammed),
					fmt.Sprintf(messageFmtProgrammed, obj.Status.NumberAvailable, obj.Status.CurrentNumberScheduled), time.Now(), gw.Generation))
			return
		}
	}

	// If there are no available replicas for the Envoy Deployment or
	// Envoy DaemonSet, don't mark the Gateway as ready yet.
	gw.Status.Conditions = MergeConditions(gw.Status.Conditions,
		newCondition(string(gwapiv1.GatewayConditionProgrammed), metav1.ConditionFalse, string(gwapiv1.GatewayReasonNoResources),
			messageNoResources, time.Now(), gw.Generation))
}
