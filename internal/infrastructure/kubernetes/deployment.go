// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	_ "embed"
	"fmt"
	"reflect"
	"strings"
	"text/template"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/pointer"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/provider/utils"
	xdsrunner "github.com/envoyproxy/gateway/internal/xds/server/runner"
)

const (
	// envoyContainerName is the name of the Envoy container.
	envoyContainerName = "envoy"
	// envoyNsEnvVar is the name of the Envoy Gateway namespace environment variable.
	envoyNsEnvVar = "ENVOY_GATEWAY_NAMESPACE"
	// envoyPodEnvVar is the name of the Envoy pod name environment variable.
	envoyPodEnvVar = "ENVOY_POD_NAME"
	// envoyCfgFileName is the name of the Envoy configuration file.
	envoyCfgFileName = "bootstrap.yaml"
	// envoyHTTPPort is the container port number of Envoy's HTTP endpoint.
	envoyHTTPPort = int32(8080)
	// envoyHTTPSPort is the container port number of Envoy's HTTPS endpoint.
	envoyHTTPSPort = int32(8443)
	// envoyGatewayXdsServerHost is the DNS name of the Xds Server within Envoy Gateway.
	// It defaults to the Envoy Gateway Kubernetes service.
	envoyGatewayXdsServerHost = "envoy-gateway"
	// envoyAdminAddress is the listening address of the envoy admin interface.
	envoyAdminAddress = "127.0.0.1"
	// envoyAdminPort is the port used to expose admin interface.
	envoyAdminPort = 19000
	// envoyAdminAccessLogPath is the path used to expose admin access log.
	envoyAdminAccessLogPath = "/dev/null"

	// rateLimitInfraName is the name for rate-limit resources.
	rateLimitInfraName = "envoy-ratelimit"
	// rateLimitInfraGRPCPort is the grpc port that the rate limit service listens on.
	rateLimitInfraGRPCPort = 8081
	// rateLimitInfraImage is the container image for the rate limit service.
	rateLimitInfraImage = "envoyproxy/ratelimit:f28024e3"
)

//go:embed bootstrap.yaml.tpl
var bootstrapTmplStr string

var bootstrapTmpl = template.Must(template.New(envoyCfgFileName).Parse(bootstrapTmplStr))

// envoyBootstrap defines the envoy Bootstrap configuration.
type bootstrapConfig struct {
	// parameters defines configurable bootstrap configuration parameters.
	parameters bootstrapParameters
	// rendered is the rendered bootstrap configuration.
	rendered string
}

// envoyBootstrap defines the envoy Bootstrap configuration.
type bootstrapParameters struct {
	// XdsServer defines the configuration of the XDS server.
	XdsServer xdsServerParameters
	// AdminServer defines the configuration of the Envoy admin interface.
	AdminServer adminServerParameters
}

type xdsServerParameters struct {
	// Address is the address of the XDS Server that Envoy is managed by.
	Address string
	// Port is the port of the XDS Server that Envoy is managed by.
	Port int32
}

type adminServerParameters struct {
	// Address is the address of the Envoy admin interface.
	Address string
	// Port is the port of the Envoy admin interface.
	Port int32
	// AccessLogPath is the path of the Envoy admin access log.
	AccessLogPath string
}

