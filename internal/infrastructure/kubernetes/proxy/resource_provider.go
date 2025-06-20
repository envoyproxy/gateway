// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package proxy

import (
	"context"
	"fmt"
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
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	gwapiresource "github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/common"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	// XdsTLSCertFilepath is the fully qualified path of the file containing Envoy's
	// xDS server TLS certificate.
	XdsTLSCertFilepath = "/certs/tls.crt"
	// XdsTLSKeyFilepath is the fully qualified path of the file containing Envoy's
	// xDS server TLS key.
	XdsTLSKeyFilepath = "/certs/tls.key"
	// XdsTLSCaFilepath is the fully qualified path of the file containing Envoy's
	// trusted CA certificate.
	XdsTLSCaFilepath = "/certs/ca.crt"

	// XdsTLSCaFileName is the file name of the xDS server TLS certificate.
	XdsTLSCaFileName = "ca.crt"
)

type ResourceRender struct {
	infra *ir.ProxyInfra

	// envoyNamespace is the namespace used for managed infra.
	envoyNamespace string

	// controllerNamespace is the namespace used for Envoy Gateway controller.
	controllerNamespace string

	// DNSDomain is the dns domain used by k8s services. Defaults to "cluster.local".
	DNSDomain string

	ShutdownManager *egv1a1.ShutdownManager

	GatewayNamespaceMode bool

	// ownerReferenceUID store the uid of its owner reference. Key is the kind of owner resource.
	// - GatewayClass when enabled ControllerNamespaceMode, merged Gateway...
	// - Gateway when enabled GatewayNamespaceMode
	ownerReferenceUID map[string]types.UID
}

// KubernetesInfraProvider provide information for initializing the proxy resource render.
type KubernetesInfraProvider interface {
	GetControllerNamespace() string
	GetDNSDomain() string
	GetEnvoyGateway() *egv1a1.EnvoyGateway
	GetOwnerReferenceUID(ctx context.Context, infra *ir.Infra) (map[string]types.UID, error)
	GetResourceNamespace(ir *ir.Infra) string
}

func NewResourceRender(ctx context.Context, kubernetesInfra KubernetesInfraProvider, infra *ir.Infra) (*ResourceRender, error) {
	ownerReference, err := kubernetesInfra.GetOwnerReferenceUID(ctx, infra)
	if err != nil {
		return nil, err
	}

	return &ResourceRender{
		envoyNamespace:       kubernetesInfra.GetResourceNamespace(infra),
		controllerNamespace:  kubernetesInfra.GetControllerNamespace(),
		DNSDomain:            kubernetesInfra.GetDNSDomain(),
		infra:                infra.GetProxyInfra(),
		ShutdownManager:      kubernetesInfra.GetEnvoyGateway().GetEnvoyGatewayProvider().GetEnvoyGatewayKubeProvider().ShutdownManager,
		GatewayNamespaceMode: kubernetesInfra.GetEnvoyGateway().GatewayNamespaceMode(),
		ownerReferenceUID:    ownerReference,
	}, nil
}

func (r *ResourceRender) Name() string {
	if r.GatewayNamespaceMode {
		return r.infra.Name
	}

	return ExpectedResourceHashedName(r.infra.Name)
}

func (r *ResourceRender) Namespace() string {
	return r.envoyNamespace
}

func (r *ResourceRender) ControllerNamespace() string {
	return r.controllerNamespace
}

func (r *ResourceRender) LabelSelector() labels.Selector {
	return labels.SelectorFromSet(r.stableSelector().MatchLabels)
}

func (r *ResourceRender) OwnerReferences() []metav1.OwnerReference {
	var ownerReferences []metav1.OwnerReference
	if r.ownerReferenceUID != nil {
		key := gwapiresource.KindGatewayClass
		if r.GatewayNamespaceMode {
			key = gwapiresource.KindGateway
		}
		if uid, ok := r.ownerReferenceUID[key]; ok {
			ownerReferences = append(ownerReferences, metav1.OwnerReference{
				APIVersion: gwapiv1.GroupVersion.String(),
				Kind:       r.infra.GetProxyMetadata().OwnerReference.Kind,
				Name:       r.infra.GetProxyMetadata().OwnerReference.Name,
				UID:        uid,
			})
		}
	}
	return ownerReferences
}

