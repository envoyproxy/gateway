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
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egcfgv1a1 "github.com/envoyproxy/gateway/api/config/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
	"github.com/envoyproxy/gateway/internal/xds/bootstrap"
)

func testExpectedProxyDeployment(t *testing.T,
	infra *ir.Infra,
	expectedResources *corev1.ResourceRequirements,
	expectedPodSecurityContext *corev1.PodSecurityContext,
	expectedSecurityContext *corev1.SecurityContext) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
	infra.Proxy.Listeners = []ir.ProxyListener{
		{
			Ports: []ir.ListenerPort{
				{
					Name:          "EnvoyHTTPPort",
					Protocol:      ir.TCPProtocolType,
					ContainerPort: envoyHTTPPort,
				},
				{
					Name:          "EnvoyHTTPSPort",
					Protocol:      ir.TCPProtocolType,
					ContainerPort: envoyHTTPSPort,
				},
			},
		},
	}

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)

	// Check the deployment name is as expected.
	assert.Equal(t, deploy.Name, expectedResourceHashedName(infra.Proxy.Name))

	// Check container details, i.e. env vars, labels, etc. for the deployment are as expected.
	container := checkContainer(t, deploy, envoyContainerName, true)
	checkContainerImage(t, container, ir.DefaultProxyImage)
	checkContainerResources(t, container, expectedResources)
	checkContainerSecurityContext(t, container, expectedSecurityContext)
	checkPodSecurityContext(t, deploy, expectedPodSecurityContext)
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
	infra.Proxy.GetProxyConfig().GetEnvoyProxyProvider().GetEnvoyProxyKubeProvider().EnvoyDeployment.Replicas = &repl

	deploy, err = kube.expectedProxyDeployment(infra)
	require.NoError(t, err)

	// Check the number of replicas is as expected.
	assert.Equal(t, repl, *deploy.Spec.Replicas)

	// Make sure no pod annotations are set by default
	checkPodAnnotations(t, deploy, nil)
}

func TestExpectedProxyDeployment(t *testing.T) {
	testExpectedProxyDeployment(t, ir.NewInfra(), egcfgv1a1.DefaultResourceRequirements(), nil, nil)
}

func TestExpectedProxyDeploymentForSpecifiedResources(t *testing.T) {
	infra := ir.NewInfra()
	requirements := corev1.ResourceRequirements{
		Limits: nil,
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("10m"),
			corev1.ResourceMemory: resource.MustParse("128Mi")},
		Claims: nil,
	}
	containerSecurityContext := corev1.SecurityContext{
		RunAsUser:                pointer.Int64(2000),
		AllowPrivilegeEscalation: pointer.Bool(false),
	}
	FSGroupChangePolicy := func(s corev1.PodFSGroupChangePolicy) *corev1.PodFSGroupChangePolicy { return &s }
	podSecurityContext := corev1.PodSecurityContext{
		RunAsUser:           pointer.Int64(1000),
		RunAsGroup:          pointer.Int64(3000),
		FSGroup:             pointer.Int64(2000),
		FSGroupChangePolicy: FSGroupChangePolicy(corev1.FSGroupChangeOnRootMismatch),
	}
	infra.Proxy.Config = &egcfgv1a1.EnvoyProxy{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: egcfgv1a1.EnvoyProxySpec{Provider: &egcfgv1a1.EnvoyProxyProvider{
			Type: egcfgv1a1.ProviderTypeKubernetes,
			Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
				EnvoyDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
					Pod: &egcfgv1a1.KubernetesPodSpec{
						SecurityContext: &podSecurityContext,
					},
					Container: &egcfgv1a1.KubernetesContainerSpec{
						Resources:       &requirements,
						SecurityContext: &containerSecurityContext,
					},
				},
			},
		}},
		Status: egcfgv1a1.EnvoyProxyStatus{},
	}

	testExpectedProxyDeployment(t, infra, &requirements, &podSecurityContext, &containerSecurityContext)
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

func TestExpectedPodAnnotations(t *testing.T) {
	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	// Set service annotations into EnvoyProxy API and ensure the same
	// value is set in the generated service.
	annotations := map[string]string{
		"key1": "val1",
		"key2": "val2",
	}
	infra.Proxy.Config = &egcfgv1a1.EnvoyProxy{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "test",
			Name:      "test",
		},
		Spec: egcfgv1a1.EnvoyProxySpec{
			Provider: &egcfgv1a1.EnvoyProxyProvider{
				Type: egcfgv1a1.ProviderTypeKubernetes,
				Kubernetes: &egcfgv1a1.EnvoyProxyKubernetesProvider{
					EnvoyDeployment: &egcfgv1a1.KubernetesDeploymentSpec{
						Pod: &egcfgv1a1.KubernetesPodSpec{
							Annotations: annotations,
						},
					},
				},
			},
		},
	}

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)
	checkPodAnnotations(t, deploy, annotations)
}

func TestExpectedContainerPort(t *testing.T) {
	const FooContainerPort, BarContainerPort = 7878, 8989

	svrCfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, svrCfg)
	infra := ir.NewInfra()

	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
	infra.Proxy.Listeners = []ir.ProxyListener{
		{
			Ports: []ir.ListenerPort{
				{
					Name:          "FooPort",
					Protocol:      ir.TCPProtocolType,
					ContainerPort: FooContainerPort,
				},
			},
		},
		{
			Ports: []ir.ListenerPort{
				{
					Name:          "BarPort",
					Protocol:      ir.UDPProtocolType,
					ContainerPort: BarContainerPort,
				},
			},
		},
	}

	deploy, err := kube.expectedProxyDeployment(infra)
	require.NoError(t, err)
	ports := []int32{FooContainerPort, BarContainerPort}
	for _, port := range ports {
		checkContainerHasPort(t, deploy, port)
	}
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
