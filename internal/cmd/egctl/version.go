// Copyright Envoy Gateway Authors
// SPDX-License-Identifier: Apache-2.0
// The full text of the Apache license is available in the LICENSE file at
// the root of the repo.

package egctl

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"

	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	"sigs.k8s.io/yaml"

	"github.com/envoyproxy/gateway/internal/cmd/options"
	"github.com/envoyproxy/gateway/internal/cmd/version"
	kube "github.com/envoyproxy/gateway/internal/kubernetes"
	"github.com/envoyproxy/gateway/internal/utils"
)

const (
	yamlOutput      = "yaml"
	jsonOutput      = "json"
	egContainerName = "envoy-gateway"
)

func newVersionCommand() *cobra.Command {
	var (
		output string
		remote bool
	)

	versionCommand := &cobra.Command{
		Use:     "version",
		Aliases: []string{"versions", "v"},
		Short:   "Show version",
		Example: `  # Show versions of both client and server.
  egctl version

  # Show versions of both client and server in JSON format.
  egctl version --output=json

  # Show version of client without server.
  egctl version --remote=false
	  `,
		Run: func(cmd *cobra.Command, args []string) {
			cmdutil.CheckErr(versions(cmd.OutOrStdout(), egContainerName, output, remote))
		},
	}

	flags := versionCommand.Flags()
	options.AddKubeConfigFlags(flags)

	versionCommand.PersistentFlags().StringVarP(&output, "output", "o", yamlOutput, "One of 'yaml' or 'json'")

	versionCommand.PersistentFlags().BoolVarP(&remote, "remote", "r", true, "If true, retrieve version from remote apiserver.")

	return versionCommand
}

type VersionInfo struct {
	ClientVersion  string           `json:"client"`
	ServerVersions []*ServerVersion `json:"server,omitempty"`
}

type ServerVersion struct {
	types.NamespacedName
	version.Info
}

func Get() VersionInfo {
	return VersionInfo{
		ClientVersion:  version.Get().EnvoyGatewayVersion,
		ServerVersions: make([]*ServerVersion, 0),
	}
}

func versions(w io.Writer, containerName, output string, remote bool) error {
	v := Get()

	if !remote {
		fmt.Fprintln(w, v.ClientVersion)
		return nil
	}

	c, err := kube.NewCLIClient(options.DefaultConfigFlags.ToRawKubeConfigLoader())
	if err != nil {
		return fmt.Errorf("failed to build kubernetes client: %w", err)
	}

	pods, err := c.PodsForSelector(metav1.NamespaceAll, "control-plane=envoy-gateway")
	if err != nil {
		return fmt.Errorf("list EG pods failed: %w", err)
	}

	for _, pod := range pods.Items {
		pod := pod
		if pod.Status.Phase != "Running" {

			fmt.Fprintf(w, "WARN: pod %s/%s is not running, skipping it.", pod.Namespace, pod.Name)
			continue
		}

		nn := utils.NamespacedName(&pod)
		stdout, _, err := c.PodExec(nn, containerName, "envoy-gateway version -ojson")
		if err != nil {
			return fmt.Errorf("pod exec on %s/%s failed: %w", nn.Namespace, nn.Name, err)
		}

		info := &version.Info{}
		if err := json.Unmarshal([]byte(stdout), info); err != nil {
			return fmt.Errorf("unmarshall pod %s/%s exec result failed: %w", nn.Namespace, nn.Name, err)
		}

		v.ServerVersions = append(v.ServerVersions, &ServerVersion{
			NamespacedName: nn,
			Info:           *info,
		})
	}

	sort.Slice(v.ServerVersions, func(i, j int) bool {
		if v.ServerVersions[i].Namespace == v.ServerVersions[j].Namespace {
			return v.ServerVersions[i].Name < v.ServerVersions[j].Name
		}

		return v.ServerVersions[i].Namespace < v.ServerVersions[j].Namespace
	})

	var out []byte
	switch output {
	case yamlOutput:
		out, err = yaml.Marshal(v)
	case jsonOutput:
		out, err = json.MarshalIndent(v, "", "  ")
	default:
		out, err = json.MarshalIndent(v, "", "  ")
	}

	if err != nil {
		return err
	}
	fmt.Fprintln(w, string(out))

	return nil
}
