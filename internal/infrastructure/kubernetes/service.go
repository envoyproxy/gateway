// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"reflect"

	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
)

func expectedProxyServiceName(proxyName string) string {
	svcName := utils.GetHashedName(proxyName)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, svcName)
}

// expectedproxyService returns the expected Service based on the provided infra.
func (i *Infra) expectedProxyService(infra *ir.Infra) (*corev1.Service, error) {
	var ports []corev1.ServicePort
	for _, listener := range infra.Proxy.Listeners {
		for _, port := range listener.Ports {
			target := intstr.IntOrString{IntVal: port.ContainerPort}
			protocol := corev1.ProtocolTCP
			if port.Protocol == ir.UDPProtocolType {
				protocol = corev1.ProtocolUDP
			}
			p := corev1.ServicePort{
				Name:       port.Name,
				Protocol:   protocol,
				Port:       port.ServicePort,
				TargetPort: target,
			}
			ports = append(ports, p)
		}
	}

	// Set the labels based on the owning gatewayclass name.
	labels := envoyLabels(infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyServiceName(infra.Proxy.Name),
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			Ports:           ports,
			Selector:        getSelector(labels).MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			// Preserve the client source IP and avoid a second hop for LoadBalancer.
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}

	return svc, nil
}

// createOrUpdateproxyService creates a Service in the kube api server based on the provided infra,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateProxyService(ctx context.Context, infra *ir.Infra) error {
	svc, err := i.expectedProxyService(infra)
	if err != nil {
		return fmt.Errorf("failed to generate expected service: %w", err)
	}

	current := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      expectedProxyServiceName(infra.Proxy.Name),
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, svc); err != nil {
				return fmt.Errorf("failed to create service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(svc.Spec, current.Spec) {
			if err := i.Client.Update(ctx, svc); err != nil {
				return fmt.Errorf("failed to update service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	}

	return nil
}

// deleteProxyService deletes the Envoy Service in the kube api server, if it exists.
func (i *Infra) deleteProxyService(ctx context.Context, infra *ir.Infra) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyServiceName(infra.Proxy.Name),
		},
	}

	return i.deleteService(ctx, svc)
}

// expectedRateLimitInfraService returns the expected rate limit Service based on the provided infra.
func (i *Infra) expectedRateLimitService(_ *ir.RateLimitInfra) *corev1.Service {
	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       rateLimitInfraHTTPPort,
			TargetPort: intstr.IntOrString{IntVal: rateLimitInfraHTTPPort},
		},
	}

	labels := rateLimitLabels()

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Type:            corev1.ServiceTypeLoadBalancer,
			Ports:           ports,
			Selector:        getSelector(labels).MatchLabels,
			SessionAffinity: corev1.ServiceAffinityNone,
			// Preserve the client source IP and avoid a second hop for LoadBalancer.
			ExternalTrafficPolicy: corev1.ServiceExternalTrafficPolicyTypeLocal,
		},
	}

	return svc
}

// createOrUpdateRateLimitService creates a Service in the kube api server based on the provided infra,
// if it doesn't exist or updates it if it does.
func (i *Infra) createOrUpdateRateLimitService(ctx context.Context, infra *ir.RateLimitInfra) error {
	svc := i.expectedRateLimitService(infra)
	return i.createOrUpdateService(ctx, svc)
}

// deleteRateLimitService deletes the rate limit Service in the kube api server, if it exists.
func (i *Infra) deleteRateLimitService(ctx context.Context, _ *ir.RateLimitInfra) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteService(ctx, svc)
}

func (i *Infra) createOrUpdateService(ctx context.Context, svc *corev1.Service) error {
	current := &corev1.Service{}
	key := types.NamespacedName{
		Namespace: svc.Namespace,
		Name:      svc.Name,
	}

	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, svc); err != nil {
				return fmt.Errorf("failed to create service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(svc.Spec, current.Spec) {
			if err := i.Client.Update(ctx, svc); err != nil {
				return fmt.Errorf("failed to update service %s/%s: %w",
					svc.Namespace, svc.Name, err)
			}
		}
	}

	return nil
}

func (i *Infra) deleteService(ctx context.Context, svc *corev1.Service) error {
	if err := i.Client.Delete(ctx, svc); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete service %s/%s: %w", svc.Namespace, svc.Name, err)
	}

	return nil
}
