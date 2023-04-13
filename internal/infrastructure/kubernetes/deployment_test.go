// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
)

func checkEnvVar(t *testing.T, deploy *appsv1.Deployment, container, name string) {
	t.Helper()

	for i, c := range deploy.Spec.Template.Spec.Containers {
		if c.Name == container {
			for _, envVar := range deploy.Spec.Template.Spec.Containers[i].Env {
				if envVar.Name == name {
					return
				}
			}
		}
	}

	t.Errorf("deployment is missing environment variable %q", name)
}

func checkContainer(t *testing.T, deploy *appsv1.Deployment, name string, expect bool) *corev1.Container {
	t.Helper()

	if deploy.Spec.Template.Spec.Containers == nil {
		t.Error("deployment has no containers")
	}

	for _, container := range deploy.Spec.Template.Spec.Containers {
		if container.Name == name {
			if expect {
				return &container
			}
			t.Errorf("deployment has unexpected %q container", name)
		}
	}

	if expect {
		t.Errorf("deployment has no %q container", name)
	}
	return nil
}

func checkContainerHasArg(t *testing.T, container *corev1.Container, arg string) {
	t.Helper()

	for _, a := range container.Args {
		if a == arg {
			return
		}
	}
	t.Errorf("container is missing argument %q", arg)
}

func checkLabels(t *testing.T, deploy *appsv1.Deployment, expected map[string]string) {
	t.Helper()

	if apiequality.Semantic.DeepEqual(deploy.Labels, expected) {
		return
	}

	t.Errorf("deployment has unexpected %q labels", deploy.Labels)
}

func checkContainerHasPort(t *testing.T, deploy *appsv1.Deployment, port int32) {
	t.Helper()

	for _, c := range deploy.Spec.Template.Spec.Containers {
		for _, p := range c.Ports {
			if p.ContainerPort == port {
				return
			}
		}
	}
	t.Errorf("container is missing containerPort %q", port)
}

func checkPodAnnotations(t *testing.T, deploy *appsv1.Deployment, expected map[string]string) {
	t.Helper()

	if apiequality.Semantic.DeepEqual(deploy.Spec.Template.Annotations, expected) {
		return
	}

	t.Errorf("deployment has unexpected %q pod annotations ", deploy.Spec.Template.Annotations)
}

func checkPodSecurityContext(t *testing.T, deploy *appsv1.Deployment, expected *corev1.PodSecurityContext) {
	t.Helper()

	if apiequality.Semantic.DeepEqual(deploy.Spec.Template.Spec.SecurityContext, expected) {
		return
	}

	t.Errorf("deployment has unexpected %q pod annotations ", deploy.Spec.Template.Spec.SecurityContext)
}

func checkContainerSecurityContext(t *testing.T, container *corev1.Container, expected *corev1.SecurityContext) {
	t.Helper()

	if apiequality.Semantic.DeepEqual(container.SecurityContext, expected) {
		return
	}

	t.Errorf("container has unexpected %q pod annotations ", container.SecurityContext)
}

func checkContainerImage(t *testing.T, container *corev1.Container, image string) {
	t.Helper()

	if container.Image == image {
		return
	}
	t.Errorf("container is missing image %q", image)
}

func checkContainerResources(t *testing.T, container *corev1.Container, expected *corev1.ResourceRequirements) {
	t.Helper()
	if apiequality.Semantic.DeepEqual(&container.Resources, expected) {
		return
	}

	t.Errorf("container has unexpected %q resources ", expected)
}
