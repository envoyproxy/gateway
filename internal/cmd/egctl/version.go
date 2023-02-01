// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/cmd/version"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
)

const (
	yamlOutput = "yaml"
)

func NewVersionsCommand() *cobra.Command {
	var (
		output          string
		egContainerName string
	)

	versionCommand := &cobra.Command{
		Use:     "versions",
		Aliases: []string{"version"},
		Short:   "Show versions",
		RunE: func(cmd *cobra.Command, args []string) error {
			return versions(cmd.OutOrStdout(), egContainerName, output)
		},
	}

	flags := versionCommand.Flags()
	options.AddKubeConfigFlags(flags)

	versionCommand.PersistentFlags().StringVarP(&output, "output", "o", yamlOutput, "One of 'yaml' or 'json'")
	versionCommand.PersistentFlags().StringVar(&egContainerName, "eg-container-name", "envoy-gateway", "Name of the Envoy Gateway container")

	return versionCommand
}

type VersionInfo struct {
	ClientVersion  string                   `json:"client"`
	ServerVersions map[string]*version.Info `json:"servers,omitempty"`
}

func Get() VersionInfo {
	return VersionInfo{
		ClientVersion:  version.Get().EnvoyGatewayVersion,
		ServerVersions: map[string]*version.Info{},
	}
}

func versions(w io.Writer, containerName, output string) error {
	v := Get()

	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return fmt.Errorf("failed to build kubernete client: %w", err)
	}

	pods, err := c.PodsForSelector(metav1.NamespaceAll, "control-plane=envoy-gateway")
	if err != nil {
		return fmt.Errorf("list EG pods failed: %w", err)
	}

	for _, pod := range pods.Items {
		nn := types.NamespacedName{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		}
		stdout, _, err := c.PodExec(nn, containerName, "envoy-gateway version -ojson")
		if err != nil {
			return fmt.Errorf("pod exec on %s/%s failed: %w", nn.Namespace, nn.Name, err)
		}

		info := &version.Info{}
		if err := json.Unmarshal([]byte(stdout), info); err != nil {
			return fmt.Errorf("unmarshall pod %s/%s exec result failed: %w", nn.Namespace, nn.Name, err)
		}

		v.ServerVersions[nn.String()] = info
	}

	var out []byte
	switch output {
	case yamlOutput:
		out, err = yaml.Marshal(v)
	default:
		out, err = json.MarshalIndent(v, "", "  ")

	}

	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(out))

	return nil
}
