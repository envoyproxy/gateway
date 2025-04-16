// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

func TestCreateOrUpdateProxyClusterRoleBinding(t *testing.T) {
	testCases := []struct {
		name    string
		ns      string
		in      *ir.Infra
		current *rbacv1.ClusterRoleBinding
		want    *rbacv1.ClusterRoleBinding
	}{
		{
			name: "default-no-op",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "create-crb-zone-discovery-enabled",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							EnableZoneDiscovery: ptr.To(true),
						},
					},
				},
			},
			want: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "envoy-test-9f86d081",
				},
				Subjects: []rbacv1.Subject{{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "envoy-test-9f86d081",
					Namespace: "test",
				}},
			},
		},
		{
			name: "create-crb-zone-discovery-disabled",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							EnableZoneDiscovery: ptr.To(false),
						},
					},
				},
			},
			want: nil,
		},
		{
			name: "crb-exists-zone-discovery-enabled",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "test",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							EnableZoneDiscovery: ptr.To(true),
						},
					},
				},
			},
			current: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "envoy-test-9f86d081",
				},
				Subjects: []rbacv1.Subject{{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "envoy-test-9f86d081",
					Namespace: "test",
				}},
			},
			want: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-test-9f86d081",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "envoy-test-9f86d081",
				},
				Subjects: []rbacv1.Subject{{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "envoy-test-9f86d081",
					Namespace: "test",
				}},
			},
		},

		{
			name: "hashed-name",
			ns:   "test",
			in: &ir.Infra{
				Proxy: &ir.ProxyInfra{
					Name: "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
					Metadata: &ir.InfraMetadata{
						Labels: map[string]string{
							gatewayapi.OwningGatewayNamespaceLabel: "default",
							gatewayapi.OwningGatewayNameLabel:      "gateway-1",
						},
					},
					Config: &egv1a1.EnvoyProxy{
						Spec: egv1a1.EnvoyProxySpec{
							EnableZoneDiscovery: ptr.To(true),
						},
					},
				},
			},
			current: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
				},
				Subjects: []rbacv1.Subject{{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "very-long-name-that-will-be-hashed-and-cut-off-because-its-too-long",
					Namespace: "test",
				}},
			},
			want: &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ClusterRoleBinding",
					APIVersion: "rbac.authorization.k8s.io/v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name: "envoy-very-long-name-that-will-be-hashed-and-cut-off-b-5bacc75e",
					Labels: map[string]string{
						"app.kubernetes.io/name":               "envoy",
						"app.kubernetes.io/component":          "proxy",
						"app.kubernetes.io/managed-by":         "envoy-gateway",
						gatewayapi.OwningGatewayNamespaceLabel: "default",
						gatewayapi.OwningGatewayNameLabel:      "gateway-1",
					},
				},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     "envoy-very-long-name-that-will-be-hashed-and-cut-off-b-5bacc75e",
				},
				Subjects: []rbacv1.Subject{{
					Kind:      rbacv1.ServiceAccountKind,
					Name:      "envoy-very-long-name-that-will-be-hashed-and-cut-off-b-5bacc75e",
					Namespace: "test",
				}},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, err := config.New(os.Stdout)
			require.NoError(t, err)
			cfg.Namespace = tc.ns

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

			r := proxy.NewResourceRender(kube.Namespace, kube.DNSDomain, tc.in.GetProxyInfra(), cfg.EnvoyGateway)
			err = kube.createOrUpdateClusterRoleBinding(context.Background(), r)
			require.NoError(t, err)

			actual := &rbacv1.ClusterRoleBinding{
				ObjectMeta: metav1.ObjectMeta{
					Name: proxy.ExpectedResourceHashedName(tc.in.Proxy.Name),
				},
			}

			err = kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual)

			if tc.want != nil {
				require.NoError(t, err)
				opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
				assert.True(t, cmp.Equal(tc.want, actual, opts), "Expected resources to be equal\n%s", cmp.Diff(tc.want, actual, opts))
			} else {
				require.True(t, errors.IsNotFound(err))
			}
		})
	}
}

func TestDeleteProxyClusterRoleBinding(t *testing.T) {
	testCases := []struct {
		name  string
		infra *ir.Infra
	}{
		{
			name:  "no-op default",
			infra: ir.NewInfra(),
		},
		{
			name:  "delete cluster role - zone discovery enabled",
			infra: newTestInfraWithZoneDiscovery(nil),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			kube := newTestInfra(t)

			tc.infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
			tc.infra.Proxy.GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = tc.infra.Proxy.Name
			r := proxy.NewResourceRender(kube.Namespace, kube.DNSDomain, tc.infra.GetProxyInfra(), kube.EnvoyGateway)

			err := kube.createOrUpdateClusterRoleBinding(context.Background(), r)
			require.NoError(t, err)

			err = kube.deleteClusterRoleBinding(context.Background(), r)
			require.NoError(t, err)
		})
	}
}
