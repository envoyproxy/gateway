// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package envoy

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	configFilePath = "/envoygateway/config.yaml"
)

func EnvoyInit(regionDiscoveryDisabled bool, regionOverride string, zoneDiscoveryDisabled bool, zoneOverride string) error {
	if regionOverride != "" && zoneOverride != "" {
		return writeConfig(regionOverride, zoneOverride)
	}

	nodeName := os.Getenv("NODE_NAME")
	if nodeName == "" {
		return fmt.Errorf("NODE_NAME environment variable is required")
	}

	node, err := getNode(nodeName)
	if err != nil {
		return fmt.Errorf("error getting node %q: %w", nodeName, err)
	}

	region, err := buildLocalityRegion(node, regionOverride, regionDiscoveryDisabled)
	if err != nil {
		return fmt.Errorf("error getting node topology region: %w", err)
	}

	zone, err := buildLocalityZone(node, zoneOverride, zoneDiscoveryDisabled)
	if err != nil {
		return fmt.Errorf("error getting node topology zone: %w", err)
	}

	// Write locality information to envoy config file
	if err := writeConfig(region, zone); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	fmt.Println("Successfully built service locality configuration.")
	return nil
}

func getNode(nodeName string) (*corev1.Node, error) {
	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return nil, fmt.Errorf("failed to build CLI client: %w", err)
	}
	return c.Kube().CoreV1().Nodes().Get(context.Background(), nodeName, metav1.GetOptions{})
}

// buildLocalityRegion configures the envoy locality region using the Kubernetes node topology labels.
func buildLocalityRegion(node *corev1.Node, override string, discoveryDisabled bool) (string, error) {
	if override != "" {
		return override, nil
	}
	if discoveryDisabled {
		return "", nil
	}

	region, exists := node.Labels[corev1.LabelTopologyRegion]
	if !exists {
		return "", fmt.Errorf("region label %q not found on node %q", corev1.LabelTopologyRegion, node.Name)
	}
	return region, nil
}

// buildLocalityZone configures the envoy locality zone using the Kubernetes node topology labels.
func buildLocalityZone(node *corev1.Node, override string, discoveryDisabled bool) (string, error) {
	if override != "" {
		return override, nil
	}
	if discoveryDisabled {
		return "", nil
	}

	zone, exists := node.Labels[corev1.LabelTopologyZone]
	if !exists {
		return "", fmt.Errorf("zone label %q not found on node %q", corev1.LabelTopologyZone, node.Name)
	}
	return zone, nil
}

// writeConfig writes the locality information to the config file used for Envoy bootstrapping.
func writeConfig(region, zone string) error {
	// Construct JSON structure as a map
	config := map[string]interface{}{
		"node": map[string]interface{}{
			"locality": map[string]interface{}{
				"region": region,
				"zone":   zone,
			},
		},
	}

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(configFilePath), 0o755); err != nil {
		return fmt.Errorf("error creating config directory: %w", err)
	}
	return os.WriteFile(configFilePath, jsonData, 0o600)
}
