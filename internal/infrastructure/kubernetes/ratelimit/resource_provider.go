// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
)

type ResourceRender struct {
	// Namespace is the Namespace used for managed infra.
	Namespace string

	infra               *ir.RateLimitInfra
	rateLimit           *egcfgv1a1.RateLimit
	rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec
}

// NewResourceRender returns a new ResourceRender.
func NewResourceRender(ns string, infra *ir.RateLimitInfra, gateway *egcfgv1a1.EnvoyGateway) *ResourceRender {
	return &ResourceRender{
		Namespace:           ns,
		infra:               infra,
		rateLimit:           gateway.RateLimit,
		rateLimitDeployment: gateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
	}
}

func (r *ResourceRender) Name() string {
	return InfraName
}

// ConfigMap returns the expected ConfigMap based on the provided infra.
func (r *ResourceRender) ConfigMap() (*corev1.ConfigMap, error) {
	labels := rateLimitLabels()
	data := make(map[string]string)

	for _, config := range r.infra.ServiceConfigs {
		data[config.Name] = config.Config
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Data: data,
	}, nil
}

// Service returns the expected rate limit Service based on the provided infra.
func (r *ResourceRender) Service() (*corev1.Service, error) {
	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       InfraGRPCPort,
			TargetPort: intstr.IntOrString{IntVal: InfraGRPCPort},
		},
	}

	labels := rateLimitLabels()

	serviceSpec := resource.ExpectedServiceSpec(egcfgv1a1.DefaultKubernetesServiceType())
	serviceSpec.Ports = ports
	serviceSpec.Selector = resource.GetSelector(labels).MatchLabels

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Spec: serviceSpec,
	}

	return svc, nil
}

// ServiceAccount returns the expected rateLimit serviceAccount.
func (r *ResourceRender) ServiceAccount() (*corev1.ServiceAccount, error) {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
		},
	}, nil
}

// Deployment returns the expected rate limit Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	containers := expectedRateLimitContainers(r.rateLimit, r.rateLimitDeployment)
	labels := rateLimitLabels()
	selector := resource.GetSelector(labels)

	var annotations map[string]string
	if r.rateLimitDeployment.Pod.Annotations != nil {
		annotations = r.rateLimitDeployment.Pod.Annotations
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: r.rateLimitDeployment.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      selector.MatchLabels,
					Annotations: annotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            InfraName,
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               r.rateLimitDeployment.Pod.SecurityContext,
					Volumes:                       expectedDeploymentVolumes(r.rateLimit),
				},
			},
		},
	}

	return deployment, nil
}
