// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package synthesizer

import (
	"errors"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// GetService renders the desired Service for the IR. The Service is a
// LoadBalancer that fronts the proxy Pods on the listener ports declared
// in the IR.
func (is *InfraSynthesizer) GetService(ir *Infra) (*corev1.Service, error) {
	svcPorts, err := buildServicePorts(ir)
	if err != nil {
		return nil, fmt.Errorf("failed to build service ports: %w", err)
	}
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: is.Namespace,
			Name:      resourceName(ir),
			Labels:    is.generateServiceLabels(ir),
		},
		Spec: corev1.ServiceSpec{
			Type:     corev1.ServiceTypeLoadBalancer,
			Ports:    svcPorts,
			Selector: ir.Proxy.Metadata.Labels,
		},
	}, nil
}

func (is *InfraSynthesizer) generateServiceLabels(ir *Infra) map[string]string {
	labels := map[string]string{
		"app.kubernetes.io/component":     "proxy",
		"app.kubernetes.io/managed-by":    "envoy-gateway",
		"app.kubernetes.io/name":          "envoy",
		"gateway.envoyproxy.io/delegated": "example-infra-server",
	}
	for k, v := range ir.Proxy.Metadata.Labels {
		labels[k] = v
	}
	return labels
}

func buildServicePorts(ir *Infra) ([]corev1.ServicePort, error) {
	var ports []corev1.ServicePort

	for _, listener := range ir.Proxy.Listeners {
		if listener == nil {
			continue
		}
		for _, port := range listener.Ports {
			p := corev1.ProtocolTCP
			if port.Protocol == "UDP" {
				p = corev1.ProtocolUDP
			}
			ports = append(ports, corev1.ServicePort{
				Name:       port.Name,
				Protocol:   p,
				Port:       port.ServicePort,
				TargetPort: intstr.FromInt32(port.ContainerPort),
			})
		}
	}

	if len(ports) == 0 {
		return nil, errors.New("proxy infra has no listener ports; at least one is required for a LoadBalancer Service")
	}
	return ports, nil
}

// mergeServiceSpec applies desired fields onto existing while preserving
// API-server-populated fields (ClusterIP/IPs, allocated nodePorts, the LB
// status, etc.). Without this, every reconcile would race the cluster and
// fail on immutable-field updates such as ClusterIP.
func mergeServiceSpec(existing, desired *corev1.Service) {
	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations

	existing.Spec.Type = desired.Spec.Type
	existing.Spec.Selector = desired.Spec.Selector
	existing.Spec.Ports = mergeServicePorts(existing.Spec.Ports, desired.Spec.Ports)
}

// mergeServicePorts preserves nodePorts allocated by kube-apiserver when the
// service type is NodePort or LoadBalancer. Ports are matched by the tuple
// (name, port, protocol).
func mergeServicePorts(existing, desired []corev1.ServicePort) []corev1.ServicePort {
	type key struct {
		name     string
		port     int32
		protocol corev1.Protocol
	}
	prev := make(map[key]int32, len(existing))
	for _, p := range existing {
		prev[key{name: p.Name, port: p.Port, protocol: p.Protocol}] = p.NodePort
	}
	out := make([]corev1.ServicePort, len(desired))
	for i, p := range desired {
		if p.NodePort == 0 {
			if np, ok := prev[key{name: p.Name, port: p.Port, protocol: p.Protocol}]; ok {
				p.NodePort = np
			}
		}
		out[i] = p
	}
	return out
}
