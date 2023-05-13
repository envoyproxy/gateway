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
	"github.com/envoyproxy/gateway/internal/ir"
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
	})
	require.NoError(t, secretErr)
}

func TestCreateRateLimitInfra(t *testing.T) {
	rateLimitInfra := new(ir.RateLimitInfra)

	testCases := []struct {
		name   string
		in     *ir.RateLimitInfra
		expect bool
	}{
		{
			name:   "nil-infra",
			in:     nil,
			expect: false,
		},
		{
			name:   "default infra",
			in:     rateLimitInfra,
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.CreateOrUpdateRateLimitInfra(context.Background(), tc.in)
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

				cm := &corev1.ConfigMap{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: kube.Namespace,
						Name:      ratelimit.InfraName,
					},
				}
				require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(cm), cm))

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
		in     *ir.RateLimitInfra
		expect bool
	}{
		{
			name:   "nil infra",
			in:     nil,
			expect: false,
		},
		{
			name:   "default infra",
			in:     new(ir.RateLimitInfra),
			expect: true,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := newTestInfra(t)

			createRateLimitTLSSecret(t, kube.Client.Client)

			err := kube.DeleteRateLimitInfra(context.Background(), tc.in)
			if !tc.expect {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
