// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/envoygateway/config"
	"github.com/envoyproxy/gateway/internal/ir"
)

func TestExpectedRateLimitServiceAccount(t *testing.T) {
	cfg, err := config.New()
	require.NoError(t, err)
	cli := fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects().Build()
	kube := NewInfra(cli, cfg)

	rateLimitInfra := new(ir.RateLimitInfra)

	// An infra without Gateway owner labels should trigger
	// an error.
	sa := kube.expectedRateLimitServiceAccount(rateLimitInfra)
	require.NotNil(t, sa)

	// Check the serviceaccount name is as expected.
	assert.Equal(t, sa.Name, rateLimitInfraName)
}

func TestCreateOrUpdateRateLimitServiceAccount(t *testing.T) {
	testCases := []struct {
		name    string
		ns      string
		in      *ir.RateLimitInfra
		current *corev1.ServiceAccount
		want    *corev1.ServiceAccount
	}{
		{
			name: "create-ratelimit-sa",
			ns:   "envoy-gateway-system",
			in:   new(ir.RateLimitInfra),
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      rateLimitInfraName,
				},
			},
		},
		{
			name: "ratelimit-sa-exists",
			ns:   "envoy-gateway-system",
			in:   new(ir.RateLimitInfra),
			want: &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					Kind:       "ServiceAccount",
					APIVersion: "v1",
				},
				ObjectMeta: metav1.ObjectMeta{
					Namespace: "envoy-gateway-system",
					Name:      rateLimitInfraName,
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			kube := &Infra{
				Namespace: tc.ns,
			}
			if tc.current != nil {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).WithObjects(tc.current).Build()
			} else {
				kube.Client = fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build()
			}
			err := kube.createOrUpdateRateLimitServiceAccount(context.Background(), tc.in)
			require.NoError(t, err)

			actual := &corev1.ServiceAccount{
				ObjectMeta: metav1.ObjectMeta{
					Namespace: kube.Namespace,
					Name:      rateLimitInfraName,
				},
			}
			require.NoError(t, kube.Client.Get(context.Background(), client.ObjectKeyFromObject(actual), actual))

			opts := cmpopts.IgnoreFields(metav1.ObjectMeta{}, "ResourceVersion")
			assert.Equal(t, true, cmp.Equal(tc.want, actual, opts))
		})
	}
}

func TestDeleteRateLimitServiceAccount(t *testing.T) {
	testCases := []struct {
		name string
	}{
		{
			name: "delete ratelimit service account",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			kube := &Infra{
				Client:    fakeclient.NewClientBuilder().WithScheme(envoygateway.GetScheme()).Build(),
				Namespace: "test",
			}
			rateLimitInfra := new(ir.RateLimitInfra)

			err := kube.createOrUpdateRateLimitService(context.Background(), rateLimitInfra)
			require.NoError(t, err)

			err = kube.deleteRateLimitServiceAccount(context.Background(), rateLimitInfra)
			require.NoError(t, err)
		})
	}
}
