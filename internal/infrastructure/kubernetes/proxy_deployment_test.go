// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
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
	"github.com/envoyproxy/gateway/internal/gatewayapi/resource"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/proxy"
	resource2 "github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/resource"
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

func deploymentWithSelectorAndLabel(deploy *appsv1.Deployment, selector *metav1.LabelSelector, additionalLabel map[string]string) *appsv1.Deployment {
	dCopy := deploy.DeepCopy()
	if selector != nil {
		dCopy.Spec.Selector = selector
	}
	for k, v := range additionalLabel {
		dCopy.Spec.Template.Labels[k] = v
	}
	return dCopy
}

func setupCreateOrUpdateProxyDeployment(gatewayNamespaceMode bool) (*appsv1.Deployment, *ir.Infra, *config.Server, error) {
	ctx := context.Background()
	cfg, err := config.New(os.Stdout)
	if err != nil {
		return nil, nil, nil, err
	}
	infra := ir.NewInfra()
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
	infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
		Kind: resource.KindGatewayClass,
		Name: testGatewayClass,
	}

	cli := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		Build()
	kube := NewInfra(cli, cfg)
	if err := setupOwnerReferenceResources(ctx, kube.Client); err != nil {
		return nil, nil, nil, err
	}

	if gatewayNamespaceMode {
		cfg.EnvoyGateway.Provider.Kubernetes.Deploy = &egv1a1.KubernetesDeployMode{
			Type: ptr.To(egv1a1.KubernetesDeployModeTypeGatewayNamespace),
		}
		infra.Proxy.Name = "gateway-1"
		infra.Proxy.Namespace = "ns1"
		infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "ns1"
		infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = "gateway-1"
		infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
			Kind: resource.KindGateway,
			Name: "gateway-1",
		}
	}

	r, err := proxy.NewResourceRender(ctx, kube, infra)
	if err != nil {
		return nil, nil, nil, err
	}
	deploy, err := r.Deployment()
	if err != nil {
		return nil, nil, nil, err
	}
	return deploy, infra, cfg, nil
}

func TestCreateOrUpdateProxyDeployment(t *testing.T) {
	deploy, infra, cfg, err := setupCreateOrUpdateProxyDeployment(false)
	require.NoError(t, err)

	gwDeploy, gwInfra, gwCfg, err := setupCreateOrUpdateProxyDeployment(true)
	require.NoError(t, err)

	testCases := []struct {
		name                 string
		cfg                  *config.Server
		in                   *ir.Infra
		gatewayNamespaceMode bool
		current              *appsv1.Deployment
		want                 *appsv1.Deployment
		wantErr              bool
	}{
		{
			name: "create deployment",
			cfg:  cfg,
			in:   infra,
			want: deploy,
		},
		{
			name:    "deployment exists",
			cfg:     cfg,
			in:      infra,
			current: deploy,
			want:    deploy,
		},
		{
			name: "update deployment image",
			cfg:  cfg,
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      infra.Proxy.Name,
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
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
		{
			name: "update deployment label",
			cfg:  cfg,
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      infra.Proxy.Name,
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							Provider: &egv1a1.EnvoyProxyProvider{
								Type: egv1a1.ProviderTypeKubernetes,
								Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
									EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												"custom-label": "version1", // added.
											},
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
			// Selector is not updated with a custom label, only pod's label is updated.
			want: deploymentWithSelectorAndLabel(deploy, nil, map[string]string{"custom-label": "version1"}),
		},
		{
			name: "the daemonset originally has a selector and label, and an user add a new label to the custom label config",
			cfg:  cfg,
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      infra.Proxy.Name,
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							Provider: &egv1a1.EnvoyProxyProvider{
								Type: egv1a1.ProviderTypeKubernetes,
								Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
									EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												"custom-label": "version1",
												// Add a new label to the custom label config.
												// It wouldn't break the deployment because the selector would still match after this label update.
												"another-custom-label": "version1", // added.
											},
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
			current: deploymentWithSelectorAndLabel(deploy, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1"}),
			// Only label is updated, selector is not updated.
			want: deploymentWithSelectorAndLabel(deploy, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1", "another-custom-label": "version1"}),
		},
		{
			name: "the deployment originally has a selector and label, and an user update an existing custom label",
			cfg:  cfg,
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      infra.Proxy.Name,
						},
						OwnerReference: &ir.ResourceMetadata{
							Kind: resource.KindGatewayClass,
							Name: testGatewayClass,
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							Provider: &egv1a1.EnvoyProxyProvider{
								Type: egv1a1.ProviderTypeKubernetes,
								Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
									EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												// Update the label value which will break the deployment
												// because the selector cannot be updated while the user wants to update the label value.
												// We cannot help this case, just emit an error and let the user recreate the envoy proxy by themselves.
												"custom-label": "version2",
											},
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
			current: deploymentWithSelectorAndLabel(deploy, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1"}),
			wantErr: true,
		},
		{
			name:                 "create deployment with gateway namespace mode",
			cfg:                  gwCfg,
			in:                   gwInfra,
			gatewayNamespaceMode: true,
			want:                 gwDeploy,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
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

			kube := NewInfra(cli, tc.cfg)
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

			r, err := proxy.NewResourceRender(ctx, kube, tc.in)
			require.NoError(t, err)
			err = kube.createOrUpdateDeployment(ctx, r)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			actual := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.GetResourceNamespace(tc.in),
					Name:      expectedName(tc.in.Proxy, tc.gatewayNamespaceMode),
				},
			}
			require.NoError(t, kube.Client.Get(ctx, client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
			require.Equal(t, tc.want.OwnerReferences, actual.OwnerReferences)
		})
	}
}

func TestDeleteProxyDeployment(t *testing.T) {
	cli := fakeclient.NewClientBuilder().
		WithScheme(envoygateway.GetScheme()).
		WithObjects().
		WithInterceptorFuncs(interceptorFunc).
		Build()
	cfg, err := config.New(os.Stdout)
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
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			kube := NewInfra(cli, cfg)
			require.NoError(t, setupOwnerReferenceResources(ctx, kube.Client))

			infra := ir.NewInfra()
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
			infra.Proxy.GetProxyMetadata().OwnerReference = &ir.ResourceMetadata{
				Kind: resource.KindGatewayClass,
				Name: testGatewayClass,
			}
			r, err := proxy.NewResourceRender(ctx, kube, infra)
			require.NoError(t, err)

			err = kube.createOrUpdateDeployment(ctx, r)
			require.NoError(t, err)
			deployment := &appsv1.Deployment{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.ControllerNamespace,
					Name:      r.Name(),
				},
			}
			err = kube.Client.Delete(ctx, deployment)
			require.NoError(t, err)
		})
	}
}
