// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
)

// ResourceKind indicates the main resources of envoy-ratelimit,
// but also the key for the uid of their ownerReference.
const (
	ResourceKindService        = "Service"
	ResourceKindDeployment     = "Deployment"
	ResourceKindServiceAccount = "ServiceAccount"
)

type ResourceRender struct {
	// Namespace is the Namespace used for managed infra.
	Namespace string

	rateLimit           *egcfgv1a1.RateLimit
	rateLimitDeployment *egcfgv1a1.KubernetesDeploymentSpec

	// ownerReferenceUID store the uid of its owner reference.
	ownerReferenceUID map[string]types.UID
}

// NewResourceRender returns a new ResourceRender.
func NewResourceRender(ns string, gateway *egcfgv1a1.EnvoyGateway, ownerReferenceUID map[string]types.UID) *ResourceRender {
	return &ResourceRender{
		Namespace:           ns,
		rateLimit:           gateway.RateLimit,
		rateLimitDeployment: gateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
		ownerReferenceUID:   ownerReferenceUID,
	}
}

func (r *ResourceRender) Name() string {
	return InfraName
}

// ConfigMap is deprecated since ratelimit supports xds grpc config server.
func (r *ResourceRender) ConfigMap() (*corev1.ConfigMap, error) {
	return nil, nil
}

// Service returns the expected rate limit Service based on the provided infra.
func (r *ResourceRender) Service() (*corev1.Service, error) {
	const apiVersion = "v1"

	ports := []corev1.ServicePort{
		{
			Name:       "http",
			Protocol:   corev1.ProtocolTCP,
			Port:       InfraGRPCPort,
			TargetPort: intstr.IntOrString{IntVal: InfraGRPCPort},
		},
	}

	labels := rateLimitLabels()
	kubernetesServiceSpec := &egcfgv1a1.KubernetesServiceSpec{
		Type: egcfgv1a1.GetKubernetesServiceType(egcfgv1a1.ServiceTypeClusterIP),
	}
	serviceSpec := resource.ExpectedServiceSpec(kubernetesServiceSpec)
	serviceSpec.Ports = ports
	serviceSpec.Selector = resource.GetSelector(labels).MatchLabels

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       ResourceKindService,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Spec: serviceSpec,
	}

	if r.ownerReferenceUID != nil {
		if uid, ok := r.ownerReferenceUID[ResourceKindService]; ok {
			svc.OwnerReferences = []metav1.OwnerReference{
				{
					Kind:       ResourceKindService,
					APIVersion: apiVersion,
					Name:       "envoy-gateway",
					UID:        uid,
				},
			}
		}
	}

	return svc, nil
}

// ServiceAccount returns the expected rateLimit serviceAccount.
func (r *ResourceRender) ServiceAccount() (*corev1.ServiceAccount, error) {
	const apiVersion = "v1"

	sa := &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       ResourceKindServiceAccount,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
		},
	}

	if r.ownerReferenceUID != nil {
		if uid, ok := r.ownerReferenceUID[ResourceKindServiceAccount]; ok {
			sa.OwnerReferences = []metav1.OwnerReference{
				{
					Kind:       ResourceKindServiceAccount,
					APIVersion: apiVersion,
					Name:       "envoy-gateway",
					UID:        uid,
				},
			}
		}
	}

	return sa, nil
}

// Deployment returns the expected rate limit Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	const apiVersion = "apps/v1"

	containers := expectedRateLimitContainers(r.rateLimit, r.rateLimitDeployment)
	labels := rateLimitLabels()
	selector := resource.GetSelector(labels)

	var annotations map[string]string
	if r.rateLimitDeployment.Pod.Annotations != nil {
		annotations = r.rateLimitDeployment.Pod.Annotations
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       ResourceKindDeployment,
			APIVersion: apiVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace,
			Name:      InfraName,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: r.rateLimitDeployment.Replicas,
			Strategy: *r.rateLimitDeployment.Strategy,
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
					Volumes:                       expectedDeploymentVolumes(r.rateLimit, r.rateLimitDeployment),
					Affinity:                      r.rateLimitDeployment.Pod.Affinity,
					Tolerations:                   r.rateLimitDeployment.Pod.Tolerations,
				},
			},
			RevisionHistoryLimit:    pointer.Int32(10),
			ProgressDeadlineSeconds: pointer.Int32(600),
		},
	}

	if r.ownerReferenceUID != nil {
		if uid, ok := r.ownerReferenceUID[ResourceKindDeployment]; ok {
			deployment.OwnerReferences = []metav1.OwnerReference{
				{
					Kind:       ResourceKindDeployment,
					APIVersion: apiVersion,
					Name:       "envoy-gateway",
					UID:        uid,
				},
			}
		}
	}

	return deployment, nil
}
