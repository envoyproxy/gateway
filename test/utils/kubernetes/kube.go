// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	egv1a1 "github.com/envoyproxy/gateway/api/v1alpha1"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

// NewKubeHelper consolidates common Kubernetes operations, including deployments, traffic management, and log probing.
func NewKubeHelper(client client.Client, kubeClient kube.CLIClient) *KubeActions {
	return &KubeActions{
		Client:    client,
		CLIClient: kubeClient,
	}
}

type KubeActions struct {
	client.Client
	kube.CLIClient
}

func (ka *KubeActions) ManageEgress(ctx context.Context, ip, namespace, policyName string, blockTraffic bool, scope map[string]string) (*networkingv1.NetworkPolicy, error) {
	// Retrieve the existing NetworkPolicy, if it exists
	existingPolicy := &networkingv1.NetworkPolicy{}
	err := ka.Get(ctx, client.ObjectKey{Name: policyName, Namespace: namespace}, existingPolicy)
	if err != nil && !kerrors.IsNotFound(err) {
		return nil, fmt.Errorf("failed to get existing NetworkPolicy: %w", err)
	}

	// Define the Egress rule based on the enforce parameter
	egressRule := networkingv1.NetworkPolicyEgressRule{
		To: []networkingv1.NetworkPolicyPeer{
			{
				IPBlock: &networkingv1.IPBlock{
					CIDR: "0.0.0.0/0",
					Except: []string{
						ip + "/32",
					},
				},
			},
		},
	}
	// Define the NetworkPolicy object
	netPolicy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      policyName,
			Namespace: namespace,
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{
				MatchLabels: scope,
			}, // Selects all pods in the namespace
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeEgress,
			},
			Egress: []networkingv1.NetworkPolicyEgressRule{
				egressRule,
			},
		},
	}

	// remove the policy
	if !blockTraffic {
		if err := ka.Delete(ctx, netPolicy); err != nil {
			return nil, fmt.Errorf("failed to delete NetworkPolicy: %w", err)
		}
		return nil, nil
	}

	if kerrors.IsNotFound(err) {
		// Create the NetworkPolicy if it doesn't exist
		if err := ka.Create(ctx, netPolicy); err != nil {
			return nil, fmt.Errorf("failed to create NetworkPolicy: %w", err)
		}
		fmt.Printf("NetworkPolicy %s created.\n", netPolicy.Name)
	} else {
		// Update the existing NetworkPolicy
		existingPolicy.Spec = netPolicy.Spec
		if err := ka.Update(ctx, existingPolicy); err != nil {
			return nil, fmt.Errorf("failed to update NetworkPolicy: %w", err)
		}
		fmt.Printf("NetworkPolicy %s updated.\n", netPolicy.Name)
	}

	return netPolicy, nil
}

func (ka *KubeActions) ScaleDeploymentAndWait(ctx context.Context, deploymentName, namespace string, replicas int32, timeout time.Duration, prefix bool) error {
	// Get the current deployment
	deployment := &appsv1.Deployment{}
	if prefix {
		var err error
		deployment, err = ka.getDepByPrefix(ctx, deploymentName, namespace)
		if err != nil {
			return err
		}
	} else {
		err := ka.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)
		if err != nil {
			return err
		}
	}

	// Update the replicas count
	deployment.Spec.Replicas = &replicas

	// Apply the update
	err := ka.Update(ctx, deployment)
	if err != nil {
		return err
	}

	fmt.Printf("Deployment %s scaled to %d replicas\n", deployment.Name, replicas)
	return ka.WaitForDeploymentReplicaCount(ctx, deployment.Name, namespace, replicas, timeout, false)
}

func (ka *KubeActions) ScaleEnvoyProxy(envoyProxyName, namespace string, replicas int32) error {
	ctx := context.Background()

	// Retrieve the existing EnvoyProxy resource
	envoyProxy := &egv1a1.EnvoyProxy{}
	err := ka.Get(ctx, types.NamespacedName{Name: envoyProxyName, Namespace: namespace}, envoyProxy)
	if err != nil {
		return fmt.Errorf("failed to get EnvoyProxy: %w", err)
	}
	envoyProxy.Spec.Provider.Kubernetes = &egv1a1.EnvoyProxyKubernetesProvider{
		EnvoyDeployment: &egv1a1.KubernetesDeploymentSpec{
			Replicas: ptr.To[int32](replicas),
		},
	}

	// Update the replicas count
	envoyProxy.Spec.Provider.Kubernetes.EnvoyDeployment.Replicas = &replicas

	// Apply the update
	err = ka.Update(ctx, envoyProxy)
	if err != nil {
		return fmt.Errorf("failed to update EnvoyProxy: %w", err)
	}

	return nil
}

