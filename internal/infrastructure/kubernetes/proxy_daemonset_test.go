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

func daemonsetWithImage(ds *appsv1.DaemonSet, image string) *appsv1.DaemonSet {
	dCopy := ds.DeepCopy()
	for i, c := range dCopy.Spec.Template.Spec.Containers {
		if c.Name == envoyContainerName {
			dCopy.Spec.Template.Spec.Containers[i].Image = image
		}
	}
	return dCopy
}

func daemonsetWithSelectorAndLabel(ds *appsv1.DaemonSet, selector *metav1.LabelSelector, additionalLabel map[string]string) *appsv1.DaemonSet {
	dCopy := ds.DeepCopy()
	if selector != nil {
		dCopy.Spec.Selector = selector
	}
	for k, v := range additionalLabel {
		dCopy.Spec.Template.Labels[k] = v
	}
	return dCopy
}

func setupCreateOrUpdateProxyDaemonSet(gatewayNamespaceMode bool) (*appsv1.DaemonSet, *ir.Infra, *config.Server, error) {
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
	infra.Proxy.Config = &egv1a1.EnvoyProxy{
		Spec: egv1a1.EnvoyProxySpec{
			Provider: &egv1a1.EnvoyProxyProvider{
				Type: egv1a1.ProviderTypeKubernetes,
				Kubernetes: &egv1a1.EnvoyProxyKubernetesProvider{
					// Use daemonset, instead of deployment.
					EnvoyDaemonSet: egv1a1.DefaultKubernetesDaemonSet(egv1a1.DefaultEnvoyProxyImage),
					EnvoyService:   egv1a1.DefaultKubernetesService(),
				},
			},
		},
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
	ds, err := r.DaemonSet()
	if err != nil {
		return nil, nil, nil, err
	}
	return ds, infra, cfg, nil
}

func TestCreateOrUpdateProxyDaemonSet(t *testing.T) {
	ds, infra, cfg, err := setupCreateOrUpdateProxyDaemonSet(false)
	require.NoError(t, err)

	gwDs, gwInfra, gwCfg, err := setupCreateOrUpdateProxyDaemonSet(true)
	require.NoError(t, err)

	testCases := []struct {
		name                 string
		cfg                  *config.Server
		in                   *ir.Infra
		gatewayNamespaceMode bool
		current              *appsv1.DaemonSet
		want                 *appsv1.DaemonSet
		wantErr              bool
	}{
		{
			name: "create daemonset",
			cfg:  cfg,
			in:   infra,
			want: ds,
		},
		{
			name:    "daemonset exists",
			cfg:     cfg,
			in:      infra,
			current: ds,
			want:    ds,
		},
		{
			name: "update daemonset image",
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
									EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
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
			current: ds,
			want:    daemonsetWithImage(ds, "envoyproxy/envoy-dev:v1.2.3"),
		},
		{
			name: "update daemonset label",
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
									EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												// Add a new label to the custom label config.
												// It wouldn't break the daemonset because the selector would still match after this label update.
												"custom-label": "version1",
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
			current: ds,
			// Selector is not updated with a custom label, only pod's label is updated.
			want: daemonsetWithSelectorAndLabel(ds, nil, map[string]string{"custom-label": "version1"}),
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
									EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												"custom-label":         "version1",
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
			current: daemonsetWithSelectorAndLabel(ds, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1"}),
			// Only label is updated, selector is not updated.
			want: daemonsetWithSelectorAndLabel(ds, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1", "another-custom-label": "version1"}),
		},
		{
			name: "the daemonset originally has a selector and label, and an user update an existing custom label",
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
									EnvoyDaemonSet: &egv1a1.KubernetesDaemonSetSpec{
										Pod: &egv1a1.KubernetesPodSpec{
											Labels: map[string]string{
												// Update the label value which will break the daemonset
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
			current: daemonsetWithSelectorAndLabel(ds, resource2.GetSelector(map[string]string{"custom-label": "version1"}), map[string]string{"custom-label": "version1"}),
			wantErr: true,
		},
		{
			name:                 "create daemonset with gateway namespace mode",
			cfg:                  gwCfg,
			in:                   gwInfra,
			gatewayNamespaceMode: true,
			want:                 gwDs,
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
			err = kube.createOrUpdateDaemonSet(ctx, r)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			actual := &appsv1.DaemonSet{
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

func expectedName(proxyInfra *ir.ProxyInfra, isGatewayNamespaceMode bool) string {
	if isGatewayNamespaceMode {
		return proxyInfra.Name
	}

	return proxy.ExpectedResourceHashedName(proxyInfra.Name)
}
