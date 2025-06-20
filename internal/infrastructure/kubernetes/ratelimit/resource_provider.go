// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package ratelimit

import (
	_ "embed"
	"strconv"

	"golang.org/x/exp/maps"
	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/utils"
)

// ResourceKind indicates the main resources of envoy-ratelimit,
// but also the key for the uid of their ownerReference.
const (
	ResourceKindService        = "Service"
	ResourceKindDeployment     = "Deployment"
	ResourceKindServiceAccount = "ServiceAccount"
	appsAPIVersion             = "apps/v1"
)

//go:embed statsd_conf.yaml
var statsConf string

type ResourceRender struct {
	// namespace is the Namespace used for managed infra.
	namespace string

	rateLimit           *egv1a1.RateLimit
	rateLimitDeployment *egv1a1.KubernetesDeploymentSpec
	rateLimitHpa        *egv1a1.KubernetesHorizontalPodAutoscalerSpec

	// ownerReferenceUID store the uid of its owner reference.
	ownerReferenceUID map[string]types.UID
}

// NewResourceRender returns a new ResourceRender.
func NewResourceRender(ns string, gateway *egv1a1.EnvoyGateway, ownerReferenceUID map[string]types.UID) *ResourceRender {
	return &ResourceRender{
		namespace:           ns,
		rateLimit:           gateway.RateLimit,
		rateLimitDeployment: gateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitDeployment,
		rateLimitHpa:        gateway.GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().RateLimitHpa,
		ownerReferenceUID:   ownerReferenceUID,
	}
}

func (r *ResourceRender) Name() string {
	return InfraName
}

func (r *ResourceRender) Namespace() string {
	return r.namespace
}

func (r *ResourceRender) LabelSelector() labels.Selector {
	return labels.SelectorFromSet(rateLimitLabels())
}

func enablePrometheus(rl *egv1a1.RateLimit) bool {
	if rl != nil &&
		rl.Telemetry != nil &&
		rl.Telemetry.Metrics != nil &&
		rl.Telemetry.Metrics.Prometheus != nil {
		return !rl.Telemetry.Metrics.Prometheus.Disable
	}

	return true
}