func (ka *KubeActions) MarkAsLeader(namespace, podName string) {
	pod, err := ka.Kube().CoreV1().Pods(namespace).Get(context.TODO(), podName, metav1.GetOptions{})
	if err != nil {
		panic(err.Error())
	}

	// Initialize the labels map if it's nil
	if pod.Labels == nil {
		pod.Labels = make(map[string]string)
	}

	// Add or update the desired label
	pod.Labels["leader"] = "true"

	// Update the Pod with the new label
	updatedPod, err := ka.Kube().CoreV1().Pods(namespace).Update(context.TODO(), pod, metav1.UpdateOptions{})
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Pod %s updated with new label.\n", updatedPod.Name)
}

func (ka *KubeActions) WaitForDeploymentReplicaCount(ctx context.Context, deploymentName, namespace string, replicas int32, timeout time.Duration, prefix bool) error {
	start := time.Now()

	for {
		// Check if the timeout has been reached
		if time.Since(start) > timeout {
			return errors.New("timeout reached waiting for deployment to scale")
		}

		// Get the current deployment status
		deployment := &appsv1.Deployment{}

		if prefix {
			var err error
			deployment, err = ka.getDepByPrefix(ctx, deploymentName, namespace)
			if err != nil {
				return err
			}
		} else {
			err := ka.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)
			if err != nil {
				return err
			}
		}

		// Check if the deployment has reached the desired number of replicas
		if deployment.Status.ReadyReplicas == replicas {
			fmt.Printf("Deployment %s scaled to %d replicas\n", deploymentName, replicas)
			return nil
		}

		// Wait before checking again
		time.Sleep(5 * time.Second)
	}
}

func (ka *KubeActions) CheckDeploymentReplicas(ctx context.Context, prefix, namespace string, expectedReplicas int, timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deployment, err := ka.getDepByPrefix(ctx, prefix, namespace)
	if err != nil {
		return err
	}

	if deployment != nil {
		// Wait for the deployment to reach the expected replica count
		for {
			select {
			case <-ctx.Done():
				return fmt.Errorf("timeout reached: deployment %q did not reach %d replicas", deployment.Name, expectedReplicas)
			default:
				// Fetch the current status of the deployment
				deployment, err := ka.Kube().AppsV1().Deployments(namespace).Get(ctx, deployment.Name, metav1.GetOptions{})
				if err != nil {
					return fmt.Errorf("failed to get deployment %q: %w", deployment.Name, err)
				}

				// Check the ready replica count
				if int(deployment.Status.ReadyReplicas) == expectedReplicas {
					fmt.Printf("Deployment %q reached %d replicas as expected.\n", deployment.Name, expectedReplicas)
					return nil
				}

				fmt.Printf("Waiting for deployment %q: ready replicas %d/%d\n",
					deployment.Name, deployment.Status.ReadyReplicas, expectedReplicas)
				time.Sleep(1 * time.Second) // Retry interval
			}
		}
	}
	return errors.New("deployment was not found")
}

func (ka *KubeActions) getDepByPrefix(ctx context.Context, prefix, namespace string) (*appsv1.Deployment, error) {
	deployments, err := ka.Kube().AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %w", err)
	}

	// Search for the deployment with the specified prefix
	for _, dep := range deployments.Items {
		if len(dep.Name) >= len(prefix) && dep.Name[:len(prefix)] == prefix {
			return &dep, nil
		}
	}
	return nil, errors.New("deployment not found")
}

func (ka *KubeActions) GetElectedLeader(ctx context.Context, namespace, leaseName string, afterTime metav1.Time, timeout time.Duration) (string, error) {
	// Create a context with a timeout
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	for {
		// Fetch the Lease object
		lease, err := ka.getLease(ctxWithTimeout, namespace, leaseName)
		if err != nil {
			return "", fmt.Errorf("failed to get lease %s in namespace %s: %w", leaseName, namespace, err)
		}

		// Check if RenewTime matches the condition
		if lease.Spec.RenewTime != nil && lease.Spec.RenewTime.After(afterTime.Time) {
			if lease.Spec.HolderIdentity == nil || *lease.Spec.HolderIdentity == "" {
				return "", fmt.Errorf("lease %s does not have a valid holderIdentity", leaseName)
			}

			// Return the leader pod name
			hi := *lease.Spec.HolderIdentity
			parts := strings.SplitN(hi, "_", 2)

			// Return the left part (pod name)
			if len(parts) > 0 {
				return parts[0], nil
			} else {
				return "", fmt.Errorf("lease %s does not have a valid holderIdentity", leaseName)
			}
		}

		// Sleep for a short interval before retrying to avoid excessive API calls
		select {
		case <-ctxWithTimeout.Done():
			return "", fmt.Errorf("timeout reached while waiting for lease renew time: %w", ctxWithTimeout.Err())
		case <-time.After(1 * time.Second):
			// Retry after a delay
		}
	}
}

func (ka *KubeActions) getLease(ctx context.Context, namespace, leaseName string) (*coordinationv1.Lease, error) {
	// Fetch the Lease object
	lease, err := ka.Kube().CoordinationV1().Leases(namespace).Get(ctx, leaseName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get lease %s in namespace %s: %w", leaseName, namespace, err)
	}

	return lease, nil
}