// ServiceAccount returns the expected proxy serviceAccount.
func (r *ResourceRender) ServiceAccount() (*corev1.ServiceAccount, error) {
	// Set the labels based on the owning gateway name.
	saLabels := r.envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(saLabels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Name:            r.Name(),
			Labels:          saLabels,
			Annotations:     r.infra.GetProxyMetadata().Annotations,
			OwnerReferences: r.OwnerReferences(),
		},
	}, nil
}

// envoyLabels returns the labels, including extraLabels, used for Envoy resources.
func (r *ResourceRender) envoyLabels(extraLabels map[string]string) map[string]string {
	appLabels := EnvoyAppLabel()
	if r.GatewayNamespaceMode {
		appLabels[gatewayapi.GatewayNameLabel] = r.Name()
	}
	maps.Copy(appLabels, extraLabels)

	return appLabels
}

// Service returns the expected Service based on the provided infra.
func (r *ResourceRender) Service() (*corev1.Service, error) {
	var ports []corev1.ServicePort
	for _, listener := range r.infra.Listeners {
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

			if port.Protocol == ir.HTTPSProtocolType {
				if listener.HTTP3 != nil {
					p := corev1.ServicePort{
						Name:       port.Name + "-h3",
						Protocol:   corev1.ProtocolUDP,
						Port:       port.ServicePort,
						TargetPort: target,
					}
					ports = append(ports, p)
				}
			}
		}
	}

	// Set the infraLabels based on the owning gatewayclass name.
	infraLabels := r.envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(infraLabels) {
		return nil, fmt.Errorf("missing owning gateway infraLabels")
	}

	// Get annotations
	annotations := map[string]string{}
	maps.Copy(annotations, r.infra.GetProxyMetadata().Annotations)

	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	envoyServiceConfig := provider.GetEnvoyProxyKubeProvider().EnvoyService
	if envoyServiceConfig.Annotations != nil {
		maps.Copy(annotations, envoyServiceConfig.Annotations)
	}
	if len(annotations) == 0 {
		annotations = nil
	}

	// Get service-specific labels
	svcLabels := map[string]string{}
	maps.Copy(svcLabels, infraLabels)
	if envoyServiceConfig.Labels != nil {
		maps.Copy(svcLabels, envoyServiceConfig.Labels)
	}
	if len(svcLabels) == 0 {
		svcLabels = nil
	}

	// Set the spec of gateway service
	serviceSpec := resource.ExpectedServiceSpec(envoyServiceConfig)
	serviceSpec.Ports = ports
	serviceSpec.Selector = resource.GetSelector(infraLabels).MatchLabels

	if (*envoyServiceConfig.Type) == egv1a1.ServiceTypeClusterIP {
		if len(r.infra.Addresses) > 0 {
			// Since K8s Service requires specify no more than one IP for each IP family
			// So we only use the first address
			// if address is not set, the automatically assigned clusterIP is used
			serviceSpec.ClusterIP = r.infra.Addresses[0]
			serviceSpec.ClusterIPs = r.infra.Addresses[0:1]
		}
	} else {
		serviceSpec.ExternalIPs = r.infra.Addresses
	}

	// Set IP family policy and families based on proxy config request
	ipFamily := r.infra.GetProxyConfig().Spec.IPFamily
	if ipFamily != nil {
		// SingleStack+IPv4 is default behavior from K8s and so is omitted
		switch *ipFamily {
		case egv1a1.IPv6:
			serviceSpec.IPFamilies = []corev1.IPFamily{corev1.IPv6Protocol}
			serviceSpec.IPFamilyPolicy = ptr.To(corev1.IPFamilyPolicySingleStack)
		case egv1a1.DualStack:
			serviceSpec.IPFamilies = []corev1.IPFamily{corev1.IPv4Protocol, corev1.IPv6Protocol}
			serviceSpec.IPFamilyPolicy = ptr.To(corev1.IPFamilyPolicyRequireDualStack)
		}
	}

	svc := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Labels:          svcLabels,
			Annotations:     annotations,
			OwnerReferences: r.OwnerReferences(),
		},
		Spec: serviceSpec,
	}

	// set name
	if envoyServiceConfig.Name != nil {
		svc.Name = *envoyServiceConfig.Name
	} else {
		svc.Name = r.Name()
	}

	// apply merge patch to service
	var err error
	if svc, err = utils.MergeWithPatch(svc, envoyServiceConfig.Patch); err != nil {
		return nil, err
	}

	return svc, nil
}

