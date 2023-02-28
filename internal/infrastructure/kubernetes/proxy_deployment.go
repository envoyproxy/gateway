// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

const (
	// envoyContainerName is the name of the Envoy container.
	envoyContainerName = "envoy"
	// envoyNsEnvVar is the name of the Envoy Gateway namespace environment variable.
	envoyNsEnvVar = "ENVOY_GATEWAY_NAMESPACE"
	// envoyPodEnvVar is the name of the Envoy pod name environment variable.
	envoyPodEnvVar = "ENVOY_POD_NAME"
	// envoyHTTPPort is the container port number of Envoy's HTTP endpoint.
	envoyHTTPPort = int32(8080)
	// envoyHTTPSPort is the container port number of Envoy's HTTPS endpoint.
	envoyHTTPSPort = int32(8443)
)

func expectedProxyDeploymentName(proxyName string) string {
	deploymentName := utils.GetHashedName(proxyName)
	return fmt.Sprintf("%s-%s", config.EnvoyPrefix, deploymentName)
}

// expectedProxyDeployment returns the expected Deployment based on the provided infra.
func (i *Infra) expectedProxyDeployment(infra *ir.Infra) (*appsv1.Deployment, error) {
	containers, err := expectedProxyContainers(infra)
	if err != nil {
		return nil, err
	}

	// Set the labels based on the owning gateway name.
	labels := envoyLabels(infra.GetProxyInfra().GetProxyMetadata().Labels)
	if len(labels[gatewayapi.OwningGatewayNamespaceLabel]) == 0 || len(labels[gatewayapi.OwningGatewayNameLabel]) == 0 {
		return nil, fmt.Errorf("missing owning gateway labels")
	}

	selector := getSelector(labels)
	// Get the EnvoyProxy config to configure the deployment.
	provider := infra.GetProxyInfra().GetProxyConfig().GetProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deployCfg := provider.GetKubeResourceProvider().EnvoyDeployment

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyDeploymentName(infra.Proxy.Name),
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: deployCfg.Replicas,
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            expectedProxyServiceAccountName(infra.Proxy.Name),
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					Volumes: []corev1.Volume{
						{
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								Secret: &corev1.SecretVolumeSource{
									SecretName: "envoy",
								},
							},
						},
						{
							Name: "sds",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: expectedProxyConfigMapName(infra.Proxy.Name),
									},
									Items: []corev1.KeyToPath{
										{
											Key:  sdsCAFilename,
											Path: sdsCAFilename,
										},
										{
											Key:  sdsCertFilename,
											Path: sdsCertFilename,
										},
									},
									DefaultMode: pointer.Int32(int32(420)),
									Optional:    pointer.Bool(false),
								},
							},
						},
					},
				},
			},
		},
	}

	return deployment, nil
}

func expectedProxyContainers(infra *ir.Infra) ([]corev1.Container, error) {
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: envoyHTTPPort,
			Protocol:      corev1.ProtocolTCP,
		},
		{
			Name:          "https",
			ContainerPort: envoyHTTPSPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	cfg, err := bootstrap.GetRenderedBootstrapConfig()
	if err != nil {
		return nil, err
	}

	containers := []corev1.Container{
		{
			Name:            envoyContainerName,
			Image:           infra.Proxy.Image,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"envoy",
			},
			Args: []string{
				fmt.Sprintf("--service-cluster %s", infra.Proxy.Name),
				fmt.Sprintf("--service-node $(%s)", envoyPodEnvVar),
				fmt.Sprintf("--config-yaml %s", cfg),
				"--log-level info",
			},
			Env: []corev1.EnvVar{
				{
					Name: envoyNsEnvVar,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.namespace",
						},
					},
				},
				{
					Name: envoyPodEnvVar,
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							APIVersion: "v1",
							FieldPath:  "metadata.name",
						},
					},
				},
			},
			Ports: ports,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      "certs",
					MountPath: "/certs",
					ReadOnly:  true,
				},
				{
					Name:      "sds",
					MountPath: "/sds",
				},
			},
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers, nil
}

// createOrUpdateProxyDeployment creates a Deployment in the kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateProxyDeployment(ctx context.Context, infra *ir.Infra) error {
	deploy, err := i.expectedProxyDeployment(infra)
	if err != nil {
		return err
	}
	return i.createOrUpdateDeployment(ctx, deploy)
}

// deleteProxyDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteProxyDeployment(ctx context.Context, infra *ir.Infra) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedProxyDeploymentName(infra.Proxy.Name),
		},
	}

	return i.deleteDeployment(ctx, deploy)
}
