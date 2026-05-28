package synthesizer

import (
	"context"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"
)

// GetDeployment renders the desired Deployment for the IR.
func (is *InfraSynthesizer) GetDeployment(ctx context.Context, ir *Infra) (*appsv1.Deployment, error) {
	containers, err := is.GetDeploymentContainers(ctx, ir)
	if err != nil {
		return nil, err
	}

	labels := ir.Proxy.Metadata.Labels

	return &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName(ir),
			Namespace: is.Namespace,
			Labels:    labels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					AutomountServiceAccountToken: ptr.To(false),
					Containers:                   containers,
					Volumes:                      is.GetDeploymentVolumes(),
				},
			},
			RevisionHistoryLimit:    ptr.To(int32(10)),
			ProgressDeadlineSeconds: ptr.To(int32(600)),
		},
	}, nil
}

// GetDeploymentContainers renders the container list (currently just envoy)
// for the proxy Pod.
func (is *InfraSynthesizer) GetDeploymentContainers(ctx context.Context, ir *Infra) ([]corev1.Container, error) {
	ports := []corev1.ContainerPort{
		{
			Name:          "readiness",
			ContainerPort: 19003,
			Protocol:      corev1.ProtocolTCP,
		},
	}

	params := templateParams{
		ServiceClusterName: ir.Proxy.Name,
	}
	buf := new(strings.Builder)
	if err := bootstrapTmpl.Execute(buf, params); err != nil {
		return nil, err
	}

	args := []string{
		"--service-cluster", ir.Proxy.Name,
		"--service-node", ir.Proxy.Name,
		"--config-yaml", buf.String(),
		"--log-level", "info",
		"--cpuset-threads",
		"--drain-strategy", "immediate",
	}

	env := []corev1.EnvVar{
		{
			Name: "ENVOY_POD_NAMESPACE",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.namespace",
				},
			},
		},
		{
			Name: "ENVOY_POD_NAME",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.name",
				},
			},
		},
	}

	volumeMounts := []corev1.VolumeMount{
		{
			Name:      "certs",
			MountPath: "/certs",
			ReadOnly:  true,
		},
	}

	readinessPort := intstr.IntOrString{Type: intstr.Int, IntVal: 19003}

	return []corev1.Container{
		{
			Name:            "envoy",
			Image:           "docker.io/envoyproxy/envoy:distroless-dev",
			ImagePullPolicy: corev1.PullIfNotPresent,
			Command:         []string{"envoy"},
			Args:            args,
			Env:             env,
			SecurityContext: &corev1.SecurityContext{
				AllowPrivilegeEscalation: ptr.To(false),
				Capabilities: &corev1.Capabilities{
					Drop: []corev1.Capability{"ALL"},
				},
				Privileged:             ptr.To(false),
				ReadOnlyRootFilesystem: ptr.To(false),
				RunAsNonRoot:           ptr.To(true),
				SeccompProfile: &corev1.SeccompProfile{
					Type: corev1.SeccompProfileTypeRuntimeDefault,
				},
				RunAsUser:  ptr.To(int64(65532)),
				RunAsGroup: ptr.To(int64(65532)),
			},
			Ports:                    ports,
			VolumeMounts:             volumeMounts,
			TerminationMessagePolicy: corev1.TerminationMessageReadFile,
			TerminationMessagePath:   "/dev/termination-log",
			StartupProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/ready",
						Port:   readinessPort,
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 30,
			},
			ReadinessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/ready",
						Port:   readinessPort,
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    5,
				SuccessThreshold: 1,
				FailureThreshold: 1,
			},
			LivenessProbe: &corev1.Probe{
				ProbeHandler: corev1.ProbeHandler{
					HTTPGet: &corev1.HTTPGetAction{
						Path:   "/ready",
						Port:   readinessPort,
						Scheme: corev1.URISchemeHTTP,
					},
				},
				TimeoutSeconds:   1,
				PeriodSeconds:    10,
				SuccessThreshold: 1,
				FailureThreshold: 3,
			},
		},
	}, nil
}

// GetDeploymentVolumes returns the pod volumes the envoy container expects.
// The envoy Secret is mounted into /certs and provides the xDS client
// material referenced from the bootstrap template.
func (is *InfraSynthesizer) GetDeploymentVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "certs",
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName:  "envoy",
					DefaultMode: ptr.To(int32(420)),
				},
			},
		},
	}
}

// mergeDeploymentSpec applies desired Deployment fields onto existing.
// Selector is intentionally left untouched after creation because it is
// immutable: changing it would force callers to delete and recreate the
// Deployment, which is out of scope for this synthesizer.
func mergeDeploymentSpec(existing, desired *appsv1.Deployment) {
	existing.Labels = desired.Labels
	existing.Annotations = desired.Annotations

	if existing.CreationTimestamp.IsZero() {
		existing.Spec.Selector = desired.Spec.Selector
	}
	existing.Spec.Replicas = desired.Spec.Replicas
	existing.Spec.RevisionHistoryLimit = desired.Spec.RevisionHistoryLimit
	existing.Spec.ProgressDeadlineSeconds = desired.Spec.ProgressDeadlineSeconds
	existing.Spec.Strategy = desired.Spec.Strategy
	existing.Spec.Template = desired.Spec.Template
}
