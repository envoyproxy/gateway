package kubernetes

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/ir"
	xdsrunner "github.com/envoyproxy/gateway/internal/xds/server/runner"
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
	deploy, err := kube.expectedDeployment(infra)
	require.NoError(t, err)

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deploy, envoyContainerName, true)
	checkContainerImage(t, container, ir.DefaultProxyImage)
	checkEnvVar(t, deploy, envoyContainerName, envoyNsEnvVar)
	checkEnvVar(t, deploy, envoyContainerName, envoyPodEnvVar)
	checkLabels(t, deploy, deploy.Labels)

	// Create a bootstrap config, render it into an arg, and ensure it's as expected.
	cfg := &bootstrapConfig{
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
	err = cfg.render()
	require.NoError(t, err)
	checkContainerHasArg(t, container, fmt.Sprintf("--config-yaml %s", cfg.rendered))

	// Check container ports for the deployment are as expected.
	ports := []int32{envoyHTTPPort, envoyHTTPSPort}
	for _, port := range ports {
		checkContainerHasPort(t, deploy, port)
	}
}

func TestCreateDeploymentIfNeeded(t *testing.T) {
	kube := NewInfra(nil)
	infra := ir.NewInfra()
	deploy, err := kube.expectedDeployment(infra)
	require.NoError(t, err)
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

func TestEnvoyPodSelector(t *testing.T) {
	cases := []struct {
		name     string
		gcName   string
		expected map[string]string
	}{
		{
			name:   "default",
			gcName: "eg",
			expected: map[string]string{
				"gatewayClass": "eg",
				"app":          "envoy",
			},
		},
	}

	for _, tc := range cases {
		t.Run("", func(t *testing.T) {
			got := EnvoyPodSelector(tc.gcName)
			require.Equal(t, tc.expected, got.MatchLabels)
		})
	}
}

func TestDeleteDeployment(t *testing.T) {
	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "delete deployment",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				mu:        sync.Mutex{},
				Namespace: "test",
			}
			err := kube.deleteDeployment(context.Background())
			require.NoError(t, err)
		})
	}
}
