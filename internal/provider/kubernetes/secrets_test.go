// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	fakeclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var (
	envoyGatewaySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-gateway",
			Namespace: "envoy-gateway-system",
		},
	}

	envoySecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy",
			Namespace: "envoy-gateway-system",
		},
	}

	envoyRateLimitSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-rate-limit",
			Namespace: "envoy-gateway-system",
		},
	}

	oidcHMACSecret = corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "envoy-oidc-hmac",
			Namespace: "envoy-gateway-system",
		},
	}

	existingSecretsWithoutHMAC = []client.Object{
		&envoyGatewaySecret,
		&envoySecret,
		&envoyRateLimitSecret,
	}

	existingSecretsWithHMAC = []client.Object{
		&envoyGatewaySecret,
		&envoySecret,
		&envoyRateLimitSecret,
		&oidcHMACSecret,
	}

	SecretsToCreate = []corev1.Secret{
		envoyGatewaySecret,
		envoySecret,
		envoyRateLimitSecret,
		oidcHMACSecret,
	}
)

func TestCreateSecretsWhenUpgrade(t *testing.T) {
	t.Run("create HMAC secret when it does not exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithoutHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, false)
		require.ErrorIs(t, err, ErrSecretExists)
		require.Len(t, created, 1)
		require.Equal(t, "envoy-oidc-hmac", created[0].Name)

		err = cli.Get(context.Background(), client.ObjectKeyFromObject(&oidcHMACSecret), &corev1.Secret{})
		require.NoError(t, err)
	})

	t.Run("skip HMAC secret when it exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, false)
		require.ErrorIs(t, err, ErrSecretExists)
		require.Emptyf(t, created, "expected no secrets to be created, got %v", created)
	})

	t.Run("update secrets when they exist", func(t *testing.T) {
		cli := fakeclient.NewClientBuilder().WithObjects(existingSecretsWithHMAC...).Build()

		created, err := CreateOrUpdateSecrets(context.Background(), cli, SecretsToCreate, true)
		require.NoError(t, err)
		require.Len(t, created, 4)
	})
}
