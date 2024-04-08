// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	"github.com/envoyproxy/gateway/internal/ir"
)

const (
	// envoyContainerName is the name of the Envoy container.
	envoyContainerName = "envoy"
)

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

	infra := ir.NewInfra()
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name

	r := proxy.NewResourceRender(cfg.Namespace, infra.GetProxyInfra())
	deploy, err := r.Deployment()
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
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							Provider: &egv1a1.EnvoyProxyProvider{
								Type: egv1a1.ProviderTypeKubernetes,
								Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
									EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
										Container: &egv1a1.KubernetesContainerSpec{
											Image: ptr.To("envoyproxy/envoy-dev:v1.2.3"),
										},
									},
								},
							},
						},
					},
					Name:      ir.DefaultProxyName,
					Listeners: ir.NewProxyListeners(),
				},
			},
			current: deploy,
			want:    deploymentWithImage(deploy, "envoyproxy/envoy-dev:v1.2.3"),
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var cli client.Client
			if tc.current != nil {
				cli = fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithObjects(tc.current).
					WithInterceptorFuncs(interceptorFunc).
					Build()
			} else {
				cli = fakeclient.NewClientBuilder().
					WithScheme(envoygateway.GetScheme()).
					WithInterceptorFuncs(interceptorFunc).
					Build()
			}

			kube := NewInfra(cli, cfg)
			r := proxy.NewResourceRender(kube.Namespace, tc.in.GetProxyInfra())
			err := kube.createOrUpdateDeployment(context.Background(), r)
			require.NoError(t, err)

			actual := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      proxy.ExpectedResourceHashedName(tc.in.Proxy.Name),
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
		})
	}
}

func TestDeleteProxyDeployment(t *testing.T) {
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	cfg, err := config.New()
	require.NoError(t, err)

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
			kube := NewInfra(cli, cfg)

			infra := ir.NewInfra()
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			r := proxy.NewResourceRender(kube.Namespace, infra.GetProxyInfra())

			err := kube.createOrUpdateDeployment(context.Background(), r)
			require.NoError(t, err)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      r.Name(),
				},
			}
			err = kube.Client.Delete(context.Background(), deployment)
			require.NoError(t, err)
		})
	}
}