// ConfigMap returns the expected rate limit ConfigMap based on the provided infra.
func (r *ResourceRender) ConfigMap(cert string) (*corev1.ConfigMap, error) {
	if !enablePrometheus(r.rateLimit) {
		return nil, nil
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace(),
			Name:      "statsd-exporter-config",
			Labels:    rateLimitLabels(),
		},
		Data: map[string]string{
			"conf.yaml": statsConf,
		},
	}, nil
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

	if enablePrometheus(r.rateLimit) {
		metricsPort := corev1.ServicePort{
			Name:       "metrics",
			Protocol:   corev1.ProtocolTCP,
			Port:       PrometheusPort,
			TargetPort: intstr.IntOrString{IntVal: PrometheusPort},
		}
		ports = append(ports, metricsPort)
	}

	labels := rateLimitLabels()
	kubernetesServiceSpec := &egv1a1.KubernetesServiceSpec{
		Type: egv1a1.GetKubernetesServiceType(egv1a1.ServiceTypeClusterIP),
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
			Namespace: r.Namespace(),
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
			Namespace: r.Namespace(),
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
	containers := expectedRateLimitContainers(r.rateLimit, r.rateLimitDeployment, r.Namespace())
	selector := resource.GetSelector(rateLimitLabels())

	podLabels := rateLimitLabels()
	if r.rateLimitDeployment.Pod.Labels != nil {
		maps.Copy(podLabels, r.rateLimitDeployment.Pod.Labels)
		// Copy overwrites values in the dest map if they exist in the src map https://pkg.go.dev/maps#Copy
		// It's applied again with the rateLimitLabels that are used as deployment selector to ensure those are not overwritten by user input
		maps.Copy(podLabels, rateLimitLabels())
	}

	var podAnnotations map[string]string
	if enablePrometheus(r.rateLimit) {
		podAnnotations = map[string]string{
			"prometheus.io/path":   "/metrics",
			"prometheus.io/port":   strconv.Itoa(PrometheusPort),
			"prometheus.io/scrape": "true",
		}
	}
	if r.rateLimitDeployment.Pod.Annotations != nil {
		if podAnnotations != nil {
			maps.Copy(podAnnotations, r.rateLimitDeployment.Pod.Annotations)
		} else {
			podAnnotations = r.rateLimitDeployment.Pod.Annotations
		}
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       ResourceKindDeployment,
			APIVersion: appsAPIVersion,
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace(),
			Labels:    rateLimitLabels(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: r.rateLimitDeployment.Replicas,
			Strategy: *r.rateLimitDeployment.Strategy,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      podLabels,
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            InfraName,
					AutomountServiceAccountToken:  ptr.To(false),
					TerminationGracePeriodSeconds: ptr.To[int64](300),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               r.rateLimitDeployment.Pod.SecurityContext,
					Volumes:                       expectedDeploymentVolumes(r.rateLimit, r.rateLimitDeployment),
					Affinity:                      r.rateLimitDeployment.Pod.Affinity,
					Tolerations:                   r.rateLimitDeployment.Pod.Tolerations,
					ImagePullSecrets:              r.rateLimitDeployment.Pod.ImagePullSecrets,
					NodeSelector:                  r.rateLimitDeployment.Pod.NodeSelector,
					TopologySpreadConstraints:     r.rateLimitDeployment.Pod.TopologySpreadConstraints,
				},
			},
			RevisionHistoryLimit:    ptr.To[int32](10),
			ProgressDeadlineSeconds: ptr.To[int32](600),
		},
	}

	// set name
	if r.rateLimitDeployment.Name != nil {
		deployment.Name = *r.rateLimitDeployment.Name
	} else {
		deployment.Name = r.Name()
	}

	if r.ownerReferenceUID != nil {
		if uid, ok := r.ownerReferenceUID[ResourceKindDeployment]; ok {
			deployment.OwnerReferences = []metav1.OwnerReference{
				{
					Kind:       ResourceKindDeployment,
					APIVersion: appsAPIVersion,
					Name:       "envoy-gateway",
					UID:        uid,
				},
			}
		}
	}

	// apply merge patch to deployment
	var err error
	deploymentConfig := r.rateLimitDeployment
	if deployment, err = utils.MergeWithPatch(deployment, deploymentConfig.Patch); err != nil {
		return nil, err
	}

	return deployment, nil
}

// TODO: implement this method
func (r *ResourceRender) DaemonSet() (*appsv1.DaemonSet, error) {
	return nil, nil
}

// HorizontalPodAutoscalerSpec returns the `HorizontalPodAutoscaler` sets spec.
func (r *ResourceRender) HorizontalPodAutoscalerSpec() (*egv1a1.KubernetesHorizontalPodAutoscalerSpec, error) {
	return r.rateLimitHpa, nil
}

func (r *ResourceRender) HorizontalPodAutoscaler() (*autoscalingv2.HorizontalPodAutoscaler, error) {
	hpaConfig, err := r.HorizontalPodAutoscalerSpec()
	if hpaConfig == nil {
		return nil, err
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: r.Namespace(),
			Name:      r.Name(),
			Labels:    rateLimitLabels(),
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
			},
			MinReplicas: hpaConfig.MinReplicas,
			MaxReplicas: ptr.Deref(hpaConfig.MaxReplicas, 1),
			Metrics:     hpaConfig.Metrics,
			Behavior:    hpaConfig.Behavior,
		},
	}

	// set deployment target ref name
	deploymentConfig := r.rateLimitDeployment
	if deploymentConfig.Name != nil {
		hpa.Spec.ScaleTargetRef.Name = *deploymentConfig.Name
	} else {
		hpa.Spec.ScaleTargetRef.Name = r.Name()
	}

	// set name
	if hpaConfig.Name != nil {
		hpa.Name = *hpaConfig.Name
	} else {
		hpa.Name = r.Name()
	}

	if hpa, err = utils.MergeWithPatch(hpa, hpaConfig.Patch); err != nil {
		return nil, err
	}

	return hpa, nil
}

func (r *ResourceRender) PodDisruptionBudget() (*policyv1.PodDisruptionBudget, error) {
	return nil, nil
}
