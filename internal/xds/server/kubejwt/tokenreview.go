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

func validateKubeJWT(ctx context.Context, clientset *kubernetes.Clientset, token string) (bool, error) {
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token: token,
		},
	}

	tokenReview, err := clientset.AuthenticationV1().TokenReviews().Create(ctx, tokenReview, metav1.CreateOptions{})
	if err != nil {
		return false, fmt.Errorf("failed to call TokenReview API to verify service account JWT: %w", err)
	}

	if !slices.Contains(tokenReview.Status.User.Groups, "system:serviceaccounts") {
		return false, fmt.Errorf("the token is not a service account")
	}

	if !tokenReview.Status.Authenticated {
		return false, fmt.Errorf("token is not authenticated")
	}

	return true, nil
}
