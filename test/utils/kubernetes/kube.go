// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package kubernetes

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"github.com/envoyproxy/gateway/api/v1alpha1"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"io"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
	"sync"
	"time"
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
	var egressRule networkingv1.NetworkPolicyEgressRule
	egressRule = networkingv1.NetworkPolicyEgressRule{
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

	//remove the policy
	if !blockTraffic {
		if err := ka.Client.Delete(ctx, netPolicy); err != nil {
			return nil, fmt.Errorf("failed to delete NetworkPolicy: %w", err)
		}
		return nil, nil
	}

	if kerrors.IsNotFound(err) {
		// Create the NetworkPolicy if it doesn't exist
		if err := ka.Client.Create(ctx, netPolicy); err != nil {
			return nil, fmt.Errorf("failed to create NetworkPolicy: %w", err)
		}
		fmt.Printf("NetworkPolicy %s created.\n", netPolicy.Name)
	} else {
		// Update the existing NetworkPolicy
		existingPolicy.Spec = netPolicy.Spec
		if err := ka.Client.Update(ctx, existingPolicy); err != nil {
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
		err := ka.Client.Get(ctx, client.ObjectKey{Name: deploymentName, Namespace: namespace}, deployment)
		if err != nil {
			return err
		}
	}

	// Update the replicas count
	deployment.Spec.Replicas = &replicas

	// Apply the update
	err := ka.Client.Update(ctx, deployment)
	if err != nil {
		return err
	}

	fmt.Printf("Deployment %s scaled to %d replicas\n", deployment.Name, replicas)
	return ka.WaitForDeploymentReplicaCount(ctx, deployment.Name, namespace, replicas, timeout, false)
}

func (ka *KubeActions) ScaleEnvoyProxy(envoyProxyName, namespace string, replicas int32) error {
	ctx := context.Background()

	// Retrieve the existing EnvoyProxy resource
	envoyProxy := &v1alpha1.EnvoyProxy{}
	err := ka.Client.Get(ctx, types.NamespacedName{Name: envoyProxyName, Namespace: namespace}, envoyProxy)
	if err != nil {
		return fmt.Errorf("failed to get EnvoyProxy: %w", err)
	}
	envoyProxy.Spec.Provider.Kubernetes = &v1alpha1.EnvoyProxyKubernetesProvider{
		EnvoyDeployment: &v1alpha1.KubernetesDeploymentSpec{
			Replicas: ptr.To[int32](replicas),
		},
	}

	// Update the replicas count
	envoyProxy.Spec.Provider.Kubernetes.EnvoyDeployment.Replicas = &replicas

	// Apply the update
	err = ka.Client.Update(ctx, envoyProxy)
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

func (ka *KubeActions) CheckConnectivityJob(targetURL string, reqs int) error {
	jobName := "ab-connectivity"
	// Check if the job already exists and delete it
	existingJob := &batchv1.Job{}
	err := ka.Get(context.Background(), client.ObjectKey{Name: jobName, Namespace: corev1.NamespaceDefault}, existingJob)
	if err == nil {
		// Job exists; delete it
		fmt.Printf("Job %s already exists. Deleting it...\n", jobName)
		if deleteErr := ka.Delete(context.Background(), existingJob); deleteErr != nil {
			return fmt.Errorf("failed to delete existing job: %w", deleteErr)
		}
		fmt.Printf("Job %s deleted.\n", jobName)
	} else if !kerrors.IsNotFound(err) {
		// Some other error occurred while checking for the job
		return fmt.Errorf("failed to check if job exists: %w", err)
	}

	// Define the Job object
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: corev1.NamespaceDefault,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:    "ab",
							Image:   "jordi/ab",
							Command: []string{"ab", "-n", fmt.Sprintf("%d", reqs), "-c", "10", targetURL},
						},
					},
					RestartPolicy: corev1.RestartPolicyNever,
				},
			},
		},
	}

	ctx := context.Background()

	// Create the Job
	if err := ka.Create(ctx, job); err != nil {
		return fmt.Errorf("failed to create job: %w", err)
	}
	fmt.Printf("Job %s created.\n", job.Name)

	// Wait for the Job to complete
	if err := waitForJobCompletion(ctx, ka, job, 5*time.Minute); err != nil {
		// Delete the Job upon failure
		deletePolicy := metav1.DeletePropagationBackground
		if deleteErr := ka.Delete(ctx, job, &client.DeleteOptions{
			PropagationPolicy: &deletePolicy,
		}); deleteErr != nil {
			return fmt.Errorf("failed to delete job after failure: %v; original error: %w", deleteErr, err)
		}
		return fmt.Errorf("job failed: %w", err)
	}

	// Delete the Job upon failure
	deletePolicy := metav1.DeletePropagationBackground
	if deleteErr := ka.Delete(ctx, job, &client.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); deleteErr != nil {
		return fmt.Errorf("failed to delete job after failure: %v; original error: %w", deleteErr, err)
	}
	fmt.Printf("Job %s deleted.\n", job.Name)

	return nil
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

