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

	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
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

func expectedDeploymentName(proxyName string) string {
	return fmt.Sprintf("%s-%s", config.EnvoyDeploymentPrefix, proxyName)
}

// expectedDeployment returns the expected Deployment based on the provided infra.
func (i *Infra) expectedDeployment(infra *ir.Infra) (*appsv1.Deployment, error) {
	containers, err := expectedContainers(infra)
	if err != nil {
		return nil, err
	}

	// Set the labels based on the owning gatewayclass name.
	labels := envoyLabels(infra.GetProxyInfra().GetProxyMetadata().Labels)
	if _, ok := labels[gatewayapi.OwningGatewayLabel]; !ok {
		return nil, fmt.Errorf("missing owning gatewayclass label")
	}

	deployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedDeploymentName(infra.Proxy.Name),
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32(1),
			Selector: envoySelector(infra.GetProxyInfra().GetProxyMetadata().Labels),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: envoySelector(infra.GetProxyInfra().GetProxyMetadata().Labels).MatchLabels,
				},
				Spec: corev1.PodSpec{
					Containers:                    containers,
					ServiceAccountName:            expectedServiceAccountName(infra.Proxy.Name),
					AutomountServiceAccountToken:  pointer.BoolPtr(false),
					TerminationGracePeriodSeconds: pointer.Int64Ptr(int64(300)),
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
										Name: config.EnvoyConfigMapName,
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
									DefaultMode: pointer.Int32Ptr(int32(420)),
									Optional:    pointer.BoolPtr(false),
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

func expectedContainers(infra *ir.Infra) ([]corev1.Container, error) {
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

// createDeployment creates a Deployment in the kube api server based on the provided
// infra, if it doesn't exist and updates it if it does.
func (i *Infra) createOrUpdateDeployment(ctx context.Context, infra *ir.Infra) error {
	deploy, err := i.expectedDeployment(infra)
	if err != nil {
		return err
	}

	current := &appsv1.Deployment{}
	key := types.NamespacedName{
		Namespace: i.Namespace,
		Name:      expectedDeploymentName(infra.Proxy.Name),
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

	if err := i.updateResource(deploy); err != nil {
		return err
	}

	return nil
}

// deleteDeployment deletes the Envoy Deployment in the kube api server, if it exists.
func (i *Infra) deleteDeployment(ctx context.Context, infra *ir.Infra) error {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: i.Namespace,
			Name:      expectedDeploymentName(infra.Proxy.Name),
		},
	}

	if err := i.Client.Delete(ctx, deploy); err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return fmt.Errorf("failed to delete deployment %s/%s: %w", deploy.Namespace, deploy.Name, err)
	}

	return nil
}
