// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/ir"
)

// expectedRateLimitInfraService returns the expected rate limit Service based on the provided infra.
func (i *Infra) expectedRateLimitService(_ *ir.RateLimitInfra) *corev1.Service {
	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       rateLimitInfraGRPCPort,
			TargetPort: intstr.IntOrString{IntVal: rateLimitInfraGRPCPort},
		},
	}

	labels := rateLimitLabels()

	serviceSpec := expectedServiceSpec(egcfgv1a1.DefaultKubernetesServiceType())
	serviceSpec.Ports = ports
	serviceSpec.Selector = getSelector(labels).MatchLabels

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
			Labels:    labels,
		},
		Spec: serviceSpec,
	}

	return svc
}

// GetRateLimitServiceURL returns the URL for the rate limit service.
func GetRateLimitServiceURL(namespace string) string {
	return fmt.Sprintf("grpc://%s.%s.svc.cluster.local:%d", rateLimitInfraName, namespace, rateLimitInfraGRPCPort)
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
