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

	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/provider/kubernetes"
)

func createRateLimitTLSSecret(t *testing.T, client client.Client) {
	_, secretErr := kubernetes.CreateOrUpdateSecrets(context.Background(), client, []corev1.Secret{
		{
			Type: corev1.SecretTypeTLS,
			TypeMeta: metav1.TypeMeta{
				Kind:       "Secret",
				APIVersion: "v1",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ratelimit-cert",
				Namespace: "envoy-gateway-system",
				Labels: map[string]string{
					"control-plane": "envoy-gateway",
				},
			},
		},
	}, false)
	require.NoError(t, secretErr)
}

func createEnvoyGatewayService(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func createEnvoyGatewayDeployment(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Deployment",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func createEnvoyGatewayServiceAccount(t *testing.T, client client.Client, ns string) {
	err := client.Create(context.Background(), &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: ns,
		},
	})
	require.NoError(t, err)
}

func TestCreateRateLimitInfra(t *testing.T) {
	testCases := []struct {
		name            string
		ownerReferences []string
		expect          bool
	}{
		{
			name: "default infra",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindDeployment,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: true,
		},
		{
			name: "default infra but missing service owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindDeployment,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: false,
		},
		{
			name: "default infra but missing deployment owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindServiceAccount,
			},
			expect: false,
		},
		{
			name: "default infra but missing service account owner reference",
			ownerReferences: []string{
				ratelimit.ResourceKindService,
				ratelimit.ResourceKindDeployment,
			},
			expect: false,
		},
		{
			name:   "default infra but missing all owner reference",
			expect: false,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			// Ratelimit infra creation require Gateway infra as owner reference.
			for _, ref := range tc.ownerReferences {
				switch ref {
				case ratelimit.ResourceKindService:
					createEnvoyGatewayService(t, kube.Client.Client, kube.Namespace)
				case ratelimit.ResourceKindDeployment:
					createEnvoyGatewayDeployment(t, kube.Client.Client, kube.Namespace)
				case ratelimit.ResourceKindServiceAccount:
					createEnvoyGatewayServiceAccount(t, kube.Client.Client, kube.Namespace)
				}
			}

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.CreateOrUpdateRateLimitInfra(context.Background())
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify all resources were created via the fake kube client.
				sa := &corev1.ServiceAccount{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(sa), sa))

				deploy := &appsv1.Deployment{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(deploy), deploy))

				svc := &corev1.Service{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(svc), svc))
			}
		})
	}
}

func TestDeleteRateLimitInfra(t *testing.T) {
	testCases := []struct {
		name   string
		expect bool
	}{
		{
			name:   "default infra",
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.DeleteRateLimitInfra(context.Background())
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
