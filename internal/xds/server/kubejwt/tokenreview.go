// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"fmt"
	"slices"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	authPodNameKey    = "authentication.kubernetes.io/pod-name"
	envoyIrKeyHeader  = "x-envoy-gateway-ir-key"
	envoyNodeIDHeader = "x-envoy-node-id"
)

// GetKubernetesClient creates a Kubernetes client using in-cluster configuration.
func GetKubernetesClient() (*kubernetes.Clientset, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return clientset, nil
}

func (i *JWTAuthInterceptor) validateKubeJWT(ctx context.Context, proxyMetadata *proxyMetadata) error {
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: proxyMetadata.token,
		},
	}

	tokenReview, err := i.clientset.AuthenticationV1().TokenReviews().Create(ctx, tokenReview, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to call TokenReview API to verify service account JWT: %w", err)
	}

	if !slices.Contains(tokenReview.Status.User.Groups, "system:serviceaccounts") {
		return fmt.Errorf("the token is not a service account")
	}

	if !tokenReview.Status.Authenticated {
		return fmt.Errorf("token is not authenticated")
	}

	if tokenReview.Status.User.Extra != nil {
		podName := tokenReview.Status.User.Extra[authPodNameKey]
		if podName[0] == "" {
			return fmt.Errorf("pod name not found in token review response")
		}

		if podName[0] != proxyMetadata.nodeId {
			return fmt.Errorf("pod name mismatch: expected %s, got %s", proxyMetadata.nodeId, podName[0])
		}

		if !i.cache.SnapshotHasIrKey(proxyMetadata.irKey) {
			return fmt.Errorf("ir key not found in cache: %s", proxyMetadata.irKey)
		}
	}

	return nil
}
