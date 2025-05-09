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

func TestCreateOrUpdateProxyDaemonSet(t *testing.T) {
	cfg, err := config.New(os.Stdout)
	require.NoError(t, err)

	infra := ir.NewInfra()
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = infra.Proxy.Name
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

	r := proxy.NewResourceRender(cfg.ControllerNamespace, cfg.ControllerNamespace, cfg.DNSDomain, infra.GetProxyInfra(), cfg.EnvoyGateway)
	ds, err := r.DaemonSet()
	require.NoError(t, err)

	testCases := []struct {
		name    string
		in      *ir.Infra
		current *appsv1.DaemonSet
		want    *appsv1.DaemonSet
		wantErr bool
	}{
		{
			name: "create daemonset",
			in:   infra,
			want: ds,
		},
		{
			name:    "daemonset exists",
			in:      infra,
			current: ds,
			want:    ds,
		},
		{
			name: "update daemonset image",
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
	}

	for _, tc := range testCases {
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
			r := proxy.NewResourceRender(kube.ControllerNamespace, cfg.ControllerNamespace, kube.DNSDomain, tc.in.GetProxyInfra(), cfg.EnvoyGateway)
			err := kube.createOrUpdateDaemonSet(context.Background(), r)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			actual := &appsv1.DaemonSet{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.ControllerNamespace,
					Name:      proxy.ExpectedResourceHashedName(tc.in.Proxy.Name),
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))
			require.Equal(t, tc.want.Spec, actual.Spec)
		})
	}
}