// ConfigMap returns the expected ConfigMap based on the provided infra.
func (r *ResourceRender) ConfigMap(cert string) (*corev1.ConfigMap, error) {
	// Set the labels based on the owning gateway name.
	cmLabels := r.envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(cmLabels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	data := map[string]string{
		common.SdsCAFilename:   common.GetSdsCAConfigMapData(XdsTLSCaFilepath),
		common.SdsCertFilename: common.GetSdsCertConfigMapData(XdsTLSCertFilepath, XdsTLSKeyFilepath),
	}
	if cert != "" {
		data[XdsTLSCaFileName] = cert
	}

	return &corev1.ConfigMap{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ConfigMap",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Name:            r.Name(),
			Labels:          cmLabels,
			Annotations:     r.infra.GetProxyMetadata().Annotations,
			OwnerReferences: r.OwnerReferences(),
		},
		Data: data,
	}, nil
}

// stableSelector returns a stable selector based on the owning gateway labels.
// "stable" here means the selector doesn't change when the infra is updated.
func (r *ResourceRender) stableSelector() *metav1.LabelSelector {
	stableLabels := map[string]string{}
	for k, v := range r.infra.GetProxyMetadata().Labels {
		if k == gatewayapi.OwningGatewayNameLabel || k == gatewayapi.OwningGatewayNamespaceLabel || k == gatewayapi.OwningGatewayClassLabel {
			stableLabels[k] = v
		}
	}

	return resource.GetSelector(r.envoyLabels(stableLabels))
}

// Deployment returns the expected Deployment based on the provided infra.
func (r *ResourceRender) Deployment() (*appsv1.Deployment, error) {
	proxyConfig := r.infra.GetProxyConfig()

	// Get the EnvoyProxy config to configure the deployment.
	provider := proxyConfig.GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment

	// If deployment config is nil, it's not Deployment installation.
	if deploymentConfig == nil {
		return nil, nil
	}

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, deploymentConfig.Container, proxyConfig.Spec.Shutdown, r.ShutdownManager, r.ControllerNamespace(), r.DNSDomain, r.GatewayNamespaceMode)
	if err != nil {
		return nil, err
	}

	dpAnnotations := r.infra.GetProxyMetadata().Annotations
	podAnnotations := r.getPodAnnotations(dpAnnotations, deploymentConfig.Pod)

	// Set the labels based on the owning gateway name.
	dpLabels, err := r.getLabels()
	if err != nil {
		return nil, err
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Labels:          dpLabels,
			Annotations:     dpAnnotations,
			OwnerReferences: r.OwnerReferences(),
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: deploymentConfig.Replicas,
			Strategy: *deploymentConfig.Strategy,
			// Deployment's selector is immutable.
			Selector: r.stableSelector(),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.getPodLabels(deploymentConfig.Pod),
					Annotations: podAnnotations,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					InitContainers:                deploymentConfig.InitContainers,
					ServiceAccountName:            r.Name(),
					TerminationGracePeriodSeconds: expectedTerminationGracePeriodSeconds(proxyConfig.Spec.Shutdown),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					SecurityContext:               deploymentConfig.Pod.SecurityContext,
					Affinity:                      deploymentConfig.Pod.Affinity,
					Tolerations:                   deploymentConfig.Pod.Tolerations,
					Volumes:                       r.expectedVolumes(deploymentConfig.Pod),
					ImagePullSecrets:              deploymentConfig.Pod.ImagePullSecrets,
					NodeSelector:                  deploymentConfig.Pod.NodeSelector,
					TopologySpreadConstraints:     deploymentConfig.Pod.TopologySpreadConstraints,
				},
			},
			RevisionHistoryLimit:    ptr.To[int32](10),
			ProgressDeadlineSeconds: ptr.To[int32](600),
		},
	}

	// set name
	if deploymentConfig.Name != nil {
		deployment.Name = *deploymentConfig.Name
	} else {
		deployment.Name = r.Name()
	}

	// apply merge patch to deployment
	if deployment, err = utils.MergeWithPatch(deployment, deploymentConfig.Patch); err != nil {
		return nil, err
	}

	return deployment, nil
}

