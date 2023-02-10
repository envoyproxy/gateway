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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/gatewayapi"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestCreateProxyInfra(t *testing.T) {
	// Infra with Gateway owner labels.
	infraWithLabels := ir.NewInfra()
	infraWithLabels.GetProxyInfra().GetProxyMetadata().Labels = envoyAppLabel()
	infraWithLabels.GetProxyInfra().GetProxyMetadata().Labels[gatewayapi.OwningGatewayNamespaceLabel] = "default"
	infraWithLabels.GetProxyInfra().GetProxyMetadata().Labels[gatewayapi.OwningGatewayNameLabel] = "test-gw"

	testCases := []struct {
		name   string
		in     *ir.Infra
		expect bool
	}{
		{
			name:   "infra-with-expected-labels",
			in:     infraWithLabels,
			expect: true,
		},
		{
			name:   "default infra without Gateway owner labels",
			in:     ir.NewInfra(),
			expect: false,
		},
		{
			name:   "nil-infra",
			in:     nil,
			expect: false,
		},
		{
			name: "nil-infra-proxy",
			in: &ir.Infra{
				Proxy: nil,
			},
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "default",
			}
			// Create or update the proxy infra.
			err := kube.CreateOrUpdateProxyInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify all resources were created via the fake kube client.
				sa := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      expectedProxyServiceAccountName(tc.in.Proxy.Name),
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(sa), sa))

				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      expectedProxyConfigMapName(tc.in.Proxy.Name),
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(cm), cm))

				deploy := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      expectedProxyDeploymentName(tc.in.Proxy.Name),
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(deploy), deploy))

				svc := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      expectedProxyServiceName(tc.in.Proxy.Name),
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(svc), svc))
			}
		})
	}
}

func TestDeleteProxyInfra(t *testing.T) {
	testCases := []struct {
		name   string
		in     *ir.Infra
		expect bool
	}{
		{
			name:   "nil infra",
			in:     nil,
			expect: false,
		},
		{
			name:   "default infra",
			in:     ir.NewInfra(),
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client: fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
			}
			err := kube.DeleteProxyInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
