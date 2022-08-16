package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
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

func checkContainerImage(t *testing.T, container *corev1.Container, image string) {
	t.Helper()

	if container.Image == image {
		return
	}
	t.Errorf("container is missing image %q", image)
}

func TestExpectedDeployment(t *testing.T) {
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli)
	infra := ir.NewInfra()
	deploy := kube.expectedDeployment(infra)

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deploy, envoyContainerName, true)
	checkContainerImage(t, container, ir.DefaultProxyImage)
	checkEnvVar(t, deploy, envoyContainerName, envoyNsEnvVar)
	checkEnvVar(t, deploy, envoyContainerName, envoyPodEnvVar)
	checkLabels(t, deploy, deploy.Labels)

	// Check container ports for the deployment are as expected.
	ports := []int32{envoyHTTPPort, envoyHTTPSPort}
	for _, port := range ports {
		checkContainerHasPort(t, deploy, port)
	}
}

func TestCreateDeploymentIfNeeded(t *testing.T) {
	kube := NewInfra(nil)
	infra := ir.NewInfra()
	deploy := kube.expectedDeployment(infra)
	deploy.ResourceVersion = "1"

	testCases := []struct {
		name    string
		in      *ir.Infra
		current *appsv1.Deployment
		out     *Resources
		expect  bool
	}{
		{
			name: "create deployment",
			in:   infra,
			out: &Resources{
				Deployment: deploy,
			},
			expect: true,
		},
		{
			name:    "deployment exists",
			in:      infra,
			current: deploy,
			out: &Resources{
				Deployment: deploy,
			},
			expect: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createDeploymentIfNeeded(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.out.Deployment, kube.Resources.Deployment)
			}
		})
	}
}