func (r *ResourceRender) DaemonSet() (*appsv1.DaemonSet, error) {
	proxyConfig := r.infra.GetProxyConfig()

	// Get the EnvoyProxy config to configure the daemonset.
	provider := proxyConfig.GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	daemonSetConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDaemonSet

	// If daemonset config is nil, it's not DaemonSet installation.
	if daemonSetConfig == nil {
		return nil, nil
	}

	// Get expected bootstrap configurations rendered ProxyContainers
	containers, err := expectedProxyContainers(r.infra, daemonSetConfig.Container, proxyConfig.Spec.Shutdown, r.ShutdownManager, r.ControllerNamespace(), r.DNSDomain, r.GatewayNamespaceMode)
	if err != nil {
		return nil, err
	}

	dsAnnotations := r.infra.GetProxyMetadata().Annotations
	podAnnotations := r.getPodAnnotations(dsAnnotations, daemonSetConfig.Pod)

	// Set the labels based on the owning gateway name.
	dsLabels, err := r.getLabels()
	if err != nil {
		return nil, err
	}

	daemonSet := &appsv1.DaemonSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "DaemonSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Labels:          dsLabels,
			Annotations:     dsAnnotations,
			OwnerReferences: r.OwnerReferences(),
		},
		Spec: appsv1.DaemonSetSpec{
			// Daemonset's selector is immutable.
			Selector:       r.stableSelector(),
			UpdateStrategy: *daemonSetConfig.Strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels:      r.getPodLabels(daemonSetConfig.Pod),
					Annotations: podAnnotations,
				},
				Spec: r.getPodSpec(containers, nil, daemonSetConfig.Pod, proxyConfig),
			},
		},
	}

	// set name
	if daemonSetConfig.Name != nil {
		daemonSet.Name = *daemonSetConfig.Name
	} else {
		daemonSet.Name = r.Name()
	}

	// apply merge patch to DaemonSet
	if daemonSet, err = utils.MergeWithPatch(daemonSet, daemonSetConfig.Patch); err != nil {
		return nil, err
	}

	return daemonSet, nil
}

func (r *ResourceRender) pdbConfig() (*egv1a1.KubernetesPodDisruptionBudgetSpec, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	podDisruptionBudget := provider.GetEnvoyProxyKubeProvider().EnvoyPDB
	if podDisruptionBudget == nil || podDisruptionBudget.MinAvailable == nil && podDisruptionBudget.MaxUnavailable == nil && podDisruptionBudget.Patch == nil {
		return nil, nil
	}

	return podDisruptionBudget, nil
}

func (r *ResourceRender) PodDisruptionBudget() (*policyv1.PodDisruptionBudget, error) {
	pdb, err := r.pdbConfig()
	// If podDisruptionBudget config is nil, ignore PodDisruptionBudget.
	if pdb == nil {
		return nil, err
	}

	pdbSpec := policyv1.PodDisruptionBudgetSpec{
		Selector: r.stableSelector(),
	}
	switch {
	case pdb.MinAvailable != nil:
		pdbSpec.MinAvailable = pdb.MinAvailable
	case pdb.MaxUnavailable != nil:
		pdbSpec.MaxUnavailable = pdb.MaxUnavailable
	default:
		pdbSpec.MinAvailable = &intstr.IntOrString{Type: intstr.Int, IntVal: 0}
	}

	podDisruptionBudget := &policyv1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			OwnerReferences: r.OwnerReferences(),
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1",
			Kind:       "PodDisruptionBudget",
		},
		Spec: pdbSpec,
	}

	// set name
	if pdb.Name != nil {
		podDisruptionBudget.Name = *pdb.Name
	} else {
		podDisruptionBudget.Name = r.Name()
	}

	// apply merge patch to PodDisruptionBudget
	if podDisruptionBudget, err = pdb.ApplyMergePatch(podDisruptionBudget); err != nil {
		return nil, err
	}

	return podDisruptionBudget, nil
}

