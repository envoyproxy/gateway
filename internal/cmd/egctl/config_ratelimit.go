// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"

	"github.com/envoyproxy/gateway/api/v1alpha1"
	"github.com/envoyproxy/gateway/internal/envoygateway"
	"github.com/envoyproxy/gateway/internal/infrastructure/kubernetes/ratelimit"
	"github.com/envoyproxy/gateway/internal/kubernetes"
)

var (
	defaultRateLimitNamespace = "envoy-gateway-system" // TODO: make this configurable until EG support
	defaultConfigMap          = "envoy-gateway-config" // TODO: make this configurable until EG support
	defaultConfigMapKey       = "envoy-gateway.yaml"   // TODO: make this configurable until EG support
)

func ratelimitConfigCommand() *cobra.Command {

	var (
		namespace string
	)

	rlConfigCmd := &cobra.Command{
		Use:     "envoy-ratelimit",
		Aliases: []string{"rl"},
		Long:    `Retrieve the relevant rate limit configuration from the Rate Limit instance`,
		Example: `  # Retrieve rate limit configuration
  egctl config envoy-ratelimit

  # Retrieve rate limit configuration with short syntax
  egctl c rl
`,
		Run: func(c *cobra.Command, args []string) {
			cmdutil.CheckErr(runRateLimitConfig(c, namespace))
		},
	}

	rlConfigCmd.Flags().StringVarP(&namespace, "namespace", "n", defaultRateLimitNamespace, "Specific a namespace to get resources")
	return rlConfigCmd
}

func runRateLimitConfig(c *cobra.Command, ns string) error {

	cli, err := getCLIClient()
	if err != nil {
		return err
	}

	out, err := retrieveRateLimitConfig(cli, ns)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintln(c.OutOrStdout(), string(out))
	return err
}

func retrieveRateLimitConfig(cli kubernetes.CLIClient, ns string) ([]byte, error) {

	// Before retrieving the rate limit configuration
	// we make sure that the global rate limit feature is enabled
	if enable, err := checkEnableGlobalRateLimit(cli); !enable {
		return nil, fmt.Errorf("global rate limit feature is not enabled")
	} else if err != nil {
		return nil, fmt.Errorf("failed to get global rate limit status: %w", err)
	}

	// Filter out all rate limit pods in the Running state
	rlNN, err := fetchRunningRateLimitPods(cli, ns, ratelimit.LabelSelector())
	if err != nil {
		return nil, err
	}

	// In fact, the configuration of multiple rate limit replicas are the same.
	// After we filter out the rate limit Pods in the Running state,
	// we can directly use the first pod.
	rlPod := rlNN[0]
	fw, err := portForwarder(cli, rlPod, rateLimitDebugPort)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize pod-forwarding for %s/%s: %w", rlPod.Namespace, rlPod.Name, err)
	}

	return extractRateLimitConfig(fw, rlPod)
}

// fetchRunningRateLimitPods gets the rate limit Pods, based on the labelSelectors.
// It further filters out only those rate limit Pods that are in "Running" state.
func fetchRunningRateLimitPods(cli kubernetes.CLIClient, namespace string, labelSelector []string) ([]types.NamespacedName, error) {

	// Since multiple replicas of the rate limit are configured to be equal,
	// we do not need to use the pod name to obtain the specified pod.
	rlPods, err := cli.PodsForSelector(namespace, labelSelector...)
	if err != nil {
		return nil, err
	}

	rlNN := []types.NamespacedName{}
	for _, rlPod := range rlPods.Items {
		rlPodNsName := types.NamespacedName{
			Namespace: rlPod.Namespace,
			Name:      rlPod.Name,
		}

		// Check that the rate limit pod is ready properly and can accept external traffic
		if !checkRateLimitPodStatusReady(rlPod.Status) {
			continue
		}

		rlNN = append(rlNN, rlPodNsName)
	}
	if len(rlNN) == 0 {
		return nil, fmt.Errorf("please check that the rate limit instance starts properly")
	}

	return rlNN, nil
}

// checkRateLimitPodStatusReady Check that the rate limit pod is ready
func checkRateLimitPodStatusReady(status corev1.PodStatus) bool {

	if status.Phase != corev1.PodRunning {
		return false
	}

	for _, condition := range status.Conditions {
		if condition.Type == corev1.PodReady &&
			condition.Status == corev1.ConditionTrue {
			return true
		}
	}

	return false
}

// extractRateLimitConfig After turning on port forwarding through PortForwarder,
// construct a request and send it to the rate limit Pod to obtain relevant configuration information.
func extractRateLimitConfig(fw kubernetes.PortForwarder, rlPod types.NamespacedName) ([]byte, error) {

	if err := fw.Start(); err != nil {
		return nil, fmt.Errorf("failed to start port forwarding for pod %s/%s: %w", rlPod.Namespace, rlPod.Name, err)
	}
	defer fw.Stop()

	out, err := rateLimitConfigRequest(fw.Address())
	if err != nil {
		return nil, fmt.Errorf("failed to send request to get rate config for pod %s/%s: %w", rlPod.Namespace, rlPod.Name, err)
	}

	return out, nil
}

// checkEnableGlobalRateLimit Check whether the Global Rate Limit function is enabled
func checkEnableGlobalRateLimit(cli kubernetes.CLIClient) (bool, error) {

	kubeCli := cli.Kube()
	cm, err := kubeCli.CoreV1().
		ConfigMaps(defaultRateLimitNamespace).
		Get(context.TODO(), defaultConfigMap, metav1.GetOptions{})
	if err != nil {
		return false, err
	}

	config, ok := cm.Data[defaultConfigMapKey]
	if !ok {
		return false, fmt.Errorf("failed to get envoy-gateway configuration")
	}

	decoder := serializer.NewCodecFactory(envoygateway.GetScheme()).UniversalDeserializer()
	obj, gvk, err := decoder.Decode([]byte(config), nil, nil)
	if err != nil {
		return false, err
	}

	if gvk.Group != v1alpha1.GroupVersion.Group ||
		gvk.Version != v1alpha1.GroupVersion.Version ||
		gvk.Kind != v1alpha1.KindEnvoyGateway {
		return false, errors.New("failed to decode unmatched resource type")
	}

	eg, ok := obj.(*v1alpha1.EnvoyGateway)
	if !ok {
		return false, errors.New("failed to convert object to EnvoyGateway type")
	}

	if eg.RateLimit == nil || eg.RateLimit.Backend.Redis == nil {
		return false, nil
	}

	return true, nil
}

func rateLimitConfigRequest(address string) ([]byte, error) {
	url := fmt.Sprintf("http://%s/rlconfig", address)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	return io.ReadAll(resp.Body)
}