func (ka *KubeActions) getDepByPrefix(ctx context.Context, prefix string, namespace string) (*appsv1.Deployment, error) {
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

func waitForJobCompletion(ctx context.Context, k8sClient client.Client, job *batchv1.Job, timeout time.Duration) error {
	key := types.NamespacedName{Name: job.Name, Namespace: job.Namespace}
	endTime := time.Now().Add(timeout)

	for time.Now().Before(endTime) {
		var j batchv1.Job
		if err := k8sClient.Get(ctx, key, &j); err != nil {
			return fmt.Errorf("failed to get job: %w", err)
		}

		if j.Status.Failed > 0 {
			return fmt.Errorf("job %s failed", job.Name)
		}

		if j.Status.Succeeded > 0 {
			fmt.Printf("Job %s completed successfully.\n", job.Name)
			return nil
		}

		fmt.Printf("Waiting for job %s to complete...\n", job.Name)
		time.Sleep(5 * time.Second)
	}

	return fmt.Errorf("job %s did not complete within the specified timeout", job.Name)
}

// MonitorDeploymentLogs monitors logs for all pods in a deployment and returns true
// if the target string is found within the given timeout.
func (ka *KubeActions) MonitorDeploymentLogs(ctx context.Context, currentTime time.Time, namespace, depName, targetString string, timeout time.Duration, prefix bool) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	deployment := &appsv1.Deployment{}
	if prefix {
		var err error
		deployment, err = ka.getDepByPrefix(ctx, depName, namespace)
		if err != nil {
			return "", err
		}
	} else {
		err := ka.Client.Get(ctx, client.ObjectKey{Name: depName, Namespace: namespace}, deployment)
		if err != nil {
			return "", err
		}
	}

	deploymentName := deployment.Name
	// Retrieve all pods for the deployment
	pods, err := ka.getPodsForDeployment(ctx, namespace, deployment)
	if err != nil {
		return "", fmt.Errorf("failed to get pods for deployment %q: %w", deploymentName, err)
	}

	if len(pods) == 0 {
		return "", fmt.Errorf("no pods found for deployment %q", deploymentName)
	}

	// Monitor logs for all pods in parallel
	results := make(chan string, len(pods))
	errors := make(chan error, len(pods))

	var wg sync.WaitGroup
	for _, pod := range pods {
		wg.Add(1)
		go func(podName string) {
			defer wg.Done()
			if found, err := ka.monitorPodLogs(ctx, currentTime, namespace, podName, targetString); err != nil {
				errors <- fmt.Errorf("error monitoring logs for pod %q: %w", podName, err)
			} else if found {
				results <- podName
			}
		}(pod.Name)
	}

	// Close result channels once all goroutines finish
	go func() {
		wg.Wait()
		close(results)
		close(errors)
	}()

	// Handle results and errors
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("timeout reached while monitoring logs for deployment %q", deploymentName)
		case found := <-results:
			if found != "" {
				return found, nil
			}
		case err := <-errors:
			return "", err
		}
	}
}

// getPodsForDeployment retrieves all pods belonging to a deployment using label selectors.
func (ka *KubeActions) getPodsForDeployment(ctx context.Context, namespace string, deployment *appsv1.Deployment) ([]corev1.Pod, error) {
	labelSelector := deployment.Spec.Selector.MatchLabels
	podList, err := ka.Kube().CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: convertMatchLabelsToSelectorString(labelSelector),
	})
	if err != nil {
		return nil, err
	}
	return podList.Items, nil
}

func convertMatchLabelsToSelectorString(matchLabels map[string]string) string {
	var sb strings.Builder
	for key, value := range matchLabels {
		if sb.Len() > 0 {
			sb.WriteString(",")
		}
		sb.WriteString(fmt.Sprintf("%s=%s", key, value))
	}
	return sb.String()
}

// monitorPodLogs streams logs from a pod and searches for the target string.
func (ka *KubeActions) monitorPodLogs(ctx context.Context, currenTime time.Time, namespace, podName, targetString string) (bool, error) {
	req := ka.Kube().CoreV1().Pods(namespace).GetLogs(podName, &corev1.PodLogOptions{Follow: true, SinceTime: ptr.To(metav1.Time{Time: currenTime})})
	stream, err := req.Stream(ctx)
	if err != nil {
		return false, fmt.Errorf("failed to stream logs for pod %q: %w", podName, err)
	}
	defer stream.Close()
	return ka.scanLogsForString(stream, targetString)
}

// scanLogsForString reads a log stream and searches for the target string.
func (ka *KubeActions) scanLogsForString(stream io.ReadCloser, targetString string) (bool, error) {
	reader := bufio.NewReader(stream)
	for {
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			continue
		}
		if err != nil {
			return false, fmt.Errorf("error reading log stream: %w", err)
		}
		if strings.Contains(line, targetString) {
			return true, nil
		}
	}
}