func (r *ResourceRender) HorizontalPodAutoscaler() (*autoscalingv2.HorizontalPodAutoscaler, error) {
	provider := r.infra.GetProxyConfig().GetEnvoyProxyProvider()
	if provider.Type != egv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}

	hpaConfig := provider.GetEnvoyProxyKubeProvider().EnvoyHpa
	if hpaConfig == nil {
		return nil, nil
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace:       r.Namespace(),
			Annotations:     r.infra.GetProxyMetadata().Annotations,
			Labels:          r.infra.GetProxyMetadata().Labels,
			OwnerReferences: r.OwnerReferences(),
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
	deploymentConfig := provider.GetEnvoyProxyKubeProvider().EnvoyDeployment
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

	var err error
	if hpa, err = utils.MergeWithPatch(hpa, hpaConfig.Patch); err != nil {
		return nil, err
	}

	return hpa, nil
}

func expectedTerminationGracePeriodSeconds(cfg *egv1a1.ShutdownConfig) *int64 {
	s := 360 // default
	if cfg != nil && cfg.DrainTimeout != nil {
		s = int(cfg.DrainTimeout.Seconds() + 300) // 5 minutes longer than drain timeout
	}
	return ptr.To(int64(s))
}

func (r *ResourceRender) getPodSpec(
	containers, initContainers []corev1.Container,
	pod *egv1a1.KubernetesPodSpec,
	proxyConfig *egv1a1.EnvoyProxy,
) corev1.PodSpec {
	return corev1.PodSpec{
		Containers:                    containers,
		InitContainers:                initContainers,
		ServiceAccountName:            r.Name(),
		TerminationGracePeriodSeconds: expectedTerminationGracePeriodSeconds(proxyConfig.Spec.Shutdown),
		DNSPolicy:                     corev1.DNSClusterFirst,
		RestartPolicy:                 corev1.RestartPolicyAlways,
		SchedulerName:                 "default-scheduler",
		SecurityContext:               pod.SecurityContext,
		Affinity:                      pod.Affinity,
		Tolerations:                   pod.Tolerations,
		Volumes:                       r.expectedVolumes(pod),
		ImagePullSecrets:              pod.ImagePullSecrets,
		NodeSelector:                  pod.NodeSelector,
		TopologySpreadConstraints:     pod.TopologySpreadConstraints,
	}
}

func (r *ResourceRender) getPodAnnotations(resourceAnnotation map[string]string, pod *egv1a1.KubernetesPodSpec) map[string]string {
	podAnnotations := map[string]string{}
	maps.Copy(podAnnotations, resourceAnnotation)
	maps.Copy(podAnnotations, pod.Annotations)

	if enablePrometheus(r.infra) {
		podAnnotations["prometheus.io/path"] = "/stats/prometheus" // TODO: make this configurable
		podAnnotations["prometheus.io/scrape"] = "true"
		podAnnotations["prometheus.io/port"] = strconv.Itoa(bootstrap.EnvoyStatsPort)
	}

	if len(podAnnotations) == 0 {
		podAnnotations = nil
	}

	return podAnnotations
}

func (r *ResourceRender) getLabels() (map[string]string, error) {
	// Set the labels based on the owning gateway name.
	resourceLabels := r.envoyLabels(r.infra.GetProxyMetadata().Labels)
	if OwningGatewayLabelsAbsent(resourceLabels) {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	return resourceLabels, nil
}

func (r *ResourceRender) getPodLabels(pod *egv1a1.KubernetesPodSpec) map[string]string {
	podLabels := r.infra.GetProxyMetadata().Labels
	maps.Copy(podLabels, pod.Labels)

	return r.envoyLabels(podLabels)
}

// OwningGatewayLabelsAbsent Check if labels are missing some OwningGatewayLabels
func OwningGatewayLabelsAbsent(labels map[string]string) bool {
	return (len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 ||
		len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0) &&
		len(labels[gatewayapi.OwningGatewayClassLabel]) == 0
}
