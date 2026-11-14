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

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	authenticationv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

func (i *JWTAuthInterceptor) validateKubeJWT(ctx context.Context, token, nodeID, nodeCluster string) error {
	tokenReview := &authenticationv1.TokenReview{
		Spec: authenticationv1.TokenReviewSpec{
			Token:     token,
			Audiences: []string{i.audience},
		},
	}

	tokenReview, err := i.clientset.AuthenticationV1().TokenReviews().Create(ctx, tokenReview, metav1.CreateOptions{})
	if err != nil {
		return status.Errorf(codes.Internal, "failed to call TokenReview API to verify service account JWT: %v", err)
	}

	if tokenReview.Status.Error != "" {
		return status.Errorf(codes.Unauthenticated, "token review found error: %s", tokenReview.Status.Error)
	}

	if !slices.Contains(tokenReview.Status.User.Groups, "system:serviceaccounts") {
		return status.Error(codes.Unauthenticated, "the token is not a service account")
	}

	if !tokenReview.Status.Authenticated {
		return status.Error(codes.Unauthenticated, "token is not authenticated")
	}

	// Check if the node ID in the request matches the pod name in the token review response.
	// This is used to prevent a client from accessing the xDS resource of another one.
	if tokenReview.Status.User.Extra != nil {
		podName := tokenReview.Status.User.Extra[serviceaccount.PodNameKey]
		if len(podName) == 0 || podName[0] == "" {
			return status.Error(codes.Unauthenticated, "pod name not found in token review response")
		}

		if podName[0] != nodeID {
			return status.Errorf(codes.Unauthenticated, "pod name mismatch: expected %s, got %s", nodeID, podName[0])
		}
	}

	// Bind the requested node cluster to the authenticated service account's namespace.
	// In GatewayNamespaceMode the node cluster is the IR key "<namespace>/<gateway>", and
	// it alone selects which Gateway's snapshot (including SDS secrets) the cache serves.
	// Peers on this xDS server are authenticated only by this interceptor (the server uses
	// server-only TLS), so without this check a proxy holding a valid token for its own pod
	// could read another namespace's resources by sending a different node.cluster. The pod
	// name check above only pins node.id, not the cluster.
	saNamespace, _, err := serviceaccount.SplitUsername(tokenReview.Status.User.Username)
	if err != nil {
		return status.Errorf(codes.Unauthenticated, "invalid service account username %q: %v", tokenReview.Status.User.Username, err)
	}
	clusterNamespace, _, ok := strings.Cut(nodeCluster, "/")
	if !ok || clusterNamespace == "" {
		return status.Errorf(codes.Unauthenticated, "node cluster %q is not namespaced", nodeCluster)
	}
	if clusterNamespace != saNamespace {
		return status.Errorf(codes.Unauthenticated, "node cluster namespace %q does not match service account namespace %q", clusterNamespace, saNamespace)
	}

	return nil
}
