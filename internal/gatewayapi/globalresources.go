// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package gatewayapi

import (
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

const envoyTLSSecretName = "envoy"

// ProcessGlobalResources processes global resources that are not tied to a specific listener or route
func (t *Translator) ProcessGlobalResources(resources *resource.Resources, xdsIRs resource.XdsIRMap, gateways []*GatewayContext) error {
	// Add the ProxyServiceCluster information for each gateway to the IR map
	for _, gateway := range gateways {
		// Get the gateway IR key and RouteDestination representing the ProxyServiceCluster
		irKey, rDest := t.processServiceClusterForGateway(gateway, resources)

		if xdsIRs[irKey] == nil {
			continue
		}
		if xdsIRs[irKey].GlobalResources == nil {
			xdsIRs[irKey].GlobalResources = &ir.GlobalResources{}
		}
		xdsIRs[irKey].GlobalResources.ProxyServiceCluster = rDest
	}

	// Get the envoy client TLS secret. It is used for envoy to establish a TLS connection with control plane components,
	// including the rate limit server and the wasm HTTP server.
	envoyTLSSecret := resources.GetSecret(t.ControllerNamespace, envoyTLSSecretName)
	if envoyTLSSecret == nil {
		return fmt.Errorf("envoy TLS secret %s/%s not found", t.ControllerNamespace, envoyTLSSecretName)
	}

	for _, xdsIR := range xdsIRs {
		if containsGlobalRateLimit(xdsIR.HTTP) || containsWasm(xdsIR.HTTP) {
			xdsIR.GlobalResources = &ir.GlobalResources{
				EnvoyClientCertificate: &ir.TLSCertificate{
					Name:        irGlobalConfigName(envoyTLSSecret),
					Certificate: envoyTLSSecret.Data[corev1.TLSCertKey],
					PrivateKey:  envoyTLSSecret.Data[corev1.TLSPrivateKeyKey],
				},
			}
		}
	}

	return nil
}

// processServiceClusterForGateway returns the matching IR key for a gateway and builds a RouteDestination to represent the ProxyServiceCluster
func (t *Translator) processServiceClusterForGateway(gateway *GatewayContext, resources *resource.Resources) (string, *ir.RouteDestination) {
	var (
		labels = make(map[string]string)
		irKey  string
	)

	if t.MergeGateways {
		irKey = string(t.GatewayClassName)
		labels[OwningGatewayClassLabel] = string(t.GatewayClassName)
	} else {
		irKey = irStringKey(gateway.Namespace, gateway.Name)
		labels[OwningGatewayNamespaceLabel] = gateway.Namespace
		labels[OwningGatewayNameLabel] = gateway.Name
	}

	svcCluster := resources.GetServiceByLabels(labels, t.ControllerNamespace)

	// Service lookup fails on first iteration
	if svcCluster == nil {
		return "", nil
	}

	bRef := gwapiv1.BackendObjectReference{
		Group:     GroupPtr(svcCluster.GroupVersionKind().Group),
		Kind:      KindPtr(svcCluster.Kind),
		Name:      gwapiv1.ObjectName(svcCluster.Name),
		Namespace: NamespacePtr(svcCluster.Namespace),
		Port:      PortNumPtr(svcCluster.Spec.Ports[0].Port),
	}
	dst := t.processServiceDestinationSetting(irKey, bRef, svcCluster.Namespace, ir.AppProtocol(svcCluster.Spec.Ports[0].Protocol), resources, resources.EnvoyProxyForGatewayClass)

	return irKey, &ir.RouteDestination{
		Name:     dst.Name,
		Settings: []*ir.DestinationSetting{dst},
	}
}

func irGlobalConfigName(object metav1.Object) string {
	return fmt.Sprintf("%s/%s", object.GetNamespace(), object.GetName())
}

func containsGlobalRateLimit(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.Traffic != nil &&
				route.Traffic.RateLimit != nil &&
				route.Traffic.RateLimit.Global != nil {
				return true
			}
		}
	}
	return false
}

func containsWasm(httpListeners []*ir.HTTPListener) bool {
	for _, httpListener := range httpListeners {
		for _, route := range httpListener.Routes {
			if route.EnvoyExtensions != nil &&
				len(route.EnvoyExtensions.Wasms) > 0 {
				return true
			}
		}
	}
	return false
}