// render the stringified bootstrap config in yaml format.
func (b *bootstrapConfig) render() error {
	buf := new(strings.Builder)
	if err := bootstrapTmpl.Execute(buf, b.parameters); err != nil {
		return fmt.Errorf("failed to render bootstrap config: %v", err)
	}
	b.rendered = buf.String()

	return nil
}

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
	// Get the EnvoyProxy config to configure the ret.
	provider := infra.GetProxyInfra().GetProxyConfig().GetProvider()
	if provider.Type != egcfgv1a1.ProviderTypeKubernetes {
		return nil, fmt.Errorf("invalid provider type %v for Kubernetes infra manager", provider.Type)
	}
	deployCfg := provider.GetKubeResourceProvider().EnvoyDeployment

	ret := &appsv1.Deployment{
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

	return ret, nil
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

	cfg := bootstrapConfig{
		parameters: bootstrapParameters{
			XdsServer: xdsServerParameters{
				Address: envoyGatewayXdsServerHost,
				Port:    xdsrunner.XdsServerPort,
			},
			AdminServer: adminServerParameters{
				Address:       envoyAdminAddress,
				Port:          envoyAdminPort,
				AccessLogPath: envoyAdminAccessLogPath,
			},
		},
	}
	if err := cfg.render(); err != nil {
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
				fmt.Sprintf("--config-yaml %s", cfg.rendered),
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

// expectedRateLimitDeployment returns the expected rate limit Deployment based on the provided infra.
func (i *Infra) expectedRateLimitDeployment(infra *ir.RateLimitInfra) *appsv1.Deployment {
	containers := expectedRateLimitContainers(infra)
	labels := rateLimitLabels()
	selector := getSelector(labels)

	ret := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(int32(1)),
			Selector: selector,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: selector.MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            rateLimitInfraName,
					AutomountServiceAccountToken:  pointer.Bool(false),
					TerminationGracePeriodSeconds: pointer.Int64(int64(300)),
					DNSPolicy:                     corev1.DNSClusterFirst,
					RestartPolicy:                 corev1.RestartPolicyAlways,
					SchedulerName:                 "default-scheduler",
					Volumes: []corev1.Volume{
						{
							Name: rateLimitInfraName,
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: rateLimitInfraName,
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

	return ret
}

func expectedRateLimitContainers(infra *ir.RateLimitInfra) []corev1.Container {
	ports := []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: rateLimitInfraGRPCPort,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	containers := []corev1.Container{
		{
			Name:            rateLimitInfraName,
			Image:           rateLimitInfraImage,
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command: []string{
				"/bin/ratelimit",
			},
			Env: []corev1.EnvVar{
				{
					Name:  "REDIS_SOCKET_TYPE",
					Value: "tcp",
				},
				{
					Name:  "REDIS_URL",
					Value: infra.Backend.Redis.URL,
				},
				{
					Name:  "RUNTIME_ROOT",
					Value: "/data",
				},
				{Name: "RUNTIME_SUBDIRECTORY",
					Value: "ratelimit",
				},
				{
					Name:  "RUNTIME_IGNOREDOTFILES",
					Value: "true",
				},
				{
					Name:  "RUNTIME_WATCH_ROOT",
					Value: "false",
				},
				{
					Name:  "LOG_LEVEL",
					Value: "debug",
				},
				{
					Name:  "USE_STATSD",
					Value: "false",
				},
			},
			Ports: ports,
			VolumeMounts: []corev1.VolumeMount{
				{
					Name:      rateLimitInfraName,
					MountPath: "/data/ratelimit/config",
					ReadOnly:  true,
				},
			},
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
		},
	}

	return containers
}

// createOrUpdateRateLimitDeployment creates a Deployment in the kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateRateLimitDeployment(ctx context.Context, infra *ir.RateLimitInfra) error {
	deploy := i.expectedRateLimitDeployment(infra)
	return i.createOrUpdateDeployment(ctx, deploy)
}

// deleteRateLimitDeployment deletes the Envoy RateLimit Deployment in the kube api server, if it exists.
func (i *Infra) deleteRateLimitDeployment(ctx context.Context, _ *ir.RateLimitInfra) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      rateLimitInfraName,
		},
	}

	return i.deleteDeployment(ctx, deploy)
}

func (i *Infra) createOrUpdateDeployment(ctx context.Context, deploy *appsv1.Deployment) error {
	current := &appsv1.Deployment{}
	key := types.NamespacedName{
		Namespace: deploy.Namespace,
		Name:      deploy.Name,
	}
	if err := i.Client.Get(ctx, key, current); err != nil {
		// Create if not found.
		if kerrors.IsNotFound(err) {
			if err := i.Client.Create(ctx, deploy); err != nil {
				return fmt.Errorf("failed to create deployment %s/%s: %w",
					deploy.Namespace, deploy.Name, err)
			}
		}
	} else {
		// Update if current value is different.
		if !reflect.DeepEqual(deploy.Spec, current.Spec) {
			if err := i.Client.Update(ctx, deploy); err != nil {
				return fmt.Errorf("failed to update deployment %s/%s: %w",
					deploy.Namespace, deploy.Name, err)
			}
		}
	}

	return nil
}

func (i *Infra) deleteDeployment(ctx context.Context, deploy *appsv1.Deployment) error {
	if err := i.Client.Delete(ctx, deploy); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete deployment %s/%s: %w", deploy.Namespace, deploy.Name, err)
	}
	return nil
}
