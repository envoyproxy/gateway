// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiequality "k8s.io/apimachinery/pkg/api/equality"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
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

func TestExpectedProxyDeployment(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)

	// Check the deployment name is as expected.
	assert.Equal(t, deploy.Name, expectedResourceHashedName(infra.Proxy.Name))

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deploy, envoyContainerName, true)
	checkContainerImage(t, container, ir.DefaultProxyImage)
	checkEnvVar(t, deploy, envoyContainerName, envoyNsEnvVar)
	checkEnvVar(t, deploy, envoyContainerName, envoyPodEnvVar)
	checkLabels(t, deploy, deploy.Labels)

	// Create a bootstrap config, render it into an arg, and ensure it's as expected.
	bstrap, err := bootstrap.GetRenderedBootstrapConfig()
	require.NoError(t, err)
	checkContainerHasArg(t, container, fmt.Sprintf("--config-yaml %s", bstrap))

	// Check container ports for the deployment are as expected.
	ports := []int32{envoyHTTPPort, envoyHTTPSPort}
	for _, port := range ports {
		checkContainerHasPort(t, deploy, port)
	}

	// Set the deployment replicas.
	repl := int32(2)
	infra.Proxy.GetProxyConfig().GetProvider().GetKubeResourceProvider().EnvoyDeployment.Replicas = &repl

	deploy, err = kube.expectedProxyDeployment(infra)
	require.NoError(t, err)

	// Check the number of replicas is as expected.
	assert.Equal(t, repl, *deploy.Spec.Replicas)
}

func TestExpectedBootstrap(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	// Set a custom bootstrap config into EnvoyProxy API and ensure the same
	// value is set as the container arg.
	bstrap := "blah"
	infra.Proxy.Config = &egcfgv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: egcfgv1a1.EnvoyProxySpec{
			Bootstrap: &bstrap,
		},
	}

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)
	container := checkContainer(t, deploy, envoyContainerName, true)
	checkContainerHasArg(t, container, fmt.Sprintf("--config-yaml %s", bstrap))
}

func deploymentWithImage(deploy *appsv1.Deployment, image string) *appsv1.Deployment {
	dCopy := deploy.DeepCopy()
	for i, c := range dCopy.Spec.Template.Spec.Containers {
		if c.Name == envoyContainerName {
			dCopy.Spec.Template.Spec.Containers[i].Image = image
		}
	}
	return dCopy
}

func TestCreateOrUpdateProxyDeployment(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)

	kube := NewInfra(nil, cfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)

	testCases := []struct {
		name    string
		in      *ir.Infra
		current *appsv1.Deployment
		want    *appsv1.Deployment
	}{
		{
			name: "create deployment",
			in:   infra,
			want: deploy,
		},
		{
			name:    "deployment exists",
			in:      infra,
			current: deploy,
			want:    deploy,
		},
		{
			name: "update deployment image",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      infra.Proxy.Name,
						},
					},
					Name:      ir.DefaultProxyName,
					Image:     "envoyproxy/gateway-dev:v1.2.3",
					Listeners: ir.NewProxyListeners(),
				},
			},
			current: deploy,
			want:    deploymentWithImage(deploy, "envoyproxy/gateway-dev:v1.2.3"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createOrUpdateProxyDeployment(context.Background(), tc.in)
			require.NoError(t, err)

			actual := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      expectedResourceHashedName(tc.in.Proxy.Name),
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
		})
	}
}

func TestDeleteProxyDeployment(t *testing.T) {
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
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			infra := ir.NewInfra()

			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

			err := kube.createOrUpdateProxyDeployment(context.Background(), infra)
			require.NoError(t, err)

			err = kube.deleteProxyDeployment(context.Background(), infra)
			require.NoError(t, err)
		})
	}
}
