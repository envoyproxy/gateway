// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubejwt

import (
	"context"
	"fmt"
	"slices"
	"strings"

	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apiserver/pkg/authentication/serviceaccount"
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

func (i *JWTAuthInterceptor) validateKubeJWT(ctx context.Context, token, nodeID string) error {
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token:     token,
			Audiences: []string{i.audience},
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

	// Check if the node ID in the request matches the pod name in the token review response.
	// This is used to prevent a client from accessing the xDS resource of another one.
	if tokenReview.Status.User.Extra != nil {
		podName := tokenReview.Status.User.Extra[serviceaccount.PodNameKey]
		if podName[0] == "" {
			return fmt.Errorf("pod name not found in token review response")
		}

		if podName[0] != nodeID {
			return fmt.Errorf("pod name mismatch: expected %s, got %s", nodeID, podName[0])
		}
	}

	// Check if the service account name in the JWT token exists in the cache.
	// This is used to verify that the token belongs to a valid Envoy managed by Envoy Gateway.
	// example: "system:serviceaccount:default:envoy-default-eg-e41e7b31"
	parts := strings.Split(tokenReview.Status.User.Username, ":")
	if len(parts) != 4 {
		return fmt.Errorf("invalid username format: %s", tokenReview.Status.User.Username)
	}
	ns, sa := parts[2], parts[3]

	irKeys := i.cache.GetIrKeys()
	for _, irKey := range irKeys {
		nn := irKey2ServiceAccountName(irKey)
		if nn.Name == sa && nn.Namespace == ns {
			return nil
		}
	}
	return fmt.Errorf("service account for Envoy %s not found in the cache", sa)
}

// this is the same logic used in infra pkg func ExpectedResourceHashedName to generate the resource name.
func irKey2ServiceAccountName(irKey string) types.NamespacedName {
	names := strings.Split(irKey, "/")
	if len(names) == 2 {
		return types.NamespacedName{
			Namespace: names[0],
			Name:      names[1],
		}
	}

	// Might be MergeGateways, should not happen
	// but just in case, return the first part as name
	return types.NamespacedName{
		Name: names[0],
	}
}
