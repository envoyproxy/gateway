// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/utils"
	"github.com/envoyproxy/gateway/internal/ir"
)

type ResourceRender struct {
	// Namespace is the Namespace used for managed infra.
	Namespace string

	infra               *ir.RateLimitInfra
	ratelimit           *egcfgv1a1.RateLimit
	rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec
}

// NewResourceRender returns a new ResourceRender.
func NewResourceRender(ns string, infra *ir.RateLimitInfra, rl *egcfgv1a1.RateLimit, deploy *egcfgv1a1.KubernetesDeploymentSpec) *ResourceRender {
	return &ResourceRender{
		Namespace:           ns,
		infra:               infra,
		ratelimit:           rl,
		rateLimitDeployment: deploy,
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

	serviceSpec := utils.ExpectedServiceSpec(egcfgv1a1.DefaultKubernetesServiceType())
	serviceSpec.Ports = ports
	serviceSpec.Selector = utils.GetSelector(labels).MatchLabels

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

// ServiceAccount returns the expected ratelimit serviceAccount.
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

// rateLimitLabels returns the labels used for all envoy rate limit resources.
func rateLimitLabels() map[string]string {
	return map[string]string{
		"app.gateway.envoyproxy.io/name": InfraName,
	}
}
